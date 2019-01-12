package svrs

import (
	"dadian/golog"
	"libfwapp_go/fwapp/fwsdef"
)

// 日志模块
var gLog *golog.Logger = golog.New("Servers")

//
// ServerWriter 接口包装了向其他服务发送事件数据的方法，用于扩展多种协议。

type ServerWriter interface {
	// 将事件数据写入目标
	// data ：待发送的事件检测数据
	// 返回： 成功返回nil，其他返回错误信息
	Write(data *fwsdef.EventDataT) error
}
