package stor

import (
	"dadian/golog"
	"libfwapp_go/fwapp/fwsdef"
)

// 日志模块
var gLog *golog.Logger = golog.New("Storage")

// 将事件数据写入本地存储，封装接口
type StorageRWer interface {
	// 将事件信息写入存储
	// edi-事件数据
	// 返回：成功返回nil，失败返回相应的错误信息
	Write(edi *fwsdef.EDItem) error

	// 负责从存储中读取待发送的事件数据
	// n表示最大读取个数
	// 返回：成功返回读取的事件数据，map的key值表示本服务接收到事件的时间（unix纳秒数）
	// Read(n int) (map[int64]*fwsdef.EDItem, error)

	// 删除事件数据及相关记录
	// ns-事件收到时的纳秒数，EDIten.Time
	Remove(ns int64) error

	// 开启管理服务，进行空间管理
	StartManager() error
	// 关闭管理服务
	StopManager()

	// 读取当前时间下，存储下的所有事件文件，后续用Next获取下一个事件数据直到返回io.EOF
	// 再次调用ReadAll，会重新获取最新文件列表
	// 返回：成功返回nil，没有文件返回io.EOF，失败返回错误类型
	ReadAll() (int, error)
	// 下一个事件文件数据信息，没有可读取的数据时返回nil,io.EOF，失败返回nil,错误信息
	Next() (*fwsdef.EDItem, error)
}
