package edmgr

import (
	"dadian/golog"
	"dadian/sort"
	"libfwapp_go/fwapp/conf"
	"libfwapp_go/fwapp/fwsdef"
	"libfwapp_go/fwapp/servers"
	"libfwapp_go/fwapp/storage"
	"sync"
	"time"
)

// 日志模块
var gLog *golog.Logger = golog.New("EevntDataMgr")

const (
	MAX_ITEM_STORAGE = 200 // 每次最大写入数量
)

// 事件列表数据，来自网络接收
type ediListNetT struct {
	ediList map[int64]*fwsdef.EDItem // 事件数据列表，key为插入事件数据时的纳秒数，方便后续的时序处理
	sync.Mutex
}

type eventDataMgr struct {
	ediGoNum       *CountT                  // 内存中并发处理个数
	edisNet        *ediListNetT             // 网络接收的事件列表数据
	edisStor       map[int64]*fwsdef.EDItem // 本地读取的之前发送失败的事件列表，key为插入事件数据时的纳秒数，方便后续的时序处理
	edStateNet     *edStateT                // 网络接收的事件发送至目标服务器的状态
	edStateStor    *edStateT                // 本地读取的事件发送至目标服务器的状态
	ediCIC         *CountT                  // 内存中在并发处理的事件个数
	cED            chan *fwsdef.EDItem      //事件数据通道，用于信息传输
	pStorage       stor.StorageRWer         // 本地存储读写接口
	bExit          bool                     // 模块退出信号
	cStartOverWait chan bool                // 等待thStart go routine退出
}

var gEDMgr *eventDataMgr = nil

const (
	MAX_LOCK_MS = 2000 //允许锁住最大毫秒数
)

// 添加新的事件数据，仅做添加处理
// 后续的管理工作由go routine负责处理
func Add(ed *fwsdef.EventDataFromCT) error {
	mgrKey := getEDListMapKey()

	// 新建数据检测单元数据
	var pEDItem *fwsdef.EDItem = &fwsdef.EDItem{
		Data:    ed.ED,
		Time:    mgrKey,
		SvrsMap: make(map[int]svrs.ServerWriter),
	}
	// 为此事件单元新建默认的http服务器
	ediSvrsMap, _ := pEDItem.SvrsMap.(map[int]svrs.ServerWriter)
	for _, url := range ed.Urls {
		key := getEDItemMapKey(pEDItem)
		// 新建http服务
		httpsvr, err := svrs.NewHttpServer(url)
		if err != nil {
			continue
		}
		ediSvrsMap[key] = svrs.ServerWriter(httpsvr)
	}
	// 为此事件数据添加需要其他发送的目标服务器
	addServers(pEDItem)

	processEDI(pEDItem)
	// 收到的事件总数增1
	gEDMgr.edStateNet.iEDCount += 1
	// 收到的事件需发送总数增加
	ediSvrsMap, _ = pEDItem.SvrsMap.(map[int]svrs.ServerWriter)
	gEDMgr.edStateNet.iSendOKCount += len(ediSvrsMap)
	return nil
}

// 为事件数据单元添加需要发送的目标服务器
func addServers(edi *fwsdef.EDItem) error {
	//
	// 在此处添加其他类型服务器，用于扩展
	//

	return nil
}

// 当前并发量未超过配额时，直接发送事件数据
// 并发量大时，放入发送队列，由发送携程负责发送
func processEDI(edi *fwsdef.EDItem) {
	iMax := gEDMgr.ediGoNum.getCount()
	iEDICIC := gEDMgr.ediCIC.getCount()
	if iEDICIC < iMax {
		// 目前并发池上有空余，扔进并发池处理
		gEDMgr.cED <- edi
		gLog.Infof("send event data to go routine OK")
		return
	}
	gLog.Warnf("Current concurrent number %d >= %d, insert to write queue", iEDICIC, iMax)
	// 并发池已满，则加入发送队列
	gEDMgr.edisNet.Lock()
	gEDMgr.edisNet.ediList[edi.Time] = edi
	gEDMgr.edisNet.Unlock()
	// gLog.Infof("add event data to write queue OK")
}

// 停止模块服务
func Stop() error {
	// 停止服务
	gEDMgr.bExit = true

	// 将内存中的事件据写入磁盘
	gLog.Infof("before event data manager service stop, write data to storage, count: %d", len(gEDMgr.edisNet.ediList))
	for k, edi := range gEDMgr.edisNet.ediList {
		err := writeToStorage(edi)
		if err != nil {
			gLog.Errorf("write to storage failed, event data time: %d", k)
		} else {
			gLog.Infof("write to storage ok, event data time: %d", k)
		}
	}

	// 等待服务停止
	<-gEDMgr.cStartOverWait
	close(gEDMgr.cStartOverWait)

	// 停止存储管理服务
	gEDMgr.pStorage.StopManager()
	gLog.Infof("Event Data Manager stoped")
	return nil
}

