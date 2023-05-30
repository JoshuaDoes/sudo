package main

import (
	"fmt"
	"strings"

	"github.com/JoshuaDoes/json"
)

var (
	jsonBuf []byte
)

func init() {
	jsonBuf = make([]byte, 0)
}

type OpCmd int
const (
	opError OpCmd = -1
	opDebug OpCmd = 255

	opEOF OpCmd = 0
	opStdout OpCmd = 1
	opStderr OpCmd = 2
	opStdin OpCmd = 3
)

type Msg struct {
	Op OpCmd `json:"c"`
	Data []byte `json:"d"`
}
func (msg *Msg) Bytes() []byte {
	data, err := json.Marshal(msg, false)
	if err != nil {
		panic(err)
	}
	data = append(data, byte('\x00'))
	return data
}
func NewMsgIn(data []byte) (msgs []*Msg, err error) {
	msgs = make([]*Msg, 0)

	msgsSplit := strings.Split(string(data), "\x00")
	for i := 0; i < len(msgsSplit); i++ {
		if msgsSplit[i] == "" {
			continue
		}
		dataBuf := append(jsonBuf, []byte(msgsSplit[i])...)

		msg := &Msg{}
		err = json.Unmarshal(dataBuf, msg)
		if err != nil {
			if i == len(msgsSplit)-1 {
				jsonBuf = dataBuf
				err = nil //We're expecting more data still...
				continue
			}
			return
		}

		msgs = append(msgs, msg)
		jsonBuf = make([]byte, 0)
	}
	return
}

func NewMsg(op OpCmd, data []byte) *Msg {
	logger.WriteString(fmt.Sprintf("[%d] %s\n", op, data))
	return &Msg{Op: op, Data: data}
}
func NewMsgS(op OpCmd, data string) *Msg {
	return NewMsg(op, []byte(data))
}
func NewMsgSF(op OpCmd, format string, data ...interface{}) *Msg {
	return NewMsgS(op, fmt.Sprintf(format, data...))
}
func NewMsgE(err error) *Msg {
	return NewMsgS(opError, err.Error())
}
func NewMsgEF(format string, data ...interface{}) *Msg {
	return NewMsgS(opError, fmt.Sprintf(format, data...))
}
func NewMsgD(data string) *Msg {
	return NewMsgS(opDebug, data)
}
func NewMsgDF(format string, data ...interface{}) *Msg {
	return NewMsgS(opDebug, fmt.Sprintf(format, data...))
}