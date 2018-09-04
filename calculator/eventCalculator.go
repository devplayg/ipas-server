package calculator

import (
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Stats interface {
	Start(wg *sync.WaitGroup) error
}

func NewStats(calculator *Calculator, stats int, from, to, mark string) Stats {
	if stats == EventStats {
		return NewEventStats(calculator, from, to, mark)

	} else if stats == ExtraStats {
		return NewExtraStats(calculator, from, to, mark)

	} else {
		return nil
	}
}

// ---------------------------------------------------------------------------------------------
type eventStatsCalculator struct {
	calculator          *Calculator
	wg                  *sync.WaitGroup
	dataMap             objs.DataMap
	dataRank            objs.DataRank
	equipStats          map[int]map[string]map[int]int
	timelineStats       map[int]map[int]map[string]map[int]int // 개발 중(org_id, group_id, hour, evt1~4)
	shockLinksStats     map[int]map[int][]string
	tables              map[string]bool
	sessionByGroupStats map[int]map[int]map[string]int
	sessionByEquipStats map[int]map[string]map[string]int
	optimeByGroupStats  map[int]map[int]map[string][]time.Time
	optimeByEquipStats  map[int]map[string]map[string][]time.Time
	from                string
	to                  string
	mark                string
	//mutex           sync.Mutex
}

func NewEventStats(calculator *Calculator, from, to, mark string) *eventStatsCalculator {
	return &eventStatsCalculator{
		calculator: calculator,
		dataMap:    make(objs.DataMap),
		dataRank:   make(objs.DataRank),
		tables: map[string]bool{ // true:전체데이터 유지, false: TopN 데이터만 유지
			"evt": true,
		},
		from: from,
		to:   to,
		mark: mark,
	}
}

func (c *eventStatsCalculator) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	start := time.Now()

	// 통계 생성
	if err := c.produceEventStats(); err != nil {
		log.Error(err)
		return err
	}

	// 통계 생성
	if err := c.produceExtraStats(); err != nil {
		log.Error(err)
		return err
	}

	if c.calculator.calType == objs.RealtimeCalculator {
		//c.mutex.Lock()
		//c.calculator.eventRank = c.dataRank
		//c.mutex.Unlock()
	}

	// DB 입력
	if err := c.insert(); err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("cal_type=%d, stats_type=%s, exec_time=%3.1fs", c.calculator.calType, StatsDesc[EventStats], time.Since(start).Seconds())
	return nil
}

func (c *eventStatsCalculator) produceEventStats() error {

	// 통계 구조체 초기화
	c.dataMap[RootId] = make(map[string]map[interface{}]int64)
	c.dataRank[RootId] = make(map[string]objs.ItemList)
	c.equipStats = make(map[int]map[string]map[int]int)
	c.timelineStats = make(map[int]map[int]map[string]map[int]int)
	c.shockLinksStats = make(map[int]map[int][]string)

	// 이벤트 로그 조회
	query := `
		select org_id, group_id, event_type, equip_id, targets, concat(substr(date, 1, 13), ':00:00') hour
		from log_ipas_event
		where date between ? and ?
	`
	rows, err := c.calculator.engine.DB.Query(query, c.from, c.to)
	if err != nil {
		log.Error(err)
		return err
	}
	defer rows.Close()

	// 이벤트 맵 생성
	for rows.Next() {

		// 이벤트 객체 생성
		e := objs.IpasEvent{}

		// 데이터 읽기
		err := rows.Scan(&e.OrgId, &e.GroupId, &e.EventType, &e.EquipId, &e.Targets, &e.Timeline)
		if err != nil {
			return err
		}

		// 이벤트 유형 통계
		c.addToEventStats(&e, "evt", e.EventType)

		// 이벤트 타입별 Src tag 통계
		if e.EventType >= 0 && e.EventType <= 4 {
			evt := strconv.Itoa(e.EventType)
			c.addToEventStats(&e, "evt"+evt+"_by_equip", e.EquipId) // eventtype1~4
			c.addToEventStats(&e, "evt"+evt+"_by_group", fmt.Sprintf("%d/%d", e.OrgId, e.GroupId))

			if e.EventType == objs.ShockEvent {
				c.addToShockLinksStats(&e)
			}
		}

		// 타임라인 통계
		c.addToTimelineStats(&e)

		// 장비(Tag) 이벤트 통계
		if _, ok := c.equipStats[e.OrgId]; !ok {
			c.equipStats[e.OrgId] = make(map[string]map[int]int)
		}
		if _, ok := c.equipStats[e.OrgId][e.EquipId]; !ok {
			c.equipStats[e.OrgId][e.EquipId] = map[int]int{
				objs.StartEvent:     0,
				objs.ShockEvent:     0,
				objs.SpeedingEvent:  0,
				objs.ProximityEvent: 0,
			}
		}
		c.equipStats[e.OrgId][e.EquipId][e.EventType]++
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
		return err
	}

	// 순위 계산
	for id, m := range c.dataMap {
		for category, data := range m {
			if keepAll, ok := c.tables[category]; ok {
				if keepAll { // 모든 순위 보관
					c.dataRank[id][category] = objs.DetermineRankings(data, 0)
				} else { // Top N 데이터만 보관
					c.dataRank[id][category] = objs.DetermineRankings(data, c.calculator.top)
				}
			} else { // Top N 데이터만 보관
				c.dataRank[id][category] = objs.DetermineRankings(data, c.calculator.top)
			}
		}
	}

	return nil
}

