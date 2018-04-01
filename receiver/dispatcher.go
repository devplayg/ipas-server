package receiver

import (
	"github.com/devplayg/ipas-server/objs"
	"time"
	log "github.com/sirupsen/logrus"
	"expvar"
	"github.com/devplayg/ipas-server"
)

var (
	stats = expvar.NewMap("engine")
)


// 처리기
type Dispatcher struct {
	size     int
	duration time.Duration

	c chan *objs.Event
}

func NewDispatcher(size int, duration time.Duration, max int) *Dispatcher {
	return &Dispatcher{
		size:     size,
		duration: duration,
		c:        make(chan *objs.Event, max),
	}
}

func (d *Dispatcher) Start(errChan chan<- error) error {
	go func() {
		batch := make([]*objs.Event, 0, d.size)
		timer := time.NewTimer(d.duration)
		timer.Stop() // Stop any first firing.

		send := func() {
			//stats.Add("batchIndexed", 1)
			stats.Add("eventsIndexed", int64(len(batch)))
			//if errChan != nil {
			//	errChan <- err
			//}
			log.Debug("### Saving")
			time.Sleep(3 * time.Second)
			log.Debugf("### Saved: %d", len(batch))
			batch = make([]*objs.Event, 0, d.size)

		}

		for {
			select {
			case event := <-d.c:
				log.Debugf("### GOT: %s", event.Received.Format(ipasserver.DateDefault))
				//spew.Dump(event)
				batch = append(batch, event)
				if len(batch) == 1 {
					timer.Reset(d.duration)
				}
				if len(batch) == d.size {
					log.Debugf("### FULL")
					timer.Stop()
					send()
				}
			case <-timer.C:
				log.Debugf("### TIMEOUT")
				stats.Add("batchTimeout", 1)
				send()
			}
		}
	}()

	return nil
}

func (d *Dispatcher) C() chan<- *objs.Event {
	return d.c
}