package impl

import (
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo"
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"gopkg.in/mgo.v2/bson"
)

type VoteDetailImpl struct {
	detail *entity.VoteDetail
	leagueId string
}

func NewVoteDetailImpl(heightor *entity.VoteDetail,orgid string) *VoteDetailImpl  {
	return &VoteDetailImpl{heightor,orgid}
}

func (svc *VoteDetailImpl) Insert() error {

	orgkey := mongokey.NewVoteDetailPrefixKey(svc.leagueId)

	session,col := mongo.KeyProvider(orgkey)
	defer session.Close()

	if svc.detail.Id == "" {
		svc.detail.Id = svc.leagueId
	}

	return col.Insert(svc.detail)
}

func (svc *VoteDetailImpl) UpdateVoteDetailById(newHeight uint64) error {
	orgkey := mongokey.NewVoteDetailPrefixKey(svc.leagueId)

	session,col := mongo.KeyProvider(orgkey)
	defer session.Close()

	return col.UpdateId(svc.leagueId,svc.detail)
}

func (svc *VoteDetailImpl) GetLeagueVoteDetail(orgid,voteid string)(*entity.VoteDetail,error) {
	orgkey := mongokey.NewVoteDetailPrefixKey(svc.leagueId)

	session,col := mongo.KeyProvider(orgkey)
	defer session.Close()

	htor := new(entity.VoteDetail)
	err := col.FindId(svc.leagueId).One(htor)
	if err != nil {
		return nil,err
	}
	return htor,nil
}


func (svc *VoteDetailImpl) GetVoteList(page,size int,height uint64,vstate uint8) (int,interface{}, error) {
	orgkey := mongokey.NewVoteDetailPrefixKey(svc.leagueId)
	session, col := mongo.KeyProvider(orgkey)
	defer session.Close()

	if page <= 0 {
		page = 1
	}

	skip := size * (page-1)

	bsm := bson.M{}

	if vstate == 1 {
		bsm = bson.M{
			"endheight":bson.M{"$gt":vstate},
		}
	}

	if vstate == 2 {
		bsm = bson.M{
			"endheight":bson.M{"$lte":vstate},
		}
	}

	var htor []*entity.VoteDetail
	query := col.Find(bsm)

	total,err := query.Count()

	if err != nil {
		return 0,nil, err
	}

	err = query.Sort("-endheight").Skip(skip).Limit(size).All(&htor)
	if err != nil {
		return 0,nil, err
	}


	return total,htor, nil
}