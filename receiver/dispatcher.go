package receiver

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
	"os"
	"path/filepath"
)

var (
	stats = expvar.NewMap("engine")
)

// 처리기
type Dispatcher struct {
	size     int
	duration time.Duration
	dataDir  string
	c        chan *objs.Event
}

func NewDispatcher(size int, duration time.Duration, max int, dataDir string) *Dispatcher {
	return &Dispatcher{
		size:     size,
		duration: duration,
		c:        make(chan *objs.Event, max),
		dataDir:  dataDir,
	}
}

func (d *Dispatcher) Start(errChan chan<- error) error {
	go func() {
		batch := make([]*objs.Event, 0, d.size)
		timer := time.NewTimer(d.duration)
		timer.Stop() // Stop any first firing.

		save := func() {
			stats.Add("eventsIndexed", int64(len(batch)))

			// 임시 파일 생성
			tmpFile, err := ioutil.TempFile("", "")
			if err != nil {
				errChan <- err
			}
			//defer os.Remove(tmpFile.Name())

			// 파일 분류 및 저장
			for _, r := range batch {
				if r.EventType == objs.StatusEvent { // 상태정보
					m := r.Parsed.(map[string]string)
					line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						r.EventType,
						r.SourceIP,
						r.Received.Format(ipasserver.DateDefault),
						m["dt"],
						m["srcid"],
						m["latitude"],
						m["longitude"],
						m["speed"],
					)
					if _, err := tmpFile.WriteString(line); err != nil {
						errChan <- err
					}

				} else if r.EventType == objs.LogEvent { // 이벤트
					m := r.Parsed.(map[string]string)
					line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						r.EventType,
						r.SourceIP,
						r.Received.Format(ipasserver.DateDefault),
						m["dt"],
						m["target"],
						m["wardist"],
						m["caudist"],
						m["v2vdist"],
					)
					if _, err := tmpFile.WriteString(line); err != nil {
						errChan <- err
					}
					log.Debug(filepath.Join(d.dataDir, tmpFile.Name() + "log"))

				} else {
					errChan <- errors.New(fmt.Sprintf("Invalid event type: %d", r.EventType))
				}
			}
			if err := tmpFile.Close(); err != nil {
				errChan <- err
			} else {
				if err := os.Rename(tmpFile.Name(), filepath.Join(d.dataDir, filepath.Base(tmpFile.Name()) + ".log")); err != nil {
					errChan <- err
				}
			}
			log.Debugf("Saved to file: %s", tmpFile.Name())

			// 파일 닫기 및 이동
			batch = make([]*objs.Event, 0, d.size)
		}

		for {
			select {
			case event := <-d.c:
				log.Debugf("### GOT[%d]: %s", event.EventType, event.Received.Format(ipasserver.DateDefault))
				//spew.Dump(event)
				batch = append(batch, event)
				if len(batch) == 1 {
					timer.Reset(d.duration)
				}
				if len(batch) == d.size {
					log.Debugf("### FULL")
					timer.Stop()
					save()
				}
			case <-timer.C:
				log.Debugf("### TIMEOUT")
				stats.Add("batchTimeout", 1)
				save()
			}
		}
	}()

	return nil
}

func (d *Dispatcher) C() chan<- *objs.Event {
	return d.c
}
