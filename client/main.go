package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.ErrorLevel)
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error:" + err.Error())
	}
	var prefix string
	for _, inter := range interfaces {
		prefix += fmt.Sprintf("|%s%s|", inter.Name, inter.HardwareAddr)
	}
	// prefix = uuid.NewV5(uuid.NewV4(), prefix).String()
	prefix = "A"
	hostAddrPtr := flag.String("host", ":9000", "tcp addr")
	usersPtr := flag.Int("users", 1000, "user count")
	topicsPtr := flag.Int("topics", 100, "topic count")
	perPtr := flag.Int("per", 20, "topics per user")
	durationPtr := flag.Int64("duration", 2, "duration in second")
	flag.Parse()
	hostAddr := *hostAddrPtr
	users := *usersPtr
	topics := *topicsPtr
	per := *perPtr
	duration := *durationPtr
	go func() {
		for i := 0; i < users; i++ {
			<-time.After(time.Millisecond * 10)
			cli := NewClient(hostAddr, prefix+"|user:"+fmt.Sprintln(i), "robot")
			go func() {
				for {
					for {
						if per > cli.TopicCount() {
							cli.Subscribe("room:" + fmt.Sprintln(rand.Intn(topics)))
						} else {
							break
						}
					}
					topicID := cli.RandomTopic()
					cli.Publish(topicID, fmt.Sprintln(time.Now().Unix()))
					<-time.After(time.Duration(int64(time.Second) * duration))
				}
			}()
			go func() {
				for {
					cli.Ping()
					<-time.After(time.Second * 30)
				}
			}()
		}
	}()
	http.ListenAndServe(":9529", nil)
}