// 开启事件数据管理服务
func Start() {
	pStorage, err := stor.NewDiskStorage()
	if err != nil {
		gLog.Errorf("stor.NewDiskStorage failed: %s", err.Error())
		pStorage = nil
	}
	// 开启存储管理服务器
	pStorage.StartManager()

	gEDMgr = &eventDataMgr{
		ediGoNum: &CountT{n: fwsconf.GetFPicMaxInCache()},
		edisNet: &ediListNetT{
			ediList: make(map[int64]*fwsdef.EDItem),
		},
		edisStor:       make(map[int64]*fwsdef.EDItem),
		edStateNet:     &edStateT{},
		edStateStor:    &edStateT{},
		ediCIC:         &CountT{n: 0},
		cED:            make(chan *fwsdef.EDItem),
		bExit:          false,
		cStartOverWait: make(chan bool),
		pStorage:       pStorage,
	}
	// 启动相关服务
	go thStart()
}

func thStart() {
	var wg sync.WaitGroup
	// 启动内存中允许最大数据量的事件个数go routine用于并发发送事件
	// 并发池创建
	iMax := gEDMgr.ediGoNum.getCount()
	for i := 0; i < iMax; i++ {
		wg.Add(1)
		go thEDIWriteToServers(&wg)
	}

	// 启动发送go routine
	wg.Add(1)
	go thEDIWriteToServersMgr(&wg)

	// 启动定时从本地读取之前发送失败的事件数据，用于继续发送
	wg.Add(1)
	go thEDIReadFormStorage(&wg)

	gLog.Infof("Event Data Manager service start")

	//
	// 等待其发起的go routine结束，其实是在Stop的调用
	//
	wg.Wait()

	// 通知Stop结束接口，本routine结束
	gEDMgr.cStartOverWait <- true
	gLog.Infof("Event Data Manager service over")
}

