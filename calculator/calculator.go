package calculator

import "github.com/devplayg/ipas-server"

type Calculator interface {
	Start() error
}

func NewCalculator(engine *ipasserver.Engine, s string) Calculator {
	if s == "event" {
		return NewEventCalculator(engine)
	} else if s == "status" {
		return NewStatusCalculator(engine)
	} else {
		return nil
	}
}