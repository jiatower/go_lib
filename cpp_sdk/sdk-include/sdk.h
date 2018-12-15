/*
 * @Author: xiongshanshan
 * @Date: 2017-11-15
 * @Last Modified by: xiongshanshan
 * @Last Modified time: 2018-10-10 10:45:46
 */

#ifndef SDK_H
#define SDK_H

#include "struct.h"

#ifdef __cplusplus
extern "C" {
#endif

Boolean yhInitSDKManager(const char *workDir, const char *ip, const char *gwmac, NETWORKTYPE networkType, int testenv);

// yhInitSDKManager之后须先调用yhSetInternalHost，然后才能调用yhSetYHAPI
void yhSetInternalHost(const char *host);

void* yhSetYHAPI(const char *appid, const char *appuid, const char *devid, const char *sid, const char *clusterid);

void yhCloseYHAPI(void *api);

void yhCloseSDKManager();

char* yhCreateDir(void *api, const char *parentFid, const char *dirName); 

long long yhUploadFid(void *api, const char *localPath, const char *parentFid, const char *name, ENCRYPTTYPE encryptType,  
            Boolean onlyWifi, yhCallback *callback, const char *ownerAppid, const char *ownerAppuid, const char *local_md5);

long long yhUploadFile(void *api, const char *localPath, const char *remotePath, ENCRYPTTYPE encryptType, Boolean onlyWifi, Boolean force,
            yhCallback *callback, const char *appid, const char *appuid, Boolean isOwner, const char *cert, const char *local_md5, const char *tags);

long long yhDownloadFile(void *api, const char *localPath, const char *fid, Boolean onlyWifi, yhCallback *callback);

char* yhGetUrl(void *api, const char *fid, Boolean onlyWifi, long long ver, const char *path);

char* yhGetThumbUrl(void *api, const char *fid, THUMBNAILTYPE type);

void deleteYhCallback(yhCallback *callback);

void yhOpenLog();

void yhCloseLog();

#ifdef __cplusplus
}
#endif

#endif //SDK_H
