package se

import (
	"sync"
	"time"

	"github.com/emsihyo/go.im/comps/bi"
	"github.com/emsihyo/go.im/comps/pr"
	"github.com/emsihyo/go.im/comps/sp"
)

//Session Session
type Session struct {
	impl        *bi.SessionImpl
	logged      bool
	consumer    *pr.Consumer
	loggedMut   sync.RWMutex
	topicIDs    map[string]string
	topicIDsMut sync.RWMutex
}

//NewSession NewSession
func NewSession(conn bi.Conn, protocol bi.Protocol, timeout time.Duration) *Session {
	return &Session{impl: bi.NewSessionImpl(conn, protocol, timeout), topicIDs: map[string]string{}}
}

//GetSessionImpl GetSessionImpl
func (sess *Session) GetSessionImpl() *bi.SessionImpl {
	return sess.impl
}

//GetLogged GetLogged
func (sess *Session) GetLogged() bool {
	sess.loggedMut.RLock()
	defer sess.loggedMut.RUnlock()
	return sess.logged
}

//SetLogged SetLogged
func (sess *Session) SetLogged(logged bool) {
	sess.loggedMut.Lock()
	defer sess.loggedMut.Unlock()
	sess.logged = logged
}

//SetConsumer SetConsumer
func (sess *Session) SetConsumer(consumer *pr.Consumer) {
	sess.loggedMut.Lock()
	defer sess.loggedMut.Unlock()
	sess.consumer = consumer
}

//GetConsumer GetConsumer
func (sess *Session) GetConsumer() *pr.Consumer {
	sess.loggedMut.RLock()
	defer sess.loggedMut.RUnlock()
	return sess.consumer
}

//Subscribe Subscribe
func (sess *Session) Subscribe(topicID string, executeFunc func(sess sp.Session, topicID string) error) error {
	sess.topicIDsMut.Lock()
	defer sess.topicIDsMut.Unlock()
	if _, ok := sess.topicIDs[topicID]; !ok {
		sess.topicIDs[topicID] = topicID
		return executeFunc(sess, topicID)
	}
	return nil
}

//Unsubscribe Unsubscribe
func (sess *Session) Unsubscribe(topicID string, executeFunc func(sess sp.Session, topicID string)) {
	sess.topicIDsMut.Lock()
	defer sess.topicIDsMut.Unlock()
	if _, ok := sess.topicIDs[topicID]; ok {
		delete(sess.topicIDs, topicID)
		executeFunc(sess, topicID)
	}
}

//UnsubscribeAll UnsubscribeAll
func (sess *Session) UnsubscribeAll(executeFunc func(sess sp.Session, topicID string)) {
	sess.topicIDsMut.Lock()
	defer sess.topicIDsMut.Unlock()
	for _, topicID := range sess.topicIDs {
		delete(sess.topicIDs, topicID)
		executeFunc(sess, topicID)
	}
}
