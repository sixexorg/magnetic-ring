package impl

import (
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo"
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"github.com/sixexorg/magnetic-ring/common"
	"gopkg.in/mgo.v2/bson"
)

type LeagueHeightImpl struct {
	Heightor *entity.LeagueHeightor
	LeagueId common.Address
}

func NewLeagueHeightImpl(heightor *entity.LeagueHeightor,orgid common.Address) *LeagueHeightImpl  {
	return &LeagueHeightImpl{heightor,orgid}
}

func (svc *LeagueHeightImpl) Insert() error {
	session,col := mongo.KeyProvider(mongokey.COL_LEAGUE_HEIGHT)
	defer session.Close()

	if svc.Heightor.Id == "" {
		svc.Heightor.Id = svc.LeagueId.ToString()
	}

	return col.Insert(svc.Heightor)
}

func (svc *LeagueHeightImpl) UpdateById(address common.Address,newHeight uint64) error {
	session,col := mongo.KeyProvider(mongokey.COL_LEAGUE_HEIGHT)
	defer session.Close()
	//svc.Heightor.Height=newHeight

	return col.Update(bson.M{"id":address.ToString()},bson.M{"$set":bson.M{"height":newHeight}})
}

func (svc *LeagueHeightImpl) GetHeightor()(*entity.LeagueHeightor,error) {
	session,col := mongo.KeyProvider(mongokey.COL_LEAGUE_HEIGHT)
	defer session.Close()

	htor := new(entity.LeagueHeightor)
	err := col.Find(bson.M{"id":svc.LeagueId.ToString()}).One(htor)
	if err != nil {
		return nil,err
	}
	return htor,nil
}