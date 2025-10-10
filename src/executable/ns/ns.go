//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func changeFormat(speed float64) string {
	unit := "k/s"
	if speed > 1024 {
		speed /= 1024
		unit = "m/s"
	}
	return fmt.Sprintf("%.2f %s", speed, unit)
}

func detectInterface() string {
	if utilsw.GetPlatform() == utilsw.LINUX {
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
type netstat struct {
	sent, recv int64
	name       string
}

func selectEth(stats []netstat) []netstat {
	res := make([]netstat, 0)
	for _, stat := range stats {
		if strw.AnyHasPrefix(stat.name, "en", "eth") {
			res = append(res, stat)
		}
	}
	return res
}

func parseNetstat(line string, i, o int) netstat {
	parts := strw.SplitByCutset(line, " \n\t")
	var stat netstat
	stat.name = parts[0]
	sent, err := strconv.ParseInt(parts[o], 10, 64)
	if err != nil {
		panic(err)
	}
	recv, err := strconv.ParseInt(parts[i], 10, 64)
	if err != nil {
		panic(err)
	}
	stat.recv = recv
	stat.sent = sent
	return stat
}

func getInputOutputIndex(header string) (int, int) {
	inputKey, outputKey := "Ibytes", "Obytes"
	iIndex, oIndex := -1, -1
	headers := strw.SplitByCutset(header, " \t\n")
	for i, header := range headers {
		switch header {
		case inputKey:
			iIndex = i
		case outputKey:
			oIndex = i
		}
	}
	return iIndex, oIndex
}

func getStats() []netstat {
	res, _ := utilsw.RunCmd("netstat -ibdnW", nil)
	lines := strw.SplitNoEmpty(res, "\n")
	header := lines[0]

	i, o := getInputOutputIndex(header)
	var stats []netstat
	for _, line := range lines[1:] {
		stats = append(stats, parseNetstat(line, i, o))
	}
	stats = selectEth(stats)
	return stats
}

func showSpeed(ch chan<- *info) {
	interval := 500
	for {
		stats := getStats()
		var sent, recv float64
		for _, stat := range stats {
			sent += float64(stat.sent)
			recv += float64(stat.recv)
		}
		ch <- &info{sent: sent, recv: recv, tp: time.Now().UnixMilli()}
		time.Sleep(time.Millisecond * time.Duration(interval))
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
			fmt.Print("Waiting...\r")
			fmt.Printf("    Ul: %s    Dl: %s%s\r", changeFormat(sent), changeFormat(float64(download)), strings.Repeat(" ", 20))
			continue
		}
		sent, download = diff(prev, curr)
		dt := float64(curr.tp-prev.tp) / 1e3
		sent, download = sent/1024/float64(dt), download/1024/float64(dt)
		fmt.Printf("    Ul: %s    Dl: %s%s\r", changeFormat(sent), changeFormat(download), strings.Repeat(" ", 20))
		prev = curr
	}
	close(ch)
}
