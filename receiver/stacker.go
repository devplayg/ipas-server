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
			tmpFile, err := ioutil.TempFile(s.tmpDir, "data_")
			if err != nil {
				errChan <- err
			}
			//log.Debug(tmpFile.Name())

			// 파일 분류 및 저장
			for _, r := range batch {
				if r.EventType == objs.LogEvent { // 이벤트
					m := r.Parsed.(map[string]string)
					line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						r.EventType, //	0
						r.SourceIP,  //	1
						r.Received.UTC().Format(ipasserver.DateDefault), //	2
						m["dt"],                                         //	3
						m["sesid"],                                      //	4
						m["cstid"],                                      //	5 **
						m["srcid"],                                      //	6
						m["dstid"],                                      //	7
						m["lat"],                                        //	8
						m["lon"],                                        //	9
						m["spd"],                                        //	10
						m["snr"],                                        //	11
						m["ctn"],                                        //	12
						m["type"],                                       //	13
						m["dist"],                                       //	14,
						m["tz"],                                         //	15
					)
					if _, err := tmpFile.WriteString(line); err != nil {
						errChan <- err
					}

				} else if r.EventType == objs.StatusEvent { // 상태정보
					m := r.Parsed.(map[string]string)
					line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						r.EventType, // 0
						r.SourceIP,  // 1
						r.Received.UTC().Format(ipasserver.DateDefault), //	2
						m["dt"],                                         //	3
						m["sesid"],                                      //	4
						m["cstid"],                                      //	5 **
						m["srcid"],                                      //	6
						m["lat"],                                        //	7
						m["lon"],                                        //	8
						m["spd"],                                        //	9
						m["snr"],                                        //	10
						m["ctn"],                                        //	11
						m["tz"],                                         //	12
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
			//log.Debugf("Received: %d", len(batch))
			log.WithFields(log.Fields{
				"received": len(batch),
				"file":     tmpFile.Name(),
			}).Debug()

			// 파일 닫기 및 이동
			batch = make([]*objs.Event, 0, s.size)
		}

		for {
			select {
			case event := <-s.c:
				//log.Debugf("### GOT[%d]: %s", event.EventType, event.Received.Format(ipasserver.DateDefault))
				log.Debug(event)
				batch = append(batch, event)
				if len(batch) == 1 {
					timer.Reset(s.duration)
				}
				if len(batch) == s.size {
					log.Debug("queue is full")
					timer.Stop()
					save()
				}
			case <-timer.C:
				log.Debug("queue timeout")
				//stats.Add("batchTimeout", 1)
				save()
			}
		}
	}()

	return nil
}

//func (s *Stacker) formatTime(timeStr, tz string) string {
//	var timeZone string
//	if len(tz) == 0 { // 타임존 정보가 없으면, 서버 타임존 참고
//		t, err := time.ParseInLocation(ipasserver.DateDefault, timeStr, s.engine.TimeZone)
//		if err != nil {
//			log.Debugf("tz=%s, str=%s", timeZone, timeStr)
//			return timeStr
//		}
//		log.Debugf("tz=%s, str=%s", timeZone, timeStr)
//		return t.UTC().Format(ipasserver.DateDefault)
//	} else if len(tz) == 4 {
//		timeZone = "+"+tz
//	} else {
//		timeZone = tz
//	}
//	log.Debugf("tz=%s, str=%s", timeZone, timeStr)
//	t, err := time.Parse("2006-01-02 15:04:05Z0700",timeStr+timeZone)
//	if err != nil {
//		log.Error(err)
//		return timeStr
//	}
//
//	return t.UTC().Format(ipasserver.DateDefault)
//}

func (s *Stacker) C() chan<- *objs.Event {
	return s.c
}