func (c *eventStatsCalculator) produceExtraStats() error {

	// 통계 구조체 초기화
	c.sessionByGroupStats = make(map[int]map[int]map[string]int)
	c.sessionByEquipStats = make(map[int]map[string]map[string]int)
	c.optimeByGroupStats = make(map[int]map[int]map[string][]time.Time)
	c.optimeByEquipStats = make(map[int]map[string]map[string][]time.Time)

	// 상태정보 조회
	query := `
		select date, org_id, group_id, equip_id, session_id
		from log_ipas_status
		where date between ? and ?
	`
	rows, err := c.calculator.engine.DB.Query(query, c.from, c.to)
	if err != nil {
		log.Error(err)
		return err
	}
	defer rows.Close()

	// 이벤트 맵 생성
	for rows.Next() {

		// 이벤트 객체 생성
		e := objs.IpasStatus{}

		// 데이터 읽기
		err := rows.Scan(&e.Date, &e.OrgId, &e.GroupId, &e.EquipId, &e.SessionId)
		if err != nil {
			log.Error(err)
			return err
		}

		// 그룹별 세션 수
		if _, ok := c.sessionByGroupStats[e.OrgId]; !ok {
			c.sessionByGroupStats[e.OrgId] = make(map[int]map[string]int)
			c.optimeByGroupStats[e.OrgId] = make(map[int]map[string][]time.Time)
		}
		if _, ok := c.sessionByGroupStats[e.OrgId][e.GroupId]; !ok {
			c.sessionByGroupStats[e.OrgId][e.GroupId] = make(map[string]int)
			c.optimeByGroupStats[e.OrgId][e.GroupId] = make(map[string][]time.Time)
		}
		c.sessionByGroupStats[e.OrgId][e.GroupId][e.SessionId]++

		// 장비별 세션 수
		if _, ok := c.sessionByEquipStats[e.OrgId]; !ok {
			c.sessionByEquipStats[e.OrgId] = make(map[string]map[string]int)
			c.optimeByEquipStats[e.OrgId] = make(map[string]map[string][]time.Time)
		}
		if _, ok := c.sessionByEquipStats[e.OrgId][e.EquipId]; !ok {
			c.sessionByEquipStats[e.OrgId][e.EquipId] = make(map[string]int)
			c.optimeByEquipStats[e.OrgId][e.EquipId] = make(map[string][]time.Time)
		}
		c.sessionByEquipStats[e.OrgId][e.EquipId][e.SessionId]++

		// 시간 측정 - 그룹별
		if _, ok := c.optimeByGroupStats[e.OrgId][e.GroupId][e.SessionId]; !ok {
			c.optimeByGroupStats[e.OrgId][e.GroupId][e.SessionId] = []time.Time{e.Date, e.Date}
		} else {
			if e.Date.Before(c.optimeByGroupStats[e.OrgId][e.GroupId][e.SessionId][0]) {
				arr := c.optimeByGroupStats[e.OrgId][e.GroupId][e.SessionId]
				arr[0] = e.Date
			} else if e.Date.After(c.optimeByGroupStats[e.OrgId][e.GroupId][e.SessionId][1]) {
				arr := c.optimeByGroupStats[e.OrgId][e.GroupId][e.SessionId]
				arr[1] = e.Date
			}
		}

		// 시간 측정 - 장비별
		if _, ok := c.optimeByEquipStats[e.OrgId][e.EquipId][e.SessionId]; !ok {
			c.optimeByEquipStats[e.OrgId][e.EquipId][e.SessionId] = []time.Time{e.Date, e.Date}
		} else {
			if e.Date.Before(c.optimeByEquipStats[e.OrgId][e.EquipId][e.SessionId][0]) {
				arr := c.optimeByEquipStats[e.OrgId][e.EquipId][e.SessionId]
				arr[0] = e.Date
			} else if e.Date.After(c.optimeByEquipStats[e.OrgId][e.EquipId][e.SessionId][1]) {
				arr := c.optimeByEquipStats[e.OrgId][e.EquipId][e.SessionId]
				arr[1] = e.Date
			}
		}
	}
	return nil
}

