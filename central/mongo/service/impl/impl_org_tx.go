package impl

import (
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo"
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"gopkg.in/mgo.v2/bson"
)

type OrgTxnsImpl struct {
	Txs      []interface{}
	LeagueId string
}

func NewOrgTxnsImpl(txs []interface{}, orgid string) *OrgTxnsImpl {
	return &OrgTxnsImpl{txs, orgid}
}

func (svc *OrgTxnsImpl) Insert() error {
	leaguetxnskey := mongokey.NewLeagueTxnsPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()
	return col.Insert(svc.Txs...)
}

func (svc *OrgTxnsImpl) GetTransferTxList(accountaddr string,page,size int) (int,interface{}, error) {
	leaguetxnskey := mongokey.NewLeagueTxnsPrefixKey(svc.LeagueId)
	session, col := mongo.KeyProvider(leaguetxnskey)
	defer session.Close()

	var htor []*entity.OrgTxDto

	querym := bson.M{
		"$or":[]bson.M{
			{"txdata.from":accountaddr},
			{"txdata.froms.tis":bson.M{"address":accountaddr}},
			{"txdata.tos.tos":bson.M{"address":accountaddr}},
		},
	}

	pg := page
	if page <= 0 {
		pg =1
	}

	skip := (pg -1 )*size

	qury := col.Find(querym)

	total,err := qury.Count()
	if err != nil {
		return 0,nil,err
	}

	err = qury.Sort("-blocktime").Skip(skip).All(&htor)

	if err != nil {
		if err != nil {
			return 0,nil,err
		}
	}
	return total,htor, nil
}
