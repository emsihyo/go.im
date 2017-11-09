package sp

import (
	"github.com/emsihyo/go.im/comps/bi"
)

//Session Session
type Session interface {
	Subscribe(topicID string, executeFunc func(sess Session, topicID string) error) error
	Unsubscribe(topicID string, executeFunc func(sess Session, topicID string))
	GetSessionImpl() *bi.SessionImpl
}
