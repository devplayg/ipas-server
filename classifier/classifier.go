package classifier

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devplayg/golibs/network"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"io"
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
	tmpDir       string
	batchSize    uint
	assetOrgMap  sync.Map
	assetIpasMap sync.Map
}

func NewClassifier(engine *ipasserver.Engine, batchSize uint) *Classifier {
	return &Classifier{
		engine:    engine,
		batchSize: batchSize,
		tmpDir:    filepath.Join(engine.ProcessDir, "tmp"),
	}
}

func (c *Classifier) loadOrgAssets(signal string) error {
	var (
		code  string
		orgId int
	)
	rows, err := c.engine.DB.Query("select code, asset_id org_id from ast_code")
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
	var (
		code    string
		equipId string
		orgId   int
		groupId int
	)
	query := `
		select t1.code, equip_id, org_id, group_id
		from ast_ipas t left outer join ast_code t1 on t1.asset_id = t.org_id
		where t1.code is not null
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

func (c *Classifier) Run() error {
	// 기관자산 로딩
	if err := c.loadOrgAssets(""); err != nil {
		log.Error(err)
	}

	// 기존 IPAS 소속정보 로딩
	if err := c.loadIpasAssets(""); err != nil {
		log.Error(err)
	}

	fileList := make([]string, 0)

	var i uint
	dir := filepath.Join(c.engine.ProcessDir, "data")
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && f.Mode().IsRegular() && strings.HasSuffix(f.Name(), ".log") {
			log.Debugf("Got it: %s", f.Name())
			fileList = append(fileList, path)
			i++
			if i == c.batchSize {
				return io.EOF
			}
		}
		return nil
	})

	for _, f := range fileList {
		c.deal(f)
	}

	return nil
}

func (c *Classifier) deal(filename string) error {
	// 파일 열기
	file, err := openFile(filename)
	if err != nil {
		log.Error(err)
		return err
	}
	defer func() {
		file.Close()
		if c.engine.IsDebug() {
			os.Rename(file.Name(), filepath.Join(c.engine.TempDir, filepath.Base(file.Name())))
		} else {
			os.Remove(file.Name())
		}
	}()

	// 데이터를 분류해서 파일에 기록
	if err := c.classify(file); err != nil {
		log.Error(err)
		os.Rename(file.Name(), file.Name()+".error")
		return err
	}

	// 파일 삭제
	return nil
}

// 이벤트 분류
func (c *Classifier) classify(file *os.File) error {
	var statusData string
	var eventData string
	var alarmData string

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
		} else {
			// 기관코드 없는 이벤트는 제외
			log.Warnf("invalid cstid: %s (%s)", r[5], scanner.Text())
			continue
		}

		// 기존 기관코드 확인
		if valOrg, ok := c.assetIpasMap.Load(r[5] + r[6]); ok { // equip_id : org(org_id + group_id)
			obj := valOrg.(objs.Org)
			//log.Debugf("[%s] %d == %d ", r[6], orgId, obj.OrgId)
			if orgId == obj.OrgId { // 기관코드과 기존과 동일하면
				groupId = obj.GroupId // 그룹코드 유지
			} // 다른 경우는, org 코드가 바뀐 것으로 간주???? - 검토필요
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

			if r[13] == "3" || r[13] == "4" { // 과속 또는 근접 이벤트이면
				category := "speeding"
				if r[13] == "4" {
					category = "proximity"
				}

				j, _ := json.Marshal(map[string]string{
					"code":     r[5],
					"equip_id": r[6],
					"date":     r[3],
				})

				alarmData += fmt.Sprintf("%d\t%s\t%d\t%d\t%s\t%s\t%s\n", groupId, r[3], 0, 4, category, j, "")
				// group_id
				// date
				// sender_id
				// priority
				// category
				// message
				// url
			}

		} else if r[0] == "2" { // 데이터 타입이 "상태정보"이면
			equipType := 0
			if strings.HasPrefix(r[4], "PT") {
				equipType = objs.PedestrianTag
			} else if strings.HasPrefix(r[4], "VT") {
				equipType = objs.VehicleTag
			} else if strings.HasPrefix(r[4], "ZT") {
				equipType = objs.ZoneTag
			}

			statusData += fmt.Sprintf("%s\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r[3], orgId, groupId, equipType, r[4], r[6], r[7], r[8], r[9], r[10], r[11], r[1], r[2])
			//1		date
			//2		org_id
			//3		group_id
			//4 	equip_type
			//5		equip_id
			//6		latitude
			//7		longitude
			//8		speed
			//9		snr
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

	// 이벤트 정보를  기록
	if len(eventData) > 0 {
		// 파일에 기록
		f, err := c.writeDataToFile(&eventData, "event_")
		if err != nil {
			return err
		}

		if !c.engine.IsDebug() {
			defer os.Remove(f.Name())
		}

		// 로그 DB테이블에 기록
		if err := c.insertIpasEventData(f.Name()); err != nil {
			return err
		}
	}

	// 상태정보를 기록
	if len(statusData) > 0 {
		// 파일에 기록
		f, err := c.writeDataToFile(&statusData, "status_")
		if err != nil {
			return err
		}
		if !c.engine.IsDebug() {
			defer os.Remove(f.Name())
		}

		// 로그 DB테이블에 기록
		if err := c.insertIpasStatusData(f.Name()); err != nil {
			return err
		}

		// 상태 DB테이블에 기록
		if err := c.insertIpasStatusDataToTemp(f.Name()); err != nil { // 상태정보에 사용
			return err
		}
		if err := c.updateIpasStatus(); err != nil { // 상태정보에 사용
			return err
		}
	}

	if len(alarmData) > 0 {
		// 파일에 기록
		f, err := c.writeDataToFile(&alarmData, "alarm_")
		if err != nil {
			return err
		}
		if !c.engine.IsDebug() {
			defer os.Remove(f.Name())
		}

		// 상태 DB테이블에 기록
		if err := c.insertIpasAlarmDataToTemp(f.Name()); err != nil { // 상태정보에 사용
			return err
		}

		// 알람 발송
		if err := c.generateAlarms(); err != nil { // 상태정보에 사용
			return err
		}
	}

	return nil
}

func (c *Classifier) insertIpasEventData(filename string) error {
	query := `
		LOAD DATA LOCAL INFILE %q
		INTO TABLE log_ipas_event
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
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
		LOAD DATA LOCAL INFILE %q
		INTO TABLE log_ipas_status
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
		(date, org_id, group_id, @dummy, session_id, equip_id, latitude, longitude, speed, snr, usim, ip, recv_date)
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
		LOAD DATA LOCAL INFILE %q
		INTO TABLE ast_ipas_temp
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
		(date,org_id,group_id,equip_type,@dummy,equip_id,latitude,longitude,speed,snr,usim,ip)
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	_, err := c.engine.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Classifier) insertIpasAlarmDataToTemp(filename string) error {
	var query string

	// 상태정보를 임시 테이블에 게록
	query = `
		LOAD DATA LOCAL INFILE %q
		INTO TABLE log_message_temp
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
		(group_id, date, sender_id, priority, category, message, url)
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	_, err := c.engine.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Classifier) generateAlarms() error {
	var query string

	// 모든 관리자에게 알람 발송
	query = `
		insert into log_message(date, status, receiver_id, sender_id, priority, category, message, url)
		select date, 1, m.member_id, 10, t.priority, t.category, t.message, t.url
		from log_message_temp t left outer join mbr_member m on true
		where m.position >= 512
		order by date asc
	`
	_, err := c.engine.DB.Exec(query)
	if err != nil {
		return err
	}

	// 일반 사용자에게 알람 발송
	query = `
		insert into log_message(date, status, receiver_id, sender_id, priority, category, message, url)
		select date, 1, m.member_id, 10, t.priority, t.category, t.message, t.url
		from log_message_temp t
	 		join mbr_asset m on m.asset_id = t.group_id
	 		left outer join mbr_member m2 on m2.member_id = m.member_id 
		where group_id > 0 and m2.position < 512
	`
	_, err = c.engine.DB.Exec(query)
	if err != nil {
		return err
	}

	query = "truncate table log_message_temp"
	if _, err := c.engine.DB.Exec(query); err != nil {
		log.Error(err)
	}

	return nil
}

func (c *Classifier) updateIpasStatus() error {
	// 상태정보 업데이트
	query := `
		insert into ast_ipas(equip_id, org_id, equip_type, latitude, longitude, speed, snr, usim, ip, updated)
		select equip_id, org_id, equip_type, latitude, longitude, speed, snr, usim, ip, date
		from ast_ipas_temp
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
	_, err := c.engine.DB.Exec(query)
	if err == nil {
		//rowsAffected, _ := rs.RowsAffected()
		//log.Debugf("table=%s, affected_rows=%d", "status", rowsAffected)
		// 테이블 비우기
		query = "truncate table ast_ipas_temp"
		if _, err := c.engine.DB.Exec(query); err != nil {
			log.Error(err)
		}
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
