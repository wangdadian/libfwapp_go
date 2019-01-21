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

	// 写入事件描述信息
	// 内容头设置
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", "form-data; name=\"vehicle\";")
	// h.Set("Content-Type", "text/plain")
	w, err := mpWr.CreatePart(h)
	if err != nil {
		gLog.Errorf("mpWr.CreatePart(h)1 failed: %s", err)
		return err
	}
	// 写入描述信息
	n := 0
	n, err = w.Write(data.DescBuf)
	if err != nil {
		gLog.Errorf("Write failed: %s", err.Error())
		return err
	}
	// 判断写入字节数是否和实际相等，是否写入完成
	if n != len(data.DescBuf) {
		gLog.Errorf("Write failed: write len=%d != %d", n, len(data.DescBuf))
		return fmt.Errorf("Write failed: write len=%d != %d", n, len(data.DescBuf))
	}
	gLog.Infof("Write Desc ok: write len=%d == %d", n, len(data.DescBuf))
	// 如果有图片数据则写入图片数据
	if len(data.PicBuf) > 0 {
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
		gLog.Infof("Write Pic ok: write len=%d == %d", n, len(data.PicBuf))
	}
	// 保存边界字符串，用于设置请求头
	strBoundary := mpWr.Boundary()
	// 多部写入缓冲关闭，自动加入边界
	mpWr.Close()

	// 建立请求客户端，并设置2秒的超时临界值，避免阻塞
	pClient := &http.Client{
		Timeout: 2000 * time.Millisecond,
	}

	// 建立请求，并将缓冲区作为body
	req, err := http.NewRequest("POST", self.Url, &bufRWer)
	if err != nil {
		gLog.Errorf("http.NewRequest failed: %s", err.Error())
		return err
	}
	defer req.Body.Close()

	// 设置请求头部信息
	req.Header.Add("HOST", req.Host)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+strBoundary)
	gLog.Infof("start to write to http server[%s]...", self.Url)
	// 开始请求
	res, err := pClient.Do(req)
	if err != nil {
		gLog.Errorf("http Client Do failed: %s", err.Error())
		return err
	}
	defer res.Body.Close()

	// gLog.Errorf("############  ReadAll")
	// 读取对端返回信息，判断是否发送成功
	byRetBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		gLog.Errorf("ioutil.ReadAll(res.Body) failed: %s", err.Error())
		return err
	}

	// 收到返回消息，解析后判断是否发送成功
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
	// 发送成功
	gLog.Infof("http server[%s] return message, it's ok", self.Url)
	return nil
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
