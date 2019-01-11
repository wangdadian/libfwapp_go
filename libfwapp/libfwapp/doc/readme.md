# libfwapp

## 1. Functions
*****
### 1.1. FW_Init
// 初始化
// type-指定底层驱动
//      0-采用libcurl（未实现，20181228）
//      1-采用go编写的库（默认模式）
// return  true-成功，false-失败
FW_API bool __stdcall FW_Init(int type=1);
*****
### 1.2. FW_Influx
// 数据传入
// 接口返回前，模块内部会复制传入的内存数据，不再依赖外部传入的内存空间，所以.....接口返回后自行销毁
// return: 0-成功，其他-失败
FW_API int __stdcall FW_Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, char** strUrls, unsigned int nUrlCount);
*****
### 1.3. FW_Cleanup
// 清理数据、结束任务，不再使用本模块时调用
// return： true-成功，false-失败
FW_API bool __stdcall FW_Cleanup();
*****
### 1.4. FW_Version
// 返回版本信息
FW_API char* __stdcall FW_Version();
*****

## 2. Usage
    header:
        fwapi.h
    compile:
        g++/gcc ..... -lfwapp -lfwapp_go -lpthread


![blockchain](https://ss1.bdstatic.com/70cFuXSh_Q1YnxGkpoWK1HF6hhy/it/u=3211196576,548193681&fm=15&gp=0.jpg)
*****