func (c *eventStatsCalculator) addToTimelineStats(e *objs.IpasEvent) error {

	// map[int]map[int]map[string]map[int]int // 개발 중
	if _, ok := c.timelineStats[e.OrgId]; !ok {
		c.timelineStats[e.OrgId] = make(map[int]map[string]map[int]int)
	}
	if _, ok := c.timelineStats[e.OrgId][e.GroupId]; !ok {
		c.timelineStats[e.OrgId][e.GroupId] = make(map[string]map[int]int)
	}
	if _, ok := c.timelineStats[e.OrgId][e.GroupId][e.Timeline]; !ok {
		c.timelineStats[e.OrgId][e.GroupId][e.Timeline] = map[int]int{
			objs.StartEvent:     0,
			objs.ShockEvent:     0,
			objs.SpeedingEvent:  0,
			objs.ProximityEvent: 0,
		}
	}
	c.timelineStats[e.OrgId][e.GroupId][e.Timeline][e.EventType] += 1

	return nil
}

func (c *eventStatsCalculator) addToShockLinksStats(e *objs.IpasEvent) error {

	if _, ok := c.shockLinksStats[e.OrgId]; !ok {
		c.shockLinksStats[e.OrgId] = make(map[int][]string)
	}

	if _, ok := c.shockLinksStats[e.OrgId][e.GroupId]; !ok {
		c.shockLinksStats[e.OrgId][e.GroupId] = make([]string, 0)
	}
	c.shockLinksStats[e.OrgId][e.GroupId] = append(c.shockLinksStats[e.OrgId][e.GroupId], fmt.Sprintf("%s/%s", e.EquipId, e.Targets))

	return nil
}

func (c *eventStatsCalculator) addToEventStats(e *objs.IpasEvent, category string, val interface{}) error {

	// 전체 통계
	if _, ok := c.dataMap[RootId][category]; !ok {
		c.dataMap[RootId][category] = make(map[interface{}]int64)
		c.dataRank[RootId][category] = nil
	}
	c.dataMap[RootId][category][val] += 1

	// 기관 통계
	if _, ok := c.dataMap[e.OrgId]; !ok {
		c.dataMap[e.OrgId] = make(map[string]map[interface{}]int64)
		c.dataRank[e.OrgId] = make(map[string]objs.ItemList)
	}
	if _, ok := c.dataMap[e.OrgId][category]; !ok {
		c.dataMap[e.OrgId][category] = make(map[interface{}]int64)
		c.dataRank[e.OrgId][category] = nil
	}
	c.dataMap[e.OrgId][category][val]++

	// 그룹 통계
	if e.GroupId > 0 {
		if _, ok := c.dataMap[e.GroupId]; !ok {
			c.dataMap[e.GroupId] = make(map[string]map[interface{}]int64)
			c.dataRank[e.GroupId] = make(map[string]objs.ItemList)
		}
		if _, ok := c.dataMap[e.GroupId][category]; !ok {
			c.dataMap[e.GroupId][category] = make(map[interface{}]int64)
			c.dataRank[e.GroupId][category] = nil
		}
		c.dataMap[e.GroupId][category][val]++

		// 사용자 소속 권한의 전체 통계
		if arr, ok := c.calculator.memberAssets[e.GroupId]; ok {
			for _, memberId := range arr {
				id := memberId * -1

				if _, ok := c.dataMap[id]; !ok {
					c.dataMap[id] = make(map[string]map[interface{}]int64)
					c.dataRank[id] = make(map[string]objs.ItemList)
				}
				if _, ok := c.dataMap[id][category]; !ok {
					c.dataMap[id][category] = make(map[interface{}]int64)
					c.dataRank[id][category] = nil
				}
				c.dataMap[id][category][val]++
			}
		}
	}

	return nil
}

