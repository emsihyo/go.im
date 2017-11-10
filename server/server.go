package main

import (
	"sync/atomic"

	"github.com/emsihyo/go.im/comps/bi"
	"github.com/emsihyo/go.im/comps/pr"
	"github.com/emsihyo/go.im/comps/se"
	"github.com/emsihyo/go.im/comps/sp"
	"github.com/emsihyo/go.im/comps/st"
)

//Server Server
type Server struct {
	bi         *bi.BI
	broker     *sp.Broker
	storage    st.Storage
	totalConn  int64
	messageIn  uint64
	messageOut uint64
}

//NewServer NewServer
func NewServer(storage st.Storage) *Server {
	serv := Server{bi: bi.NewBI(), storage: storage}
	serv.init()
	serv.broker = sp.NewBroker(uint32(1024), serv.broadcast)
	return &serv
}

//Handle Handle
func (serv *Server) Handle(sess *se.Session) {
	serv.bi.Handle(sess)
}

func (serv *Server) init() {

	serv.bi.On(bi.Connection, func(sess *se.Session) {
		atomic.AddInt64(&serv.totalConn, 1)
	})

	serv.bi.On(bi.Disconnection, func(sess *se.Session) {
		if sess.GetLogged() {
			sess.SetLogged(false)
			sess.SetConsumer(nil)
			sess.UnsubscribeAll(func(sess sp.Session, topicID string) {
				serv.broker.Unsubscribe(topicID, sess.GetSessionImpl().GetID())
			})
		}
		atomic.AddInt64(&serv.totalConn, -1)
	})

	serv.bi.On(pr.Type_Ping.String(), func(sess *se.Session, req *pr.ReqPing) *pr.RespPing {
		atomic.AddUint64(&serv.messageIn, 1)
		defer func() {
			atomic.AddUint64(&serv.messageOut, 1)
		}()
		return &pr.RespPing{}
	})

	serv.bi.On(pr.Type_Login.String(), func(sess *se.Session, req *pr.ReqLogin) *pr.RespLogin {
		atomic.AddUint64(&serv.messageIn, 1)
		defer func() {
			atomic.AddUint64(&serv.messageOut, 1)
		}()
		logged := true
		if !logged {
			return &pr.RespLogin{Code: 1, Desc: "login at other device"}
		}
		sess.SetConsumer(&pr.Consumer{ID: req.GetUserID()})
		sess.SetLogged(true)
		return &pr.RespLogin{Code: 0, Desc: "成功"}
	})

	serv.bi.On(pr.Type_Logout.String(), func(sess *se.Session, req *pr.ReqLogout) *pr.RespLogout {
		atomic.AddUint64(&serv.messageIn, 1)
		defer func() {
			atomic.AddUint64(&serv.messageOut, 1)
		}()
		if !sess.GetLogged() {
			return &pr.RespLogout{Code: 1, Desc: ""}
		}
		sess.SetLogged(false)
		sess.SetConsumer(nil)
		sess.UnsubscribeAll(func(sess sp.Session, topicID string) {
			serv.broker.Unsubscribe(topicID, sess.GetSessionImpl().GetID())
		})
		return &pr.RespLogout{Code: 0, Desc: ""}
	})

	serv.bi.On(pr.Type_Subscribe.String(), func(sess *se.Session, req *pr.ReqSubscribe) *pr.RespSubscribe {
		atomic.AddUint64(&serv.messageIn, 1)
		defer func() {
			atomic.AddUint64(&serv.messageOut, 1)
		}()
		if !sess.GetLogged() {
			return &pr.RespSubscribe{Code: 1, Desc: ""}
		}
		canSubscribe := true
		if !canSubscribe {
			return &pr.RespSubscribe{Code: 1, Desc: ""}
		}
		if 0 == len(req.GetTopicID()) {
			return &pr.RespSubscribe{Code: 1, Desc: ""}
		}
		var messages []*pr.Message
		var err error
		err = sess.Subscribe(req.GetTopicID(), func(sess sp.Session, topicID string) error {
			c := sp.NewConsumer(sess)
			serv.broker.Subscribe(topicID, c)
			minSID := req.GetMinSID()
			if 0 <= minSID {
				if 0 < req.GetMaxCount() {
					messages, _, err = serv.storage.Fetch(req.GetTopicID(), req.GetMinSID(), req.GetMaxCount())
					if nil != err {
						return err
					}
					if 0 != len(messages) {
						minSID = messages[0].GetSID()
					}
				}
				go func() {
					c.Ready(minSID)
				}()
			} else {
				go func() {
					c.Ready(0)
				}()
			}
			return nil
		})
		if nil != err {
			sess.Unsubscribe(req.GetTopicID(), func(sess sp.Session, topicID string) {
				serv.broker.Unsubscribe(topicID, sess.GetSessionImpl().GetID())
			})
			return &pr.RespSubscribe{Code: 1, Desc: ""}
		}
		return &pr.RespSubscribe{Code: 0, Desc: "", Histories: messages}
	})

	serv.bi.On(pr.Type_Unsubscribe.String(), func(sess *se.Session, req *pr.ReqUnsubscribe) *pr.RespUnsubscribe {
		atomic.AddUint64(&serv.messageIn, 1)
		defer func() {
			atomic.AddUint64(&serv.messageOut, 1)
		}()
		if !sess.GetLogged() {
			return &pr.RespUnsubscribe{Code: 1, Desc: ""}
		}
		canUnsubscribe := true
		if !canUnsubscribe {
			return &pr.RespUnsubscribe{Code: 1, Desc: ""}
		}
		if 0 == len(req.GetTopicID()) {
			return &pr.RespUnsubscribe{Code: 1, Desc: ""}
		}
		sess.Unsubscribe(req.GetTopicID(), func(sess sp.Session, topicID string) {
			serv.broker.Unsubscribe(topicID, sess.GetSessionImpl().GetID())
		})
		return &pr.RespUnsubscribe{Code: 0, Desc: ""}
	})

	serv.bi.On(pr.Type_Deliver.String(), func(sess *se.Session, req *pr.ReqDeliver) *pr.RespDeliver {
		atomic.AddUint64(&serv.messageIn, 1)
		defer func() {
			atomic.AddUint64(&serv.messageOut, 1)
		}()
		if !sess.GetLogged() {
			return &pr.RespDeliver{Code: 1, Desc: ""}
		}
		canDeliver := true
		if !canDeliver {
			return &pr.RespDeliver{Code: 1, Desc: ""}
		}
		m := req.GetMessage()
		if 0 == len(m.GetTo().GetID()) || 0 == len(m.GetCID()) {
			return &pr.RespDeliver{Code: 1, Desc: ""}
		}
		isNew, err := serv.storage.Store(m)
		if nil != err {
			return &pr.RespDeliver{Code: 1, Desc: err.Error()}
		}
		if isNew {
			serv.broker.Publish(m.GetTo().GetID(), sess, req.GetMessage())
			if false == req.GetMessage().GetTo().GetGroup() {
				serv.broker.Publish(m.GetFrom().GetID(), sess, req.GetMessage())
			}
		}
		return &pr.RespDeliver{Code: 0, Desc: "", SID: m.SID, At: m.At}
	})
}

