package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	msgEOF *Msg = &Msg{Op: opEOF}

	argIndex = 0
	bufSize int64 = 65536
	buffer int64 = 65536
	version bool = false
	check bool = false
	debug bool = false
)

func init() {
	flag.Int64Var(&buffer, "buffer", buffer, "Set max size of all byte buffers")
	flag.BoolVar(&version, "version", version, "Set to display version info and exit")
	flag.BoolVar(&check, "check", check, "Set to check for admin privileges and exit")
	flag.BoolVar(&debug, "debug", debug, "Set to enable debug logging")
	flag.Parse()

	bufSize = buffer - int64(len(msgEOF.Bytes())) //Ensure the size of a msg is accounted for
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("sudo: error: no arguments")
		return
	}

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i][0] == '-' {
			continue
		}
		if os.Args[i] == fmt.Sprintf("%d", buffer) {
			continue
		}
		argIndex = i
		break
	}
	if argIndex == 0 {
		fmt.Println("sudo: error: must specify program")
	}

	if version {
		fmt.Println("sudo for windows v0")
		return
	}
	if check {
		fmt.Println("argIndex:", argIndex)
		fmt.Println("isAdmin:", isAdmin())
		fmt.Println("buffer:", buffer)
		fmt.Println("bufSize:", bufSize)
		fmt.Println("debug:", debug)
		return
	}

	_, err := exec.LookPath(os.Args[argIndex])
	if err != nil {
		fmt.Println("sudo: error: unable to find", os.Args[argIndex], "in PATH")
		return
	}

	if isAdmin() {
		spawnClient()
		return
	}

	spawnServer()
}

func isAdmin() bool {
	_, err := os.Open("C:\\Program Files\\WindowsApps")
	if err != nil {
		return false
	}
	return true
}