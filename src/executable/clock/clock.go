package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	start := time.Now()
	go func(start time.Time) {
		for range c {
			stop := time.Now()
			fmt.Printf("\n%s\n", color.HiRedString(stop.Format("2006-01-02 15:04:05.999")))
			fmt.Printf("elapsed: %s\n", stop.Sub(start))
			os.Exit(0)
		}
	}(start)
	fmt.Println(color.GreenString(start.Format("2006-01-02 15:04:05.999")))
	for {
		time.Sleep(1 * time.Millisecond)
		fmt.Printf("\r%s", time.Now().Format("2006-01-02 15:04:05.999"))
	}
}
