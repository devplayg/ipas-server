package classifier

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/devplayg/golibs/network"
	"github.com/devplayg/ipas-server"
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

var count uint64
var tagMap sync.Map

type org struct {
	orgId   int
	groupId int
}

type Classifier struct {
	engine  *ipasserver.Engine
	watcher *fsnotify.Watcher
	tmpDir  string
}

func NewClassifier(engine *ipasserver.Engine) *Classifier {
	return &Classifier{
		engine: engine,
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

func (c *Classifier) Start() error {
	ch := make(chan bool, 2)
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
						go c.deal(ch, event.Name)
						log.Debug(c.engine.LogPrefix, event)
					}
				}
			case err := <-c.watcher.Errors:
				log.Error(err)
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
		log.Error(err)
		os.Rename(file.Name(), file.Name()+".error")
		return err
	}

	// 파일 삭제
	return nil
}

func (c *Classifier) classify(file *os.File) error {
	var statusData string
	var eventData string

	// Todo : Org/Group 분류
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r := strings.Split(scanner.Text(), "\t")
		r[1] = strconv.FormatUint(uint64(network.IpToInt32(net.ParseIP(r[1]))), 10)

		if r[0] == "1" { // 이벤트 정보면
			belongTo, ok := tagMap.Load(r[5]) // Tag ID
			if ok {
				b := belongTo.(org)
				r = append(r, string(b.orgId), string(b.groupId))
			} else {
				r = append(r, "0", "0") //
			}
			eventData += strings.Join(r, "\t") + "\n"
		} else if r[0] == "2" { // 상태정보면
			belongTo, ok := tagMap.Load(r[5]) // Tag ID
			if ok {
				b := belongTo.(org)
				r = append(r, string(b.orgId), string(b.groupId))
			} else {
				r = append(r, "0", "0") //
			}
			statusData += strings.Join(r, "\t") + "\n"
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
		defer os.Remove(f.Name()) // clean up

		// DB에 입력
		if err := insertEventData(f.Name()); err != nil {
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
		defer os.Remove(f.Name()) // clean up

		// DB에 입력
		if err := insertStatusData(f.Name()); err != nil {
			return err
		}
	}

	return nil
}

func insertEventData(filename string) error {
	query := `
		LOAD DATA LOCAL INFILE '%s'
		INTO TABLE log_ipas_event
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' (@dummy, ip, date, recv_date, session_id, equip_id, targets, latitude, longitude, speed, snr, 	usim, event_type, distance, org_id, group_id)
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	o := orm.NewOrm()
	rs, err := o.Raw(query).Exec()
	if err != nil {
		return err
	}
	rowsAffected, _ := rs.RowsAffected()
	log.Debugf("type=%s, affected_rows=%d", "event", rowsAffected)
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

func insertStatusData(filename string) error {
	query := `
		LOAD DATA LOCAL INFILE '%s'
		INTO TABLE log_ipas_status
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' (@dummy, ip, date, recv_date, session_id, equip_id, latitude, longitude, speed, snr, usim, org_id, group_id)
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	o := orm.NewOrm()
	rs, err := o.Raw(query).Exec()
	if err != nil {
		return err
	}
	rowsAffected, _ := rs.RowsAffected()
	log.Debugf("type=%s, affected_rows=%d", "status", rowsAffected)
	return nil
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
