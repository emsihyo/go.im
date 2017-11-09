package bi

import (
	"errors"
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
	id             string
	b              *BI
	extension      interface{}
	protocol       Protocol
	conn           Conn
	qa             *QA
	isClosed       bool
	closedErr      error
	closedMut      sync.RWMutex
	didDisconnects []chan error
}

//NewSessionImpl NewSessionImpl
func NewSessionImpl(conn Conn, protocol Protocol) *SessionImpl {
	return &SessionImpl{id: uuid.NewV3(uuid.NewV4(), conn.RemoteAddr()).String(), conn: conn, protocol: protocol, qa: newQA(), didDisconnects: []chan error{}}
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
		return protocol.Unmarshal(respBytes, resp)
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
	sess.conn.Write(marshalledData)
}

func (sess *SessionImpl) handle(b *BI, extension interface{}) {
	sess.b = b
	sess.extension = extension
	var err error
	var data []byte
	sess.b.onEmit(sess.extension, Connection, sess.protocol, nil)
	for {
		if data, err = sess.conn.Read(); nil != err {
			break
		}
		if nil == data {
			continue
		}
		m := Message{}
		if err = sess.protocol.Unmarshal(data, &m); nil != err {
			break
		}
		switch m.T {
		case Message_Emit:
			sess.b.onEmit(sess.extension, m.M, sess.protocol, m.A)
		case Message_Request:
			resp, _ := sess.b.onRequest(sess.extension, m.M, sess.protocol, m.A)
			m.T = Message_Response
			m.A = resp
			m.M = ""
			sess.sendMessage(&m)
		case Message_Response:
			q := sess.qa.getQ(m.I)
			if nil != q {
				q <- m.A
			}
		}
	}
	sess.closedMut.Lock()
	defer sess.closedMut.Unlock()
	sess.conn.Close()
	sess.isClosed = true
	if nil != sess.closedErr {
		err = sess.closedErr
	}
	for _, didDisconnect := range sess.didDisconnects {
		didDisconnect <- err
	}
	sess.b.onEmit(sess.extension, Disconnection, sess.protocol, nil)
}

func (sess *SessionImpl) sendMessage(m *Message) {
	marshalledData, err := sess.protocol.Marshal(m)
	if nil != err {
		return
	}
	sess.conn.Write(marshalledData)
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
