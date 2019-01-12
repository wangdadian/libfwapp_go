package stor

import (
	"dadian/endian"
	"dadian/sort"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"libfwapp_go/fwapp/conf"
	"libfwapp_go/fwapp/fwsdef"
	"libfwapp_go/fwapp/servers"
	"os"
	"strconv"
	"time"
)

type DiskStorage struct {
	pathPic      string  // 图片存储的本地磁盘路径
	iKeyFilesArr []int64 // 存储路径下的文件列表，文件名是事件时间纳秒数，和事件时间相同
	iIndex       int     // 索引值，指向iKeyFile的位置
	bExit        bool
}

//
//
//
var gDiskStorage *DiskStorage = nil

// 新建磁盘存储对象
func NewDiskStorage() (*DiskStorage, error) {
	if gDiskStorage != nil {
		return gDiskStorage, nil
	}
	pathPic := fwsconf.GetFPicStorPath()
	_, err := os.Stat(pathPic)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(pathPic, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("os.MkdirAll dir[%s] failed: %s", pathPic, err.Error())
			}
		} else {
			return nil, fmt.Errorf("os.Stat dir[%s] failed: %s", pathPic, err.Error())
		}
	}

	gDiskStorage := &DiskStorage{
		pathPic:      pathPic,
		iKeyFilesArr: nil,
		iIndex:       0,
		bExit:        false,
	}

	return gDiskStorage, nil
}

// 开启管理服务，进行空间管理
func (self *DiskStorage) StartManager() error {
	go self.thStorageManager()
	return nil
}

