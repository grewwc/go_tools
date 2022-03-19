package main

import "os/exec"

func main() {
	cmd := exec.Command("cmd.exe", "/c", "start", ".")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
