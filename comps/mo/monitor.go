package mo

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"runtime/pprof"

	n "github.com/shirou/gopsutil/net"
)

//Snapshot Snapshot
type Snapshot struct {
	CPUPercent        float64
	MemPercent        float64
	AverageNetIn      uint64
	AverageNetOut     uint64
	Thread            uint64
	Goroutine         uint64
	TotalConn         uint64
	TotalTopic        uint64
	TotalConsumer     uint64
	MaxConsumer       uint64
	TotalMessageIn    uint64
	AverageMessageIn  uint64
	TotalMessageOut   uint64
	AverageMessageOut uint64
}

//DataSource DataSource
type DataSource interface {
	GetTotalConsumer() map[string]uint64
	GetTotalConn() uint64
	GetTotalTopic() uint64
	GetMessageIn() uint64
	GetMessageOut() uint64
}

//Monitor Monitor
type Monitor struct {
	dataSource DataSource
}

//NewMonitor NewMonitor
func NewMonitor(dataSource DataSource) *Monitor {
	return &Monitor{dataSource: dataSource}
}

//Monitor Monitor
func (mo *Monitor) Monitor(duration time.Duration) *Snapshot {
	snap := Snapshot{}
	netIn1, netOut1 := mo.netMo()
	messageIn := mo.dataSource.GetMessageIn()
	messageOut := mo.dataSource.GetMessageOut()
	snap.CPUPercent = mo.cpuMo(duration)
	netIn2, netOut2 := mo.netMo()
	snap.MemPercent = mo.memMo()
	snap.AverageNetIn = uint64(float64(netIn2-netIn1) / duration.Seconds())
	snap.AverageNetOut = uint64(float64(netOut2-netOut1) / duration.Seconds())
	snap.TotalConn = mo.dataSource.GetTotalConn()
	snap.MaxConsumer, snap.TotalConsumer = mo.consumerMo()
	snap.TotalTopic = mo.dataSource.GetTotalTopic()
	snap.TotalMessageIn = mo.dataSource.GetMessageIn()
	snap.AverageMessageIn = uint64(float64(snap.TotalMessageIn-messageIn) / duration.Seconds())
	snap.TotalMessageOut = mo.dataSource.GetMessageOut()
	snap.AverageMessageOut = uint64(float64(snap.TotalMessageOut-messageOut) / duration.Seconds())
	snap.Goroutine = uint64(runtime.NumGoroutine())
	snap.Thread = uint64(pprof.Lookup("threadcreate").Count())
	return &snap
}

func (mo *Monitor) cpuMo(duration time.Duration) (percent float64) {
	v, err := cpu.Percent(duration, false)
	if nil == err && len(v) > 0 {
		return v[0]
	}
	return 0
	// return v[0]
	// <-time.After(duration)
	// return 0
}

func (mo *Monitor) consumerMo() (maximum uint64, total uint64) {
	c := mo.dataSource.GetTotalConsumer()
	for _, consumer := range c {
		total += consumer
		if consumer > maximum {
			maximum = consumer
		}
	}
	return maximum, total
}

func (mo *Monitor) netMo() (in uint64, out uint64) {
	c, err := n.IOCounters(false)
	if nil == err && len(c) > 0 {
		return c[0].BytesRecv, c[0].BytesSent
	}
	return 0, 0
}

func (mo *Monitor) memMo() (usedPercent float64) {
	v, _ := mem.VirtualMemory()
	return v.UsedPercent
}
