//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/utilsW"
	"github.com/shirou/gopsutil/net"
)

func selectEn0(stats []net.IOCountersStat, name string) *net.IOCountersStat {
	for _, stat := range stats {
		if stat.Name == name {
			return &stat
		}
	}
	return nil
}

func changeFormat(speed float64) string {
	unit := "k/s"
	if speed > 1024 {
		speed /= 1024
		unit = "m/s"
	}
	return fmt.Sprintf("%.2f %s", speed, unit)
}

func detectInterface() string {
	if utilsW.GetPlatform() == utilsW.LINUX {
		return "eth0"
	}
	return "en0"
}

func diff(prev, now *info) (float64, float64) {
	sentDiff := float64((now.sent - prev.sent))
	downloadDiff := float64(now.recv - prev.recv)

	return sentDiff, downloadDiff
}

type info struct {
	sent, recv float64
	tp         int64
}

func showSpeed(ch chan<- *info) {
	interval := 1
	stats, err := net.IOCounters(true)
	if err != nil {
		panic(err)
	}
	interface_ := detectInterface()
	for {
		stats, _ = net.IOCounters(true)
		currStat := selectEn0(stats, interface_)
		ch <- &info{sent: float64(currStat.BytesSent), recv: float64(currStat.BytesRecv), tp: time.Now().UnixMilli()}
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func main() {
	ch := make(chan *info)
	go showSpeed(ch)
	var prev *info
	var sent, download float64
	for curr := range ch {
		if prev == nil {
			prev = curr
			fmt.Printf("    Ul: %s    Dl: %s%s\r", changeFormat(sent), changeFormat(float64(download)), strings.Repeat(" ", 20))
			continue
		}
		sent, download = diff(prev, curr)
		dt := float64(curr.tp-prev.tp) / 1e3
		sent, download = sent/1024/float64(dt), download/1024/float64(dt)
		fmt.Printf("    Ul: %s    Dl: %s%s\r", changeFormat(sent), changeFormat(download), strings.Repeat(" ", 20))
		prev = curr
	}
}