func (c *eventStatsCalculator) insert() error {
	fm := make(map[string]*os.File)
	defer func() {
		if !c.calculator.engine.IsDebug() {
			for _, file := range fm {
				os.Remove(file.Name())
			}
		}
	}()

	// 통계별 파일 생성
	for id, m := range c.dataRank {
		for category, list := range m {
			if _, ok := fm[category]; !ok {
				tempFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_"+category+"_")
				if err != nil {
					return err
				}
				fm[category] = tempFile
			}

			for idx, item := range list {
				str := fmt.Sprintf("%s\t%d\t%v\t%d\t%d\n", c.mark, id, item.Key, item.Count, idx+1)
				fm[category].WriteString(str)
			}
		}
	}
	for _, file := range fm {
		file.Close()
	}

	// 통계 Bulk insert
	for category, file := range fm {
		file.Close()
		query := fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_%s", file.Name(), category)
		_, err := c.calculator.engine.DB.Exec(query)
		if err == nil {
			//num, _ := rs.RowsAffected()
			//log.Debugf("cal_type=%d, stats_type=1, category=%s, affected_rows=%d", c.calculator.calType, category, num)
		} else {
			log.Error(err)
			return err
		}
	}

	// 장비별 추이 통계
	tempFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_equip_")
	if err != nil {
		return err
	}
	if !c.calculator.engine.IsDebug() {
		defer os.Remove(tempFile.Name())
	}
	for orgId, m := range c.equipStats {
		for tag, m2 := range m {
			line := fmt.Sprintf("%s\t%d\t%s\t%d,%d,%d,%d\n", c.mark, orgId, tag, m2[objs.StartEvent], m2[objs.ShockEvent], m2[objs.SpeedingEvent], m2[objs.ProximityEvent])
			tempFile.WriteString(line)
		}
	}
	tempFile.Close()
	query := fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_equip_trend", tempFile.Name())
	_, err = c.calculator.engine.DB.Exec(query)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsEvent, "equip", num)
	} else {
		log.Error(err)
		return err
	}

	// 타임라인 통계
	tempTimelineFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_timeline_")
	if err != nil {
		return err
	}
	if !c.calculator.engine.IsDebug() {
		defer os.Remove(tempTimelineFile.Name())
	}
	for orgId, m := range c.timelineStats { //  map[int]map[string]map[int]int
		for groupId, m1 := range m {
			for time, m2 := range m1 {
				line := fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%d\t%d\t%d\n", c.mark, orgId, groupId, time, m2[objs.StartEvent], m2[objs.ShockEvent], m2[objs.SpeedingEvent], m2[objs.ProximityEvent])
				tempTimelineFile.WriteString(line)
			}
		}
	}
	tempTimelineFile.Close()
	query = fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_timeline", tempTimelineFile.Name())
	_, err = c.calculator.engine.DB.Exec(query)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsEvent, "timeline", num)
	} else {
		log.Error(err)
		return err
	}

	// Shock links
	tempShowLinksFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_shocklinks_")
	if err != nil {
		return err
	}
	if !c.calculator.engine.IsDebug() {
		defer os.Remove(tempShowLinksFile.Name())
	}
	for orgId, m := range c.shockLinksStats {
		for groupId, arr := range m {
			line := fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%d\t%d\t%d\n", c.mark, orgId, groupId, strings.Join(arr, ","))
			tempShowLinksFile.WriteString(line)
		}
	}
	tempShowLinksFile.Close()
	query = fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_shocklinks", tempShowLinksFile.Name())
	_, err = c.calculator.engine.DB.Exec(query)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsEvent, "timeline", num)
	} else {
		log.Error(err)
		return err
	}

	// Session count by group
	tempSessionByGroupFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_activated_group_")
	if err != nil {
		return err
	}
	if !c.calculator.engine.IsDebug() {
		defer os.Remove(tempSessionByGroupFile.Name())
	}
	for orgId, m := range c.sessionByGroupStats {
		for groupId, arr := range m {
			optime := getOperatingTime(c.optimeByGroupStats[orgId][groupId])
			line := fmt.Sprintf("%s\t%d\t%d\t%d\t%3.0f\n", c.mark, orgId, groupId, len(arr), optime)
			tempSessionByGroupFile.WriteString(line)
		}
	}
	tempSessionByGroupFile.Close()
	query = fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_activated_group", tempSessionByGroupFile.Name())
	_, err = c.calculator.engine.DB.Exec(query)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsEvent, "timeline", num)
	} else {
		log.Error(err)
		return err
	}

	// Session count by equip
	tempSessionByEquipCountFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_activated_equip_")
	tempSessionByEquipFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_operation_record_")
	if err != nil {
		return err
	}
	if !c.calculator.engine.IsDebug() {
		defer os.Remove(tempSessionByEquipCountFile.Name())
	}
	for orgId, m := range c.sessionByEquipStats {
		for equipId, arr := range m {
			optime := getOperatingTime(c.optimeByEquipStats[orgId][equipId])
			line := fmt.Sprintf("%s\t%d\t%s\t%d\t%3.0f\n", c.mark, orgId, equipId, len(arr), optime)
			tempSessionByEquipCountFile.WriteString(line)

			// 운행시간, 실사용시간(이동시간+작업시간) 계산
			for sessionId, time := range c.optimeByEquipStats[orgId][equipId] {
				line := fmt.Sprintf("%s\t%s\t%s\t%d\t%s\t%s\t%3.0f\t%3.0f\t%3.0f\n",
					c.mark,
					time[0].Format(ipasserver.DateDefault),
					time[1].Format(ipasserver.DateDefault),
					orgId,
					equipId,
					sessionId,
					time[1].Sub(time[0]).Seconds(),
					0.0,
					0.0,
				)
				tempSessionByEquipFile.WriteString(line)
			}
		}
	}

	// 장비별 운행 기록(건수)
	tempSessionByEquipCountFile.Close()
	query = fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_activated_equip", tempSessionByEquipCountFile.Name())
	_, err = c.calculator.engine.DB.Exec(query)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsEvent, "timeline", num)
	} else {
		log.Error(err)
		return err
	}

	tempSessionByEquipFile.Close()
	query = fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_operation_record", tempSessionByEquipFile.Name())
	_, err = c.calculator.engine.DB.Exec(query)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, ExtraStats, "operation_record", num)
	} else {
		log.Error(err)
		return err
	}
	//if !c.calculator.engine.IsDebug() {
	return nil
}

