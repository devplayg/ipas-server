package receiver

import (
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//type Receiver interface {
//	Start(chan<- *objs.Event) error
//}

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
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		event := objs.NewEvent(objs.StatusEvent, host)
		u, _ := url.ParseRequestURI(req.RequestURI)
		//spew.Println("Req: " + req.RequestURI)
		parsed, _ := url.ParseQuery(u.RawQuery)
		//spew.Println("Raw: " + u.RawQuery)
		//spew.Dump(parsed)
		status := objs.IpasStatus{}
		status.Date, _ = time.Parse(ipasserver.DateDefault, parsed.Get("dt"))
		status.ID = parsed.Get("srcid")
		latitude, _ := strconv.ParseFloat(parsed.Get("lat"), 32)
		status.Latitude = float32(latitude)
		longitude, _ := strconv.ParseFloat(parsed.Get("lat"), 32)
		status.Longitude = float32(longitude)
		speed, _ := strconv.ParseFloat(parsed.Get("spd"), 32)
		status.Speed = float32(speed)
		event.Parsed = status
		event.Received = status.Date

		c <- event
	})
	return nil
}

//func (r *StatusReceiver) Handler(resp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
//	//event, _ := r.parseRequest(req)
//}

//func (r *StatusReceiver) parseRequest(req *http.Request) (*objs.Event, error) {
//	host, _, _ := net.SplitHostPort(req.RemoteAddr)
//	event := objs.NewEvent(objs.StatusEvent, host)
//	u, _ := url.ParseRequestURI(req.RequestURI)
//	parsed, _ := url.ParseQuery(u.RawQuery)
//	status := objs.IpasStatus{}
//	status.Date, _ = time.Parse(ipasserver.DateDefault, parsed.Get("dt"))
//	status.ID = parsed.Get("srcid")
//	latitude, _ := strconv.ParseFloat(parsed.Get("lat"), 32)
//	status.Latitude = float32(latitude)
//	longitude, _ := strconv.ParseFloat(parsed.Get("lat"), 32)
//	status.Longitude = float32(longitude)
//	speed, _ := strconv.ParseFloat(parsed.Get("spd"), 32)
//	status.Speed = float32(speed)
//	event.Parsed = status
//	return event, nil
//}

// 이벤트 수신기
type EventReceiver struct {
	router *httprouter.Router
}

func NewEventReceiver(router *httprouter.Router) *EventReceiver {
	receiver := EventReceiver{router}
	return &receiver
}

func (e *EventReceiver) Start(c chan<- *objs.Event) error {
	e.router.GET("/event", func(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		event := objs.NewEvent(objs.StatusEvent, host)
		u, _ := url.ParseRequestURI(req.RequestURI)
		parsed, _ := url.ParseQuery(u.RawQuery)
		event.Parsed = parsed
		date, _ := time.Parse(ipasserver.DateDefault, parsed.Get("dt"))
		event.Received = date

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
