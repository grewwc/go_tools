package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	conf := utilsW.GetAllConfig()
	if conf == nil {
		log.Fatalln(conf)
		return
	}
	cmd := conf.GetOrDefault("utils.oo.cmd", nil)
	if cmd == nil {
		log.Fatalf("need to set utils.oo.cmd in ~/.configW")
		return
	}
	cmdStr := cmd.(string)
	cmdStr += " " + strings.Join(os.Args[1:], " ")
	_, err := utilsW.RunCmdWithTimeout(cmdStr, time.Second*60)
	if err != nil {
		log.Fatalln(err)
		return
	}
}
