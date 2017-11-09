package st

import "github.com/emsihyo/go.im/comps/pr"

//Storage Storage
type Storage interface {
	Store(message *pr.Message) (isNew bool, err error)
	Fetch(topicID string, minSID int64, maxCount int64) ([]*pr.Message, int64, error)
}
