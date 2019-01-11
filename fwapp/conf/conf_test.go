package fwsconf

import (
	"testing"
)

func TestGetListenPort(t *testing.T) {
	err := InitConf("../fwserver-config.json")
	if err != nil {
		t.Errorf("InitConf failed: %s\n", err.Error())
		return
	}
	t.Logf("Listen port: %d\n", GetListenPort())
	t.Logf("Log file path: %s\n", GetLogFilePath())
	t.Logf("Log to stdout: %v\n", IsLogToStdout())
	t.Logf("Log to file: %v\n", IsLogToFile())
	t.Logf("Log file keep days: %d\n", GetLogFileKeepdays())
	t.Logf("Pic max count in cache: %d\n", GetFPicMaxInCache())
	t.Logf("Pic file path: %s\n", GetFPicStorPath())
	t.Logf("Pic file storage max MB: %d\n", GetFPicStorMaxMB())
}
