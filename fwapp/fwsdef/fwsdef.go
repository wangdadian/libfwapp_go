package fwsdef

import (
	"bytes"
	"fmt"
)

// 检测事件数据
type EventDataT struct {
	DescBuf []byte
	PicBuf  []byte
}

// 从客户端发来的事件检测数据
type EventDataFromCT struct {
	ED   *EventDataT // 单次检测数据
	Urls []string    // 需要发送的目标URL列表
}

// 事件检测数据单元
type EDItem struct {
	// 单个检测事件数据
	Data *EventDataT
	// 事件接收到的时间，纳秒数
	Time int64
	// 需要写入事件数据的服务器列表，每成功发送一个至目标服务器，则删除。
	// 当为map长度为0时，说明事件数据全部发送完毕，删除此事件数据
	// 此处SvrsMap为map[int]svrs.ServerWriter类型，为了防止循环import，
	// 改为interface{}类型，每次使用SvrsMap需要进行类型断言
	SvrsMap interface{}
}

// 从字节流中获取url列表
// 判断方法读取非0值
func GetUrlsFromBytes(b []byte) ([]string, error) { // FWS_URL_MAXSIZE
	MIN_URL_LEN := len("http://0.0.0.0")
	iLen := len(b)
	if iLen <= MIN_URL_LEN {
		return nil, fmt.Errorf("invalid bytes buffer length.")
	}
	var urls []string = nil
	bb := bytes.Split(b, []byte{0})
	for _, bi := range bb {
		if len(bi) > MIN_URL_LEN {
			urls = append(urls, string(bi))
		}
	}
	return urls, nil
}
