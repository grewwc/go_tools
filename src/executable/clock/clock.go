package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	start := time.Now()
	go func(start time.Time) {
		for range c {
			stop := time.Now()
			utilsW.Printf("\n%s\n", color.HiRedString(stop.Format("2006-01-02 15:04:05.999")))
			utilsW.Printf("elapsed: %s\n", stop.Sub(start))
			os.Exit(0)
		}
	}(start)
	utilsW.Println(color.GreenString(start.Format("2006-01-02 15:04:05.999")))
	for {
		time.Sleep(1 * time.Millisecond)
		utilsW.Printf("\r%s", time.Now().Format("2006-01-02 15:04:05.999"))
	}
}
