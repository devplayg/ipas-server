package receiver

import (
	"github.com/devplayg/ipas-server/objs"
	"github.com/julienschmidt/httprouter"
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
		m := make(map[string]string)
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
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

		c <- event
	})

	r.router.GET("/status", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		m := make(map[string]string)
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
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
		m["cstid"] = req.Form.Get("cstid")
		m["srcid"] = req.Form.Get("srcid")
		m["dstid"] = req.Form.Get("dstid")
		m["lat"] = req.Form.Get("lat")
		m["lon"] = req.Form.Get("lon")
		m["spd"] = req.Form.Get("spd")
		m["snr"] = req.Form.Get("snr")
		m["ctn"] = req.Form.Get("ctn")
		m["type"] = req.Form.Get("type")
		//if m["type"] == "1" {
		//	return
		//}
		m["dist"] = req.Form.Get("dist")
		m["sesid"] = req.Form.Get("sesid")
		event.Parsed = m

		c <- event
	})

	e.router.GET("/event", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		m := make(map[string]string)
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
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
		//if m["type"] == "1" {
		//	return
		//}
		m["dist"] = req.Form.Get("dist")
		m["sesid"] = req.Form.Get("sesid")
		event.Parsed = m

		c <- event
	})
	return nil
}
