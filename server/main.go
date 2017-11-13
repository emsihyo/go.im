package main

import (
	_ "expvar"
	"flag"
	"log"
	"net"
	"net/http"

	_ "net/http/pprof"

	"github.com/emsihyo/go.im/comps/mo"
	"github.com/emsihyo/go.im/comps/st"

	"time"

	"github.com/emsihyo/go.im/comps/bi"
	"github.com/emsihyo/go.im/comps/se"
	"github.com/gorilla/websocket"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	// f, _ := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY, 0666)
	// log.SetOutput(f)
	// dbAddrPtr := flag.String("db", "127.0.0.1", "db addr")
	tcpPortPtr := flag.String("tcp", "10001", "tcp port,default 10001")
	httpPortPtr := flag.String("http", "10002", "http port,default 10002")
	httpNamespacePtr := flag.String("namespace", "chat", "http namespace,default chat")
	flag.Parse()
	// serv := NewServer(st.NewMongoStorage(*dbAddrPtr))
	serv := NewServer(st.NewStorageImpl())
	go handleTCP(serv, ":"+*tcpPortPtr)
	go handleHTTP(serv, ":"+*httpPortPtr, *httpNamespacePtr)
	monitor := mo.NewMonitor(serv)
	at := time.Now().UnixNano()
	for {
		snap := monitor.Monitor(time.Second * 5)
		duration := float64(time.Now().UnixNano()-at) / float64(time.Second)
		log.Printf(`
CPU:                 %.2f%%
MEM:                 %.2f%%
THREAD:              %d
GOROUTINE:           %d
NET_IN:              %.2fMB/s
NET_OUT:             %.2fMB/s
CONNECTION:          %d
TOPIC:               %d
CONSUMER_TOTAL:      %d
CONSUMER_MAXIMUM:    %d
MESSAGE_IN_TOTAL:    %d
MESSAGE_IN_TOTAL_S:  %d
MESSAGE_IN_S:        %d
MESSAGE_OUT_TOTAL:   %d
MESSAGE_OUT_TOTAL_S: %d
MESSAGE_OUT_S:       %d
			`, snap.CPUPercent, snap.MemPercent, snap.Thread, snap.Goroutine, float64(snap.AverageNetIn)/1024.0/1024.0, float64(snap.AverageNetOut)/1024.0/1024.0, snap.TotalConn, snap.TotalTopic, snap.TotalConsumer, snap.MaxConsumer, snap.TotalMessageIn, int64(float64(snap.TotalMessageIn)/duration), snap.AverageMessageIn, snap.TotalMessageOut, int64(float64(snap.TotalMessageOut)/duration), snap.AverageMessageOut)
	}
}

func handleTCP(serv *Server, port string) {
	pp := bi.ProtobufProtocol{}
	var err error
	var a *net.TCPAddr
	var l *net.TCPListener
	if a, err = net.ResolveTCPAddr("tcp", port); nil == err {
		if l, err = net.ListenTCP("tcp", a); nil == err {
			log.Println("[LISTEN]", "TCP:", port)
			for {
				c, err := l.AcceptTCP()
				if nil != err {
					continue
				}
				go func() {
					conn := bi.NewTCPConn(c)
					sess := se.NewSession(conn, &pp, time.Minute)
					serv.Handle(sess)
				}()
			}
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Fatalln(err)
	}
}

func handleHTTP(serv *Server, port string, namespace string) {
	pj := bi.JSONProtocol{}
	up := websocket.Upgrader{}
	f := func(w http.ResponseWriter, r *http.Request) {
		var err error
		var c *websocket.Conn
		if c, err = up.Upgrade(w, r, nil); nil == err {
			defer c.Close()
			conn := bi.NewWebsocketConn(c)
			sess := se.NewSession(conn, &pj, time.Minute)
			serv.Handle(sess)
		} else {
			log.Print("[WEBSOCKET]:", "upgrade:", err)
		}
	}
	http.HandleFunc("/"+namespace, f)
	log.Println("[LISTEN]", "HTTP:", port)
	http.ListenAndServe(port, nil)
}
