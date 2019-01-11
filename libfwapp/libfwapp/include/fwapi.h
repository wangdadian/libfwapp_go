#ifndef FWAPI_H_20181225
#define FWAPI_H_20181225

#if (defined(_WIN32)) //windows
	#ifdef FW_API_EXPORTS
		#define FW_API extern "C" __declspec(dllexport)
	#else
		#define FW_API extern "C" __declspec(dllimport)
	#endif
#elif defined(__linux__) //linux
#define __stdcall
//#define FW_API extern "C"
#define FW_API
#endif


// 初始化
// type-指定底层驱动
//      0-采用libcurl（未实现，20181228）
//      1-采用go编写的库（默认模式）
// return  true-成功，false-失败
FW_API bool __stdcall FW_Init(int type=1);

// 数据传入
// 接口返回前，模块内部会复制传入的内存数据，不再依赖外部传入的内存空间，所以.....接口返回后自行销毁
// return: 0-成功，其他-失败
FW_API int __stdcall FW_Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, char** strUrls, unsigned int nUrlCount);

// 清理数据、结束任务，不再使用本模块时调用
// return： true-成功，false-失败
FW_API bool __stdcall FW_Cleanup();

// 返回版本信息
FW_API char* __stdcall FW_Version();

// 测试用接口
FW_API void __stdcall FW_Test();


#endif // FWAPI_H_20181225



