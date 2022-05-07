package main

import (
	"fmt"

	"github.com/JoshuaDoes/json"
)

type OpCmd int
const (
	opError OpCmd = -1
	opDebug OpCmd = 255

	opEOF OpCmd = iota //0
	opStdout           //1
	opStderr           //2
	opStdin            //3
)

type Msg struct {
	Op OpCmd `json:"op"`
	Data []byte `json:"data"`
}
func (msg *Msg) Bytes() []byte {
	data, err := json.Marshal(msg, false)
	if err != nil {
		panic(err)
	}
	return data
}
func NewMsgIn(data []byte) (*Msg, error) {
	msg := &Msg{}
	err := json.Unmarshal(data, msg)
	return msg, err
}

func NewMsg(op OpCmd, data []byte) *Msg {
	return &Msg{Op: op, Data: data}
}
func NewMsgS(op OpCmd, data string) *Msg {
	return &Msg{Op: op, Data: []byte(data)}
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