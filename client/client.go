package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/emsihyo/go.im/comps/se"
	"github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"

	"github.com/emsihyo/go.im/comps/bi"
	"github.com/emsihyo/go.im/comps/pr"
)

//Client Client
type Client struct {
	b           *bi.BI
	sess        bi.Session
	addr        string
	id          string
	platform    string
	password    string
	isConnected bool
	isLogged    bool
	topics      map[string]string
	sids        map[string]int64
	mut         sync.RWMutex
}

//NewClient NewClient
func NewClient(addr string, id string, platform string) *Client {
	b := bi.NewBI()
	cli := Client{b: b, addr: addr, id: id, platform: platform, topics: map[string]string{}, sids: map[string]int64{}}
	cli.b.On(pr.Type_Send.String(), func(sess bi.Session, req *pr.EmitSend) {
		go func() {
			cli.mut.Lock()
			defer cli.mut.Unlock()
			sid := req.GetMessage().GetSID()
			id := req.GetMessage().GetTo().GetID()
			if sid > cli.sids[id] {
				cli.sids[id] = sid
				logrus.Debug("[REV]", req.GetMessage().GetTo().GetID(), req.GetMessage().GetFrom().GetID(), req.GetMessage().GetBody())
			}
		}()
	})
	return &cli
}

func (cli *Client) preproccess() {
	for {
		if false == cli.isConnected {
			if err := cli.connect(); nil != err {
				continue
			} else {
				cli.isConnected = true
			}
		}
		if false == cli.isLogged {
			if err := cli.login(); nil != err {
				continue
			} else {
				cli.isLogged = true
				for _, topicID := range cli.topics {
					if _, err := cli.subscribe(topicID); nil != err {
						cli.reset()
						goto label
					}
				}
				return
			}
		} else {
			return
		}
	label:
	}
}

func (cli *Client) connect() error {
	a, err := net.ResolveTCPAddr("tcp", cli.addr)
	if nil != err {
		logrus.Debug("[CONNECT]", cli.id, err)
		return err
	}
	c, err := net.DialTCP("tcp", nil, a)
	if nil != err {
		logrus.Debug("[CONNECT]", cli.id, err)
		return err
	}
	conn := bi.NewTCPConn(c, time.Hour)
	sess := se.NewSession(conn, &bi.ProtobufProtocol{})
	cli.sess = sess
	go cli.b.Handle(sess)
	return nil
}

func (cli *Client) login() error {
	resp := pr.RespLogin{}
	err := cli.sess.GetSessionImpl().Request(pr.Type_Login.String(), &pr.ReqLogin{UserID: cli.id, Password: cli.password, Platform: cli.platform}, &resp, time.Hour)
	if nil != err {
		logrus.Debug("[LOGIN]", cli.id, err)
		return err
	}
	if 0 != resp.Code {
		err = fmt.Errorf("code:%d, desc:%s", resp.Code, resp.Desc)
		logrus.Debug("[LOGIN]", cli.id, err)
		return err
	}
	return nil
}

//Subscribe Subscribe
func (cli *Client) Subscribe(topicID string) {
	cli.mut.Lock()
	defer cli.mut.Unlock()
	for {
		cli.preproccess()
		_, err := cli.subscribe(topicID)
		if nil != err {
			cli.reset()
		} else {
			return
		}
	}
}

//Unsubscribe Unsubscribe
// func (cli *Client) Unsubscribe(topicID string) {
// 	cli.mut.Lock()
// 	defer cli.mut.Unlock()
// 	for {
// 		cli.preproccess()
// 		err := cli.unsubscribe(topicID)
// 		if nil != err {
// 			logrus.Debug("[UNSUBSCRIBE]", err)
// 			cli.reset()
// 		} else {
// 			return
// 		}
// 	}
// }

//Publish Publish
func (cli *Client) Publish(topicID string, body string) {
	cli.mut.Lock()
	defer cli.mut.Unlock()
	for {
		cli.preproccess()
		err := cli.publish(topicID, body)
		if nil != err {
			cli.reset()
		} else {
			return
		}
	}
}

