package classifier

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/devplayg/golibs/network"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Classifier struct {
	engine       *ipasserver.Engine
	watcher      *fsnotify.Watcher
	tmpDir       string
	worker       int
	assetOrgMap  sync.Map
	assetIpasMap sync.Map
}

func NewClassifier(engine *ipasserver.Engine, worker int) *Classifier {
	return &Classifier{
		engine: engine,
		worker: worker,
		tmpDir: filepath.Join(engine.ProcessDir, "tmp"),
	}
}

func (c *Classifier) Stop() error {
	if c.watcher != nil {
		if err := c.watcher.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Classifier) loadOrgAssets(signal string) error {
	if len(signal) > 0 {
		defer func() {
			log.Debug("reloading IPAS complete")
			time.Sleep(300 * time.Microsecond)
			err := os.Remove(signal)
			if err != nil {
				log.Warn(err)
			}
		}()
	}

	var (
		code  string
		orgId int
	)
	rows, err := c.engine.DB.Query("select code, asset_id org_id from ast_asset where class = ? and type1 = ?", 1, 1)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&code, &orgId)
		if err != nil {
			return err
		}
		c.assetOrgMap.Store(code, orgId)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Classifier) loadIpasAssets(signal string) error {
	if len(signal) > 0 {
		defer func() {
			log.Debug("reloading assets complete")
			time.Sleep(300 * time.Microsecond)
			err := os.Remove(signal)
			if err != nil {
				log.Warn(err)
			}
		}()
	}

	var (
		code    string
		equipId string
		orgId   int
		groupId int
	)
	query := `
		select t1.code, equip_id, org_id, group_id
		from ast_ipas t left outer join (
			select asset_id, code
			from ast_asset
			where class = 1 and type1 = 1
		) t1 on t1.asset_id = t.org_id
		where code is not null
	`
	rows, err := c.engine.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&code, &equipId, &orgId, &groupId)
		if err != nil {
			return err
		}
		c.assetIpasMap.Store(code+equipId, objs.Org{orgId, groupId})
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Classifier) Start() error {
	// 기관자산 로딩
	if err := c.loadOrgAssets(""); err != nil {
		log.Error(err)
	}

	// 기존 IPAS 소속정보 로딩
	if err := c.loadIpasAssets(""); err != nil {
		log.Error(err)
	}

	ch := make(chan bool, c.worker)
	var err error
	c.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event := <-c.watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					if strings.HasSuffix(event.Name, ".log") {
						ch <- true
						log.Debug("deal")
						go c.deal(ch, event.Name) // 로그 분류

					} else if filepath.Base(event.Name) == "ast_ipas.sig" { // 자산(기관/그룹) 변경 이벤트
						log.Debug("reload IPAS")
						go c.loadIpasAssets(event.Name)

					} else if filepath.Base(event.Name) == "ast_asset.sig" { // IPAS 자산 변경 이벤트
						log.Debug("reloading assets")
						go c.loadIpasAssets(event.Name)
					}
				}
			case err := <-c.watcher.Errors:
				if err != nil {
					log.Error(err)
				}
			}
		}
	}()

	err = c.watcher.Add(filepath.Join(c.engine.ProcessDir, "data"))
	if err != nil {
		return err
	}

	return nil
}

func (c *Classifier) deal(ch <-chan bool, filename string) error {
	time.Sleep(10 * time.Millisecond)
	defer func() {
		<-ch
	}()

	// 파일 읽기
	file, err := openFile(filename)
	if err != nil {
		log.Error(err)
		return err
	}
	defer func() {
		file.Close()
		//os.Remove(file.Name())
	}()

	// 메모리에서 데이터 분류 및 파일 적재
	if err := c.classify(file); err != nil {
		log.Error(err.Error())
		os.Rename(file.Name(), file.Name()+".error")
		return err
	}

	// 파일 삭제
	return nil
}

