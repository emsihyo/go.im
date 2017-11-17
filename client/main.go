package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/emsihyo/go.im/comps/mo"
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
	hostAddrPtr := flag.String("host", "120.79.29.75:10001", "tcp addr")
	usersPtr := flag.Int("users", 10, "user count")
	topicsPtr := flag.Int("topics", 5000, "topic count")
	perPtr := flag.Int("per", 1, "topics per user")
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
					t := time.Duration(int64(time.Second) * *durationPtr)
					tm := time.NewTimer(t)
					for {
						tm.Reset(t)
						<-tm.C
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
	go func() {
		http.ListenAndServe(":"+*portAddrPtr, nil)
	}()
	monitor := mo.NewMonitor(nil)
	for {
		snap := monitor.Monitor(time.Second * 5)
		log.Printf(`
CPU:                 %.2f%%
MEM:                 %.2f%%
THREAD:              %d
GOROUTINE:           %d
NET_IN:              %.2fMB/s
NET_OUT:             %.2fMB/s
			`, snap.CPUPercent, snap.MemPercent, snap.Thread, snap.Goroutine, float64(snap.AverageNetIn)/1024.0/1024.0, float64(snap.AverageNetOut)/1024.0/1024.0)
	}
}
