#include "fwapi.h"
#include <string.h>
#include <stdlib.h>
#include "fwdriverfactory.h"
#include "fwdriver.h"


// 唯一实例
CFWDriver* gFwApp = NULL;

// 初始化
bool __stdcall FW_Init(int type){
    // 已经初始化，无需再次初始化
    if( gFwApp != NULL ){
        //printf("FW_Init: no need to call me again!\n");
        return true;
    }
    //
    // 初始化
    //
    gFwApp = CFWDriverFactory::CreateFWDriver(type);
    if(gFwApp == NULL){
        printf("FW_Init: failed to allocate memory.");
        exit(1);
        return false;
    }

	bool bInit = gFwApp->Init();
	if (bInit == false){
        //
		return false;
	}

    //printf("FW_Init: Hi!\n");
    return true;
}

// 输入数据
int __stdcall FW_Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, \
                        char** strUrls, unsigned int nUrlCount){
    // printf("FW_Influx: describe info len: %d, picture size: %d\n", uiDescLen, uiPicLen);
    if(gFwApp == NULL){
        //printf("FW_Influx: call FW_Init() first.");
        return -1;
    }
	int iRet = gFwApp->Influx(szDesc, uiDescLen, pPic, uiPicLen, strUrls, nUrlCount);
	if( iRet != 0 ){
		//
        return iRet;
	}

    return 0;
}

// 清理
bool __stdcall FW_Cleanup(){
    if( gFwApp != NULL){
		gFwApp->Cleanup();
        delete gFwApp;
        gFwApp = NULL;
    }
    //printf("FW_Cleanup: bye!\n");

    return true;
}

// 打印版本
char* __stdcall FW_Version(){
	static char szInfo[256] = {0};
	memset(szInfo, 0, 256 * sizeof(char));
	char szVer[80]={0};
	char szDate[80]={0};

    #if defined(VERSION) || defined(BUILDDATE)
	// 版本
    #ifdef VERSION
    //printf("Version : %s\n", VERSION);
	sprintf(szVer, "%s",VERSION);
    #endif

	// 编译日期
    #ifdef BUILDDATE
    //printf("Build at: %s\n", BUILDDATE);
	sprintf(szDate, "%s",BUILDDATE);
    #endif
    sprintf(szInfo, "%s,%s\n", szVer, szDate);
	#else
    sprintf(szInfo, "%s", "");
    #endif
    return szInfo;
}

// 测试
void  __stdcall FW_Test(){
    // print a heart
    # define U 0.12
	# define V 0.05
	# define M 1.1
	# define N 1.2
	float x, y;
	for ( y=2; y>=-2; y-=U ){
		for ( x=-1.2; x<=1.2; x+=V){
			if ( ( ( (x*x + y*y - 1)*(x*x + y*y - 1)*(x*x + y*y - 1) - x*x*y*y*y ) <= 0 ) ) {
				printf("*");
            }else{
				printf(" ");
            }
		}
		printf("\n");
	}
    printf("FW_Test(): don't call me again!\n");
    exit(1);
}
