package receiver

import (
	"github.com/devplayg/ipas-server"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"net/http"
)

type EventReceiver struct {
	e *ipasserver.Engine
	r *httprouter.Router
}

func NewEventReceiver(engine *ipasserver.Engine, router *httprouter.Router) *EventReceiver {
	receiver := EventReceiver{engine, router}
	router.GET("/event", receiver.Handler)
	return &receiver
}

func (c *EventReceiver) Start() error {
	return nil
}

func (c *EventReceiver)Handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "Event handler")
}


type StatusReceiver struct {
	engine *ipasserver.Engine
	router *httprouter.Router
}

func NewStatusReceiver(engine *ipasserver.Engine, router *httprouter.Router) *StatusReceiver {
	receiver := StatusReceiver{engine, router}
	router.GET("/status", receiver.Handler)
	return &receiver
}

func (s *StatusReceiver) Start() error {
	return nil
}

func (s *StatusReceiver)Handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "Status handler")
}