func (serv *Server) broadcast(topicID string, to map[string]*sp.Consumer, from sp.Session, message interface{}) {
	m, ok := message.(*pr.Message)
	if !ok {
		return
	}
	datas := map[bi.Protocol][]byte{}
	for _, consumer := range to {
		protocol := consumer.GetSession().GetSessionImpl().GetProtocol()
		data, ok := datas[protocol]
		if false == ok {
			e := pr.EmitSend{Message: m}
			a, err := protocol.Marshal(&e)
			if nil != err {
				datas[protocol] = []byte{}
				continue
			}
			data, err = protocol.Marshal(&bi.Message{M: pr.Type_Send.String(), T: bi.Message_Emit, A: a})
			if nil != err {
				datas[protocol] = []byte{}
				continue
			}
			datas[protocol] = data
		} else {
			if 0 == len(data) {
				continue
			}
		}
		consumer.SendMarshalledData(data, m.At)
		atomic.AddUint64(&serv.messageOut, 1)
	}
}

//GetTotalConsumer GetTotalConsumer
func (serv *Server) GetTotalConsumer() map[string]uint64 {
	return serv.broker.GetTotalConsumer()
}

//GetTotalConn GetTotalConn
func (serv *Server) GetTotalConn() uint64 {
	return uint64(atomic.LoadInt64(&serv.totalConn))
}

//GetTotalTopic GetTotalTopic
func (serv *Server) GetTotalTopic() uint64 {
	return serv.broker.GetTotalTopic()
}

//GetMessageIn GetMessageIn
func (serv *Server) GetMessageIn() uint64 {
	return atomic.LoadUint64(&serv.messageIn)
}

//GetMessageOut GetMessageOut
func (serv *Server) GetMessageOut() uint64 {
	return atomic.LoadUint64(&serv.messageOut)
}
