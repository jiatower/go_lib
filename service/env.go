package service

import (
	"errors"
	"net/http"

	"github.com/jiatower/go_lib/log"
	"github.com/jiatower/go_lib/utils"
)

type Env struct {
	Log       *log.MLogger
	ModuleEnv interface{}
}

type Session struct {
	Uid uint32
	Sid string
}

func NewEnv(data interface{}) *Env {
	return &Env{nil, data}
}

type Config struct {
	IpPort      string
	LogDir      string
	LogLevel    string
	GetEnv      func(module string) *Env
	IsValidUser func(r *http.Request) (s *Session, e error) //如果不是合法用户，需要返回""
}

type Result struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Detail string      `json:"detail"`
	Code   uint        `json:"code"`
	Res    interface{} `json:"res"`
	Unread interface{} `json:"unread"`
}

func NewResult() Result {
	return Result{"ok", "", "", ERR_NOERR, map[string]interface{}{}, map[string]interface{}{}}
}

func (r *Result) Set(key string, value interface{}) error {
	switch v := r.Res.(type) {
	case map[string]interface{}:
		v[key] = value
		return nil
	default:
		return errors.New("res type is not map[string]interface{}")
	}
}

func (r *Result) Get(key string) (value interface{}, found bool) {
	switch v := r.Res.(type) {
	case map[string]interface{}:
		value, found = v[key]
		return
	default:
		return value, false
	}
}

func (r *Result) GetString(key string) (value string, found bool) {
	v, f := r.Get(key)
	return utils.ToString(v), f
}

func (r *Result) GetBool(key string) (value bool, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToBool(v)
	return
}

func (r *Result) GetInt(key string) (value int, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToInt(v)
	return
}

func (r *Result) GetUint64(key string) (value uint64, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToUint64(v)
	return
}

func (r *Result) GetUint32(key string) (value uint32, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToUint32(v)
	return
}

func (r *Result) GetInt64(key string) (value int64, found bool, e error) {
	v, found := r.Get(key)
	value, e = utils.ToInt64(v)
	return
}

func (r *Result) CheckIsOk() (ok bool) {
	if r.Status == RESULT_STATE_OK {
		return true
	}
	return false
}
