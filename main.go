package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

type AppMem struct {
	Sys        float64 `json:"sys"`
	TotalAlloc float64 `json:"total_alloc"`
	Alloc      float64 `json:"alloc"`
	HeapSys    float64 `json:"heap_sys"`
	HeapAlloc  float64 `json:"heap_alloc"`
}

func getAppMem() AppMem {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return AppMem{
		Sys:        float64(m.Sys) / MB,
		TotalAlloc: float64(m.TotalAlloc) / MB,
		Alloc:      float64(m.Alloc) / MB,
		HeapSys:    float64(m.HeapSys) / MB,
		HeapAlloc:  float64(m.HeapAlloc) / MB,
	}
}

type NodeUsage struct {
	CPU []float64 `json:"cpu"`
	Mem NodeMem   `json:"memory"`
	Net NodeNet   `json:"net"`
}

type NodeMem struct {
	Total     float64 `json:"total"`
	Available float64 `json:"available"`
	Used      float64 `json:"used"`
	Percent   float64 `json:"percent"`
}

type NodeNet struct {
	SendByte uint64 `json:"send_bytes"`
	RecByte  uint64 `json:"recv_bytes"`
}

func getNodeMem() NodeMem {
	v, _ := mem.VirtualMemory()
	return NodeMem{
		Total:     float64(v.Total) / GB,
		Available: float64(v.Available) / GB,
		Used:      float64(v.Used) / GB,
		Percent:   v.UsedPercent,
	}
}

func getNodeNet(interval time.Duration) NodeNet {
	io1, _ := net.IOCounters(false)
	time.Sleep(interval)
	io2, _ := net.IOCounters(false)

	recvPerSec := (io2[0].BytesRecv - io1[0].BytesRecv) / uint64(interval.Seconds())
	sendPerSec := (io2[0].BytesSent - io1[0].BytesSent) / uint64(interval.Seconds())
	return NodeNet{
		SendByte: sendPerSec,
		RecByte:  recvPerSec,
	}
}

func getNodeCPUs() []float64 {
	usage, _ := cpu.Percent(0, false)
	return usage
}

func GetNodeUsage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	node := NodeUsage{
		CPU: getNodeCPUs(),
		Mem: getNodeMem(),
		Net: getNodeNet(time.Second),
	}
	jsonData, err := json.Marshal(node)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling JSON: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func main() {
	router := httprouter.New()
	router.GET("/node", GetNodeUsage)
	http.ListenAndServe(":8080", router)
}
