package st

import (
	"time"

	"github.com/emsihyo/go.im/comps/pr"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	//DB DB
	DB = "im"
	//TableMessage TableMessage
	TableMessage = "message"
	//TableSeq TableSeq
	TableSeq = "message_seq"
)

//MongoStorage MongoStorage
type MongoStorage struct {
	sess  *mgo.Session
	group chan struct{}
}

//NewMongoStorage NewMongoStorage
func NewMongoStorage(url string) *MongoStorage {
	if sess, err := mgo.Dial(url); nil == err {
		sess.SetPoolLimit(50)
		sess.DB(DB).C(TableMessage).EnsureIndexKey("cid")
		sess.DB(DB).C(TableMessage).EnsureIndexKey("to.id", "sid")
		return &MongoStorage{sess: sess, group: make(chan struct{}, 50)}
	}
	return nil
}

//Store Store
func (s *MongoStorage) Store(message *pr.Message) (isNew bool, err error) {
	s.group <- struct{}{}
	defer func() {
		<-s.group
	}()
	sess := s.sess.Copy()
	defer sess.Close()
	db := sess.DB(DB)
	result := bson.M{}
	err = db.C(TableMessage).Find(bson.M{"cid": message.GetCID()}).One(&result)
	if nil == err {
		message.SID = result["sid"].(int64)
		message.At = result["at"].(int64)
		return false, nil
	}
	if mgo.ErrNotFound != err {
		return false, err
	}
	_, err = db.C(TableSeq).Find(bson.M{"_id": message.GetTo().GetID()}).Apply(mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": int64(1)}},
		Upsert:    true,
		ReturnNew: true,
	}, &result)
	if nil != err {
		return false, err
	}
	message.SID = result["seq"].(int64)
	message.At = time.Now().Unix()
	err = db.C(TableMessage).Insert(message)
	return true, err
}

//Fetch Fetch
func (s *MongoStorage) Fetch(topicID string, minSID int64, maxCount int64) ([]*pr.Message, int64, error) {
	s.group <- struct{}{}
	defer func() {
		<-s.group
	}()
	if 0 >= maxCount {
		return nil, 0, nil
	}
	sess := s.sess.Copy()
	defer sess.Close()
	var results []*pr.Message
	query := sess.DB(DB).C(TableMessage).Find(bson.M{"to.id": topicID, "sid": bson.M{"$gt": minSID}})
	count, err := query.Count()
	if nil != err {
		return nil, 0, err
	}
	err = query.Sort("-sid").Limit(int(maxCount)).All(&results)
	if nil != err {
		return nil, 0, err
	}
	return results, int64(count), nil
}
