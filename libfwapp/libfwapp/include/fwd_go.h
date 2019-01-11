#ifndef FW_GO_DRIVER_H_
#define FW_GO_DRIVER_H_

#include "fwdriver.h"
#include <stdio.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>
#include <tchar.h>
typedef struct { void *data; long long len; long long cap; } GoSlice;
#else
#include "libfwapp_go.h"
#endif

typedef int (*FWGO_Init_T)();
typedef int (*FWGO_Influx_T)(GoSlice, unsigned int, GoSlice, unsigned int, GoSlice,unsigned int );
typedef int (*FWGO_Cleanup_T)();

class CFwdGo: public CFWDriver{
public:
    CFwdGo();
    ~CFwdGo();

    bool Init();
	int Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, \
                char** strUrls, unsigned int nUrlCount);
	bool Cleanup();
private:


private:
    FWGO_Init_T m_fFWGO_Init;
    FWGO_Influx_T m_fFWGO_Influx;
    FWGO_Cleanup_T m_fFWGO_Cleanup;
    #ifdef _WIN32
    HMODULE m_hDLL;
    #else
    void* m_hDLL;
    #endif
};

#endif // FW_GO_DRIVER_H_
