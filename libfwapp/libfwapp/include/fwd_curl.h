#ifndef FW_CURL_DRIVER_H_
#define FW_CURL_DRIVER_H_

#include "fwdriver.h"
#include <stdio.h>
class CFwdUrl: public CFWDriver {
public:
    CFwdUrl(){

    }
    ~CFwdUrl(){

    }

    bool Init();
	int Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, \
                char** strUrls, unsigned int nUrlCount);
	bool Cleanup();
protected:

};

#endif // FW_CURL_DRIVER_H_
