package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
)

type Client struct {
	net.Conn
}
func (cli *Client) WriteMsg(msg *Msg) (n int, err error) {
	if len(msg.Data) > 0 {
		for i := int64(0); i < int64(len(msg.Data)); i += bufSize {
			var data []byte

			if i > int64(len(msg.Data)) - bufSize {
				data = msg.Data[i:]
			} else {
				data = msg.Data[i:i+bufSize]
			}

			written, writeErr := cli.Write(NewMsg(msg.Op, data).Bytes())
			n += written
			if writeErr != nil {
				return n, writeErr
			}
		}
		//cli.Write(NewMsgDF("%d", n).Bytes())
		return n, nil
	}
	//cli.Write(NewMsgDF("%d", len(msg.Bytes())).Bytes())
	return cli.Write(msg.Bytes())
}

func spawnClient() {
	cli, err := net.Dial("tcp", "127.0.0.1:4488")
	if err != nil {
		fmt.Println("sudo: error: unable to connect to server")
		return
	}
	defer cli.Close()
	client := &Client{cli}

	client.WriteMsg(NewMsgD("client: Creating cmd..."))
	cmd := exec.Command(os.Args[argIndex])
	if len(os.Args) > argIndex+1 {
		cmd = exec.Command(os.Args[argIndex], os.Args[argIndex+1:]...)
	}

	cmd.Dir, err = os.Getwd()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdDir: %v", err))
	}

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdIn: %v", err))
	}
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdOut: %v", err))
	}
	cmdErr, err := cmd.StderrPipe()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdErr: %v", err))
	}
	go cmdReadOutput(cmdOut, client)
	go cmdReadError(cmdErr, client)
	go cmdReadInput(cmdIn, client)

	client.WriteMsg(NewMsgD("client: Spawning target process..."))
	err = cmd.Start()
	if err != nil {
		client.WriteMsg(NewMsgEF("unable to spawn target process"))
		return
	}

	client.WriteMsg(NewMsgD("client: Waiting for process to exit..."))
	err = cmd.Wait()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmd: %v", err))
	}

	client.WriteMsg(NewMsgD("client: Sending EOF to server..."))
	client.WriteMsg(msgEOF)
}

func cmdReadOutput(cmdOut io.ReadCloser, client *Client) {
	defer cmdOut.Close()
	for {
		buf := make([]byte, bufSize)

		n, err := cmdOut.Read(buf)
		if err != nil {
			if err == io.EOF {
				client.WriteMsg(msgEOF)
			} else {
				client.WriteMsg(NewMsgEF("cmdReadOutput: %v", err))
			}
			return
		}

		if n > 0 {
			msg := NewMsg(opStdout, buf[:n])
			client.WriteMsg(msg)
		}
	}
}
func cmdReadError(cmdErr io.ReadCloser, client *Client) {
	defer cmdErr.Close()
	for {
		buf := make([]byte, bufSize)

		n, err := cmdErr.Read(buf)
		if err != nil {
			if err == io.EOF {
				client.WriteMsg(msgEOF)
			} else {
				client.WriteMsg(NewMsgEF("cmdReadError: %v", err))
			}
			return
		}

		if n > 0 {
			msg := NewMsg(opStderr, buf[:n])
			client.WriteMsg(msg)
		}
	}
}
func cmdReadInput(cmdIn io.WriteCloser, client *Client) {
	defer cmdIn.Close()
	for {
		buf := make([]byte, bufSize)

		n, err := client.Read(buf)
		if err != nil {
			client.WriteMsg(NewMsgEF("cmdReadInput: %v", err))
			return
		}

		if n > 0 {
			msgs, err := NewMsgIn(buf[:n])
			if err != nil {
				client.WriteMsg(NewMsgEF("NewMsgIn: %v", err))
				continue
			}

			for i := 0; i < len(msgs); i++ {
				msg := msgs[i]
				switch msg.Op {
				case opEOF:
					client.WriteMsg(NewMsgD("client: Goodbye!"))
					return
				case opStdin:
					client.WriteMsg(NewMsgD("client: Writing stdin..."))
					cmdIn.Write(msg.Data)
				default:
					client.WriteMsg(NewMsgEF("cmdReadInput: unknown op: %d", msg.Op))
				}
			}
		}
	}
}