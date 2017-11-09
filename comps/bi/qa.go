package bi

import (
	"sync"
)

//QA QA
type QA struct {
	id    uint64
	idMtx sync.Mutex
	qs    map[uint64]chan []byte
	qsMtx sync.RWMutex
}

func newQA() *QA {
	return &QA{qs: map[uint64]chan []byte{}}
}

func (qa *QA) nextID() uint64 {
	qa.idMtx.Lock()
	defer qa.idMtx.Unlock()
	qa.id++
	return qa.id
}

func (qa *QA) addQ(id uint64, ch chan []byte) {
	qa.qsMtx.Lock()
	qa.qs[id] = ch
	qa.qsMtx.Unlock()
}

func (qa *QA) removeQ(id uint64) {
	qa.qsMtx.Lock()
	delete(qa.qs, id)
	qa.qsMtx.Unlock()
}

func (qa *QA) getQ(id uint64) chan []byte {
	qa.qsMtx.RLock()
	defer qa.qsMtx.RUnlock()
	return qa.qs[id]
}
