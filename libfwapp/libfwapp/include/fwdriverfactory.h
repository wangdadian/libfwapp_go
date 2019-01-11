#ifndef FWDRIVER_FACTORY_H_2018
#define FWDRIVER_FACTORY_H_2018

#include "fwdriver.h"
#include "fwd_go.h"
#include "fwd_curl.h"
#include <stdlib.h>
class CFWDriverFactory {
public:
    static CFWDriver* CreateFWDriver(int driver){
        CFWDriver* pFWDriver = NULL;
        switch(driver){
            case 0:
                pFWDriver = new CFwdUrl();
                break;
            case 1:
                pFWDriver = new CFwdGo();
                break;
            default:
                printf("unsupport driver type[%d]\n", driver);
                exit(1);
        }
        return pFWDriver;
    }
private:
    CFWDriverFactory(){

    }
    ~CFWDriverFactory(){

    }
};

#endif // FWDRIVER_FACTORY_H_2018
