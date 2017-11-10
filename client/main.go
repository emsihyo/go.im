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
	prefix = uuid.NewV5(uuid.NewV4(), prefix).String()
	portAddrPtr := flag.String("port", "10000", "http port")
	hostAddrPtr := flag.String("host", ":10001", "tcp addr")
	usersPtr := flag.Int("users", 5000, "user count")
	topicsPtr := flag.Int("topics", 100, "topic count")
	perPtr := flag.Int("per", 20, "topics per user")
	durationPtr := flag.Int64("duration", 20, "duration in second")
	flag.Parse()
	ch := make(chan struct{}, 4)
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
					<-time.After(time.Millisecond * 5)
				}
				<-ch
				go func() {
					for {
						topicID := cli.RandomTopic()
						cli.Publish(topicID, fmt.Sprintln(time.Now().Unix()))
						<-time.After(time.Duration(int64(time.Second) * *durationPtr))
					}
				}()
			}()
			go func() {
				for {
					cli.Ping()
					<-time.After(time.Second * 30)
				}
			}()
		}
	}()
	log.Print(http.ListenAndServe(":"+*portAddrPtr, nil))
}