// ---------------------------------------------------------------------------------------------

type extraStatsCalculator struct {
	calculator *Calculator
	wg         *sync.WaitGroup
	dataMap    objs.DataMap
	dataRank   objs.DataRank
	tables     map[string]bool
	from       string
	to         string
	mark       string
}

func NewExtraStats(calculator *Calculator, from, to, mark string) *extraStatsCalculator {
	return &extraStatsCalculator{
		calculator: calculator,
		dataMap:    make(objs.DataMap),
		dataRank:   make(objs.DataRank),
		tables:     map[string]bool{},
		from:       from,
		to:         to,
		mark:       mark,
	}
}

func (c *extraStatsCalculator) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	start := time.Now()

	// 일별 자산추이 개수 기록
	query := `
			insert into stats_equip_count
			select ?, org_id, group_id, equip_type, count(*) count
			from ast_ipas
			where created <= ?
			group by org_id, group_id, equip_type
		`
	_, err := c.calculator.engine.DB.Exec(query, c.mark, c.mark)
	if err == nil {
		//num, _ := rs.RowsAffected()
		//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsStatus, "equip_count", num)
	} else {
		log.Error(err)
		return err
	}

	log.Debugf("cal_type=%d, stats_type=%s, exec_time=%3.1fs", c.calculator.calType, StatsDesc[ExtraStats], time.Since(start).Seconds())
	return nil
}

func getOperatingTime(m map[string][]time.Time) float64 {
	var sum float64
	for _, t := range m {
		sum += t[1].Sub(t[0]).Seconds()
	}
	return sum
}