// 关闭管理服务
func (self *DiskStorage) StopManager() {
	self.bExit = false
}
func (self *DiskStorage) thStorageManager() {
	for {
		if self.bExit {
			return
		}
		for i := 0; i < 300; i++ {
			if self.bExit {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		self.storageManager()
	}
}

// 每个 fwsdef.EDItem 数据单元写入一个文件，文件名称为事件接收事件的纳秒数
// 所有4字节的整型数据采用网络模式的大端方式存储（方便不通平台转移），取出时需转换
// 文件格式如下
/*
| 4字节 |  // 文件名长度
| 文件名|  // 文件名称
| 4字节 |  // 描述文件信息长度
| 4字节 |  // 图片数据大小
| 4字节 |  // 多少个待发送的目标服务器，n个
| 4字节 |  // 每个目标服务器的json信息（服务器结构体序列化后的信息），重复n个4字节
| 描述信息  | // 描述信息字节流
| 图片      | // 图片数据字节流
| 服务器信息 | 重复n个服务器信息

*/
// 写入本地磁盘
func (self *DiskStorage) Write(edi *fwsdef.EDItem) error {
	ediSvrsMap, ok := edi.SvrsMap.(map[int]svrs.ServerWriter)
	if edi == nil || ok == false || len(ediSvrsMap) <= 0 {
		return fmt.Errorf("data is nil, or no servers specified")
	}

	var szFile string     // 全路径
	var szFileName string // 文件名
	var err error

	szFileName = fmt.Sprintf("%d", edi.Time)
	szFile = self.pathPic + szFileName
	_, err = os.Stat(szFile)
	if err == nil {
		// 文件已存在，无需再次重写
		gLog.Warnf("file[%d] has been existed, no need write again.", edi.Time)
		return nil
	}
	gLog.Infof("now to create and write to file[%s]...", szFile)
	file, err := os.Create(szFile)
	if err != nil {
		gLog.Errorf("Create file[%s] faiiled: %s", szFile, err.Error())
		return err
	}
	defer func() {
		file.Close()
		// 接口执行中，出现错误，删除文件
		if err != nil {
			gLog.Errorf("delete file[%s], because of a error occoured.", file.Name())
			os.Remove(szFileName)
		}
	}()
	iFileNameLen := len(szFileName)
	iDescLen := len(edi.Data.DescBuf)
	iPicLen := len(edi.Data.PicBuf)
	iServersCount := len(ediSvrsMap)
	var bySvrs [][]byte
	for _, v := range ediSvrsMap {
		bySvr, err := json.Marshal(v)
		if err != nil {
			gLog.Errorf("json.Marshal failed: %s", err.Error())
			return err
		}
		bySvrs = append(bySvrs, bySvr)
	}
	//
	// 写入文件
	//
	// 首先写入文件长度
	iFileNameLen = int(endian.HTONL(uint32(iFileNameLen)))
	byBuf, _ := endian.IntToBytes(iFileNameLen, 4)
	_, err = file.Write(byBuf)
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 文件名称
	_, err = file.Write([]byte(szFileName))
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 描述信息长度
	iDescLen = int(endian.HTONL(uint32(iDescLen)))
	byBuf, _ = endian.IntToBytes(iDescLen, 4)
	_, err = file.Write(byBuf)
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 图片数据长度
	iPicLen = int(endian.HTONL(uint32(iPicLen)))
	byBuf, _ = endian.IntToBytes(iPicLen, 4)
	_, err = file.Write(byBuf)
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 服务器个数
	iServersCount = int(endian.HTONL(uint32(iServersCount)))
	byBuf, _ = endian.IntToBytes(iServersCount, 4)
	_, err = file.Write(byBuf)
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 每个服务器序列化后的长度
	for _, bs := range bySvrs {
		bsLen := len(bs)
		bsLen = int(endian.HTONL(uint32(bsLen)))
		byBuf, _ = endian.IntToBytes(bsLen, 4)
		_, err = file.Write(byBuf)
		if err != nil {
			gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
			return err
		}
	}
	// 描述信息字节流
	_, err = file.Write(edi.Data.DescBuf)
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 图片数据字节流
	_, err = file.Write(edi.Data.PicBuf)
	if err != nil {
		gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
		return err
	}
	// 每个服务器的序列化字节流
	for _, bs := range bySvrs {
		_, err = file.Write(bs)
		if err != nil {
			gLog.Errorf("Write file[%s] failed: %s", szFile, err.Error())
			return err
		}
	}
	gLog.Infof("create and write event data to file[%s] ok", szFile)
	return nil
}
func (self *DiskStorage) Remove(ns int64) error {
	szFile := self.pathPic + fmt.Sprintf("%d", ns)
	err := os.Remove(szFile)
	if err != nil {
		gLog.Errorf("remove file [%s] failed: %s", szFile, err.Error())
		return err
	}
	return nil
}

// 从本地磁盘读取事件数据
func (self *DiskStorage) Read(n int) (map[int64]*fwsdef.EDItem, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid number of count to read")
	}
	fiList, err := ioutil.ReadDir(self.pathPic)
	if err != nil {
		gLog.Errorf("ioutil.ReadDir[%s] failed: %s", self.pathPic, err.Error())
		return nil, err
	}
	if len(fiList) <= 0 {
		// 目录下没有文件
		// gLog.Warnf("no files in local storage, path; %s", self.pathPic)
		return nil, io.EOF
	}
	// 生成事件时间--文件名称的对应，用于排序
	mapFile := make(map[int64]string)
	var KeySlice []int64
	for _, f := range fiList {
		// 跳过目录、小文件
		if f.IsDir() || f.Size() <= 20 {
			continue
		}

		szFile := self.pathPic + f.Name()
		iTime, err := strconv.ParseInt(f.Name(), 10, 64)
		if err != nil || iTime <= 0 {
			// 跳过文件名称解析失败的文件
			gLog.Warnf("file[%s] is not a event data file, process next file", szFile)
			continue
		}
		// 跳过非事件文件
		ok := isEventFile(szFile, iTime)
		if !ok {
			continue
		}
		mapFile[iTime] = szFile
		KeySlice = append(KeySlice, iTime)
	}
	// 按照时间顺序逆序排序
	KeySlice = ddsort.SortReverseInt64(KeySlice)

	// 待读取的个数
	var iToReadCount int = n
	if n > len(KeySlice) {
		iToReadCount = len(KeySlice)
	}

	// 读取结果
	mapED := make(map[int64]*fwsdef.EDItem)
	iReaded := 0 // 已经读取OK的个数
	// 采用i < len(KeySlice)，而不是i < iToReadCount
	// 可以跳过读取失败或者文件不合规范的文件
	for i := 0; i < len(KeySlice); i++ {
		if iReaded >= iToReadCount {
			break
		}
		key := KeySlice[i]
		szFileL := mapFile[key]
		edi, err := self.readFile(key, szFileL)
		if err != nil {
			continue
		}
		iReaded++
		mapED[key] = edi
		// 读取完毕，删除文件
		gLog.Warnf("read event data file [%s] ok.", szFileL)
	}
	// gLog.Infof("read editem size=%d", len(mapED))
	return mapED, nil
}

// 读取当前时间下，存储下的所有事件文件，后续用Next获取下一个事件数据直到返回io.EOF
// 返回： 成功返回事件文件个数,nil，失败返回0,错误信息
func (self *DiskStorage) ReadAll() (int, error) {
	self.iIndex = 0
	self.iKeyFilesArr = nil
	fiList, err := ioutil.ReadDir(self.pathPic)
	if err != nil {
		gLog.Errorf("ioutil.ReadDir[%s] failed: %s", self.pathPic, err.Error())
		return 0, err
	}
	if len(fiList) <= 0 {
		// 目录下没有文件
		// gLog.Warnf("no files in local storage, path; %s", self.pathPic)
		return 0, io.EOF
	}
	// 生成事件时间--文件名称的对应，用于排序
	var KeySlice []int64
	for _, f := range fiList {
		// 跳过目录、小文件
		if f.IsDir() || f.Size() <= 20 {
			continue
		}

		szFile := self.pathPic + f.Name()
		iTime, err := strconv.ParseInt(f.Name(), 10, 64)
		if err != nil || iTime <= 0 {
			// 跳过文件名称解析失败的文件
			gLog.Warnf("file[%s] is not a event data file, process next file", szFile)
			continue
		}
		// 跳过非事件文件
		ok := isEventFile(szFile, iTime)
		if !ok {
			continue
		}
		KeySlice = append(KeySlice, iTime)
	}

	// 按照时间顺序逆序排序
	if len(KeySlice) <= 0 {
		return 0, io.EOF
	}
	KeySlice = ddsort.SortReverseInt64(KeySlice)
	self.iKeyFilesArr = KeySlice
	// 初始化索引至开始位置
	self.iIndex = 0
	return len(self.iKeyFilesArr), nil
}

// 下一个事件文件数据信息，没有可读取的数据时返回nil,io.EOF，失败返回nil,错误信息
func (self *DiskStorage) Next() (*fwsdef.EDItem, error) {
	if self.iKeyFilesArr == nil || len(self.iKeyFilesArr) <= 0 || self.iIndex >= len(self.iKeyFilesArr) {
		return nil, io.EOF
	}
	var pEDI *fwsdef.EDItem = nil
	var err error
	for {
		if self.iIndex >= len(self.iKeyFilesArr) {
			pEDI = nil
			err = io.EOF
			// 读到结尾，清理掉之前的数据
			self.iKeyFilesArr = nil
			self.iIndex = 0
			break
		}
		iTime := self.iKeyFilesArr[self.iIndex]
		szFile := self.pathPic + fmt.Sprintf("%d", iTime)
		self.iIndex += 1
		pEDI, err = self.readFile(iTime, szFile)
		if err == nil {
			// 读取完毕，删除文件
			gLog.Warnf("read event data file [%s] ok.", szFile)
			break
		}
	}

	return pEDI, err
}

func (self *DiskStorage) readFile(key int64, szFile string) (*fwsdef.EDItem, error) {
	fi, err := os.Stat(szFile)
	if err != nil {
		return nil, err
	}
	tModTime := fi.ModTime()
	tNow := time.Now()
	if int(tNow.Sub(tModTime).Seconds()) > MAX_EDFILE_KEEP_DAYS*24*3600 {
		return nil, fmt.Errorf("file[%s] has lasted more than %d days, no need to send.", szFile, MAX_EDFILE_KEEP_DAYS)
	}
	file, err := os.Open(szFile)
	if err != nil {
		gLog.Errorf("os.Open file [%s] failed: %s", szFile, err.Error())
		return nil, err
	}
	defer file.Close()
	byRdBuf := make([]byte, 1024)
	// 读取4字节长度的文件名长度
	_, err = file.Read(byRdBuf[:4])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	iFileNameLen, _ := endian.BytesToInt(byRdBuf[:4], true)
	iFileNameLen = int(endian.NTOHL(uint32(iFileNameLen)))
	// 读取文件名称
	_, err = file.Read(byRdBuf[:iFileNameLen])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	szNameL := string(byRdBuf[:iFileNameLen])
	if szNameL != fmt.Sprintf("%d", key) {
		// 文件名称和实际不匹配
		gLog.Warnf("file[%s] is not a Event Data Item file", file.Name())
		return nil, fmt.Errorf("not a event data file")
	}

	// 读取描述文件信息长度
	_, err = file.Read(byRdBuf[:4])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	iDescLen, _ := endian.BytesToInt(byRdBuf[:4], true)
	iDescLen = int(endian.NTOHL(uint32(iDescLen)))

	// 读取图片数据大小
	_, err = file.Read(byRdBuf[:4])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	iPicLen, _ := endian.BytesToInt(byRdBuf[:4], true)
	iPicLen = int(endian.NTOHL(uint32(iPicLen)))

	// 读取目标服务器个数
	_, err = file.Read(byRdBuf[:4])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	iServersCount, _ := endian.BytesToInt(byRdBuf[:4], true)
	iServersCount = int(endian.NTOHL(uint32(iServersCount)))
	// 读取每个服务器的序列化信息长度
	var arrSvrsInfoLen []int
	for j := 0; j < iServersCount; j++ {
		_, err = file.Read(byRdBuf[:4])
		if err != nil && err != io.EOF {
			gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
			return nil, err
		}
		infoLen, _ := endian.BytesToInt(byRdBuf[:4], true)
		infoLen = int(endian.NTOHL(uint32(infoLen)))
		arrSvrsInfoLen = append(arrSvrsInfoLen, infoLen)
	}
	var ed fwsdef.EventDataT
	// 读取描述信息字节流
	descBuf := make([]byte, iDescLen)
	_, err = file.Read(descBuf)
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	ed.DescBuf = descBuf

	// 读取图片信息字节流
	picBuf := make([]byte, iPicLen)
	_, err = file.Read(picBuf)
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return nil, err
	}
	ed.PicBuf = picBuf
	edi := fwsdef.EDItem{
		Data:    &ed,
		Time:    key,
		SvrsMap: make(map[int]svrs.ServerWriter),
	}

	mapSvrs := make(map[int]svrs.ServerWriter)
	// 读取每个服务器的信息字节流
	for j := 0; j < iServersCount; j++ {
		bySvr := make([]byte, arrSvrsInfoLen[j])
		_, err = file.Read(bySvr)
		if err != nil && err != io.EOF {
			gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
			return nil, err
		}
		var v interface{}
		err = json.Unmarshal(bySvr, &v)
		if err != nil {
			gLog.Errorf("json.Unmarshal failed: %s", err.Error())
			return nil, err
		}
		// 类型断言
		vv, ok := v.(map[string]interface{})
		if !ok {
			gLog.Errorf("unknown type!!!\n")
			return nil, err
		}
		if _, ok := vv["ID"]; !ok {
			gLog.Errorf("unknown struct!!!\n")
			return nil, err
		}
		// httpServer
		if vv["ID"].(string) == "ServerHttp" {
			svr := svrs.ServerHttp{}
			err = json.Unmarshal(bySvr, &svr)
			if err != nil {
				gLog.Errorf("json.Unmarshal failed: %s", err.Error())
				return nil, err
			}
			mapSvrs[j] = svrs.ServerWriter(&svr)
		}
	}
	edi.SvrsMap = mapSvrs
	return &edi, nil
}

