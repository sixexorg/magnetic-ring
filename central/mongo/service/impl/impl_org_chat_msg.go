package impl

import (
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo"
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"gopkg.in/mgo.v2/bson"
)

type OrgChatMsgImpl struct {
	ChatMsgs      []interface{}
	LeagueId string
}

func NewOrgChatMsgImpl(txs []interface{}, orgid string) *OrgChatMsgImpl {
	return &OrgChatMsgImpl{txs, orgid}
}

func (svc *OrgChatMsgImpl) Insert() error {
	leaguetxnskey := mongokey.NewLeagueTxnsPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()
	return col.Insert(svc.ChatMsgs...)
}

func (svc *OrgChatMsgImpl) GetChatList(page,size int) (int,interface{}, error) {
	leaguetxnskey := mongokey.NewLeagueTxnsPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()

	if page <= 0 {
		page = 1
	}

	skip := size * (page-1)

	var htor []*entity.OrgChatDto
	query := col.Find(bson.M{})

	total,err := query.Count()

	if err != nil {
		return 0,nil, err
	}

	err = query.Sort("-blockTime").Skip(skip).Limit(size).All(&htor)
	if err != nil {
		return 0,nil, err
	}


	return total,htor, nil
}
