package mongokey

import (
	"fmt"
)

const (
	COL_LEAGUE_HEIGHT MongoKey = MongoKey("col_league_height")
)

type MongoKey string

func NewVoteDetailPrefixKey(orgid string) MongoKey {
	key := fmt.Sprintf("col_vote_%s",orgid)
	return MongoKey(key)
}

func NewLeagueTxnsPrefixKey(orgid string) MongoKey {
	key := fmt.Sprintf("col_league_txns_%s",orgid)
	return MongoKey(key)
}

func NewLeagueMemberPrefixKey(orgid string) MongoKey {
	key := fmt.Sprintf("col_league_members_%s",orgid)
	return MongoKey(key)
}

func (k MongoKey) Value() string {
	return string(k)
}


func NewVoteSatisticsKey(orgid,voteid string) MongoKey {
	key := fmt.Sprintf("col_vtststks_%s_%s",orgid,voteid)
	return MongoKey(key)
}