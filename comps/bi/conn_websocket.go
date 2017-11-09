package bi

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

const (
	biWebsocketMessageType = -9527
)

type websocketReceivedRes struct {
	data []byte
	err  error
}

//WebsocketConn WebsocketConn
type WebsocketConn struct {
	conn       *websocket.Conn
	didReceive chan *websocketReceivedRes
	timeout    time.Duration
	t          *time.Timer
}

//NewWebsocketConn NewWebsocketConn
func NewWebsocketConn(conn *websocket.Conn, timeout time.Duration) *WebsocketConn {
	c := WebsocketConn{conn: conn, timeout: timeout, t: time.NewTimer(timeout), didReceive: make(chan *websocketReceivedRes)}
	go c.handle()
	return &c
}

//Close Close
func (conn *WebsocketConn) Close() {
	conn.conn.Close()
}

//RemoteAddr RemoteAddr
func (conn *WebsocketConn) RemoteAddr() string {
	return conn.conn.RemoteAddr().String()
}

func (conn *WebsocketConn) handle() {
	for {
		t, data, err := conn.conn.ReadMessage()
		if biWebsocketMessageType == t {
			conn.didReceive <- &websocketReceivedRes{data: data, err: err}
		}
	}
}

//Write Write
func (conn *WebsocketConn) Write(data []byte) error {
	return conn.conn.WriteMessage(biWebsocketMessageType, data)
}

//Read Read
func (conn *WebsocketConn) Read() ([]byte, error) {
	conn.t.Reset(conn.timeout)
	select {
	case <-conn.t.C:
		conn.conn.Close()
		return nil, errors.New("websocket conn timeout")
	case res := <-conn.didReceive:
		return res.data, res.err
	}
}
