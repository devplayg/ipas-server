package calculator

import (
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
)

type Stats interface {
	Start(wg *sync.WaitGroup) error
}

func NewStats(calculator *Calculator, stats int, from, to, mark string) Stats {
	if stats == StatsEvent {
		return NewEventStats(calculator, from, to, mark)

	} else if stats == StatsStatus {
		return NewStatusStats(calculator, from, to, mark)

	} else {
		return nil
	}
}

// ---------------------------------------------------------------------------------------------

type eventStatsCalculator struct {
	calculator    *Calculator
	wg            *sync.WaitGroup
	dataMap       objs.DataMap
	dataRank      objs.DataRank
	equipStats    map[string]map[int]int
	timelineStats map[int]map[string]map[int]int
	tables        map[string]bool
	from          string
	to            string
	mark          string
	mutex         sync.Mutex
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
	if err := c.produceStats(); err != nil {
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
	log.Debugf("cal_type=%d, stats_type=%d, exec_time=%3.1fs", c.calculator.calType, StatsEvent, time.Since(start).Seconds())
	return nil
}

func (c *eventStatsCalculator) produceStats() error {

	// 통계 구조체 초기화
	c.dataMap[RootId] = make(map[string]map[interface{}]int64)
	c.dataRank[RootId] = make(map[string]objs.ItemList)
	c.equipStats = make(map[string]map[int]int)
	c.timelineStats = make(map[int]map[string]map[int]int)

	// 데이터 조회
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
	// 데이터 맵 생성
	for rows.Next() {

		// 이벤트 객체 생성
		e := objs.IpasEvent{}

		// 데이터 읽기
		err := rows.Scan(&e.OrgId, &e.GroupId, &e.EventType, &e.EquipId, &e.Targets, &e.Timeline)
		if err != nil {
			log.Error(err)
			return err
		}

		// 이벤트 유형 통계
		c.addToStats(&e, "evt", e.EventType)

		// 이벤트 타입별 Src tag 통계
		if e.EventType >= 0 && e.EventType <= 4 {
			evt := strconv.Itoa(e.EventType)
			c.addToStats(&e, "evt"+evt+"_by_equip", e.EquipId) // eventtype1~4
			c.addToStats(&e, "evt"+evt+"_by_group", fmt.Sprintf("%d/%d", e.OrgId, e.GroupId))
		}

		// 타임라인 통계
		c.addToTimelineStats(&e)

		// 장비(Tag) 이벤트 통계
		if _, ok := c.equipStats[e.EquipId]; !ok {
			c.equipStats[e.EquipId] = map[int]int{
				objs.StartEvent:     0,
				objs.ShockEvent:     0,
				objs.SpeedingEvent:  0,
				objs.ProximityEvent: 0,
			}
		}
		c.equipStats[e.EquipId][e.EventType]++
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
		return err
	}

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
func (c *eventStatsCalculator) addToTimelineStats(e *objs.IpasEvent) error {

	// 전체 통계
	if _, ok := c.timelineStats[RootId]; !ok {
		c.timelineStats[RootId] = make(map[string]map[int]int)
	}
	if _, ok := c.timelineStats[RootId][e.Timeline]; !ok {
		c.timelineStats[RootId][e.Timeline] = map[int]int{
			objs.StartEvent:     0,
			objs.ShockEvent:     0,
			objs.SpeedingEvent:  0,
			objs.ProximityEvent: 0,
		}
	}
	c.timelineStats[RootId][e.Timeline][e.EventType] += 1

	// 기관 통계
	if _, ok := c.timelineStats[e.OrgId]; !ok {
		c.timelineStats[e.OrgId] = make(map[string]map[int]int)
	}
	if _, ok := c.timelineStats[e.OrgId][e.Timeline]; !ok {
		c.timelineStats[e.OrgId][e.Timeline] = map[int]int{
			objs.StartEvent:     0,
			objs.ShockEvent:     0,
			objs.SpeedingEvent:  0,
			objs.ProximityEvent: 0,
		}
	}
	c.timelineStats[e.OrgId][e.Timeline][e.EventType]++

	if e.GroupId > 0 {
		// 그룹 통계
		if _, ok := c.timelineStats[e.GroupId]; !ok {
			c.timelineStats[e.GroupId] = make(map[string]map[int]int)
		}
		if _, ok := c.timelineStats[e.GroupId][e.Timeline]; !ok {
			c.timelineStats[e.GroupId][e.Timeline] = map[int]int{
				objs.StartEvent:     0,
				objs.ShockEvent:     0,
				objs.SpeedingEvent:  0,
				objs.ProximityEvent: 0,
			}
		}
		c.timelineStats[e.GroupId][e.Timeline][e.EventType]++
	}

	return nil
}

func (c *eventStatsCalculator) addToStats(e *objs.IpasEvent, category string, val interface{}) error {

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
		for _, file := range fm {
			file.Close()
			//os.Remove(file.Name())
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

	// Tag 통계
	tempFile, err := ioutil.TempFile(c.calculator.tmpDir, "stats_equip_")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	for tag, m := range c.equipStats {
		line := fmt.Sprintf("%s\t%s\t%d,%d,%d,%d\n", c.mark, tag, m[objs.StartEvent], m[objs.ShockEvent], m[objs.SpeedingEvent], m[objs.ProximityEvent])
		tempFile.WriteString(line)
	}
	tempFile.Close()
	query := fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_equip", tempFile.Name())
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
	defer os.Remove(tempTimelineFile.Name())
	for assetId, timeline := range c.timelineStats { //  map[int]map[string]map[int]int
		for time, m := range timeline {
			line := fmt.Sprintf("%s\t%d\t%s\t%d,%d,%d,%d\n", c.mark, assetId, time, m[objs.StartEvent], m[objs.ShockEvent], m[objs.SpeedingEvent], m[objs.ProximityEvent])
			tempTimelineFile.WriteString(line)
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
	return nil
}

// ---------------------------------------------------------------------------------------------

type statusStatsCalculator struct {
	calculator *Calculator
	wg         *sync.WaitGroup
	dataMap    objs.DataMap
	dataRank   objs.DataRank
	tables     map[string]bool
	from       string
	to         string
	mark       string
}

func NewStatusStats(calculator *Calculator, from, to, mark string) *statusStatsCalculator {
	return &statusStatsCalculator{
		calculator: calculator,
		dataMap:    make(objs.DataMap),
		dataRank:   make(objs.DataRank),
		tables:     map[string]bool{},
		from:       from,
		to:         to,
		mark:       mark,
	}
}

func (c *statusStatsCalculator) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	start := time.Now()
	t1, _ := time.Parse(ipasserver.DateDefault, c.from)

	if c.calculator.calType == objs.RealtimeCalculator || t1.Format("2006-01-02") == time.Now().Add(-24*time.Hour).Format("2006-01-02") { // 실시간 통계 또는 어제 통계이면
		query := `
			insert into stats_equip_count
			select ?, org_id, group_id, equip_type, count(*) count
			from ast_ipas
			group by org_id, group_id, equip_type
		`
		_, err := c.calculator.engine.DB.Exec(query, c.mark)
		if err == nil {
			//num, _ := rs.RowsAffected()
			//log.Debugf("cal_type=%d, stats_type=%d, category=%s, affected_rows=%d", c.calculator.calType, StatsStatus, "equip_count", num)
		} else {
			log.Error(err)
			return err
		}
	}

	log.Debugf("cal_type=%d, stats_type=%d, exec_time=%3.1f", c.calculator.calType, StatsStatus, time.Since(start).Seconds())
	return nil
}