func (self *DiskStorage) storageManager() error {
	fiList, err := ioutil.ReadDir(self.pathPic)
	if err != nil {
		gLog.Errorf("StorageMgr: ioutil.ReadDir[%s] failed: %s", self.pathPic, err.Error())
		return err
	}
	if len(fiList) <= 0 {
		// gLog.Warnf("no files in local storage, path; %s", self.pathPic)
		return nil
	}

	// 所有的事件文件列表，key值为事件时间
	mapFile := make(map[int64]string)
	// 时间时间key值列表，用于排序
	var KeySlice []int64

	// 获取限额使用量,MB
	iMaxSize := fwsconf.GetFPicStorMaxMB()
	// 获取总使用容量,MB
	var fTotalSize float64
	tModTime := time.Now()
	tNow := time.Now()
	for _, fi := range fiList {
		// 目录或者文件过小，跳过
		if fi.IsDir() || fi.Size() <= 20 {
			continue
		}
		// 解析文件名称出错，判定为非事件文件，跳过
		szFile := self.pathPic + fi.Name()
		iTime, err := strconv.ParseInt(fi.Name(), 10, 64)
		if err != nil || iTime <= 0 {
			gLog.Warnf("file [%s] is not a event data file, process next file", szFile)
			continue
		}

		// 判断是否为事件文件
		if ok := isEventFile(szFile, iTime); !ok {
			continue
		}

		// 删除超过MAX_EDFILE_KEEP_DAYS天的事件文件
		tModTime = fi.ModTime()
		if int(tNow.Sub(tModTime).Seconds()) > MAX_EDFILE_KEEP_DAYS*24*3600 {
			gLog.Warnf("event data file[%s] has lasted more than %d days, need to remove.", szFile, MAX_EDFILE_KEEP_DAYS)
			err = os.Remove(szFile)
			if err != nil {
				gLog.Errorf("remove event data file[%s] failed: %s", szFile, err.Error())
				continue
			} else {
				gLog.Infof("remove event data file[%s] ok.", szFile)
			}
			// 继续下一个文件
			continue
		}

		// 所有事件文件大小,MB
		fTotalSize += float64(fi.Size()) / 1024.0 / 1024.0
		KeySlice = append(KeySlice, iTime)
		mapFile[iTime] = szFile

	}

	iTotalSize := int(fTotalSize + 1.0)
	if iTotalSize < iMaxSize {
		// 未超额
		gLog.Infof("event data file total size=%d MB < %d MB, no need to delete.", iTotalSize, iMaxSize)
		return nil
	}

	//
	// 超额：删除最旧文件，删除量为 20%
	//
	gLog.Infof("event data file total size=%d > %d MB, now to delete some oldest files.", iTotalSize, iMaxSize)
	// 删除大小MB,目标值
	fToDeleteSize := fTotalSize * 0.2
	// 已删除的大小
	var fDeletedSize float64 = 0.0
	// 按照时间顺序排序，然后顺序删除
	KeySlice = ddsort.SortInt64(KeySlice)
	for i := 0; i < len(KeySlice); i++ {
		key := KeySlice[i]
		szFileNameL := mapFile[key]

		fi, err := os.Stat(szFileNameL)
		if err != nil {
			gLog.Errorf("os.Stat file[%s] failed: %s", szFileNameL, err.Error())
			continue
		}
		err = os.Remove(szFileNameL)
		if err != nil {
			gLog.Errorf("os.Remove file[%s] failed: %s", szFileNameL, err.Error())
			continue
		}
		gLog.Infof("delete event data file [%s] ok.", szFileNameL)
		fDeletedSize += float64(fi.Size()) / 1024.0 / 1024.0
		if fDeletedSize >= fToDeleteSize {
			break
		} else {
			continue
		}
	}
	gLog.Infof("has deleted some event data files, deleted size=%.02f MB", fDeletedSize)
	return nil
}

// 判定文件是否为事件文件
func isEventFile(szFile string, key int64) bool {
	file, err := os.Open(szFile)
	if err != nil {
		gLog.Errorf("os.Open file [%s] failed: %s", szFile, err.Error())
		return false
	}
	defer file.Close()
	byRdBuf := make([]byte, 1024)
	// 读取4字节长度的文件名长度
	_, err = file.Read(byRdBuf[:4])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return false
	}
	iFileNameLen, _ := endian.BytesToInt(byRdBuf[:4], true)
	iFileNameLen = int(endian.NTOHL(uint32(iFileNameLen)))
	// 读取文件名称
	_, err = file.Read(byRdBuf[:iFileNameLen])
	if err != nil && err != io.EOF {
		gLog.Errorf("Read file[%s] failed: %s", file.Name(), err.Error())
		return false
	}
	szNameL := string(byRdBuf[:iFileNameLen])
	if szNameL != fmt.Sprintf("%d", key) {
		// 文件名称和实际不匹配
		gLog.Warnf("file[%s] is not a Event Data Item file", file.Name())
		return false
	}
	return true
}