// 负责发送数据至服务器的go routine
// 根据map中key值代表事件时序发送，先发送最后收到的数据，逆序排列key值
func thEDIWriteToServersMgr(wg *sync.WaitGroup) {
	defer wg.Done()
	var iEDListLen int = 0
	var tStart time.Time
	// 储存时序值用于排序
	var keySlice []int64

	for {
		if gEDMgr.bExit {
			return
		}
		// 睡一会儿
		time.Sleep(100 * time.Millisecond)
		gEDMgr.edisNet.Lock()
		// 记录加锁时间，用于避免锁住太久
		tStart = time.Now()
		iEDListLen = len(gEDMgr.edisNet.ediList)
		if iEDListLen <= 0 {
			// 发送队列中没有待发送的事件数据，继续...
			gEDMgr.edisNet.Unlock()
			continue
		}
		gLog.Infof("current write queue is %d, now to write to server", iEDListLen)
		iMax := gEDMgr.ediGoNum.getCount()
		iEDICIC := gEDMgr.ediCIC.getCount()
		if iEDICIC >= iMax {
			// 目前并发数过大
			gEDMgr.edisNet.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		// 清空之前的时序列表
		keySlice = nil
		for key, _ := range gEDMgr.edisNet.ediList {
			keySlice = append(keySlice, key)
		}
		// 按照时序逆序排序，最早的事件在最后
		keySlice = ddsort.SortReverseInt64(keySlice)
		for i := 0; i < iEDListLen; i++ {
			// 如果锁超时，解锁
			if isLockTimeout(tStart) {
				break
			}
			edi := gEDMgr.edisNet.ediList[keySlice[i]]
			iMax := gEDMgr.ediGoNum.getCount()
			iEDICIC := gEDMgr.ediCIC.getCount()
			if iEDICIC < iMax {
				// 将事件单元数据通过通道发送给并发协程处理
				// gLog.Errorf("waiting for read...")
				gEDMgr.cED <- edi
				// gLog.Errorf("read ok")
				// 删除此元素
				delete(gEDMgr.edisNet.ediList, keySlice[i])
				// 每次发送3个
				if i < 3 {
					continue
				} else {
					break
				}
			} else {
				// 目前并发数过大
				break
			}
		} // for i := 0; i < iEDListLen; i++ {
		gEDMgr.edisNet.Unlock()
	} // for {

}

// 并发池go routine
// 负责处理事件数据的转发
func thEDIWriteToServers(wg *sync.WaitGroup) {
	defer wg.Done()
	// 并发总数减少
	defer gEDMgr.ediGoNum.addCount(-1)
	var pEDI *fwsdef.EDItem = nil
	for {
		if gEDMgr.bExit {
			return
		}
		time.Sleep(1 * time.Millisecond)
		select {
		case pEDI = <-gEDMgr.cED: // 收到待处理的事件数据
			// 并发数增1
			gEDMgr.ediCIC.addCount(1)
			break
		default:
			time.Sleep(2 * time.Millisecond)
			continue
		}
		// 收到数据，处理。。。
		if pEDI == nil {
			// 无效数据，继续接收
			gEDMgr.ediCIC.addCount(-1)
			continue
		}
		ediSvrsMap, ok := pEDI.SvrsMap.(map[int]svrs.ServerWriter)
		if ok == false || len(ediSvrsMap) <= 0 {
			// 无效数据或者已发送完毕无需再次发送，删除后，继续下一个
			gLog.Warnf("event data item has invalid member [edi=%v, type ok=%v, len(server)=%d], deleted", pEDI, ok, len(ediSvrsMap))
			pEDI = nil
			gEDMgr.ediCIC.addCount(-1)
			continue
		}
		// 循环发送，多次失败写入硬盘
		//
		MAX_SEND_NUM := 10
		for i := 0; i < MAX_SEND_NUM; {
			i++
			iSvrs, iOK := writeToServers(pEDI)
			// 修改发送成功事件个数
			gEDMgr.edStateNet.addSendOKCount(iOK)

			// 发送全部成功
			if iSvrs == iOK {
				pEDI = nil
				// 并发数减少
				gEDMgr.edStateNet.addEDSendOKCount(1)
				gEDMgr.ediCIC.addCount(-1)
				break
			}

			// 没有全部发送成功
			// 当前内存中并发过大，或者超过最大循环发送次数，则写入磁盘
			iEDICIC := gEDMgr.ediCIC.getCount()
			iMax := gEDMgr.ediGoNum.getCount()
			if iEDICIC >= iMax || i >= MAX_SEND_NUM || gEDMgr.bExit {
				// 写入磁盘，不关心结果
				writeToStorage(pEDI)
				// 并发数减少
				gEDMgr.ediCIC.addCount(-1)
				break
			}
			// 等待下一次发送
			for n := 0; n < 50; n++ {
				if gEDMgr.bExit {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		} // for i := 0; i < MAX_SEND_NUM; {
	} // for {
}

// 负责定时从本地存储读取之前发送失败额数据
// 每次读取一个
func thEDIReadFormStorage(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if gEDMgr.bExit {
			return
		}
		// 睡一大会儿
		for i := 0; i < 100; i++ {
			if gEDMgr.bExit {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

		// 初始化当前存储下的所有数据条目
		iTotal, err := gEDMgr.pStorage.ReadAll()
		if err != nil || iTotal <= 0 {
			continue
		}
		gEDMgr.edStateStor.addEDCount(iTotal)
		// 循环读取每个事件文件进行重新发送
		for {
			time.Sleep(50 * time.Millisecond)
			pEDI, err := gEDMgr.pStorage.Next()
			if gEDMgr.bExit || err != nil || pEDI == nil {
				break
			}
			MAX_SEND_NUM := 2
			for i := 0; i < MAX_SEND_NUM; {
				i++
				//
				iSvrs, iOK := writeToServers(pEDI)
				// 修改发送成功事件个数
				gEDMgr.edStateStor.addSendOKCount(iOK)

				// 发送全部成功
				if iSvrs == iOK {
					// 发送成功后，删除文件
					gEDMgr.pStorage.Remove(pEDI.Time)
					pEDI = nil
					break
				}

				// 没有全部发送成功
				// 超过最大循环发送次数，继续处理下一个事件
				if i >= MAX_SEND_NUM {
					break
				}

				// 等待下一次发送
				for n := 0; n < 50; n++ {
					if gEDMgr.bExit {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			} // for i := 0; i < MAX_SEND_NUM; {
		} // for {
	} // for {
}

// 新增事件时，获取事件的key值，用于存入map。
// key值为当前unix时间纳秒数
func getEDListMapKey() int64 {
	var key int64
	for {
		key = time.Now().UnixNano()
		gEDMgr.edisNet.Lock()
		_, ok := gEDMgr.edisNet.ediList[key]
		gEDMgr.edisNet.Unlock()
		if ok {
			key++
			continue
		}
		break
	}
	return key
}

// 为当前事件单元中的服务器map列表增加key值
func getEDItemMapKey(edi *fwsdef.EDItem) int {
	if edi == nil {
		return -1
	}
	ediSvrsMap, ok := edi.SvrsMap.(map[int]svrs.ServerWriter)
	if ok == false {
		return -1
	}
	key := len(ediSvrsMap)
	for {
		_, ok := ediSvrsMap[key]
		if ok {
			key++
			if key > 0x0fffffff {
				key = 0
			}
			continue
		}
		break
	}
	return key
}

// 判断是否超时
func isLockTimeout(tStart time.Time) bool {
	tNow := time.Now()
	// 如果锁住超过预设最大毫秒数，则释放锁
	secValue := int(tNow.Sub(tStart).Nanoseconds() / time.Millisecond.Nanoseconds())
	if secValue >= MAX_LOCK_MS {
		return true
	}
	return false
}
