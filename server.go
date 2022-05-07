package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

func spawnServer() {
	srv, err := net.Listen("tcp", "127.0.0.1:4488")
	if err != nil {
		fmt.Println("sudo: error: unable to spawn server")
		return
	}
	defer srv.Close()

	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 0 //0 = SW_HIDE, 1 = SW_NORMAL

	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println("sudo: error:", err)
		return
	}

	conn, err := srv.Accept()
	if err != nil {
		fmt.Println("sudo: error: unable to accept client")
		return
	}

	//Stdio buffers
	input := make(chan []byte, bufSize)
	output := make(chan []byte, bufSize)

	//Syscall notifications for terminating child processes
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	signal.Notify(sc, syscall.SIGKILL)

	go readInput(input)
	go readOutput(output, conn)

	for {
		finished := false
		select {
		case o, ok := <-output:
			if !ok {
				finished = true
				break
			}

			msg, err := NewMsgIn(o)
			if err != nil {
				fmt.Println("sudo: error: unable to process msg:", err)
			}
			switch msg.Op {
			case opError:
				finished = true
				fmt.Printf("sudo: error: %s\n", string(msg.Data))
			case opEOF:
				finished = true
				break
			case opStdout:
				os.Stdout.Write(msg.Data)
			case opStderr:
				os.Stderr.Write(msg.Data)
			case opDebug:
				if debug {
					fmt.Printf("sudo: debug: %s\n", string(msg.Data))
				}
			default:
				fmt.Printf("sudo: error: unknown op: %d\n", msg.Op)
			}
		case i, ok := <-input:
			if !ok {
				finished = true
				break
			}

			i = NewMsg(opStdin, i).Bytes()
			conn.Write(i)
		case _, ok := <-sc:
			if ok {
				//fmt.Println("TERMINATED!")
				finished = true
				break
			}
		default:
			time.Sleep(time.Millisecond * 1)
		}
		if finished {
			break
		}
	}
}

func readInput(input chan []byte) {
	for {
		in := make([]byte, bufSize)

		n, err := os.Stdin.Read(in)
		if err != nil {
			//fmt.Printf("readInput: %v\n", err)
			return
		}

		if n > 0 {
			in = in[:n]
			input <- in
		}
	}
}
func readOutput(output chan []byte, conn net.Conn) {
	for {
		out := make([]byte, bufSize)

		n, err := conn.Read(out)
		if err != nil {
			//fmt.Printf("readOutput: %v\n", err)
			return
		}

		if n > 0 {
			out = out[:n]
			output <- out
		}
	}
}