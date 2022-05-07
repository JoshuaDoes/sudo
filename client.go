package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
)

func spawnClient() {
	cli, err := net.Dial("tcp", "127.0.0.1:4488")
	if err != nil {
		fmt.Println("sudo: error: unable to connect to server")
		return
	}
	defer cli.Close()

	cli.Write(NewMsgD("client: Creating cmd...").Bytes())
	cmd := exec.Command(os.Args[argIndex])
	if len(os.Args) > argIndex+1 {
		cmd = exec.Command(os.Args[argIndex], os.Args[argIndex+1:]...)
	}

	cmd.Dir, err = os.Getwd()
	if err != nil {
		cli.Write(NewMsgEF("cmdDir: %v", err).Bytes())
	}

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		cli.Write(NewMsgEF("cmdIn: %v", err).Bytes())
	}
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		cli.Write(NewMsgEF("cmdOut: %v", err).Bytes())
	}
	cmdErr, err := cmd.StderrPipe()
	if err != nil {
		cli.Write(NewMsgEF("cmdErr: %v", err).Bytes())
	}
	go cmdReadOutput(cmdOut, cli)
	go cmdReadError(cmdErr, cli)
	go cmdReadInput(cmdIn, cli)

	cli.Write(NewMsgD("client: Spawning target process...").Bytes())
	err = cmd.Start()
	if err != nil {
		cli.Write(NewMsgEF("unable to spawn target process").Bytes())
		return
	}

	cli.Write(NewMsgD("client: Waiting for process to exit...").Bytes())
	err = cmd.Wait()
	if err != nil {
		cli.Write(NewMsgEF("cmd: %v", err).Bytes())
	}

	cli.Write(NewMsgD("client: Sending EOF to server...").Bytes())
	cli.Write(msgEOF.Bytes())
}

func cmdReadOutput(cmdOut io.ReadCloser, cli net.Conn) {
	defer cmdOut.Close()
	for {
		buf := make([]byte, bufSize)

		n, err := cmdOut.Read(buf)
		if err != nil {
			if err == io.EOF {
				cli.Write(msgEOF.Bytes())
			} else {
				cli.Write(NewMsgEF("cmdReadOutput: %v", err).Bytes())
			}
			return
		}

		if n > 0 {
			msg := NewMsg(opStdout, buf[:n])
			cli.Write(msg.Bytes())
		}
	}
}
func cmdReadError(cmdErr io.ReadCloser, cli net.Conn) {
	defer cmdErr.Close()
	for {
		buf := make([]byte, bufSize)

		n, err := cmdErr.Read(buf)
		if err != nil {
			if err == io.EOF {
				cli.Write(msgEOF.Bytes())
			} else {
				cli.Write(NewMsgEF("cmdReadError: %v", err).Bytes())
			}
			return
		}

		if n > 0 {
			msg := NewMsg(opStderr, buf[:n])
			cli.Write(msg.Bytes())
		}
	}
}
func cmdReadInput(cmdIn io.WriteCloser, cli net.Conn) {
	defer cmdIn.Close()
	for {
		buf := make([]byte, bufSize)

		n, err := cli.Read(buf)
		if err != nil {
			cli.Write(NewMsgEF("cmdReadInput: %v", err).Bytes())
			return
		}

		if n > 0 {
			msg, err := NewMsgIn(buf[:n])
			if err != nil {
				cli.Write(NewMsgEF("NewMsgIn: %v", err).Bytes())
				continue
			}

			switch msg.Op {
			case opEOF:
				cli.Write(NewMsgD("client: Goodbye!").Bytes())
				return
			case opStdin:
				cli.Write(NewMsgD("client: Writing stdin...").Bytes())
				cmdIn.Write(msg.Data)
			default:
				cli.Write(NewMsgEF("cmdReadInput: unknown op: %d", msg.Op).Bytes())
			}
		}
	}
}