package objs

import "time"

const (
	LogEvent    = 1
	StatusEvent = 2
	RuleEvent   = 3

	StartEvent     = 1 // 시동
	ShockEvent     = 2 // 충격
	SpeedingEvent  = 3 // 과속
	ProximityEvent = 4 // 근접
)

type Event struct {
	Parsed    interface{} // 데이터
	Received  time.Time   // 수신 시간
	EventType int         // 이벤트 타입
	SourceIP  string      // 요청 IP
}

func NewEvent(eventType int, ipStr string) *Event {
	return &Event{
		Received:  time.Now(),
		EventType: eventType,
		SourceIP:  ipStr,
	}
}

type IpasEvent struct {
	OrgId     int
	GroupId   int
	EventType int
	EquipId   string
	Targets   string
	Timeline  string
}

type IpasStatus struct {
	Date      time.Time
	OrgId     int
	GroupId   int
	EquipId   string
	SessionId string
}
