/*
 * @Author: xiongshanshan
 * @Date: 2017-11-22
 * @Last Modified by: xiongshanshan
 * @Last Modified time: 2018-10-10 10:46:37
 */

#ifndef CHAINEDBOX_STRUCT_H
#define CHAINEDBOX_STRUCT_H

#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum
{
    False = 0,
    True = 1
} Boolean;

typedef enum 
{
    ENCRYPT_SRC = -1,
    ENCRYPT_NO = 0,
    ENCRYPT_AES_ECB = 1,
    ENCRYPT_AES_CBC = 2
} ENCRYPTTYPE;

typedef enum
{
    SINGLESTATUS_UNKNOWN = 0,
    SINGLESTATUS_WAITING,
    SINGLESTATUS_QUEUEING,
    SINGLESTATUS_TASKBEGIN,
    SINGLESTATUS_TASKNOWIFI,
    SINGLESTATUS_UPLOADING,
    SINGLESTATUS_TASKFAILED,
    SINGLESTATUS_UPLOADCOMPLETE,
    SINGLESTATUS_DOWNLOADING,
    SINGLESTATUS_DOWNLOADCOMPLETE
} FileTaskStatus;

typedef enum
{
    THUMBNAIL_ORI = 0,
    THUMBNAIL_200 = 1
} THUMBNAILTYPE;

typedef enum
{
    NETWORK_NONE = 0,
    NETWORK_3G = 1,
    NETWORK_WIFI = 2
} NETWORKTYPE;

typedef struct
{
    FileTaskStatus fileStatus;
    int speed;
    int percent;
    int fidLength;
    int errorMsgLength;
    char *fid;
    char *errorMsg;
} yhCallback;

#ifdef __cplusplus
}
#endif

#endif //CHAINEDBOX_STRUCT_H