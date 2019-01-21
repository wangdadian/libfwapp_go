package main

import "C"
import (
	"dadian/golog"
	"fmt"
	"libfwapp_go/fwapp/conf"
	"libfwapp_go/fwapp/eventdatamgr"
	"libfwapp_go/fwapp/fwsdef"
	"libfwapp_go/fwapp/logmgr"
	"os"
	"runtime"
	"time"
)

var gLog *golog.Logger = golog.New("FWAPP")

// 退出信号
var gchExit chan bool

func init() {
	iCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(iCpu)
	gchExit = make(chan bool)
}

// 日志文件每天一个，自动进行更替
func logJob() {
	// 如果不写入日志文件，则直接退出
	if ok := fwsconf.IsLogToFile(); ok == false {
		return
	}

	// 根据配置，如果需要记入日志文件的情况下
	// 判断路径是否存在，不存在则创建，创建失败直接panic
	szPath := fwsconf.GetLogFilePath()
	_, err := os.Stat(szPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(szPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	// 日志文件全路径
	tLast := time.Now()
	tNow := tLast
	szFile := szPath + "fwserver_log_" + tLast.Format("20060102_150405") + ".log"
	// 设置日志文件
	golog.SetLogFile(".", szFile)
	gLog.Infof("new log file[%s]", szFile)
	// 定时每天记录新文件
	for {
		select {
		case <-gchExit:
			goto goto_ret
		default:
		}
		tNow = time.Now()
		if tNow.Day() != tLast.Day() {
			tLast = tNow
			szFile = szPath + "fwserver_log_" + tLast.Format("20060102_150405") + ".log"
			golog.CloseLogFile(".")
			golog.SetLogFile(".", szFile)
			gLog.Infof("new log file[%s]", szFile)
		}

		time.Sleep(500 * time.Millisecond)
	}
goto_ret:
}

//
// ******************************************************************************
// *********************   不要修改函数前的export注释   ***************************
// ******************************************************************************
//

func main() {}

//export FWGO_Init
func FWGO_Init() int32 {
	err := fwsconf.InitConf("libfwapp_go-config.json")
	if err != nil {
		return -1
	}
	// 生效日志颜色打印，"."匹配所有日志对象
	golog.EnableColorLogger(".", true)
	// 根据配置文件设置是否终端打印
	golog.SetStdout(".", fwsconf.IsLogToStdout())
	// 设置log存储路径以及文件信息
	if fwsconf.IsLogToFile() {
		go logJob()
	} else {
		golog.CloseLogFile(".")
	}
	// 启动日志管理
	logmgr.Start()
	// 启动数据管理服务
	edmgr.Start()
	gLog.Infof("FW_GO_APP Init OK.")
	return 0
}

//export FWGO_Influx
func FWGO_Influx(byDesc []byte, uiDescLen uint32, byPic []byte, uiPicLen uint32, byUrls []byte, uiUrlLen uint32) int32 {
	gLog.Infof("####FWGO_Influx get event data:")
	gLog.Infof("    byDesc bytes length=%d, uiDescLen=%d\n", len(byDesc), uiDescLen)
	gLog.Infof("    byPic bytes length=%d, uiPicLen=%d\n", len(byPic), uiPicLen)
	gLog.Infof("    byUrls bytes length=%d, uiUrlLen=%d\n", len(byUrls), uiUrlLen)

	var err error

	//
	// 复制数据
	//

	// 描述信息
	var byDescNew []byte = nil
	if uiDescLen > 0 {
		byDescNew = make([]byte, uiDescLen)
		copy(byDescNew, byDesc)
	}

	// 图片数据
	var byPicNew []byte = nil
	if uiPicLen > 0 {
		byPicNew = make([]byte, uiPicLen)
		copy(byPicNew, byPic)
	}

	// url列表
	var byUrlsNew []byte = nil
	var urls []string = nil
	if uiUrlLen > 0 {
		byUrlsNew = make([]byte, uiUrlLen)
		copy(byUrlsNew, byUrls)
		// url列表取出
		urls, err = fwsdef.GetUrlsFromBytes(byUrlsNew)
		if err != nil {
			gLog.Errorf("fwsdef.GetUrlsFromBytes failed: %s, invalid urls data", err.Error())
			return -1
		}
	}

	var ed *fwsdef.EventDataFromCT = &fwsdef.EventDataFromCT{
		ED: &fwsdef.EventDataT{
			DescBuf: byDescNew,
			PicBuf:  byPicNew,
		},
		Urls: urls,
	}

	// 扔给数据管理模块处理
	edmgr.Add(ed)
	return 0
}

//export FWGO_Cleanup
func FWGO_Cleanup() int32 {
	// 关闭本模块启动的协程
	if ok := fwsconf.IsLogToFile(); ok == true {
		gchExit <- true
	}

	// 停止日志管理模块
	logmgr.Stop()

	// 关闭数据管理模块
	edmgr.Stop()

	gLog.Infof("to be close log file.")
	// 关闭日志输出
	golog.Close(".", nil)

	time.Sleep(500 * time.Millisecond)

	gLog.Infof("#### fwapp exit ####")
	return 0
}

//export FWGO_Test
func FWGO_Test() {
	fmt.Println("it's FW_Test()")
}
