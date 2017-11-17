package bi

import (
	"github.com/gorilla/websocket"
)

const (
	biWebsocketMessageType = -9527
)

//WebsocketConn WebsocketConn
type WebsocketConn struct {
	conn *websocket.Conn
}

//NewWebsocketConn NewWebsocketConn
func NewWebsocketConn(conn *websocket.Conn) *WebsocketConn {
	c := WebsocketConn{conn: conn}
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

//Write Write
func (conn *WebsocketConn) Write(data []byte) error {
	return conn.conn.WriteMessage(biWebsocketMessageType, data)
}

//Read Read
func (conn *WebsocketConn) Read() ([]byte, error) {
	for {
		t, data, err := conn.conn.ReadMessage()
		// if nil != err {
		// 	log.Println(err)
		// }
		if biWebsocketMessageType == t {
			return data, err
		}
	}
}
