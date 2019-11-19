package impl

import (
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo"
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"gopkg.in/mgo.v2/bson"
)

type MemberImpl struct {
	record *entity.Member
	LeagueId string
}

func NewMemberImpl(member *entity.Member, orgid string) *MemberImpl {
	return &MemberImpl{member, orgid}
}

func (svc *MemberImpl) Insert() error {
	leaguetxnskey := mongokey.NewLeagueMemberPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()
	return col.Insert(svc.record)
}

func (svc *MemberImpl) Delete() error {
	leaguetxnskey := mongokey.NewLeagueMemberPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()
	return col.Remove(bson.M{"addr":svc.record.Addr})
}

func (svc *MemberImpl) GetMebers(page,size int) (int,interface{}, error) {

	leaguetxnskey := mongokey.NewLeagueMemberPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()

	if page <= 0 {
		page = 1
	}

	skip := size * (page-1)

	var htor []*entity.Member
	query := col.Find(bson.M{})

	total,err := query.Count()

	if err != nil {
		return 0,nil, err
	}

	err = query.Sort("addr").Skip(skip).Limit(size).All(&htor)
	if err != nil {
		return 0,nil, err
	}


	return total,htor, nil
}
