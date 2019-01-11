#include "fwd_go.h"

CFwdGo::CFwdGo(){
    m_fFWGO_Init = NULL;
    m_fFWGO_Influx = NULL;
    m_fFWGO_Cleanup = NULL;
    m_hDLL = NULL;
}
CFwdGo::~CFwdGo(){
    m_fFWGO_Init = NULL;
    m_fFWGO_Influx = NULL;
    m_fFWGO_Cleanup = NULL;
    m_hDLL = NULL;
}

bool CFwdGo::Init(){
    #ifdef _WIN32
    m_hDLL = LoadLibrary(_T("libfwapp_go.dll"));
    if(m_hDLL == NULL){
        printf("LoadLibrary [libfwapp_go.dll] failed!\n");
        return false;
    }
    // 获取接口
    m_fFWGO_Init = (FWGO_Init_T)GetProcAddress(m_hDLL, "FW_Init");
    if (m_fFWGO_Init == NULL) {
		printf("could not find function: FW_Init\n");
        return false;
	}

	m_fFWGO_Influx = (FWGO_Influx_T)GetProcAddress(m_hDLL, "FW_Influx");
    if (m_fFWGO_Influx != NULL) {
		printf("could not find function: FW_Influx\n");
        return false;
	}

    m_fFWGO_Cleanup = (FWGO_Cleanup_T)GetProcAddress(m_hDLL, "FW_Cleanup");
    if (m_fFWGO_Cleanup != NULL) {
        printf("could not find function: FW_Cleanup\n");
        return false;
	}
    #else
    m_fFWGO_Init = FWGO_Init;
    m_fFWGO_Influx = FWGO_Influx;
    m_fFWGO_Cleanup = FWGO_Cleanup;
    #endif
    int ret = m_fFWGO_Init();
    if( ret != 0){
        return false;
    }
    return true;
}

int CFwdGo::Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, \
            char** strUrls, unsigned int nUrlCount){
    if(uiDescLen <= 0 || uiPicLen <= 0 || nUrlCount <= 0 || m_fFWGO_Influx == NULL){
        return -1;
    }
    GoSlice desc, pic, urls;

    desc.data = (void*)szDesc;
    desc.len = uiDescLen;
    desc.cap = desc.len;

    pic.data = pPic;
    pic.len = uiPicLen;
    pic.cap = pic.len;

    urls.len = 256 * nUrlCount;

    urls.data = (void*)new char[urls.len];
    if (urls.data == NULL){
        printf("new memory failed!\n");
        return -1;
    }
    for(unsigned int i=0; i<nUrlCount; i++){
        memcpy((char*)urls.data+i*256, strUrls[i], strlen(strUrls[i])*sizeof(char));
    }
    urls.cap = urls.len;

    int ret = m_fFWGO_Influx(desc, desc.len, pic, pic.len, urls, urls.len);
    if(urls.data != NULL){
        char* p=(char*)urls.data;
        delete [] p;
        urls.data = NULL;
        p = NULL;
    }
    if (ret != 0){
        return -1;
    }
    return 0;
}

bool CFwdGo::Cleanup(){
    int ret = 0;
    if (m_fFWGO_Cleanup != NULL){
        ret = m_fFWGO_Cleanup();
    }
    if(m_hDLL != NULL){
        #ifdef _WIN32
        FreeLibrary(m_hDLL);
        #else

        #endif
        m_hDLL = NULL;
    }
    m_fFWGO_Init = NULL;
    m_fFWGO_Influx = NULL;
    m_fFWGO_Cleanup = NULL;
    return (ret==0);
}
