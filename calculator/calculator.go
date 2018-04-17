package calculator

import (
	"fmt"
	"github.com/devplayg/ipas-server"
	//"github.com/devplayg/ipas-server/objs"
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
	//dataMap         objs.DataMap
	//_rank           objs.DataRank
	memberAssets    map[int][]int
	mutex           *sync.RWMutex
	eventStatsKeys  []string
	statusStatsKeys []string
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

func (c *Calculator) removeStats(date string) error {
	log.Debugf("Removing old stats for %s", date)
	query := "delete from stats_%s where date >= ? and date <= ?"
	for _, k := range append(c.eventStatsKeys, c.statusStatsKeys...) {
		_, err := c.engine.DB.Query(fmt.Sprintf(query, k), date+" 00:00:00", date+" 23:59:59")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Calculator) createTables() error {
	query := `
		CREATE TABLE IF NOT EXISTS stats_%s (
			date datetime NOT NULL,
			group_id int(11) NOT NULL,
			%s varchar(64) NOT NULL,
			count int(10) unsigned NOT NULL,
			rank int(10) unsigned NOT NULL,
			KEY ix_stats_%s_date (date),
			KEY ix_stats_%s_groupid (date,group_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
	for _, k := range append(c.eventStatsKeys, c.statusStatsKeys...) {
		_, err := c.engine.DB.Query(fmt.Sprintf(query, k, k, k, k))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Calculator) Start() error {
	c.eventStatsKeys = strings.Split("eventtype,srctag,dsttag", ",")
	//c.statusStatsKeys = nil

	if err := c.createTables(); err != nil {
		log.Fatal(err)
	}

	if len(c.date) > 0 { // 지정한 날짜에 대한 통계
		t, err := time.Parse("2006-01-02", c.date)
		if err != nil {
			return err
		}
		log.Debugf("Calculating statistics for %s", c.date)

		// 기존 통계 삭제
		if err := c.removeStats(t.Format("2006-01-02")); err != nil {
			return err
		}

		// 통계 산출
		if err := c.calculate(
			t.Format("2006-01-02")+" 00:00:00",
			t.Format("2006-01-02")+" 23:59:59",
			t.Format("2006-01-02")+" 00:00:00",
		); err != nil {
			return err
		}
	} else if len(c.report) > 0 { // 특정 기간에 대한 통계 (추후 개발)
		//timeArr := strings.Split(c.report, ",")
		//from, err := time.Parse("2006-01-02", timeArr[0])
		//if err != nil {
		//	return err
		//}
		//to, err := time.Parse("2006-01-02", timeArr[1])
		//if err != nil {
		//	return err
		//}
		//mark, err := time.Parse(ipasserver.DateDefault, timeArr[2])
		//if err != nil {
		//	return err
		//}
		//log.Debugf("Calculating statistics for %s", c.report)
		//c.calculate(
		//	from.Format("2006-01-02")+" 00:00:00",
		//	to.Format("2006-01-02")+" 00:00:00",
		//	mark.Format(ipasserver.DateDefault),
		//)

	} else { // 실시간 통계(당일)
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
	}

	return nil
}

func (c *Calculator) calculate(from, to, mark string) error {
	var err error

	// 관리자를 제외한 사용자 자산 조회
	c.memberAssets, err = c.getMemberAssets()
	if err != nil {
		log.Error(err)
	}

	// 데이터 초기화
	//c.dataMap = make(objs.DataMap) // 데이터 맵
	//c._rank = make(objs.DataRank)  // 순위
	//c.dataMap[RootId] = make(map[string]map[interface{}]int64)
	//c._rank[RootId] = make(map[string]objs.ItemList)

	start := time.Now()
	log.Debugf("Calculating statistics(%s ~ %s) will be marked as %s", from, to, mark)
	//wg := new(sync.WaitGroup)
	//
	//// 이벤트 통계 산출
	//wg.Add(1)
	//go c.calculateEvents(wg, from, to)
	//
	//// 상태정보 통계 산출
	//wg.Add(1)
	//go c.calculateStatus(wg, from, to)
	//
	//// 위 두 개의 통계 작업이 완료될 때까지 대기
	//wg.Wait()

	if len(mark) > 0 {
		//
	}

	// Write signature
	log.Debugf("Done. Execution time: %3.1f", time.Since(start).Seconds())
	return nil
}
//
//func (c *Calculator) calculateEvents(wg *sync.WaitGroup, from, to string) error {
//	defer wg.Done()
//	log.Debug("Calculating event..")
//
//	// 통계 구조체 초기화
//	dataMap := make(objs.DataMap)
//	rank := make(objs.DataRank)
//	dataMap[RootId] = make(map[string]map[interface{}]int64)
//	rank[RootId] = make(map[string]objs.ItemList)
//
//	// 데이터 조회
//	query := `
//		select org_id, group_id, event_type, equip_id, targets
//		from log_ipas_event
//		where date between ? and ?
//		limit 100
//	`
//	rows, err := c.engine.DB.Query(query, from, to)
//	if err != nil {
//		log.Error(err)
//	}
//	defer rows.Close()
//
//	// 데이터 맵 생성
//	for rows.Next() {
//		e := IpasEvent{}
//		err := rows.Scan(&e.orgId, &e.groupId, &e.eventType, &e.equipId, &e.targets)
//		if err != nil {
//			log.Error(err)
//		}
//		c.addToStats(&e, "eventtype", e.eventType)
//		c.addToStats(&e, "srctag", e.equipId)
//		c.addToStats(&e, "dsttag", e.targets)
//	}
//	err = rows.Err()
//	if err != nil {
//		log.Error(err)
//	}
//
//	var rankAll = map[string]bool{
//		"eventtype": true,
//	}
//	for id, m := range c.dataMap {
//		for category, data := range m {
//			if _, ok := rankAll[category]; ok {
//				c._rank[id][category] = objs.DetermineRankings(data, 0)
//			} else {
//				c._rank[id][category] = objs.DetermineRankings(data, c.top)
//			}
//		}
//	}
//
//	return nil
//}
//
//func (c *Calculator) calculateStatus(wg *sync.WaitGroup, from, to string) error {
//	defer wg.Done()
//	log.Debug("Calculating status..")
//	//time.Sleep(2 * time.Second)
//	return nil
//}
//
//func (c *Calculator) addToStats(r *IpasEvent, category string, val interface{}) error {
//
//	// 전체 통계
//	if _, ok := c.dataMap[RootId][category]; !ok {
//		c.dataMap[RootId][category] = make(map[interface{}]int64)
//		c._rank[RootId][category] = nil
//	}
//	c.dataMap[RootId][category][val] += 1
//
//	// 기관 통계
//	if _, ok := c.dataMap[r.orgId]; !ok {
//		c.dataMap[r.orgId] = make(map[string]map[interface{}]int64)
//		c._rank[r.orgId] = make(map[string]objs.ItemList)
//	}
//	if _, ok := c.dataMap[r.orgId][category]; !ok {
//		c.dataMap[r.orgId][category] = make(map[interface{}]int64)
//		c._rank[r.orgId][category] = nil
//	}
//	c.dataMap[r.orgId][category][val] += 1
//
//	// 그룹 통계
//	if r.groupId > 0 {
//		if _, ok := c.dataMap[r.groupId]; !ok {
//			c.dataMap[r.groupId] = make(map[string]map[interface{}]int64)
//			c._rank[r.groupId] = make(map[string]objs.ItemList)
//		}
//		if _, ok := c.dataMap[r.groupId][category]; !ok {
//			c.dataMap[r.groupId][category] = make(map[interface{}]int64)
//			c._rank[r.groupId][category] = nil
//		}
//		c.dataMap[r.groupId][category][val] += 1
//	}
//
//	// 사용자 소속 권한의 전체 통계
//	if arr, ok := c.memberAssets[r.orgId]; ok {
//		for _, memberId := range arr {
//			id := memberId * -1
//
//			if _, ok := c.dataMap[id]; !ok {
//				c.dataMap[id] = make(map[string]map[interface{}]int64)
//				c._rank[id] = make(map[string]objs.ItemList)
//			}
//			if _, ok := c.dataMap[id][category]; !ok {
//				c.dataMap[id][category] = make(map[interface{}]int64)
//				c._rank[id][category] = nil
//			}
//			c.dataMap[id][category][val] += 1
//		}
//	}
//
//	return nil
//}

// 사용자 자산 조회
func (c *Calculator) getMemberAssets() (map[int][]int, error) {
	m := make(map[int][]int)
	var (
		memberId int
		assetId  int
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

func (c *Calculator) cleanStats(date string) error {
	//query = "delete from "
	return nil
	//rows, err := c.engine.DB.Query(query)
	//if err != nil {
	//	log.Error(err)
	//}
	//defer rows.Close()
	//for rows.Next() {
	//	err := rows.Scan(&memberId, &assetId)
	//	if err != nil {
	//		log.Error(err)
	//	}
	//
	//	if _, ok := m[assetId]; !ok {
	//		m[assetId] = make([]int, 0)
	//	}
	//	m[assetId] = append(m[assetId], memberId)
	//}
	//err = rows.Err()
	//if err != nil {
	//	log.Error(err)
	//}
	//
	//return m, nil
}
