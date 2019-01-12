package svrs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"libfwapp_go/fwapp/fwsdef"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// // 测试写入本地用的import
// import (
// 	"encoding/json"
// 	"fmt"
// 	"fwserver/conf"
// 	"fwserver/fwsdef"
// 	"os"
// 	"time"
// )

type ServerHttp struct {
	ID  string `json:"ID"` // 赋值时填写本类型名称，用于反序列化时的类型匹配，扩展其他类型服务器时也需要相同的声明
	Url string `json:"url"`
}

// 从http服务器收到的返回消息，用于判断是否发送成功
type RetBodyT struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

// 新建HTTP服务器
func NewHttpServer(url string) (*ServerHttp, error) {
	pSvrHttp := &ServerHttp{
		ID:  "ServerHttp",
		Url: url,
	}
	return pSvrHttp, nil
}

func (self *ServerHttp) Write(data *fwsdef.EventDataT) error {
	var bufRWer bytes.Buffer
	mpWr := multipart.NewWriter(&bufRWer)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", "form-data; name=\"vehicle\";")
	// h.Set("Content-Type", "text/plain")
	w, err := mpWr.CreatePart(h)
	if err != nil {
		gLog.Errorf("mpWr.CreatePart(h)1 failed: %s", err)
		return err
	}
	n := 0
	n, err = w.Write(data.DescBuf)
	if err != nil {
		gLog.Errorf("Write failed: %s", err.Error())
		return err
	}
	if n != len(data.DescBuf) {
		gLog.Errorf("Write failed: write len=%d != %d", n, len(data.DescBuf))
		return fmt.Errorf("Write failed: write len=%d != %d", n, len(data.DescBuf))
	}
	h = make(textproto.MIMEHeader)
	h.Set("Content-Disposition", "form-data; name=\"qjtp\";")
	h.Set("Content-Type", "image/jpeg")
	w, err = mpWr.CreatePart(h)
	if err != nil {
		gLog.Errorf("mpWr.CreatePart(h)1 failed: %s", err)
		return err
	}
	n, err = w.Write(data.PicBuf)
	if err != nil {
		gLog.Errorf("Write failed: %s", err.Error())
		return err
	}
	if n != len(data.PicBuf) {
		gLog.Errorf("Write failed: writed len=%d != %d", n, len(data.PicBuf))
		return fmt.Errorf("Write failed: writed len=%d != %d", n, len(data.PicBuf))
	}
	strBoundary := mpWr.Boundary()
	mpWr.Close()
	pClient := &http.Client{
		Timeout: 2000 * time.Millisecond,
	}
	req, err := http.NewRequest("POST", self.Url, &bufRWer)
	if err != nil {
		gLog.Errorf("http.NewRequest failed: %s", err.Error())
		return err
	}
	defer req.Body.Close()

	// ip, port, err := getInfoFromUrl(self.Url)
	// if err != nil {
	// 	gLog.Errorf("getInfoFromUrl failed: %s", err.Error())
	// 	return err
	// }

	// req.Header.Add("HOST", fmt.Sprintf("%s:%d", ip, port))
	req.Header.Add("HOST", req.Host)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+strBoundary)
	// gLog.Errorf("############  Do")
	gLog.Infof("start to write to http server[%s]...", self.Url)
	res, err := pClient.Do(req)
	if err != nil {
		gLog.Errorf("http Client Do failed: %s", err.Error())
		return err
	}
	defer res.Body.Close()

	// gLog.Errorf("############  ReadAll")
	byRetBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		gLog.Errorf("ioutil.ReadAll(res.Body) failed: %s", err.Error())
		return err
	}

	gLog.Infof("received message: %s", string(byRetBody))
	retBody := RetBodyT{}
	err = json.Unmarshal(byRetBody, &retBody)
	if err != nil {
		gLog.Errorf("json.Unmarshal failed: %s", err.Error())
		return err
	}
	if retBody.Code != "0" {
		gLog.Errorf("http server[%s] return message, it's not ok:[%s] %s", self.Url, retBody.Code, retBody.Msg)
		return fmt.Errorf("http server[%s] return message, it's not ok:[%s] %s", self.Url, retBody.Code, retBody.Msg)
	}
	gLog.Infof("http server[%s] return message, it's ok", self.Url)
	return nil
}

func getInfoFromUrl(szUrl string) (string, int, error) {
	if len(szUrl) <= len("http://0.0.0.0") {
		return "", -1, fmt.Errorf("invalid url")
	}
	u, err := url.Parse(szUrl)
	if err != nil {
		return "", -1, err
	}
	host := u.Host
	ip := ""
	port := 0
	sIpPort := strings.Split(host, ":")
	if len(sIpPort) == 1 {
		ip = sIpPort[0]
		port = 80
	} else if len(sIpPort) == 2 {
		ip = sIpPort[0]
		port, err = strconv.Atoi(sIpPort[1])
		if err != nil {
			return "", -1, fmt.Errorf("parse ip:port failed: %s", err.Error())
		}
	} else {
		return "", -1, fmt.Errorf("invalid url string")
	}

	if err != nil {
		return "", -1, fmt.Errorf("parse ip:port failed: %s", err.Error())
	}
	return ip, port, nil
}

// //
// // ***** 测试用
// //
// // 实现ServerWriter接口
// // 用于向http服务器写入事件数据
// func (self *ServerHttp) Write(data *fwsdef.EventDataT) error {
// 	//
// 	// 测试代码，写入本地先
// 	//
// 	TEST := true
// 	if TEST {
// 		timeNow := time.Now().UnixNano()
// 		pathPic := fwsconf.GetFPicStorPath()
// 		_, err := os.Stat(pathPic)
// 		if err != nil {
// 			if os.IsNotExist(err) {
// 				err = os.MkdirAll(pathPic, os.ModePerm)
// 				if err != nil {
// 					panic(err)
// 				}
// 			} else {
// 				panic(err)
// 			}
// 		}
// 		// 描述信息
// 		descFile := pathPic + fmt.Sprintf("%d.txt", timeNow)
// 		f, err := os.Create(descFile)
// 		if err != nil {
// 			gLog.Errorf("create file[%s] failed: %s", descFile, err.Error())
// 			return err
// 		}
// 		f.Write(data.DescBuf)
// 		f.Close()
// 		gLog.Infof("write desc file[%s] ok.", descFile)

// 		// 图片
// 		picFile := pathPic + fmt.Sprintf("%d.bmp", timeNow)
// 		f, err = os.Create(picFile)
// 		if err != nil {
// 			gLog.Errorf("create file[%s] failed: %s", picFile, err.Error())
// 			return err
// 		}
// 		f.Write(data.PicBuf)
// 		f.Close()
// 		gLog.Infof("write picture file[%s] ok.", picFile)

// 		// url
// 		urlFile := pathPic + fmt.Sprintf("%d.url", timeNow)
// 		f, err = os.Create(urlFile)
// 		if err != nil {
// 			gLog.Errorf("create file[%s] failed: %s", urlFile, err.Error())
// 			return err
// 		}
// 		b, err := json.Marshal(self)
// 		if err != nil {
// 			gLog.Errorf("json.Marshal failed of http server, %v", self)
// 		} else {
// 			f.Write(b)
// 			gLog.Infof("write picture file[%s] ok.", urlFile)
// 		}
// 		f.Close()
// 	}
// 	return nil
// }
