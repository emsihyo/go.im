package st

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/emsihyo/go.im/comps/pr"
)

// func TestStorage(t *testing.T) {
// 	storage := NewStorageImpl("127.0.0.1:27017")
// 	finish := make(chan struct{})
// 	count := 0
// 	var mut sync.RWMutex
// 	for i := 0; i < 20; i++ {
// 		for j := 0; j < 2000; j++ {
// 			mut.Lock()
// 			count++
// 			mut.Unlock()
// 			cid := fmt.Sprintf("cid: %d|%d", i, j)
// 			uid := fmt.Sprintf("uid: %d", i)
// 			rid := fmt.Sprintf("rid: %d", i)
// 			go func() {
// 				message := Message{
// 					CID:     cid,
// 					To:      rid,
// 					From:    uid,
// 					At:      time.Now().UnixNano(),
// 					Content: fmt.Sprintf("from: %s, to: %s", uid, rid),
// 				}
// 				if err := storage.Store(&message); nil != err {
// 					t.Fatal(err)
// 				}
// 				mut.Lock()
// 				count--
// 				if 0 == count {
// 					finish <- struct{}{}
// 				}
// 				mut.Unlock()
// 			}()
// 		}
// 	}
// 	<-finish
// }
var ii = 0

func BenchmarkStorage(b *testing.B) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	storage := NewMongoStorage("localhost")
	max := 1000
	ii++
	finish := make(chan struct{}, max)
	for i := ii; i < ii+1; i++ {
		for j := 0; j < max; j++ {
			cid := fmt.Sprintf("cid: %d|%d", i, j)
			uid := fmt.Sprintf("uid: %d", i)
			rid := fmt.Sprintf("rid: %d", i)
			go func() {
				message := pr.Message{
					CID:  cid,
					To:   &pr.Topic{ID: rid},
					From: &pr.Consumer{ID: uid},
					At:   time.Now().UnixNano(),
					Body: fmt.Sprintf("from: %s, to: %s", uid, rid),
				}
				_, err := storage.Store(&message)
				if nil != err {
					log.Println(err)
				}
				finish <- struct{}{}
			}()
		}
	}
	for i := 0; i < max; i++ {
		<-finish
	}
	_, _, err := storage.Fetch(fmt.Sprintf("rid: %d", ii), int64(max-100), 1)
	if nil != err {
		log.Println(err)
	} else {
	}
}
