package calculator

import (
	"github.com/devplayg/ipas-server"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type IpasEvent struct {
	orgId     int
	groupId   int
	eventType int
	equipId   string
	targets   string
}

const (
	RootId = -1
)

type Calculator struct {
	engine   *ipasserver.Engine
	top      int
	interval int64
	date     string
	report   string
	assetMap map[int][]int
}

func NewCalculator(engine *ipasserver.Engine, top int, interval int64, date, report string) *Calculator {
	return &Calculator{
		engine:   engine,
		top:      top,
		interval: interval,
		date:     date,
		report:   report,
	}
}

func (c *Calculator) Start() error {
	if len(c.date) > 0 { // 특정 지정한 날짜에 대한 통계 생성
		t, err := time.Parse("2006-01-02", c.date)
		if err != nil {
			return err
		}
		log.Debugf("Calculating statistics for %s", c.date)
		err = c.calculate(
			t.Format("2006-01-02")+" 00:00:00",
			t.Format("2006-01-02")+" 23:59:59",
			t.Format("2006-01-02")+" 23:59:59",
		)
		return err

	} else if len(c.report) > 0 { // 시스템에서 생성하는 보고서 생성 시
		timeArr := strings.Split(c.report, ",")
		from, err := time.Parse("2006-01-02", timeArr[0])
		if err != nil {
			return err
		}
		to, err := time.Parse("2006-01-02", timeArr[1])
		if err != nil {
			return err
		}
		mark, err := time.Parse(ipasserver.DateDefault, timeArr[2])
		if err != nil {
			return err
		}
		log.Debugf("Calculating statistics for %s", c.report)
		c.calculate(
			from.Format("2006-01-02")+" 00:00:00",
			to.Format("2006-01-02")+" 00:00:00",
			mark.Format(ipasserver.DateDefault),
		)

		return nil
	}

	go func() {
		for {
			t := time.Now()

			err := c.calculate(
				t.Format("2006-01-02")+" 00:00:00",
				t.Format("2006-01-02")+" 23:59:59",
				t.Format(ipasserver.DateDefault),
			)
			if err != nil {
				log.Error(err)
			}
			time.Sleep(time.Duration(c.interval) * time.Millisecond)
		}
	}()

	return nil
}

func (c *Calculator) calculate(from, to, mark string) error {
	var err error
	c.assetMap, err = c.getMemberAssets()
	if err != nil {
		log.Error(err)
	}

	start := time.Now()
	log.Debugf("Calculating statistics, [%s ~ %s] Mark as %s", from, to, mark)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go c.calculateEvents(wg, from, to, mark)
	wg.Add(1)
	go c.calculateStatus(wg, from, to, mark)
	wg.Wait()

	// Write signature
	log.Debugf("Done. Execution time: %3.1f", time.Since(start).Seconds())
	return nil
}

func (c *Calculator) calculateEvents(wg *sync.WaitGroup, from, to, mark string) error {
	defer wg.Done()
	log.Debug("Calculating event..")

	//var (
	//	orgId     int
	//	groupId   int
	//	eventType int
	//	equipId   string
	//	targets   string
	//)

	query := `
		select  org_id, group_id, event_type, equip_id, targets
		from log_ipas_event
		where date between ? and ?
		limit 10
	`
	//logs := make([]IpasEvent, 0)
	rows, err := c.engine.DB.Query(query, from, to)
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		e := IpasEvent{}
		err := rows.Scan(&e.orgId, &e.groupId, &e.eventType, &e.equipId, &e.targets)
		if err != nil {
			log.Error(err)
		}

		//logs = append(logs, e)
		//log.Debugf("%s\t%s", equipId, targets)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
	}

	return nil
}

func (c *Calculator) calculateStatus(wg *sync.WaitGroup, from, to, mark string) error {
	defer wg.Done()
	log.Debug("Calculating status..")
	//time.Sleep(2 * time.Second)
	return nil
}

func (c *Calculator) addToStats(r *IpasEvent, category string, val interface{}) error {

//	// By sensor
//	if r.SensorId > 0 {
//		if _, ok := f.dataMap[r.SensorId]; !ok {
//			f.dataMap[r.SensorId] = make(map[string]map[interface{}]int64)
//			f._rank[r.SensorId] = make(map[string]stats.ItemList)
//		}
//		if _, ok := f.dataMap[r.SensorId][category]; !ok {
//			f.dataMap[r.SensorId][category] = make(map[interface{}]int64)
//			f._rank[r.SensorId][category] = nil
//		}
//		f.dataMap[r.SensorId][category][val] += 1
//	}
//
//	// By group
//	if r.IppoolSrcGcode > 0 {
//		if _, ok := f.dataMap[r.IppoolSrcGcode]; !ok {
//			f.dataMap[r.IppoolSrcGcode] = make(map[string]map[interface{}]int64)
//			f._rank[r.IppoolSrcGcode] = make(map[string]stats.ItemList)
//		}
//		if _, ok := f.dataMap[r.IppoolSrcGcode][category]; !ok {
//			f.dataMap[r.IppoolSrcGcode][category] = make(map[interface{}]int64)
//			f._rank[r.IppoolSrcGcode][category] = nil
//		}
//		f.dataMap[r.IppoolSrcGcode][category][val] += 1
//	}
//
//	// To all
//	if _, ok := f.dataMap[RootId][category]; !ok {
//		f.dataMap[RootId][category] = make(map[interface{}]int64)
//		f._rank[RootId][category] = nil
//	}
//	f.dataMap[RootId][category][val] += 1
//
//	// By member
//	if arr, ok := f.memberAssets[r.IppoolSrcGcode]; ok {
//		for _, memberId := range arr {
//			id := memberId * -1
//
//			if _, ok := f.dataMap[id]; !ok {
//				f.dataMap[id] = make(map[string]map[interface{}]int64)
//				f._rank[id] = make(map[string]stats.ItemList)
//			}
//			if _, ok := f.dataMap[id][category]; !ok {
//				f.dataMap[id][category] = make(map[interface{}]int64)
//				f._rank[id][category] = nil
//			}
//			f.dataMap[id][category][val] += 1
//		}
//	}
//
	return nil
}


func (c *Calculator) getMemberAssets() (map[int][]int, error) {
	m := make(map[int][]int)
	var (
		memberId int
		assetId int
	)
	query := `
		select member_id, asset_id
		from mbr_asset
		where member_id in (select member_id from mbr_member where position < 256)
	`
	rows, err := c.engine.DB.Query(query)
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&memberId, &assetId)
		if err != nil {
			log.Error(err)
		}

		if _, ok := m[assetId]; !ok {
			m[assetId] = make([]int, 0)
		}
		m[assetId] = append(m[assetId], memberId)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
	}

	return m, nil
}