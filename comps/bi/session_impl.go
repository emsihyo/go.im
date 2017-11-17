package bi

import (
	"errors"
	"log"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	//Connection Connection
	Connection = "Connection"
	//Disconnection Disconnection
	Disconnection = "Disconnection"
)

//SessionImpl SessionImpl
type SessionImpl struct {
	id                 string
	b                  *BI
	extension          interface{}
	protocol           Protocol
	conn               Conn
	qa                 *QA
	isClosed           bool
	closedErr          error
	closedMut          sync.Mutex
	sendMarshalledData chan []byte
	didReceiveMessage  chan *Message
	didReceiveError    chan error
	didDisconnects     []chan error
	timeout            time.Duration
	timer              *time.Timer
}

//NewSessionImpl NewSessionImpl
func NewSessionImpl(conn Conn, protocol Protocol, timeout time.Duration) *SessionImpl {
	return &SessionImpl{id: uuid.NewV3(uuid.NewV4(), conn.RemoteAddr()).String(), conn: conn, protocol: protocol, qa: newQA(), didDisconnects: []chan error{}, didReceiveMessage: make(chan *Message), didReceiveError: make(chan error), sendMarshalledData: make(chan []byte, 512), timeout: timeout, timer: time.NewTimer(timeout)}
}

//GetID GetID
func (sess *SessionImpl) GetID() string {
	return sess.id
}

//GetProtocol GetProtocol
func (sess *SessionImpl) GetProtocol() Protocol {
	return sess.protocol
}

//Close Close
func (sess *SessionImpl) Close() {
	sess.conn.Close()
}

//Emit Emit
func (sess *SessionImpl) Emit(method string, args interface{}) error {
	var a []byte
	var err error
	if nil != args {
		if a, err = sess.protocol.Marshal(args); nil != err {
			log.Println(err)
			return err
		}
	}
	m := Message{
		T: Message_Emit,
		M: method,
		A: a,
	}
	sess.sendMessage(&m)
	return nil
}

//Request Request
func (sess *SessionImpl) Request(method string, args interface{}, resp interface{}, timeout time.Duration) error {
	var a []byte
	var err error
	protocol := sess.protocol
	if nil != args {
		if a, err = protocol.Marshal(args); nil != err {
			log.Println(err)
			return err
		}
	}
	qa := sess.qa
	m := Message{
		T: Message_Request,
		I: qa.nextID(),
		M: method,
		A: a,
	}
	q := make(chan []byte, 1)
	qa.addQ(m.I, q)
	defer qa.removeQ(m.I)
	disconnection := sess.waitForDisconnection()
	sess.sendMessage(&m)
	select {
	case respBytes := <-q:
		err = protocol.Unmarshal(respBytes, resp)
		if nil != err {
			log.Println(err)
		}
		return err
	case err = <-disconnection:
		if nil != err {
			return err
		}
		return errors.New("bi: session.impl.closed")
	case <-time.After(timeout):
		return errors.New("bi: session.impl.timeout")
	}
}

//SendMarshalledData SendMarshalledData
func (sess *SessionImpl) SendMarshalledData(marshalledData []byte) {
	sess.sendMarshalledData <- marshalledData
}

func (sess *SessionImpl) handle(b *BI, extension interface{}) {
	sess.b = b
	sess.extension = extension
	sess.b.onEmit(sess.extension, Connection, sess.protocol, nil)
	waitForDisconnection := sess.waitForDisconnection()
	go func() {
	loop1:
		for {
			select {
			case <-waitForDisconnection:
				break loop1
			case data := <-sess.sendMarshalledData:
				sess.conn.Write(data)
			}

		}
	}()
	go func() {
		var err error
		var data []byte
		for {
			if data, err = sess.conn.Read(); nil != err {
				sess.didReceiveError <- err
				break
			}
			if nil == data {
				continue
			}
			m := Message{}
			err = sess.protocol.Unmarshal(data, &m)
			if nil != err {
				log.Println(err)
				sess.didReceiveError <- err
				break
			} else {
				sess.didReceiveMessage <- &m
			}
		}
	}()
	var err error
	var ignore bool
loop2:
	for {
		sess.timer.Reset(sess.timeout)
		select {
		case <-sess.timer.C:
			ignore = true
			sess.conn.Close()
		case m := <-sess.didReceiveMessage:
			if true != ignore {
				switch m.T {
				case Message_Emit:
					sess.b.onEmit(sess.extension, m.M, sess.protocol, m.A)
				case Message_Request:
					resp, _ := sess.b.onRequest(sess.extension, m.M, sess.protocol, m.A)
					m.T = Message_Response
					m.A = resp
					m.M = ""
					sess.sendMessage(m)
				case Message_Response:
					q := sess.qa.getQ(m.I)
					if nil != q {
						q <- m.A
					}
				}
			}
		case err = <-sess.didReceiveError:
			break loop2
		}
	}
	sess.closedMut.Lock()
	defer sess.closedMut.Unlock()
	sess.closedErr = err
	sess.isClosed = true
	sess.conn.Close()
	for _, didDisconnect := range sess.didDisconnects {
		didDisconnect <- err
	}
	sess.b.onEmit(sess.extension, Disconnection, sess.protocol, nil)
}

func (sess *SessionImpl) sendMessage(m *Message) {
	marshalledData, err := sess.protocol.Marshal(m)
	if nil != err {
		log.Println(err)
		return
	}
	sess.sendMarshalledData <- marshalledData
}

func (sess *SessionImpl) waitForDisconnection() chan error {
	didDisconnect := make(chan error, 1)
	sess.closedMut.Lock()
	defer sess.closedMut.Unlock()
	if true == sess.isClosed {
		go func() {
			didDisconnect <- sess.closedErr
		}()
	} else {
		sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	}
	return didDisconnect
}
