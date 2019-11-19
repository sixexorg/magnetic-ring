package mongo

import (
	"github.com/sixexorg/magnetic-ring/central/mongokey"
	"github.com/sixexorg/magnetic-ring/config"
	"log"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
)

const (
	col_pay     string = "vote"
	col_tx     string = "tx"
)

var (
	db_blockChain string = "mgoSession"
	mgoSession    *mgo.Session
	keys          []string = []string{"Mongo"}
	dbinit = false
)


func InitMongo() {
	if dbinit {
		return
	}

	err := loadMongoSession()
	if err != nil {
		panic(err)
	}
	dbinit = true
	log.Println("database mongodb init sucess!!!!!!!")
}
func loadMongoSession() error {
	mongoDBInfo := &mgo.DialInfo{
		Addrs:     strings.Split(config.GlobalConfig.MongoCfg.Addr, ","),
		Timeout:   config.GlobalConfig.MongoCfg.Timeout * time.Second,
		PoolLimit: config.GlobalConfig.MongoCfg.PoolLimit,
		Database:  config.GlobalConfig.MongoCfg.Database,
	}
	session, err := mgo.DialWithInfo(mongoDBInfo)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	mgoSession = session
	db_blockChain = mongoDBInfo.Database
	return nil
}

func VoteProvider() (*mgo.Session, *mgo.Collection) {
	session := mgoSession.Clone()
	col := session.DB(db_blockChain).C(col_pay)
	return session, col
}


func TxProvider() (*mgo.Session, *mgo.Collection) {
	session := mgoSession.Clone()
	col := session.DB(db_blockChain).C(col_tx)
	return session, col
}

func KeyProvider(prifexkey mongokey.MongoKey) (*mgo.Session, *mgo.Collection) {
	session := mgoSession.Clone()
	col := session.DB(db_blockChain).C(prifexkey.Value())
	return session, col
}