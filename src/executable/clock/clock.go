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

	go func() {
		for range c {
			fmt.Printf("\n%s\n", color.RedString(time.Now().Format("2006-01-02 15:04:05.999")))
			os.Exit(0)
		}
	}()

	fmt.Println(color.GreenString(time.Now().Format("2006-01-02 15:04:05.999")))
	for {
		time.Sleep(1 * time.Millisecond)
		fmt.Printf("\r%s", time.Now().Format("2006-01-02 15:04:05.999"))
	}
}
