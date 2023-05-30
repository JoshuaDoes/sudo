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
	srv, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Println("sudo: error: unable to spawn server")
		return
	}
	defer srv.Close()

	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := "-client " + strings.Join(os.Args[1:], " ")
	if !strings.Contains(args, fmt.Sprintf("%d", port)) {
		args = fmt.Sprintf("-port %d %s", port, args)
	}

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
		select {
		case o, ok := <-output:
			if !ok {
				continue
			}

			msgs, err := NewMsgIn(o)
			if err != nil {
				fmt.Println("sudo: error: unable to process msgs:", err)
				fmt.Println(string(o))
			}
			for i := 0; i < len(msgs); i++ {
				msg := msgs[i]
				switch msg.Op {
				case opError:
					fmt.Printf("sudo: error: %s\n", string(msg.Data))
				case opEOF:
					return
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
			}
		case i, ok := <-input:
			if !ok {
				continue
			}

			i = NewMsg(opStdin, i).Bytes()
			conn.Write(i)
		case _, ok := <-sc:
			if ok {
				return
			}
		default:
			time.Sleep(time.Millisecond * 1)
		}
	}
}

func readInput(input chan []byte) {
	for {
		in := make([]byte, bufSize)

		n, err := os.Stdin.Read(in)
		if err != nil {
			fmt.Printf("readInput: %v\n", err)
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
		out := make([]byte, buffer)

		n, err := conn.Read(out)
		if err != nil {
			fmt.Printf("readOutput: %v\n", err)
			return
		}

		if n > 0 {
			out = out[:n]
			output <- out
		}
	}
}