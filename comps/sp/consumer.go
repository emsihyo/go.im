package sp

import (
	"sync"
)

type buffer struct {
	data []byte
	sid  int64
}

func newbuffer(data []byte, sid int64) *buffer {
	return &buffer{data: data, sid: sid}
}

//Consumer Consumer
type Consumer struct {
	bfs  []*buffer
	sess Session
	min  int64
	mut  sync.RWMutex
}

//NewConsumer NewConsumer
func NewConsumer(sess Session) *Consumer {
	return &Consumer{sess: sess, min: -1, bfs: []*buffer{}}
}

//GetSession GetSession
func (c *Consumer) GetSession() Session {
	return c.sess
}

//Ready Ready
func (c *Consumer) Ready(min int64) {
	c.mut.Lock()
	c.min = min
	for _, b := range c.bfs {
		if min < b.sid {
			c.sess.GetSessionImpl().SendMarshalledData(b.data)
		}
	}
	c.mut.Unlock()
}

//SendMarshalledData SendMarshalledData
func (c *Consumer) SendMarshalledData(data []byte, sid int64) {
	c.mut.RLock()
	if 0 > c.min {
		c.bfs = append(c.bfs, newbuffer(data, sid))
	} else {
		if sid > c.min {
			c.sess.GetSessionImpl().SendMarshalledData(data)
		}
	}
	c.mut.RUnlock()
}
