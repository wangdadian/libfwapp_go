package edmgr

import (
	"sync"
)

//事件数据发送至目标服务器的状态
type edStateT struct {
	iEDCount       int // 总的事件个数
	iEDSendOKCount int // 全部发送成功的事件数
	iToSendCount   int // 需要发送的总数（一个数据可能需要发送至多目标）
	iSendOKCount   int // 发送成功个数（基于iToSendCount）
	sync.RWMutex
}

// 增加或减去事件总数，n为负数表示减去
func (self *edStateT) addEDCount(n int) {
	if n == 0 {
		return
	}
	self.Lock()
	self.iEDCount += n
	self.Unlock()
}

// 读取事件总数
func (self *edStateT) getEDCount() int {
	self.RLock()
	n := self.iEDCount
	self.RUnlock()
	return n
}

// 增加或减去事件发送成功总数，n为负数表示减去
func (self *edStateT) addEDSendOKCount(n int) {
	if n == 0 {
		return
	}
	self.Lock()
	self.iEDCount += n
	self.Unlock()
}

// 读取事件发送成功的总数
func (self *edStateT) getEDSendOKCount() int {
	self.RLock()
	n := self.iEDCount
	self.RUnlock()
	return n
}

// 增加或减去需发送总数
func (self *edStateT) addToSendCount(n int) {
	if n == 0 {
		return
	}
	self.Lock()
	self.iToSendCount += n
	self.Unlock()
}

// 读取需发送的总数
func (self *edStateT) getToSendCount() int {
	self.RLock()
	n := self.iToSendCount
	self.RUnlock()
	return n
}

// 增加或减去需发送总数
func (self *edStateT) addSendOKCount(n int) {
	if n == 0 {
		return
	}
	self.Lock()
	self.iSendOKCount += n
	self.Unlock()
}

// 读取发送成功的总数
func (self *edStateT) getSendOKCount() int {
	self.RLock()
	n := self.iSendOKCount
	self.RUnlock()
	return n
}

type CountT struct {
	n int
	sync.RWMutex
}

// 增加或减去
func (self *CountT) addCount(n int) {
	self.Lock()
	self.n += n
	self.Unlock()
}

// 读取个数
func (self *CountT) getCount() int {
	self.RLock()
	n := self.n
	self.RUnlock()
	return n
}
