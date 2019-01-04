package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jiatower/go_lib/log"
	"github.com/jiatower/go_lib/utils"
)

//Env 环境变量基类
type Env struct {
	Log       *log.MLogger
	ModuleEnv interface{}
}

//Session 会话对象
type Session struct {
	Uid uint32
	Sid string
}

//NewEnv 创建环境变量
func NewEnv(data interface{}) *Env {
	return &Env{nil, data}
}

//Config 配置对象
type Config struct {
	IPPort      string
	LogDir      string
	LogLevel    string
	GetEnv      func(module string) *Env
	IsValidUser func(r *http.Request) (s *Session, e error) //如果不是合法用户，需要返回""
}

//Result 返回结果的基础结构
type Result struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Detail string      `json:"detail"`
	Code   uint        `json:"code"`
	Res    interface{} `json:"res"`
}

//NewResult 创建默认的Result对象
func NewResult() Result {
	return Result{"ok", "", "", ERR_NOERR, map[string]interface{}{}}
}

func (r *Result) String() string {
	return fmt.Sprintf("status=%s, code=%d, detail=%s, msg=%s, res=%v", r.Status, r.Code, r.Detail, r.Msg, r.Res)
}

//Set 设置Result的某一项值
func (r *Result) Set(key string, value interface{}) error {
	switch v := r.Res.(type) {
	case map[string]interface{}:
		v[key] = value
		return nil
	default:
		return errors.New("res type is not map[string]interface{}")
	}
}

//Get 获取Result的某个值
func (r *Result) Get(key string) (value interface{}, found bool) {
	switch v := r.Res.(type) {
	case map[string]interface{}:
		value, found = v[key]
		return
	default:
		return value, false
	}
}

//GetString 获取字符串类型的返回值
func (r *Result) GetString(key string) (value string, found bool) {
	v, f := r.Get(key)
	return utils.ToString(v), f
}

//GetBool 获取布尔类型的返回值
func (r *Result) GetBool(key string) (value bool, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToBool(v)
	return
}

//GetInt 获取整型的返回值
func (r *Result) GetInt(key string) (value int, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToInt(v)
	return
}

//GetUint64 获取无符号长整型的返回值
func (r *Result) GetUint64(key string) (value uint64, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToUint64(v)
	return
}

//GetUint32 获取无符号整型的返回值
func (r *Result) GetUint32(key string) (value uint32, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToUint32(v)
	return
}

//GetInt64 获取长整型的返回值
func (r *Result) GetInt64(key string) (value int64, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToInt64(v)
	return
}

//CheckIsOk 判断返回值是否是成功
func (r *Result) CheckIsOk() (ok bool) {
	if r.Status == RESULT_STATE_OK {
		return true
	}
	return false
}

//ToError 把result转换成error，如果返回结果没有错误，返回nil
func (r *Result) ToError() (e error) {
	if r.CheckIsOk() {
		return nil
	} else {
		return errors.New(r.String())
	}
}
