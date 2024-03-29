package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

var (
	msgEOF *Msg = &Msg{Op: opEOF}

	argIndex = 0
	port int64 = 0
	bufSize int64 = 65536
	buffer int64 = 65536
	version bool = false
	check bool = false
	debug bool = false
	isClient bool = false

	logFile string
	logger *os.File
)

func init() {
	flag.Int64Var(&buffer, "buffer", buffer, "Set max size of all byte buffers")
	flag.BoolVar(&version, "version", version, "Set to display version info and exit")
	flag.BoolVar(&check, "check", check, "Set to check for admin privileges and exit")
	flag.BoolVar(&debug, "debug", debug, "Set to enable debug logging")
	flag.BoolVar(&isClient, "client", isClient, "Set to act as client")
	flag.Int64Var(&port, "port", port, "Custom TCP port for session (0 for randomization)")
	flag.StringVar(&logFile, "log", logFile, "File to log packets to when debugging")
	flag.Parse()

	bufSize = buffer - int64(len(msgEOF.Bytes())) //Ensure the size of a msg is accounted for
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("sudo: error: no arguments")
		return
	}

	if port <= 0 {
		port = rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(65536)
	}

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i][0] == '-' {
			continue
		}
		if os.Args[i] == fmt.Sprintf("%d", buffer) {
			continue
		}
		if os.Args[i] == fmt.Sprintf("%d", port) {
			continue
		}
		if os.Args[i] == fmt.Sprintf("%s", logFile) {
			continue
		}
		argIndex = i
		break
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
		fmt.Println("port:", port)
		fmt.Println("log:", logFile)
		return
	}

	if argIndex == 0 {
		fmt.Println("sudo: error: must specify program")
		return
	}

	_, err := exec.LookPath(os.Args[argIndex])
	if err != nil {
		fmt.Println("sudo: error: unable to find", os.Args[argIndex], "in PATH")
		return
	}

	if isClient {
		if debug {
			logger, err = os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("sudo: error: unable to open log file:", err)
				return
			}
		}
		spawnClient()
		return
	}

	//fmt.Println(os.Args[argIndex:])
	spawnServer()
}

func isAdmin() bool {
	_, err := os.Open("C:\\Program Files\\WindowsApps")
	if err != nil {
		return false
	}
	return true
}
