package receiver

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	stats = expvar.NewMap("engine")
)

// 처리기
type Stacker struct {
	size     int
	duration time.Duration
	c        chan *objs.Event
	engine   *ipasserver.Engine
	tmpDir   string
}

func NewStacker(size int, duration time.Duration, max int, engine *ipasserver.Engine) *Stacker {
	return &Stacker{
		size:     size,
		duration: duration,
		c:        make(chan *objs.Event, max),
		engine:   engine,
		tmpDir:   filepath.Join(engine.ProcessDir, "tmp"),
	}
}

func (s *Stacker) Start(errChan chan<- error) error {
	go func() {
		batch := make([]*objs.Event, 0, s.size)
		timer := time.NewTimer(s.duration)
		timer.Stop() // Stop any first firing.

		save := func() {
			// 임시 파일 생성
			tmpFile, err := ioutil.TempFile(s.tmpDir, "")
			if err != nil {
				errChan <- err
			}
			//log.Debug(tmpFile.Name())

			// 파일 분류 및 저장
			for _, r := range batch {
				if r.EventType == objs.LogEvent { // 이벤트
					m := r.Parsed.(map[string]string)
					line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						r.EventType,
						r.SourceIP,
						r.Received.Format(ipasserver.DateDefault),
						m["dt"],
						m["orgcode"],
						m["sesid"],
						m["srcid"],
						m["dstid"],
						m["lat"],
						m["lon"],
						m["spd"],
						m["snr"],
						m["ctn"],
						m["type"],
						m["dist"],
					)
					if _, err := tmpFile.WriteString(line); err != nil {
						errChan <- err
					}

				} else if r.EventType == objs.StatusEvent { // 상태정보
					m := r.Parsed.(map[string]string)
					line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						r.EventType,
						r.SourceIP,
						r.Received.Format(ipasserver.DateDefault),
						m["dt"],
						m["orgcode"],
						m["sesid"],
						m["srcid"],
						m["lat"],
						m["lon"],
						m["spd"],
						m["snr"],
						m["ctn"],
					)
					if _, err := tmpFile.WriteString(line); err != nil {
						errChan <- err
					}

				} else {
					errChan <- errors.New(fmt.Sprintf("Invalid event type: %d", r.EventType))
				}
			}
			if err := tmpFile.Close(); err != nil {
				errChan <- err
			} else {
				if err := os.Rename(tmpFile.Name(), filepath.Join(s.engine.ProcessDir, "data", filepath.Base(tmpFile.Name())+".log")); err != nil {
					errChan <- err
				}
			}
			log.Debugf("Received: %d", len(batch))

			// 파일 닫기 및 이동
			batch = make([]*objs.Event, 0, s.size)
		}

		for {
			select {
			case event := <-s.c:
				log.Debugf("### GOT[%d]: %s", event.EventType, event.Received.Format(ipasserver.DateDefault))
				//spew.Dump(event)
				batch = append(batch, event)
				if len(batch) == 1 {
					timer.Reset(s.duration)
				}
				if len(batch) == s.size {
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

func (s *Stacker) C() chan<- *objs.Event {
	return s.c
}
