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
	for {
		snap := monitor.Monitor(time.Second * 5)
		log.Printf("\n\nCPU:%.2f%%\nMEM:%.2f%%\nNET:in:%.2fMB/s out:%.2fMB/s\nTHREAD:%d\nGO:%d\nCONN:%d\nTOPIC:%d\nCONSUMER:maximum:%d total:%d\nMSG_IN:avg:%d total:%d\nMSG_OUT:avg:%d total:%d\n\n", snap.CPUPercent, snap.MemPercent, float64(snap.AverageNetIn)/1024.0/1024.0, float64(snap.AverageMessageOut)/1024.0/1024.0, snap.Thread, snap.Goroutine, snap.TotalConn, snap.TotalTopic, snap.MaxConsumer, snap.TotalConsumer, snap.AverageMessageIn, snap.TotalMessageIn, snap.AverageMessageOut, snap.TotalMessageOut)
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
