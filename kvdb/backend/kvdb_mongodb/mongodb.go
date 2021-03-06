package kvdb_mongo

import (
	"gopkg.in/mgo.v2"

	"io"

	"github.com/xiaonanln/goworld/gwlog"
	. "github.com/xiaonanln/goworld/kvdb/types"
	"gopkg.in/mgo.v2/bson"
)

const (
	DEFAULT_DB_NAME = "goworld"
	VAL_KEY         = "_"
)

type MongoKVDB struct {
	s *mgo.Session
	c *mgo.Collection
}

func OpenMongoKVDB(url string, dbname string, collectionName string) (KVDBEngine, error) {
	gwlog.Debug("Connecting MongoDB ...")
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	if dbname == "" {
		// if db is not specified, use default
		dbname = DEFAULT_DB_NAME
	}
	db := session.DB(dbname)
	c := db.C(collectionName)
	return &MongoKVDB{
		s: session,
		c: c,
	}, nil
}

func (kvdb *MongoKVDB) Put(key string, val string) error {
	_, err := kvdb.c.UpsertId(key, map[string]string{
		VAL_KEY: val,
	})
	return err
}

func (kvdb *MongoKVDB) Get(key string) (val string, err error) {
	q := kvdb.c.FindId(key)
	var doc map[string]string
	err = q.One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		return
	}
	val = doc[VAL_KEY]
	return
}

type MongoKVIterator struct {
	it *mgo.Iter
}

func (it *MongoKVIterator) Next() (KVItem, error) {
	var doc map[string]string
	ok := it.it.Next(&doc)
	if ok {
		return KVItem{
			Key: doc["_id"],
			Val: doc["_"],
		}, nil
	} else {
		err := it.it.Close()
		if err != nil {
			return KVItem{}, err
		} else {
			return KVItem{}, io.EOF
		}
	}
}

func (kvdb *MongoKVDB) Find(beginKey string, endKey string) Iterator {
	q := kvdb.c.Find(bson.M{"_id": bson.M{"$gte": beginKey, "$lt": endKey}})
	it := q.Iter()
	return &MongoKVIterator{
		it: it,
	}
}

func (kvdb *MongoKVDB) Close() {
	kvdb.s.Close()
}

func (kvdb *MongoKVDB) IsEOF(err error) bool {
	return err == io.EOF || err == io.ErrUnexpectedEOF
}
