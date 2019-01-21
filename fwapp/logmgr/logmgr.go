package logmgr

import (
	"dadian/compress"
	"dadian/golog"
	"fmt"
	"io/ioutil"
	"libfwapp_go/fwapp/conf"
	"os"
	"runtime"
	"strings"
	"time"
)

type logMgr struct {
	path string // 日志存储路径
	days int    // 日志保留天数
	exit bool
	c    chan bool
}

func (self *logMgr) String() string {
	return fmt.Sprintf("log file path: %s, log file keep days: %d", self.path, self.days)
}

var gLogMgr *logMgr = nil
var gLog *golog.Logger = golog.New("LogMgr")

// 启动
func Start() error {
	if gLogMgr != nil {
		gLog.Warnf("Log Manager has been Start, no need to start again.")
		return nil
	}
	// 获取配置信息
	gLogMgr = &logMgr{
		path: fwsconf.GetLogFilePath(),
		days: fwsconf.GetLogFileKeepdays(),
		exit: false,
		c:    make(chan bool),
	}
	gLog.Infof("Log file configure information: %s", gLogMgr.String())
	// 启动log日志文件管理go routine
	go thLogManagerWorker()
	gLog.Infof("Log Manager start OK.")
	return nil
}

// 停止
func Stop() {
	gLogMgr.exit = true
	<-gLogMgr.c
	close(gLogMgr.c)
	gLog.Infof("Log Manager stoped")
}

// 日志文件管理协程
func thLogManagerWorker() {
	gLog.Infof("thLogManagerWorker start")
	for {
		if gLogMgr.exit {
			break
		}
		logManager()
		// 睡眠
		for i := 0; i < 150; i++ {
			if gLogMgr.exit {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}

	}
	gLogMgr.c <- true
	gLog.Infof("thLogManagerWorker end")
}

func logManager() error {
	// 判断文件夹访问是否正常
	_, err := os.Stat(gLogMgr.path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件夹不存在，不处理直接返回
			return err
		}
		// 文件夹访问异常
		gLog.Errorf("access dir[%s] failed: %s", gLogMgr.path, err.Error())
		return err
	}

	// 获取文件夹下的文件
	fList, err := getFileListAtDir(gLogMgr.path)
	if err != nil {
		return err
	}
	// 判断文件时间是否需要删除
	for _, szFile := range fList {
		fi, err := os.Stat(szFile)
		if err != nil {
			continue
		}
		tModTime := fi.ModTime()
		tNow := time.Now()
		if int(tNow.Sub(tModTime).Seconds()) > gLogMgr.days*24*3600 {
			gLog.Warnf("log file[%s] has lasted more than %d days, need to remove.", szFile, gLogMgr.days)
			// 超过最大天数，则删除文件
			err = os.Remove(szFile)
			if err != nil {
				gLog.Errorf("remove log file[%s] failed: %s", szFile, err.Error())
				continue
			} else {
				gLog.Infof("remove log file[%s] ok.", szFile)
			}
			// 继续下一个文件
			continue
		}
		//
		// 如果没超过最大天数，则压缩1天前的日志文件
		//
		if int(tNow.Sub(tModTime).Seconds()) > 1*24*3600 {
			// 压缩日志文件
			// 已经压缩了，继续下一个文件
			if strings.HasSuffix(szFile, ".zip") || strings.HasSuffix(szFile, ".gz") {
				continue
			}
			destF, err := compressLog(szFile)
			if err != nil {
				gLog.Errorf("compress log file[%s] failed: %s", szFile, err.Error())
				continue
			}
			gLog.Infof("compress log file[%s] to [%s] OK", szFile, destF)
			// 更新压缩文件的更新时间，为日志文件的更新时间
			// 当超过最大保留天数后，可以正常删除
			err = os.Chtimes(destF, tModTime, tModTime)
			if err != nil {
				gLog.Warnf("change file[%s] modify time failed: %s", destF, err.Error())
			}
			gLog.Infof("change file[%s] modify time[%v] ok", destF, tModTime)
			// 删除原日志文件
			err = os.Remove(szFile)
			if err != nil {
				gLog.Infof("after compress ok, remove log file[%s] failed: %s", szFile, err.Error())
			}
			gLog.Infof("after compress ok, remove log file[%s] ok", szFile)
		}

	}
	return nil
}

func getFileListAtDir(path string) ([]string, error) {
	var fList []string = nil
	fiList, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, fi := range fiList {
		if fi.IsDir() {
			// 忽略文件夹
			continue
		}
		// 文件
		fName := fi.Name()
		szDelim := ""
		b := []byte(path)
		if runtime.GOOS == "windows" {
			if b[len(b)-1] != '\\' {
				b = append(b, []byte{'\\', '\\'}...)
			} else {
				if b[len(b)-2] != '\\' {
					b = append(b, '\\')
				}
			}
		} else if runtime.GOOS == "linux" {
			if b[len(b)-1] != '/' {
				b = append(b, '/')
			}
		}
		szFile := string(b) + szDelim + fName
		fList = append(fList, szFile)
	}
	return fList, nil
}

func compressLog(szFile string) (string, error) {
	szDestFile := szFile
	szDestFile, err := compress.Compress([]string{szFile}, szDestFile, compress.CT_AUTO)
	if err != nil {
		return "", err
	}
	return szDestFile, nil
}
