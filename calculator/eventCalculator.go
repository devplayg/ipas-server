package calculator

import (
	"fmt"
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
	calculator *Calculator
	wg         *sync.WaitGroup
	dataMap    objs.DataMap
	dataRank   objs.DataRank
	tables     map[string]bool
	from       string
	to         string
	mark       string
}

func NewEventStats(calculator *Calculator, from, to, mark string) *eventStatsCalculator {
	return &eventStatsCalculator{
		calculator: calculator,
		dataMap:    make(objs.DataMap),
		dataRank:   make(objs.DataRank),
		tables: map[string]bool{
			"eventtype":  true,
			"eventtype1": false,
			"eventtype2": false,
			"eventtype3": false,
			"eventtype4": false,
		},
		from: from,
		to:   to,
		mark: mark,
	}
}

func (c *eventStatsCalculator) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	start := time.Now()
	if err := c.produceStats(); err != nil {
		log.Error(err)
		return err
	}

	if err := c.insert(); err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("cal_type=%d, stats_type=%d, exec_time=%3.1f", c.calculator.calType, StatsEvent, time.Since(start).Seconds())
	return nil
}

func (c *eventStatsCalculator) produceStats() error {
	// 통계 구조체 초기화
	c.dataMap[RootId] = make(map[string]map[interface{}]int64)
	c.dataRank[RootId] = make(map[string]objs.ItemList)

	// 데이터 조회
	query := `
		select org_id, group_id, event_type, equip_id, targets
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
		e := objs.IpasEvent{}
		err := rows.Scan(&e.OrgId, &e.GroupId, &e.EventType, &e.EquipId, &e.Targets)
		if err != nil {
			log.Error(err)
			return err
		}
		// 이벤트 유형 통계
		c.addToStats(&e, "eventtype", e.EventType)
		
		// 이벤트 타입별 Src tag 통계
		c.addToStats(&e, "eventtype"+strconv.Itoa(e.EventType), e.EquipId) // eventtype1~4
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
		return err
	}

	for id, m := range c.dataMap {
		for category, data := range m {
			if isTotalStats, ok := c.tables[category]; ok {
				if isTotalStats {
					c.dataRank[id][category] = objs.DetermineRankings(data, 0)
				} else {
					c.dataRank[id][category] = objs.DetermineRankings(data, c.calculator.top)
				}
			}
		}
	}

	return nil
}

func (c *eventStatsCalculator) addToStats(r *objs.IpasEvent, category string, val interface{}) error {

	// 전체 통계
	if _, ok := c.dataMap[RootId][category]; !ok {
		c.dataMap[RootId][category] = make(map[interface{}]int64)
		c.dataRank[RootId][category] = nil
	}
	c.dataMap[RootId][category][val] += 1

	// 기관 통계
	if _, ok := c.dataMap[r.OrgId]; !ok {
		c.dataMap[r.OrgId] = make(map[string]map[interface{}]int64)
		c.dataRank[r.OrgId] = make(map[string]objs.ItemList)
	}
	if _, ok := c.dataMap[r.OrgId][category]; !ok {
		c.dataMap[r.OrgId][category] = make(map[interface{}]int64)
		c.dataRank[r.OrgId][category] = nil
	}
	c.dataMap[r.OrgId][category][val]++

	// 그룹 통계
	if r.GroupId > 0 {
		if _, ok := c.dataMap[r.GroupId]; !ok {
			c.dataMap[r.GroupId] = make(map[string]map[interface{}]int64)
			c.dataRank[r.GroupId] = make(map[string]objs.ItemList)
		}
		if _, ok := c.dataMap[r.GroupId][category]; !ok {
			c.dataMap[r.GroupId][category] = make(map[interface{}]int64)
			c.dataRank[r.GroupId][category] = nil
		}
		c.dataMap[r.GroupId][category][val]++
	}

	// 사용자 소속 권한의 전체 통계
	if arr, ok := c.calculator.memberAssets[r.OrgId]; ok {
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

	return nil
}

func (c *eventStatsCalculator) insert() error {
	fm := make(map[string]*os.File)
	defer func() {
		for _, file := range fm {
			file.Close()
			os.Remove(file.Name())
		}
	}()
	for id, m := range c.dataRank {
		for category, list := range m {
			if _, ok := fm[category]; !ok {
				tempFile, err := ioutil.TempFile("", category+"_")
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

	for category, file := range fm {
		file.Close()
		query := fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE stats_%s", file.Name(), category)
		rs, err := c.calculator.engine.DB.Exec(query)
		if err == nil {
			num, _ := rs.RowsAffected()
			log.Debugf("cal_type=%d, category=%s, affected_rows=%d", c.calculator.calType, category, num)
		} else {
			log.Debug(err)
			return err
		}
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
	log.Debugf("cal_type=%d, stats_type=%d, exec_time=%3.1f", c.calculator.calType, StatsStatus, time.Since(start).Seconds())
	return nil
}
