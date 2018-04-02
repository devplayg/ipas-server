package receiver

import (
	"github.com/devplayg/ipas-server/objs"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"net/url"
)

// http://127.0.0.1:8080/status?dt=2006-01-02%2015%3A04%3A05&srcid=VTSAMPLE&lat=126.886559&lon=37.480888&spd=1234.1

// 상태정보 수신기
type StatusReceiver struct {
	router *httprouter.Router
}

func NewStatusReceiver(router *httprouter.Router) *StatusReceiver {
	receiver := StatusReceiver{router}
	return &receiver
}

func (r *StatusReceiver) Start(c chan<- *objs.Event) error {
	r.router.GET("/status", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		m := make(map[string]string)
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		event := objs.NewEvent(objs.StatusEvent, host)

		u, _ := url.ParseRequestURI(req.RequestURI)
		parsed, _ := url.ParseQuery(u.RawQuery)

		m["dt"] = parsed.Get("dt")
		m["srcid"] = parsed.Get("srcid")
		m["latitude"] = parsed.Get("lat")
		m["longitude"] = parsed.Get("lon")
		m["speed"] = parsed.Get("spd")
		event.Parsed = m

		c <- event
	})
	return nil
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
		m := make(map[string]string)
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		event := objs.NewEvent(objs.LogEvent, host)

		req.ParseForm()
		m["dt"] = req.Form.Get("dt")
		m["target"] = req.Form.Get("target")
		m["wardist"] = req.Form.Get("wardist")
		m["caudist"] = req.Form.Get("caudist")
		m["v2vdist"] = req.Form.Get("v2vdist")
		event.Parsed = m

		c <- event
	})
	return nil
}

//func (r *EventReceiver) Handler(resp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
//	event, _ := r.parseRequest(req)
//	spew.Dump(event)
//}
//
//func (r *EventReceiver) parseRequest(req *http.Request) (*objs.Event, error) {
//	host, _, _ := net.SplitHostPort(req.RemoteAddr)
//	event := objs.NewEvent(objs.LogEvent, host)
//	u, _ := url.ParseRequestURI(req.RequestURI)
//	parsed, _ := url.ParseQuery(u.RawQuery)
//	event.Parsed = parsed
//	return event, nil
//}
