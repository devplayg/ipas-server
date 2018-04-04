package classifier

import (
	"bufio"
	"github.com/devplayg/ipas-server"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"net"
	"time"
	"strings"
	"io/ioutil"
	"fmt"
	"github.com/astaxie/beego/orm"
	"sync"
	"github.com/devplayg/golibs/network"
	"strconv"
)

var count uint64
var tagMap sync.Map
type org struct {
	orgId int
	groupId int
}

type Classifier struct {
	engine  *ipasserver.Engine
	watcher *fsnotify.Watcher
}

func NewClassifier(engine *ipasserver.Engine) *Classifier {
	return &Classifier{
		engine: engine,
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
					ch <- true
					go deal(ch, event.Name)
					//log.Debugf("event: %s", event)
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

func deal(ch <-chan bool, filename string) error {
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
	defer file.Close()

	// 메모리에서 데이터 분류
	classify(file)

	// 파일 저장 및 DB 입력

	// 파일 삭제

	//new := atomic.AddUint64(&count, 1)
	//log.Debugf("done: %d", new)
	return nil
	//log.Debug("start: " + name)
	//time.Sleep(1000 * time.Millisecond)
	//

	//new := atomic.AddUint64(&count, 1)
	//log.Debugf("done: %d", new)

}

func classify(file *os.File) error {
	var statusData string
	var eventData string

	// Todo : Org/Group 분류
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r := strings.Split(scanner.Text(), "\t")
		ip := net.ParseIP(r[1])
		u := network.IpToInt32(ip)
		r[1] = strconv.FormatUint(uint64(u), 10)

		if r[0] == "1" { // event
			eventData += scanner.Text()+"\n"

		} else if r[0] == "2" { // status
			belongTo, ok := tagMap.Load(r[4]) // Tag ID
			if ok {
				b := belongTo.(org)
				r = append(r, string(b.orgId), string(b.groupId))
			} else {
				r = append(r, "1", "2")
			}
			statusData += strings.Join(r, "\t")+"\n"
		}
	}


	// 상태정보 입력
	if len(statusData) > 0 {
		tmpFile, err := ioutil.TempFile("c:/temp", "")
		if err != nil {
			return err
		}
		//defer os.Remove(tmpFile.Name()) // clean up

		// 파일에 기록
		if _, err := tmpFile.WriteString(statusData); err != nil {
			return err
		}

		// 닫기
		if err := tmpFile.Close(); err != nil {
			return err
		}

		// DB에 입력
		if err := insertStatusData(tmpFile.Name()); err != nil {
			return err
		}
	}

	// 이벤트 입력
	if len(eventData) > 0 {

	}

	return nil
	//if err := scanner.Err(); err != nil {
	//	return nil, err
	//}
}


func insertStatusData(filename string) error {
	query := `
		LOAD DATA LOCAL INFILE '%s'
		INTO TABLE log_ipas_status
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' (@dummy, ip, date, recv_date, equip_id, latitude, longitude, speed, snr, usim, org_id, group_id)
	`
	query = fmt.Sprintf(query, filepath.ToSlash(filename))
	o := orm.NewOrm()
	rs, err := o.Raw(query).Exec()
	if err == nil {
		rowsAffected, _ := rs.RowsAffected()
		log.Debugf("Type: 1, Affected rows: %d", rowsAffected)
	}
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
		//if i == 299 {
		//	return nil, errors.New(err.Error() + ": " + filename)
		//}
		//log.Debug(i)
		time.Sleep(100 * time.Millisecond)
	}

	return file, err
	//defer file.Close()

	//scanner := bufio.NewScanner(file)
	//for scanner.Scan() {
	//	//fmt.Println(scanner.Text())
	//}
	//
	//if err := scanner.Err(); err != nil {
	//	return nil, err
	//}

}
