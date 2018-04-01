package objs

import "time"

const (
	LogEvent    = 1
	StatusEvent = 2
	RuleEvent   = 3
)

type Event struct {
	Parsed    interface{} // 데이터
	Received  time.Time              // 수신 시간
	eventType int                    // 이벤트 타입
	sourceIP  string                 // 요청 IP
}

func NewEvent(eventType int, ipStr string) *Event {
	return &Event{
		Received:  time.Now(),
		eventType: eventType,
		sourceIP:  ipStr,
	}
}

type IpasStatus struct {
	Date       time.Time
	ID         string
	Speed      float32
	Latitude   float32
	Longitude  float32
	ShockCount int
}
