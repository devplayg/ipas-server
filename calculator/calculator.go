package calculator

import (
	"github.com/devplayg/ipas-server"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Calculator struct {
	engine   *ipasserver.Engine
	top      int
	interval int64
	date     string
	report   string
}

func NewCalculator(engine *ipasserver.Engine, top int, interval int64, date, report string) *Calculator {
	return &Calculator{
		engine:   engine,
		top:      top,
		interval: interval,
		date:     date,
		report:   report,
	}
}

func (c *Calculator) calculate(from, to, mark string) error {
	log.Debugf("[%s ~ %s] Mark as %s", from, to, mark)
	return nil
}

func (c *Calculator) Start() error {
	log.Debug("start")
	if len(c.date) > 0 { // 특정 지정한 날짜에 대한 통계 생성
		t, err := time.Parse("2006-01-02", c.date)
		if err != nil {
			return err
		}
		log.Debugf("Target date: %s", c.date)
		err = c.calculate(
			t.Format("2006-01-02")+" 00:00:00",
			t.Format("2006-01-02")+" 23:59:59",
			t.Format("2006-01-02")+" 23:59:59",
		)
		return err

	} else if len(c.report) > 0 { // 시스템에서 생성하는 보고서 생성 시
		timeArr := strings.Split(c.report, ",")
		from, err := time.Parse("2006-01-02", timeArr[0])
		if err != nil {
			return err
		}
		to, err := time.Parse("2006-01-02", timeArr[1])
		if err != nil {
			return err
		}
		mark, err := time.Parse(ipasserver.DateDefault, timeArr[2])
		if err != nil {
			return err
		}
		log.Debugf("Target date: %s", c.report)
		c.calculate(
			from.Format("2006-01-02")+" 00:00:00",
			to.Format("2006-01-02")+" 00:00:00",
			mark.Format(ipasserver.DateDefault),
		)

		return nil
	}

	go func() {
		for {
			t := time.Now()

			err := c.calculate(
				t.Format("2006-01-02")+" 00:00:00",
				t.Format("2006-01-02")+" 23:59:59",
				t.Format(ipasserver.DateDefault),
			)
			if err != nil {
				log.Error(err)
			}
			time.Sleep(time.Duration(c.interval) * time.Millisecond)
		}
	}()

	return nil
}

//type Calculator interface {
//	Start() error
//}

//func NewCalculator(engine *ipasserver.Engine, logType, top int, interval int64, date, manual string) Calculator {
//	if logType == objs.LogEvent {
//		return NewEventCalculator(engine, top, interval, date, manual)
//
//	} else if logType == objs.StatusEvent {
//		return NewStatusCalculator(engine, top, interval, date, manual)
//
//	} else {
//		return nil
//	}
//}
