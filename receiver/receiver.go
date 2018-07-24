package receiver

import (
	"github.com/devplayg/ipas-server/objs"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"

)

// 상태정보 수신기
type StatusReceiver struct {
	router *httprouter.Router
}

func NewStatusReceiver(router *httprouter.Router) *StatusReceiver {
	receiver := StatusReceiver{router}
	return &receiver
}

func (r *StatusReceiver) Start(c chan<- *objs.Event) error {
	r.router.POST("/status", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		event := createStatusObj(req)
		c <- event
	})

	r.router.GET("/status", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		event := createStatusObj(req)
		c <- event
	})
	return nil
}

func createStatusObj(req *http.Request) *objs.Event {
	m := make(map[string]string)
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Error(err)
	}
	event := objs.NewEvent(objs.StatusEvent, host)

	req.ParseForm()

	m["dt"] = req.Form.Get("dt")
	m["cstid"] = req.Form.Get("cstid")
	m["srcid"] = req.Form.Get("srcid")
	m["lat"] = req.Form.Get("lat")
	m["lon"] = req.Form.Get("lon")
	m["spd"] = req.Form.Get("spd")
	m["snr"] = req.Form.Get("snr")
	m["ctn"] = req.Form.Get("ctn")
	m["sesid"] = req.Form.Get("sesid")

	event.Parsed = m
	return event
}

// 이벤트 수신기
type EventReceiver struct {
	router *httprouter.Router
}

func NewEventReceiver(router *httprouter.Router) *EventReceiver {
	receiver := EventReceiver{router}
	return &receiver
}

func (e *EventReceiver) Start(c chan<- *objs.Event) error {
	e.router.POST("/event", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		event := createEventObj(req)
		c <- event
	})

	e.router.GET("/event", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		event := createEventObj(req)
		c <- event
	})
	return nil
}

func createEventObj(req *http.Request) *objs.Event {
	m := make(map[string]string)
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Error(err)
	}
	event := objs.NewEvent(objs.LogEvent, host)

	req.ParseForm()
	m["dt"] = req.Form.Get("dt")
	m["cstid"] = req.Form.Get("cstid")
	m["srcid"] = req.Form.Get("srcid")
	m["dstid"] = req.Form.Get("dstid")
	m["lat"] = req.Form.Get("lat")
	m["lon"] = req.Form.Get("lon")
	m["spd"] = req.Form.Get("spd")
	m["snr"] = req.Form.Get("snr")
	m["ctn"] = req.Form.Get("ctn")
	m["type"] = req.Form.Get("type")
	m["dist"] = req.Form.Get("dist")
	m["sesid"] = req.Form.Get("sesid")
	event.Parsed = m

	return event
}
