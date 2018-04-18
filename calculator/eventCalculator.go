package calculator

import (
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Stats interface {
	Start(wg *sync.WaitGroup) error
}

func NewStats(s string, calculator *Calculator) Stats {
	if s == "event" {
		return NewEventStats(calculator)
	} else if s == "status" {
		return NewStatusStats(calculator)
	} else {
		return nil
	}
}

// ---------------------------------------------------------------------------------------------

type eventStats struct {
	calculator *Calculator
	wg         *sync.WaitGroup
	dataMap    objs.DataMap
	rank       objs.DataRank
}

func NewEventStats(calculator *Calculator) *eventStats {
	return &eventStats{
		calculator: calculator,
	}
}

func (s *eventStats) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	log.Debug("eventStats start")
	time.Sleep(2 * time.Second)
	return nil
}

// ---------------------------------------------------------------------------------------------

type statusStats struct {
	calculator *Calculator
	wg         *sync.WaitGroup
}

func NewStatusStats(calculator *Calculator) *statusStats {
	return &statusStats{
		calculator: calculator,
	}
}

func (s *statusStats) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	log.Debug("eventStats start")
	time.Sleep(2 * time.Second)
	return nil
}
