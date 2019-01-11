#ifndef FWDRIVER_H_
#define FWDRIVER_H_

class CFWDriver {
public:
    CFWDriver(){

    }
    virtual ~CFWDriver(){

    }

    virtual bool Init()=0;
	virtual int Influx(char *szDesc, unsigned int uiDescLen, void* pPic, unsigned int uiPicLen, \
                char** strUrls, unsigned int nUrlCount)=0;
	virtual bool Cleanup()=0;
protected:

};

#endif //FWDRIVER_H_
