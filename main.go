package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	//"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"system-monitor/pkg/metrics"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

var (
	NodeName string
)

func init() {
	NodeName = os.Getenv("NODE_NAME")
	if NodeName == "" {
		NodeName = "unknown"
	}
}

type AppMem struct {
	Sys        float64 `json:"sys"`
	TotalAlloc float64 `json:"total_alloc"`
	Alloc      float64 `json:"alloc"`
	HeapSys    float64 `json:"heap_sys"`
	HeapAlloc  float64 `json:"heap_alloc"`
}

func GetAppMem() AppMem {
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
	GPU NodeGPUs  `json:"gpus"`
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

type NodeGPUs []NodeGPU
type NodeGPU struct {
	Used  float64
	Total float64
	Usage float64
}

/*
func getGPUStatus() NodeGPUs {
	result := make(NodeGPUs, 0)
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return result
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			return
		}
	}()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return result
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		name, _ := device.GetName()
		memInfo, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			continue
		}
		usedMB := float64(memInfo.Used) / MB
		totalMB := float64(memInfo.Total) / MB
		usagePercent := float64(memInfo.Used) / float64(memInfo.Total) * 100
		result = append(result, NodeGPU{
			Used:  usedMB,
			Total: totalMB,
			Usage: usagePercent,
		})
		fmt.Printf("GPU %d: %s\n", i, name)
		fmt.Printf("  VRAM Used: %f MB / %f MB (%.2f%%)\n", usedMB, totalMB, usagePercent)
	}
	return result
}
*/

func getNodeMem() NodeMem {
	v, _ := mem.VirtualMemory()
	metrics.Mem_usage.WithLabelValues(NodeName).Set(v.UsedPercent)
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
	metrics.Net_usage.WithLabelValues(NodeName, "send").Set(float64(sendPerSec))
	metrics.Net_usage.WithLabelValues(NodeName, "recv").Set(float64(recvPerSec))
	return NodeNet{
		SendByte: sendPerSec,
		RecByte:  recvPerSec,
	}
}

func getNodeCPUs() []float64 {
	percent, _ := cpu.Percent(0, false)
	metrics.CPU_usage.WithLabelValues(NodeName).Set(percent[0])
	return percent
}

func GetNodeUsage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	node := NodeUsage{
		CPU: getNodeCPUs(),
		Mem: getNodeMem(),
		Net: getNodeNet(time.Second),
		//GPU: getGPUStatus(),
	}
	jsonData, err := json.Marshal(node)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling JSON: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetSystemUsage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	app := GetAppMem()
	jsonData, err := json.Marshal(app)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling JSON: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Print("cpu")
			_ = getNodeCPUs()
		}
	}()
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Print("net")
			_ = getNodeNet(time.Second)
		}
	}()
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Print("mem")
			_ = getNodeMem()
		}
	}()
	router := httprouter.New()
	router.GET("/node", metrics.MetricsMiddleware(GetNodeUsage))
	router.GET("/system", metrics.MetricsMiddleware(GetSystemUsage))
	router.Handler("GET", "/metrics", promhttp.Handler())
	period := time.Duration(5) * time.Second
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
