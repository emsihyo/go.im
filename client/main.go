package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_ "net/http/pprof"

	uuid "github.com/satori/go.uuid"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error:" + err.Error())
	}
	var prefix string
	for _, inter := range interfaces {
		prefix += fmt.Sprintf("|%s%s|", inter.Name, inter.HardwareAddr)
	}
	prefix = uuid.NewV5(uuid.NewV4(), prefix).String()
	portAddrPtr := flag.String("port", "10000", "http port")
	hostAddrPtr := flag.String("host", ":10001", "tcp addr")
	usersPtr := flag.Int("users", 10000, "user count")
	topicsPtr := flag.Int("topics", 1000, "topic count")
	perPtr := flag.Int("per", 2, "topics per user")
	durationPtr := flag.Int64("duration", 60, "duration in second")
	flag.Parse()
	ch := make(chan struct{}, 10)
	go func() {
		for i := 0; i < *usersPtr; i++ {
			ch <- struct{}{}
			cli := NewClient(*hostAddrPtr, prefix+"|user:"+fmt.Sprintln(i), "robot")
			idx := i
			go func() {
				slot := idx % (*topicsPtr / *perPtr)
				start := (*perPtr) * slot
				for i := start; i < start+*perPtr; i++ {
					cli.Subscribe("room:" + fmt.Sprintln(i))
					<-time.After(time.Millisecond * 10)
				}
				<-time.After(time.Millisecond * 200)
				<-ch
				go func() {
					for {
						<-time.After(time.Duration(int64(time.Second) * *durationPtr))
						topicID := cli.RandomTopic()
						cli.Publish(topicID, fmt.Sprintln(time.Now().Unix()))
					}
				}()
			}()
			go func() {
				for {
					cli.Ping()
					<-time.After(time.Second * 60)
				}
			}()
		}
	}()
	log.Print(http.ListenAndServe(":"+*portAddrPtr, nil))
}