func (c *Classifier) classify(file *os.File) error {
	var statusData string
	var eventData string

	// 라인 단위로 파일 읽기
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		// 파싱
		//	1	127.0.0.1	2018-04-11 17:01:03	2018-04-11 17:01:03	VT_SAM_8_20180411170103_1	SAM	VT_SAM_8	VT_SAM_0	37.19359	128.70250	9	8	2-423-618-38-65	4	9
		//	1	127.0.0.1	2018-04-11 17:01:03	2018-04-11 17:01:03	PT_LG_6_20180411170103_1	LG	PT_LG_6		ZT_LG_2		37.66667	127.72560	22	1	7-677-105-37-04	1	1
		r := strings.Split(scanner.Text(), "\t")

		// 문자열 IP를 정수형 IP로 변환
		r[1] = strconv.FormatUint(uint64(network.IpToInt32(net.ParseIP(r[1]))), 10)

		// Ipas 분류
		var (
			orgId   int
			groupId int
		)

		// 기관코드 정의
		if valOrg, ok := c.assetOrgMap.Load(r[5]); ok { // code : asset_id
			orgId = valOrg.(int)
		}

		// 기존 기관코드 확인
		if valOrg, ok := c.assetIpasMap.Load(r[5] + r[6]); ok { // equip_id : org(org_id + group_id)
			obj := valOrg.(objs.Org)
			//log.Debugf("[%s] %d == %d ", r[6], orgId, obj.OrgId)
			if orgId == obj.OrgId { // 기관코드과 기존과 동일하면
				groupId = obj.GroupId // 그룹코드 유지
			}
		}

		if r[0] == "1" { // 데이터 타입이 "이벤트" 이면
			eventData += fmt.Sprintf("%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r[3], orgId, groupId, r[13], r[4], r[6], r[7], r[8], r[9], r[10], r[11], r[12], r[14], r[1], r[2])
			//1	date
			//2	org_id
			//3	group_id
			//4	event_type
			//5	session_id /
			//6	equip_id
			//7	targets
			//8	latitude
			//9	longitude
			//10	speed
			//11	snr
			//12	usim
			//13	distance
			//14	ip
			//15	recv_date

		} else if r[0] == "2" { // 데이터 타입이 "상태정보"이면
			statusData += fmt.Sprintf("%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r[3], orgId, groupId, r[4], r[6], r[7], r[8], r[9], r[10], r[11], r[1], r[2])
			//1	date
			//2	org_id
			//3	group_id
			//4	session_id
			//5	equip_id
			//6	latitude
			//7	longitude
			//8	speed
			//9	snr
			//10	usim
			//11	ip
			//12	recv_date

			//	2	127.0.0.1	2018-04-11 17:01:03	2018-04-11 17:01:03	VT_SAM_8_20180411170103_1	SAM	VT_SAM_8	37.19359	128.70250	9	8	2-423-618-38-65
			//	2	127.0.0.1	2018-04-11 17:01:03	2018-04-11 17:01:03	PT_LG_6_20180411170103_1	LG	PT_LG_6		37.66667	127.72560	22	1	7-677-105-37-04
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// 이벤트 정보 입력
	if len(eventData) > 0 {
		// 파일에 기록
		f, err := c.writeDataToFile(&eventData, "event_")
		if err != nil {
			return err
		}
		//defer os.Remove(f.Name()) // clean up

		// DB에 입력
		if err := c.insertIpasEventData(f.Name()); err != nil {
			log.Debug("##1")
			return err
		}
	}

	// 상태정보 입력
	if len(statusData) > 0 {
		// 파일에 기록
		f, err := c.writeDataToFile(&statusData, "status_")
		if err != nil {
			return err
		}
		//defer os.Remove(f.Name()) // clean up

		// DB에 입력
		if err := c.insertIpasStatusData(f.Name()); err != nil {
			return err
		}
		if err := c.insertIpasStatusDataToTemp(f.Name()); err != nil { // 상태정보에 사용
			return err
		}
		if err := c.updateIpasStatus(f.Name()); err != nil { // 상태정보에 사용
			return err
		}
	}

	return nil
}

func (c *Classifier) insertIpasEventData(filename string) error {
	query := `
		LOAD DATA LOCAL INFILE '%s'
		INTO TABLE log_ipas_event
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	//o := orm.NewOrm()

	rs, err := c.engine.DB.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, _ := rs.RowsAffected()
	log.Debugf("table=%s, affected_rows=%d", "log_ipas_event", rowsAffected)
	return nil
}

func (c *Classifier) writeDataToFile(str *string, prefix string) (*os.File, error) {
	if len(*str) < 1 {
		return nil, errors.New("no data")
	}

	tmpFile, err := ioutil.TempFile(c.tmpDir, prefix)
	if err != nil {
		return nil, err
	}

	// 파일에 기록
	if _, err := tmpFile.WriteString(*str); err != nil {
		return nil, err
	}

	// 닫기
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func (c *Classifier) insertIpasStatusData(filename string) error {
	query := `
		LOAD DATA LOCAL INFILE '%s'
		INTO TABLE log_ipas_status
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	rs, err := c.engine.DB.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, _ := rs.RowsAffected()
	log.Debugf("table=%s, affected_rows=%d", "log_ipas_status", rowsAffected)
	return nil
}

func (c *Classifier) insertIpasStatusDataToTemp(filename string) error {
	var query string

	// 상태정보를 임시 테이블에 게록
	query = `
		LOAD DATA LOCAL INFILE '%s'
		INTO TABLE log_ipas_status_temp
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
		(date,org_id,group_id,session_id,equip_id,latitude,longitude,speed,snr,usim,ip,recv_date)
		SET filename = '%s';
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename), filepath.Base(filename))
	_, err := c.engine.DB.Exec(query)
	if err != nil {
		return err
	}
	//rowsAffected, _ := rs.RowsAffected()
	//log.Debugf("type=%s, affected_rows=%d", "status", rowsAffected)

	return nil
}

func (c *Classifier) updateIpasStatus(fp string) error {
	name := filepath.Base(fp)
	// 상태정보 업데이트
	query := `
		insert into ast_ipas(equip_id, org_id, equip_type, latitude, longitude, speed, snr, usim, ip, updated)
		select equip_id, org_id, 0, latitude, longitude, speed, snr, usim, ip, date
		from log_ipas_status_temp
		where filename = ?
		on duplicate key update
			org_id = values(org_id),
			equip_type = values(equip_type),
			latitude = values(latitude),
			longitude = values(longitude),
			speed = values(speed),
			snr = values(snr),
			usim = values(usim),
			ip = values(ip),
			updated = values(updated);
	`
	rs, err := c.engine.DB.Exec(query, name)
	rowsAffected, _ := rs.RowsAffected()
	log.Debugf("table=%s, affected_rows=%d", "status", rowsAffected)
	if err == nil {
		// 테이블 비우기
		query = "delete from log_ipas_status_temp where filename = ?"
		c.engine.DB.Exec(query)
	}

	return err
}

//date, equip_id, latitude, longitude, speed, snr, usim, event_type, distance
func openFile(filename string) (*os.File, error) {
	// 파일 읽기
	var file *os.File
	var err error

	for i := 0; i < 300; i++ {
		file, err = os.Open(filename)
		if err == nil {
			break
		} else {
			log.Debug("Waiting: ", filename)
		}
		time.Sleep(100 * time.Millisecond)
	}

	return file, err
}
