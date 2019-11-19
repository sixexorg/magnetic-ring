package impl

import (
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo"
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"gopkg.in/mgo.v2/bson"
)

type VoteStatisticsImpl struct {
	vu *entity.VoteUnit
	leagueId string
}

func NewVoteStatisticsImpl(vu *entity.VoteUnit,leagueId string) *VoteStatisticsImpl  {
	return &VoteStatisticsImpl{vu:vu,leagueId:leagueId}
}

func (svc *VoteStatisticsImpl) Add() error {
	leaguetxnskey := mongokey.NewVoteSatisticsKey(svc.leagueId,svc.vu.VoteId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()

	return col.Insert(svc.vu)
}

func (svc *VoteStatisticsImpl) IsAccountVoted(account string) bool {
	leaguetxnskey := mongokey.NewVoteSatisticsKey(svc.leagueId,svc.vu.VoteId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()

	result := new(entity.VoteUnit)
	err := col.Find(bson.M{"account":account}).One(result)
	if err != nil {
		return false
	}

	return true
}


func (svc *VoteStatisticsImpl) FindVoteUnitList(leagueId,voteId string,page,pgsize int)(int,interface{},error) {
	leaguetxnskey := mongokey.NewVoteSatisticsKey(leagueId,voteId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()

	bsonm := bson.M{
		"voteid":voteId,
	}

	query := col.Find(bsonm)

	total,err := query.Count()
	if err != nil {
		return 0,nil,err
	}

	if page <= 0 {
		page = 1
	}

	if pgsize < 10{
		pgsize=10
	}

	skip := pgsize * (page-1)
	var result []entity.VoteUnit

	err = query.Sort("-endblockheight").Skip(skip).Limit(pgsize).All(&result)
	if err != nil {
		return 0,nil,err
	}

	return total,result,nil
}