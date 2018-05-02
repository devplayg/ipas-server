package calculator

import (
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/objs"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
	"time"
)

const (
	RootId = -1

	StatsEvent  = 1
	StatsStatus = 2
)

type Calculator struct {
	engine          *ipasserver.Engine // 엔진
	top             int                // Top N 순위
	interval        time.Duration      // 실행 주기(실시간 모드에서 사용)
	calType         int                // 산출기 타입(실시간, 특정날짜, 특정기간)
	targetDate      string             // 대상 날짜
	memberAssets    map[int][]int
	eventTableKeys  []string
	statusTableKeys []string
	tmpDir          string
}

func NewCalculator(engine *ipasserver.Engine, top int, interval time.Duration, calType int, targetDate string) *Calculator {
	return &Calculator{
		engine:          engine,
		top:             top,
		interval:        interval,
		calType:         calType,
		targetDate:      targetDate,
		tmpDir:          filepath.Join(engine.ProcessDir, "tmp"),
		eventTableKeys:  []string{"eventtype", "eventtype1", "eventtype2", "eventtype3", "eventtype4"},
		statusTableKeys: []string{},
	}
}

func (c *Calculator) removeStats(date string, isToday bool) error {
	query := "delete from stats_%s where date >= ? and date <= ?"
	if isToday {
		query += " and date <> (select value_s from sys_config where section = 'stats' and keyword = 'last_updated')"
	}
	from := date + " 00:00:00"
	to := date + " 23:59:59"
	for _, k := range append(c.eventTableKeys, c.statusTableKeys...) {
		_, err := c.engine.DB.Exec(fmt.Sprintf(query, k), from, to)
		if err != nil {
			log.Error(err)
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
		) ENGINE=InnoDB DEFAULT CHARSET=utf8
	`
	for _, k := range append(c.eventTableKeys, c.statusTableKeys...) {
		_, err := c.engine.DB.Query(fmt.Sprintf(query, k, k, k, k))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Calculator) Start() error {
	if err := c.createTables(); err != nil {
		log.Fatal(err)
	}

	log.Debugf("cal_type=%d, inverval=%d(ms)", c.calType, c.interval)
	if c.calType == objs.SpecificDateCalculator {
		t, err := time.Parse("2006-01-02", c.targetDate)
		if err != nil {
			return err
		}

		// 기존 통계 삭제
		if err := c.removeStats(t.Format("2006-01-02"), false); err != nil {
			log.Error(err)
			return err
		}

		// 통계 산출
		if err := c.calculate(
			t.Format("2006-01-02")+" 00:00:00",
			t.Format("2006-01-02")+" 23:59:59",
			t.Format("2006-01-02")+" 00:00:00",
		); err != nil {
			log.Error(err)
			return err
		}
	} else if c.calType == objs.DateRangeCalculator {
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

	} else if c.calType == objs.RealtimeCalculator { // 실시간 통계(당일)
		go func() {
			for {
				t := time.Now()

				// 통계산출
				if err := c.calculate(
					t.Format("2006-01-02")+" 00:00:00",
					t.Format("2006-01-02")+" 23:59:59",
					t.Format(ipasserver.DateDefault),
				); err == nil {
					// 최종 통계산출 시간 업데이트
					if err := c.engine.UpdateConfig("stats", "last_updated", t.Format(ipasserver.DateDefault), 0); err == nil {
						// 직전에 산출한 통계 삭제
						if err := c.removeStats(t.Format("2006-01-02"), true); err != nil {
							log.Error(err)
						}
					} else {
						log.Error(err)
					}
				} else {
					log.Error(err)
				}
				time.Sleep(c.interval)
			}
		}()
	}

	return nil
}


func (c *Calculator) calculate(from, to, mark string) error {
	var err error

	// 사용자 자산 조회
	c.memberAssets, err = c.getMemberAssets()
	if err != nil {
		log.Error(err)
	}

	start := time.Now()
	log.Debugf("cal_type=%d, stats_from=%s, stats_to=%s, stats_mark=%s", c.calType, from, to, mark)
	wg := new(sync.WaitGroup)

	// 이벤트 통계
	s1 := NewStats(c, StatsEvent, from, to, mark)
	wg.Add(1)
	go s1.Start(wg)

	// 상태 통계
	s2 := NewStats(c, StatsStatus, from, to, mark)
	wg.Add(1)
	go s2.Start(wg)

	// 통계산출 완료까지 대기
	wg.Wait()
	log.Debugf("cal_type=%d, total_exec_time=%3.1f", c.calType, time.Since(start).Seconds())
	return nil
}

// 사용자 자산 조회
func (c *Calculator) getMemberAssets() (map[int][]int, error) {
	m := make(map[int][]int)
	var (
		memberId int
		assetId  int
	)

	// Administrator는 모든 자산 데이터에 대한 접근 권한을 가짐
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
