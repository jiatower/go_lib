package service

import "fmt"

const (
	SERVER_REDIRECT    = 302 // 302跳转特殊错误码
	SERVER_WRITE_IMAGE = 303 //输出图片
	SERVER_WRITE_TEXT  = 304 //输出文本
	SERVER_WRITE_XML   = 305 //输出xml格式
	SERVER_WRITE_MP4   = 306 //输出mp4视频
	SERVER_WRITE_DOC   = 331
	SERVER_WRITE_DOCX  = 332
	SERVER_WRITE_XLS   = 333
	SERVER_WRITE_XLSX  = 334
	SERVER_WRITE_PPT   = 335
	SERVER_WRITE_PPTX  = 336
	SERVER_WRITE_PDF   = 337

	ERR_NOERR      = 0    //没有错误
	ERR_UNKNOWN    = 1001 //未知错误
	ERR_INTERNAL   = 1002 //内部错误
	ERR_MYSQL      = 1003 //mysql错误
	ERR_REDIS      = 1004 //redis错误
	ERR_NOT_FOUND  = 1005 //未找到
	ERR_INVALID_IP = 1006 //ip非法

	ERR_INVALID_PARAM         = 2001 //请求参数错误
	ERR_INVALID_FORMAT        = 2002 //格式错误
	ERR_ENCRYPT_ERROR         = 2003 //加密错误
	ERR_INVALID_REQUEST       = 2004 //不合法的请求
	ERR_VERIFY_FAIL           = 2005 //验证失败
	ERR_VCODE_TIMEOUT         = 2006 //验证码超时
	ERR_INVALID_USER          = 2007 //用户验证不通过
	ERR_PERMISSION_DENIED     = 2008 //权限不足
	ERR_VCODE_ERROR           = 2009 //验证码错误
	ERR_TOO_MANY              = 2010 //次数过多
	ERR_IN_BLACKLIST          = 2011 //在黑名单中
	ERR_MUSTINFO_NOT_COMPLETE = 2012 //必填项没有填写完整
	ERR_INVALID_IMG           = 2013 // 图片检查未通过
	ERR_NOT_ONLINE            = 2014 //不在线
	ERR_TIMEOUT               = 2015 //超时
	ERR_INVALID_CERTIFICATE   = 2016 //证书不合法
	ERR_INVALID_PWD           = 2017 //密码错误
	ERR_USER_EXIST            = 2018 //帐号已存在
	ERR_USER_NOT_EXIST        = 2019 //帐号不存在
	ERR_SID_TIMEOUT           = 2020 //SID过期
	ERR_REMOVE_BIND           = 2021 //解绑手机和邮箱错误（解绑唯一帐号时）
	ERR_CERTIFICATE_TIMEOUT   = 2022 //证书过期
	ERR_DELETE_FILE_TIPS      = 2023 //删除存储，存在未备份文件提示
	ERR_DEVICE_FORBID         = 2024 //登录设备被禁止，需要重新登录
	ERR_DELETE_USER_NOFORBID  = 2025 //需要先禁用，在删除
	ERR_NO_SPACE              = 2026 //剩余空间不足
	ERR_NO_DISK               = 2027 //未接入磁盘
	ERR_LIMIT_USER            = 2028 //超过人数上线

	ERR_BINDDEV_INVALID = 90001 //绑定设备不可用
	ERR_AUTH_TIMEOUT    = 90004 //第三方授权失效
	ERR_UPDATE_SID      = 90005 //网络环境变化，需要更换通信sid
	ERR_USER_REBIND     = 90006 //用户绑定时间改变
	ERR_TOO_MANY_DEVICE = 90007 //用用户设备绑定太多

	//小程序相关错误码
	ERR_WECHAT_USER_LIMIT     = 96001 // 微信成员超限制
	ERR_WECHAT_ALBUM_CANCELED = 96002 // 影集取消共享
	ERR_WECHAT_ALBUM_INFO     = 96003 // 从servApp获取影集信息失败
	ERR_WECHAT_UPLOAD_FAIL    = 96004 // 文件上传失败
	ERR_WECHAT_API_FAIL       = 96005 // 文件上传失败

	ERR_POP_NOTIFY = 10001 //特殊错误码，需要解析desc字段，并做弹窗处理

	//p2p 相关错误码
	ERR_P2P_FILE_NOT_FOUND        = 300001 //文件不存在
	ERR_P2P_TASK_OTHER_NODE_DOING = 300002 //任务其他节点正在完成
	ERR_P2P_FILE_ALREADY_EXIST    = 300003 // 文件已经添加了

	//oss 相关错误码
	ERR_OSS_FILE_DELETE    = 200001 // 文件被删除
	ERR_HTTP_SERVICE_ERROR = 200002
)