//Ping Ping
func (cli *Client) Ping() int64 {
	cli.mut.Lock()
	defer cli.mut.Unlock()
	for {
		cli.preproccess()
		delay, err := cli.ping()
		if nil != err {
			cli.reset()
		} else {
			return delay
		}
	}
}

func (cli *Client) subscribe(topicID string) ([]*pr.Message, error) {
	if _, ok := cli.topics[topicID]; ok {
		return nil, nil
	}
	sid := cli.sids[topicID]
	resp := pr.RespSubscribe{}
	err := cli.sess.GetSessionImpl().Request(pr.Type_Subscribe.String(), &pr.ReqSubscribe{TopicID: topicID, MinSID: sid, MaxCount: 20}, &resp, time.Hour)
	if nil != err {
		logrus.Debug("[SUBSCRIBE]", cli.id, topicID, err)
		return nil, err
	}
	if 0 != resp.Code {
		err = fmt.Errorf("code:%d, desc:%s", resp.Code, resp.Desc)
		logrus.Debug("[SUBSCRIBE]", cli.id, topicID, err)
		return nil, err
	}
	cli.topics[topicID] = topicID
	if 0 != len(resp.Histories) {
		cli.sids[topicID] = resp.Histories[len(resp.Histories)-1].GetSID()
	}
	logrus.Debug("[SUBSCRIBE]", cli.id, topicID)
	return resp.Histories, nil
}

// func (cli *Client) unsubscribe(topicID string) error {
// 	if _, ok := cli.topics[topicID]; !ok {
// 		return nil
// 	}
// 	resp := pr.RespUnsubscribe{}
// 	err := cli.sess.Request(int32(pr.Type_Unsubscribe), &pr.ReqUnsubscribe{TopicID: topicID}, &resp, time.Hour)
// 	if nil != err {
// 		return err
// 	}
// 	if 0 != resp.Code {
// 		return fmt.Errorf("code:%d, desc:%s", resp.Code, resp.Desc)
// 	}
// 	delete(cli.topics, topicID)
// 	return nil
// }

func (cli *Client) publish(topicID string, body string) error {
	message := pr.Message{CID: uuid.NewV3(uuid.NewV4(), topicID+"|"+cli.id).String(), To: &pr.Topic{ID: topicID}, Body: body, From: &pr.Consumer{ID: cli.id}}
	resp := pr.RespDeliver{}
	err := cli.sess.GetSessionImpl().Request(pr.Type_Deliver.String(), &pr.ReqDeliver{Message: &message}, &resp, time.Hour)
	if nil != err {
		logrus.Debug("[PUBLISH]", cli.id, topicID, body, err)
		return err
	}
	if 0 != resp.Code {
		err = fmt.Errorf("code:%d, desc:%s", resp.Code, resp.Desc)
		logrus.Debug("[PUBLISH]", cli.id, topicID, body, err)
		return err
	}
	logrus.Debug("[PUBLISH]", cli.id, topicID, body)
	return nil
}

func (cli *Client) ping() (int64, error) {
	resp := pr.RespPing{}
	now := time.Now().UnixNano()
	err := cli.sess.GetSessionImpl().Request(pr.Type_Ping.String(), &pr.ReqPing{}, &resp, time.Hour)
	if nil != err {
		logrus.Debug("[PING]", cli.id, err)
		return 0, err
	}
	// delay := (time.Now().UnixNano() - now) / time.Millisecond.Nanoseconds()
	delay := (time.Now().UnixNano() - now)
	logrus.Debug("[PING]", cli.id, delay)
	return delay, nil
}

//RandomTopic RandomTopic
func (cli *Client) RandomTopic() string {
	cli.mut.Lock()
	defer cli.mut.Unlock()
	var topic string
	for _, r := range cli.topics {
		if strings.EqualFold(r, cli.id) {
			continue
		}
		topic = r
		break
	}
	return topic
}

//TopicCount TopicCount
func (cli *Client) TopicCount() int {
	cli.mut.Lock()
	defer cli.mut.Unlock()
	return len(cli.topics)
}

func (cli *Client) reset() {
	cli.isConnected = false
	cli.isLogged = false
	cli.sess = nil
}
