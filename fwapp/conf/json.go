package fwsconf

import (
	"dadian/jsonconf"
	"encoding/json"
	"runtime"
)

//
// 坑爹的玩意儿：
// **** 结构体字段后面的json标签中，"josn:"与标签名称中间不能有空格，不能有空格！！！
//
type JCNetwork struct {
	ListenPort int `json:"listen_port"`
}
type JClogFilePath struct {
	LFPLinux   string `json:"tfp_linux"`
	LFPWindows string `json:"tfp_windows"`
}
type JCLog struct {
	ToStdout        bool          `json:"to_stdout"`
	ToFile          bool          `json:"to_file"`
	LogPath         JClogFilePath `json:"to_file_path"`
	LogFileKeepdays int           `json:"to_file_keep_days"`
}
type JCFPicCache struct {
	FPicMaxInCache int `json:"pic_items_max"`
}
type JCpicStorPath struct {
	PSPLinux   string `json:"psp_linux"`
	PSPWindows string `json:"psp_windows"`
}
type JCFPicStor struct {
	PicPath      JCpicStorPath `json:"pic_storage_path"`
	PicStorMaxMB int           `json:"pic_storage_max_mb"`
}
type JsonConf struct {
	bInitOK bool        // 是否已经初始化成功，小写字母开头不会被json解析
	Net     JCNetwork   `json:"network"`
	Log     JCLog       `json:"log"`
	Picc    JCFPicCache `json:"failure_pic_cache"`
	Pics    JCFPicStor  `json:"failure_pic_storage"`
}

func (self *JsonConf) InitConf(szFile string) error {
	self.bInitOK = false
	var byJsonBuff []byte
	var err error
	for {
		byJsonBuff, err = jsonconf.ReadConf(szFile)
		if err != nil {
			gLog.Errorf("read json config file[%s] failed: %s", szFile, err.Error())
			break
		}
		err = json.Unmarshal(byJsonBuff, self)
		if err != nil {
			gLog.Errorf("parse json config file[%s] failed: %s", szFile, err.Error())
			break
		}
		break
	}
	if err != nil {
		gLog.Warnf("using default json configuration.")
		// 读取配置错误，使用默认配置，返回nil
		return nil
	}
	gLog.Infof("read from json config file[%s] ok.", szFile)
	self.bInitOK = true
	return nil
}

// 获取服务端口
func (self *JsonConf) GetListenPort() int {
	if self.bInitOK && self.Net.ListenPort > 0 {
		return self.Net.ListenPort
	}
	return 1282
}

// 日志信息是否终端打印，默认true
func (self *JsonConf) IsLogToStdout() bool {
	if self.bInitOK {
		return self.Log.ToStdout
	}
	return true
}

// 日志信息是否保存至日志文件，默认false
func (self *JsonConf) IsLogToFile() bool {
	if self.bInitOK {
		return self.Log.ToFile
	}
	return false
}

// 获取日志文件路径（不含文件名称）
func (self *JsonConf) GetLogFilePath() string {
	pathLinux := "/tmp/fwserver/log/"
	pathWindows := "c:\\fwserver\\log\\"
	if self.bInitOK {
		if runtime.GOOS == "linux" {
			if self.Log.LogPath.LFPLinux != "" {
				return self.Log.LogPath.LFPLinux
			}
		} else if runtime.GOOS == "windows" {
			if self.Log.LogPath.LFPWindows != "" {
				return self.Log.LogPath.LFPWindows
			}
		}
	}
	if runtime.GOOS == "linux" {
		return pathLinux
	} else if runtime.GOOS == "windows" {
		return pathWindows
	}
	return ""
}

// 获取日志文件最大保存天数
func (self *JsonConf) GetLogFileKeepdays() int {
	if self.bInitOK && self.Log.LogFileKeepdays > 0 {
		return self.Log.LogFileKeepdays
	}
	return 15
}

/*
* 发送失败的情况下，图片本地保存策略
 */
// 发送失败后，超过一定数量后，图片写入磁盘。获取内存最大保存图片数量
func (self *JsonConf) GetFPicMaxInCache() int {
	if self.bInitOK && self.Picc.FPicMaxInCache >= 20 {
		return self.Picc.FPicMaxInCache
	}
	return 800
}

// 获取图片存储路径，默认：
// windows: "/tmp/fwserver/pic/"
// linux: "c:\\fwserver\\pic\\"
func (self *JsonConf) GetFPicStorPath() string {
	pathLinux := "/tmp/fwserver/pic/"
	pathWindows := "c:\\fwserver\\pic\\"
	if self.bInitOK {
		if runtime.GOOS == "linux" {
			if self.Pics.PicPath.PSPLinux != "" {
				return self.Pics.PicPath.PSPLinux
			}
		} else if runtime.GOOS == "windows" {
			if self.Pics.PicPath.PSPWindows != "" {
				return self.Pics.PicPath.PSPWindows
			}
		}
	}
	if runtime.GOOS == "linux" {
		return pathLinux
	} else if runtime.GOOS == "windows" {
		return pathWindows
	}
	return ""
}

// 保存的图片占用当前磁盘的空间值（单位MB）
func (self *JsonConf) GetFPicStorMaxMB() int {
	if self.bInitOK && self.Pics.PicStorMaxMB >= 128 {
		return self.Pics.PicStorMaxMB
	}
	return 4096
}
