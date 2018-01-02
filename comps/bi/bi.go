package bi

import (
	"sync"
)

//BI BI
type BI struct {
	mut     sync.RWMutex
	callers map[string]*caller
}

//NewBI NewBI
func NewBI() *BI {
	return &BI{callers: map[string]*caller{}}
}

//On On
func (bi *BI) On(m string, f interface{}) {
	bi.mut.Lock()
	bi.callers[m] = newCaller(f)
	bi.mut.Unlock()
}

//Handle Handle
func (bi *BI) Handle(sess Session) {
	sess.GetSessionImpl().handle(bi, sess)
}

func (bi *BI) onRequest(from interface{}, m string, p Protocol, a []byte) ([]byte, error) {
	bi.mut.RLock()
	caller, ok := bi.callers[m]
	bi.mut.RUnlock()
	if ok {
		return caller.call(from, p, a)
	}
	return nil, nil
}
