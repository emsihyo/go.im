package st

import (
	"sync"
	"time"

	"github.com/emsihyo/go.im/comps/pr"
)

//StorageImpl StorageImpl
type StorageImpl struct {
	sid int64
	mut sync.Mutex
}

//NewStorageImpl NewStorageImpl
func NewStorageImpl() *StorageImpl {
	return &StorageImpl{}
}

//Store Store
func (s *StorageImpl) Store(message *pr.Message) (isNew bool, err error) {
	s.mut.Lock()
	s.sid++
	message.SID = s.sid
	s.mut.Unlock()
	message.At = time.Now().Unix()
	return true, nil
}

//Fetch Fetch
func (s *StorageImpl) Fetch(topicID string, minSID int64, maxCount int64) ([]*pr.Message, int64, error) {
	return nil, 0, nil
}
