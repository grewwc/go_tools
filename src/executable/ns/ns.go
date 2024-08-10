package main

import (
	"log"
	"os"
	"strings"

	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	conf := utilsW.GetAllConfig()
	if conf == nil {
		log.Fatalln(conf)
		return
	}
	cmd := conf.GetOrDefault("utils.ns.cmd", nil)
	if cmd == nil {
		log.Fatalf("need to set utils.ns.cmd in ~/.configW")
		return
	}
	cmdStr := cmd.(string)
	cmdStr += " " + strings.Join(os.Args[1:], " ")

	if err := utilsW.RunCmd(cmdStr); err != nil {
		log.Fatalln(err)
	}
}
