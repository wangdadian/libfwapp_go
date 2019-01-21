package edmgr

import (
	"libfwapp_go/fwapp/fwsdef"
	"libfwapp_go/fwapp/servers"
)

// 按照事件数据中配置的服务器列表分别发送至各个服务器
// 针对发送成功的，将此服务器从服务器列表中清除
// 返回：发送总服务器数，以及发送成功的个数
func writeToServers(edi *fwsdef.EDItem) (int, int) {
	ediSvrsMap, ok := edi.SvrsMap.(map[int]svrs.ServerWriter)
	if edi == nil || ok == false || len(ediSvrsMap) <= 0 {
		// 理论上不会进入
		gLog.Errorf("edi == nil || ok == false || len(ediSvrsMap) <= 0")
		return 0, 0
	}

	iSvrs := 0
	iOK := 0
	// 循环向每个服务器发送
	for k, svr := range ediSvrsMap {
		// 删除无效服务器，理论上不会进入
		if svr == nil {
			delete(ediSvrsMap, k)
			continue
		}
		iSvrs++
		err := svr.Write(edi.Data)
		// 删除发送成功服务器
		if err == nil {
			iOK++
			delete(ediSvrsMap, k)
		}
	}
	return iSvrs, iOK
}