//文档类文件响应头的content-type映射
var DocContentTypeMap map[uint]string = map[uint]string{
	SERVER_WRITE_DOC:  "application/msword",
	SERVER_WRITE_DOCX: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	SERVER_WRITE_XLS:  "application/vnd.ms-excel",
	SERVER_WRITE_XLSX: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	SERVER_WRITE_PPT:  "application/vnd.ms-powerpoint",
	SERVER_WRITE_PPTX: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	SERVER_WRITE_PDF:  "application/pdf",
}

var DocExtRespCodeMap map[string]uint = map[string]uint{
	"DOC":  SERVER_WRITE_DOC,
	"DOCX": SERVER_WRITE_DOCX,
	"XLS":  SERVER_WRITE_XLS,
	"XLSX": SERVER_WRITE_XLSX,
	"PPT":  SERVER_WRITE_PPT,
	"PPTX": SERVER_WRITE_PPTX,
	"PDF":  SERVER_WRITE_PDF,
}

//通用返回结果ok，fail
const (
	RESULT_STATE_OK   = "ok"   //返回ok
	RESULT_STATE_FAIL = "fail" //返回失败
)

// 常用result-> msg 返回信息
const (
	MSG_DEF               = "系统繁忙"
	MSG_INVALID_USER      = "用户验证失败"
	MSG_INVALID_PARAM     = "参数错误"
	MSG_MYSQL_ERROR       = "数据获取失败"
	MSG_REDIS_ERROR       = "获取缓存数据失败"
	MSG_INVALID_CERTIFY   = "证书错误"
	MSG_CERTIFY_TIMEOUT   = "证书失效"
	MSG_PERMISSION_DENIED = "无权限"
	MSG_VCODE_ERROR       = "验证码错误"
	MSG_INVALID_PWD       = "密码错误"

	MSG_STORAGE_MAC_NOT_FOUND        = "未找到该设备信息，请联系在线客服。"
	MSG_STORAGE_RCODE_NOTMATCH_MODEL = "客户端与设备类型不匹配"
)

//302跳转key
const (
	SERVER_REDIRECT_KEY = "redirect_url"
)

type Error struct {
	Code uint
	Desc string
	Show string //客户端显示的内容
}

func NewError(ecode uint, desc string, show ...string) (err Error) {

	if len(show) > 0 {
		err = Error{ecode, desc, show[0]}
	} else {
		switch ecode {
		case ERR_INVALID_PARAM:
			err = Error{ecode, desc, "参数错误"}
		case ERR_INVALID_REQUEST:
			err = Error{ecode, desc, "不合法的请求"}
		case ERR_MYSQL, ERR_REDIS:
			err = Error{ecode, desc, "数据库错误"}
		default:
			err = Error{ecode, desc, "内部错误"}
		}
	}
	return
}

func NewSimpleError(ecode uint, desc string) (err Error) {
	return NewError(ecode, desc, desc)
}

func (e Error) Error() (re string) {
	return fmt.Sprintf("ecode=%v, desc=%v", e.Code, e.Desc)
}
