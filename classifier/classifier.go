package classifier

import (
	"bufio"
	"errors"
	"github.com/devplayg/ipas-server"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

var count uint64

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

	var file *os.File
	// 파일 읽기
	if file, err := openFile(filename); err != nil {
		log.Error(err)
		return err
	}

	file.



	// 메모리에서 데이터 분류

	// 파일 저장 및 DB 입력

	// 파일 삭제

	new := atomic.AddUint64(&count, 1)
	log.Debugf("done: %d", new)
	return nil
	//log.Debug("start: " + name)
	//time.Sleep(1000 * time.Millisecond)
	//

	//new := atomic.AddUint64(&count, 1)
	//log.Debugf("done: %d", new)

}

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
		if i == 299 {
			return nil, errors.New(err.Error() + ": " + filename)
		}
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
