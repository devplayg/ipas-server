package objs

import (
	"sort"
)

const (
	RealtimeCalculator     = 1
	SpecificDateCalculator = 2
	DateRangeCalculator    = 3
)

type Item struct {
	Key   interface{}
	Count int64
}
type ItemList []Item

type DataMap map[int]map[string]map[interface{}]int64 // Code / Category / Key / Count
type DataRank map[int]map[string]ItemList             // Code / Category / Key / Ranking

func (p ItemList) Len() int { return len(p) }
func (p ItemList) Less(i, j int) bool {
	//return p[i].Count < p[j].Count  // && p[i].Key < p[j].Key
	switch p[i].Key.(type) {
	case int:
		if p[i].Count < p[j].Count {
			return true
		} else if p[i].Count == p[j].Count {
			return p[i].Key.(int) < p[j].Key.(int)
		} else {
			return false
		}
	case string:
		if p[i].Count < p[j].Count {
			return true
		} else if p[i].Count == p[j].Count {
			return p[i].Key.(string) > p[j].Key.(string)
		} else {
			return false
		}
	default:
		return p[i].Count < p[j].Count
	}
}

func (p ItemList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func DetermineRankings(m map[interface{}]int64, top int) ItemList {
	list := make(ItemList, len(m))
	i := 0
	for k, v := range m {
		list[i] = Item{k, v}
		i++
	}
	sort.Sort(sort.Reverse(list))
	if top > 0 && len(list) > top {
		return list[0:top]
	} else {
		return list
	}
}
