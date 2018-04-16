package calculator

import (
	"github.com/devplayg/ipas-server"
)

type eventCalculator struct {
	engine *ipasserver.Engine
}

func NewEventCalculator(engine *ipasserver.Engine) Calculator {
	return &eventCalculator{
		engine: engine,
	}
}

func (c *eventCalculator) Start() error {
	return nil
}


type statusCalculator struct {
	engine *ipasserver.Engine
}

func NewStatusCalculator(engine *ipasserver.Engine) Calculator {
	return &statusCalculator{
		engine: engine,
	}
}

func (c *statusCalculator) Start() error {
	return nil
}