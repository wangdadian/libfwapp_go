package fwsdef

import (
	"fmt"
)

const MAX_BUFF_SIZE = 1024

// 以5个连续"\r\n"结尾的消息体
var MSG_END []byte = []byte("\r\n\r\n\r\n\r\n\r\n")

//消息类型
const (
	_             int = iota
	REQ_HEARTBEAT     = 0xF001 // 心跳发送
	RET_HEARTBEAT     = 0xF002 // 心跳返回
	REQ_EVENT         = 0xF003 // 检测事件数据发送
	RET_EVENT         = 0xF004 // 检测事件数据返回
)

// URL最大长度
const FWS_URL_MAXSIZE = 256

// json消息以5个连续“\r\n”结尾，数据除外
const FWS_MSG_VERSION = 0x1F010001

// 消息头
type FwsHeaderT struct {
	VER  int `json:"version"`
	MSG  int `json:"message"`
	WFID int `json:"workflowid"`
}

func (self *FwsHeaderT) String() string {
	return fmt.Sprintf("version: 0x%X, message: 0x%X, workflowid: %d", self.VER, self.MSG, self.WFID)
}

// 消息头，用于解析通用json消息体中的header
type FwHeaderExT struct {
	Header FwsHeaderT `json:"header"`
}

// 通用返回消息体
type FwRetT struct {
	V int    `json:"retvalue"`
	S string `json:"retinfo"`
}
type FwMsgRetT struct {
	Header FwsHeaderT `json:"header"`
	Ret    FwRetT     `json:"ret"`
}

// 事件消息体
type FwEventT struct {
	DescLen int `json:"desc_len"`
	PicLen  int `json:"pic_len"`
	UrlLen  int `json:"url_len"`
}
type FwEventMsgT struct {
	Header FwsHeaderT `json:"header"`
	Event  FwEventT   `json:"event"`
}

// 心跳间隔以及最大超时个数
const (
	INTERVAL_HEARTBEAT = 5 // 心跳间隔
	MAX_HB_TIMEOUT     = 3 // 心跳超时个数
)
