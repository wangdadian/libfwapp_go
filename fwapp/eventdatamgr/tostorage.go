package edmgr

import (
	"fmt"
	"libfwapp_go/fwapp/fwsdef"
)

func writeToStorage(edi *fwsdef.EDItem) error {
	// 写入本地磁盘
	if gEDMgr.pStorage == nil {
		return fmt.Errorf("pStorage == nil")
	}
	return gEDMgr.pStorage.Write(edi)
}
