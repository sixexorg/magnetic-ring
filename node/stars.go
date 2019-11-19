package node

import (
	"sort"
	"sync"
)

type starsMgr struct {
	stars []string
	earth string
	m     sync.RWMutex
}

func NewStarsMgr() *starsMgr {
	return &starsMgr{}
}

func (s *starsMgr) curStars() []string {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.stars
	/*return []string{
		"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",

		//"045d54e7cebc80e03c52e53a9a3b7601041183514230ae5776996a131942f8f25425144adc67bbe266200fc7357781a9cc579b3814b37b4f46b3de67cf12177da8",
		//"04a3a9d49d883984c61e3f902f73c4fd0b4067bef3199b3f248a1973d02879e09a89b1b869ea03d40c1816a01fcc2d0a95288c98cfd4386883d9df3015e3f2d72d",
	}*/
}
func (s *starsMgr) pushStars(pubStrs []string) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, v := range pubStrs {
		isFound := false
		for _, vi := range s.stars {
			if vi == v {
				isFound = true
				break
			}
		}
		if isFound{
			continue
		}
		s.stars = append(s.stars, v)
	}
}
func (s *starsMgr) popStars(pubStrs []string) {
	s.m.Lock()
	defer s.m.Unlock()
	//linq.From(s.stars).Except(linq.From(pubStrs)).ToSlice(&s.stars)
	idx := make([]int, 0, len(pubStrs))
	for _, v := range pubStrs {
		for k, vi := range s.stars {
			if vi == v {
				idx = append(idx, k)
				break
			}
		}
	}
	sort.Ints(idx)
	l := len(idx)
	dynamic := len(s.stars) - 1
	for i := 0; i < l; i++ {
		tmp := idx[i] - i
		if tmp == 0 {
			s.stars = s.stars[1:]
		} else if tmp == dynamic {
			s.stars = s.stars[:tmp]
		} else {
			s.stars = append(s.stars[0:idx[i]-1], s.stars[idx[i]:]...)
		}
		dynamic--
	}
}

var st *starsMgr = NewStarsMgr()
