#include "../libfwapp/include/fwapi.h"
#ifdef _WIN32
#include <windows.h>
#else
#endif

#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <fstream>
#include <sys/types.h>
#include <math.h>
#include <iostream>
#include <signal.h>
#include <sys/time.h>
#include <fcntl.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <string.h>
using namespace std;
// 程序是否退出标示
volatile bool g_bExit = false;

int getFileSize(char* szFile){
    int size = -1;

    #ifdef _WIN32

    FILE *fp=fopen(szFile, "r");
    if(fp==NULL) {
        return -1;
    }
    fseek(fp, 0L, SEEK_END);
    size = ftell(fp);
    fclose(fp);

    #else

    struct stat st;
    if(lstat(szFile, &st) < 0){
        printf("lstat file[%s] failed ", szFile);
        perror("");
        return -1;
    }
    size = st.st_size;

    #endif
    return size;
}

// 事件数据
typedef struct _structEventDataT{
    char* pDesc;
    unsigned int uiDescLen;
    char* pPic;
    unsigned int uiPicLen;
    char** pUrls;
    unsigned int uiUrlsCount;
}EventDataT;

EventDataT* gpEDList = NULL;
int giEDListCount = 0;

int GetEventData(){
    if (gpEDList != NULL){
        return 0;
    }
    #define FILE_COUNT  3
    char szPicList[FILE_COUNT][80]={"pic/1.jpg","pic/2.jpg","pic/3.jpg"};
    char szDescList[FILE_COUNT][80]={"pic/1.txt","pic/2.txt","pic/3.txt"};
    #define URL_COUNT 1
    char szUrlList[URL_COUNT][256]={
        // "http://192.168.1.205/api/data/v1/forward/push"
        "http://192.168.1.218:6688/push"
        };
    giEDListCount = FILE_COUNT;
    gpEDList = new EventDataT[giEDListCount];
    if (gpEDList==NULL) {
        printf("memory failed!\n");
        exit(1);
    }
    for(int i=0; i<FILE_COUNT; i++){
        gpEDList[i].uiDescLen = getFileSize(szDescList[i]);
        gpEDList[i].uiPicLen = getFileSize(szPicList[i]);
        gpEDList[i].uiUrlsCount = URL_COUNT;
        gpEDList[i].pDesc = new char[gpEDList[i].uiDescLen];
        if (gpEDList[i].pDesc == NULL){
            printf("memory failed!\n");
            exit(1);
        }
        FILE* fdesc = fopen(szDescList[i], "r");
        if (fdesc == NULL){
            printf("open file[%s] failed.", szDescList[i]);
            exit(1);
        }
        fread(gpEDList[i].pDesc, 1, gpEDList[i].uiDescLen, fdesc);
        fclose(fdesc);
        fdesc = NULL;

        gpEDList[i].pPic = new char[gpEDList[i].uiPicLen];
        if (gpEDList[i].pPic == NULL){
            printf("memory failed!\n");
            exit(1);
        }
        FILE* fpic = fopen(szPicList[i], "r");
        if (fpic == NULL){
            printf("open file[%s] failed.\n", szPicList[i]);
            exit(1);
        }
        fread(gpEDList[i].pPic, 1, gpEDList[i].uiPicLen, fpic);
        fclose(fpic);

        gpEDList[i].pUrls = new char*[gpEDList[i].uiUrlsCount];
        if (gpEDList[i].pUrls == NULL){
            printf("memory failed!\n");
            exit(1);
        }
        for(unsigned int n=0; n<gpEDList[i].uiUrlsCount; n++){
            int size = strlen(szUrlList[n]);
            gpEDList[i].pUrls[n] = new char[size];
            if (gpEDList[i].pUrls[n] == NULL){
                printf("memory failed\n");
                exit(1);
            }
            sprintf(gpEDList[i].pUrls[n], "%s", szUrlList[n]);
        }
    }
    return 0;
}

int main(int argc, char** argv){
    int MAX_ED_COUNT = 1;
    if (argc > 1){
        int n = atoi(argv[1]);
        if (n > 0){
            MAX_ED_COUNT = n;
            printf("get arg = %d\n", MAX_ED_COUNT);
        }
    }
    GetEventData();

    bool bInit = FW_Init();
    if( bInit == false ) {
        printf("init failed!\nexit!\n");
        return -1;
    }
    printf("Version: %s\n", FW_Version());

    int index = 0;
    int iRet = 0;
    char byInput = 0;
    while(g_bExit==false){
        for(int i=0; i<MAX_ED_COUNT; i++){
            if(g_bExit){
                break;
            }
            index = i % giEDListCount;
            iRet = FW_Influx(gpEDList[index].pDesc, gpEDList[index].uiDescLen, gpEDList[index].pPic, gpEDList[index].uiPicLen, \
            gpEDList[index].pUrls, gpEDList[index].uiUrlsCount);
            if (iRet != 0){
                printf("FW_Influx failed [NO.%d]\n", i);
            }
            #ifdef _WIN32
            Sleep((index+1)*100);
            #else
            usleep((index+1)*100*1000);
            #endif
        }
        byInput = getchar();
        /*
         * 按q回车退出程序，直接回车或其他字符后回车，则继续发送MAX_ED_COUNT个事件数据
         */
        if(byInput == 'q'){
            g_bExit = true;
            break;
        }
    }

    while ( g_bExit == false )
	{
		usleep(100*1000);
	}
    //FW_Test();
    FW_Cleanup();
    printf("#### Exit! ####\n\n");
    return 0;
}

// end of file

