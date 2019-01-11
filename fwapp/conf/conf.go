package fwsconf

import (
	"dadian/golog"
	"errors"
	"path"
	"strings"
)

type fwsConfIntf interface {
	/*
	 * 初始化
	 */
	InitConf(szFile string) error

	/*
	 * 网络信息
	 */
	// 获取服务端口
	GetListenPort() int

	/*
	 * 日志类信息
	 */
	// 日志信息是否终端打印，默认true
	IsLogToStdout() bool

	// 日志信息是否保存至日志文件，默认false
	IsLogToFile() bool

	// 获取日志文件路径（不含文件名称），
	// 默认值：
	//      windows: "c:\\fwserver\\log\\""
	//      linux ： "/tmp/fwserver/log/"
	GetLogFilePath() string

	// 获取日志文件最大保存天数，默认15天
	GetLogFileKeepdays() int

	/*
	 * 发送失败的情况下，图片本地保存策略
	 */
	// 发送失败后，超过一定数量后，图片写入磁盘。获取内存最大保存图片数量，默认800
	GetFPicMaxInCache() int

	// 获取图片存储路径，默认：
	// windows: "/tmp/fwserver/pic/"
	// linux: "c:\\fwserver\\pic\\"
	GetFPicStorPath() string

	// 保存的图片占用当前磁盘的最大MB数，默认4096MB(4GB)
	GetFPicStorMaxMB() int
}

const (
	CONFIG_FILE = "fwserver-config.json"
)

var gFwsConf fwsConfIntf = nil
var gLog *golog.Logger = golog.New("Config")

var NOSUP error = errors.New("unsupport extension of configuration file ")

func InitConf(szFile string) error {
	// 根据文件名判断使用哪种配置驱动
	// .json使用JSON解析
	// .ini 使用INI解析
	szExt := path.Ext(szFile)
	szExt = strings.ToLower(szExt)
	switch szExt {
	case ".json":
		gFwsConf = &JsonConf{}
		break
	case ".ini":
		fallthrough
	case ".xml":
		fallthrough
	default:
		return NOSUP
	}
	return gFwsConf.InitConf(szFile)
}

/*
 * 网络信息
 */
// 获取服务端口
func GetListenPort() int {
	return gFwsConf.GetListenPort()
}

/*
 * 日志类信息
 */
// 日志信息是否终端打印，默认true
func IsLogToStdout() bool {
	if gFwsConf == nil {

	}
	return gFwsConf.IsLogToStdout()
}

// 日志信息是否保存至日志文件，默认false
func IsLogToFile() bool {
	if gFwsConf == nil {

	}
	return gFwsConf.IsLogToFile()
}

// 获取日志文件路径（不含文件名称），
func GetLogFilePath() string {
	if gFwsConf == nil {

	}
	return gFwsConf.GetLogFilePath()
}

// 获取日志文件最大保存天数
func GetLogFileKeepdays() int {
	if gFwsConf == nil {

	}
	return gFwsConf.GetLogFileKeepdays()
}

/*
 * 发送失败的情况下，图片本地保存策略
 */
// 发送失败后，超过一定数量后，图片写入磁盘。获取内存最大保存图片数量
func GetFPicMaxInCache() int {
	if gFwsConf == nil {

	}
	return gFwsConf.GetFPicMaxInCache()
}

// 获取图片存储路径，默认：
func GetFPicStorPath() string {
	if gFwsConf == nil {

	}
	return gFwsConf.GetFPicStorPath()
}

// 保存的图片占用当前磁盘的最大空间值（单位MB）
func GetFPicStorMaxMB() int {
	if gFwsConf == nil {

	}
	return gFwsConf.GetFPicStorMaxMB()
}
