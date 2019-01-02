package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/jiatower/go_lib/utils"
)

const MAX_PS = 1000

//HTTPRequest 封装的Http请求
type HTTPRequest struct {
	body    map[string]interface{}
	BodyRaw []byte
	request *http.Request
	Session *Session
}

//GetRequest 获取HTTPRequest对象
func (hr *HTTPRequest) GetRequest() *http.Request {
	return hr.request
}

//GetParam 获取name参数值
func (hr *HTTPRequest) GetParam(name string) string {
	return hr.request.URL.Query().Get(name)
}

//IP 获取请求者IP地址
func (hr *HTTPRequest) IP() string {
	ips := strings.Split(hr.request.Header.Get("X-Forwarded-For"), ",")
	if len(ips[0]) > 3 {
		return ips[0]
	} else {
		addr := strings.Split(hr.request.RemoteAddr, ":")
		return addr[0]
	}
}

//GetCookie 获取名称为name的cookie值
func (hr *HTTPRequest) GetCookie(name string) (*http.Cookie, error) {
	/*	fmt.Println("getCooike: --" + hr.request.Header.Get("Cookie"))
		fmt.Println("header---" + utils.ToString(hr.request.Header))
		fmt.Println(hr.request.Cookies())
	*/
	return hr.request.Cookie(name)
}

//GetCookieStr 获取某cookie的值
func (hr *HTTPRequest) GetCookieStr(name string) (string, error) {
	c, e := hr.GetCookie(name)
	if e != nil {
		return "", e
	}
	return c.Value, nil
}

//GetAppChannel 获取App的渠道
func (hr *HTTPRequest) GetAppChannel() string {
	ch, e := hr.GetCookieStr("channel")
	if e != nil {
		ch = "official"
	}
	return ch
}

//Body 获得map类型的json body
func (hr *HTTPRequest) Body() (map[string]interface{}, error) {
	if hr.body == nil {
		hr.body = make(map[string]interface{})
		if e := json.Unmarshal(hr.BodyRaw, &hr.body); e != nil {
			return nil, NewError(ERR_INVALID_PARAM, "read body error : "+e.Error())
		}
	}
	return hr.body, nil
}

//EnsureBody 检查Body中的字段是否齐全
func (hr *HTTPRequest) EnsureBody(keys ...string) (string, bool) {
	b, e := hr.Body()
	if e != nil {
		return "", false
	}
	for _, key := range keys {
		if _, ok := b[key]; !ok {
			return key, false
		}
	}
	return "", true
}

//ParseObj 把body解析成为一个对象
func (hr *HTTPRequest) ParseObj(obj interface{}) error {
	if obj == nil {
		return fmt.Errorf("param obj is nil")
	}
	return json.Unmarshal(hr.BodyRaw, obj)
}

//ParseOpt 带默认值的解析
func (hr *HTTPRequest) ParseOpt(params ...interface{}) error {
	if len(params)%3 != 0 {
		return errors.New("params count invalid")
	}
	b, e := hr.Body()
	if e != nil {
		return e
	}
	for i := 0; i < len(params); i += 3 {
		key := utils.ToString(params[i])
		v, ok := b[key]
		var e error
		switch ref := params[i+1].(type) {
		case *string:
			if ok {
				*ref = utils.ToString(v)
			} else {
				*ref = utils.ToString(params[i+2])
			}
		case *float64:
			if ok {
				*ref, e = utils.ToFloat64(v)
			} else {
				*ref, e = utils.ToFloat64(params[i+2])
			}
		case *int:
			if ok {
				*ref, e = utils.ToInt(v)
			} else {
				*ref, e = utils.ToInt(params[i+2])
			}
		case *uint32:
			if ok {
				*ref, e = utils.ToUint32(v)
			} else {
				*ref, e = utils.ToUint32(params[i+2])
			}
		case *uint64:
			if ok {
				*ref, e = utils.ToUint64(v)
			} else {
				*ref, e = utils.ToUint64(params[i+2])
			}
		case *int64:
			if ok {
				*ref, e = utils.ToInt64(v)
			} else {
				*ref, e = utils.ToInt64(params[i+2])
			}
		case *int8:
			if ok {
				*ref, e = utils.ToInt8(v)
			} else {
				*ref, e = utils.ToInt8(params[i+2])
			}
		case *uint:
			if ok {
				*ref, e = utils.ToUint(v)
			} else {
				*ref, e = utils.ToUint(params[i+2])
			}
		case *bool:
			if ok {
				*ref, e = utils.ToBool(v)
			} else {
				*ref, e = utils.ToBool(params[i+2])
			}
		case *[]string:
			if ok {
				*ref, e = utils.ToStringSlice(v)
			} else {
				*ref = params[i+2].([]string)
			}
		case *map[string]interface{}:
			if ok {
				switch m := v.(type) {
				case map[string]interface{}:
					*ref = m
				default:
					e = fmt.Errorf("%v is not map[string]iterface{}, but is %v", key, reflect.TypeOf(v))
				}
			} else {
				*ref = params[i+2].(map[string]interface{})
			}
		case *interface{}:
			if ok {
				*ref = v
			} else {
				*ref = params[i+2]
			}
		default:
			return fmt.Errorf("key [%v] with unknown type %v ", key, reflect.TypeOf(ref))
		}
		if e != nil {
			return fmt.Errorf("parse [%v] error:%v", key, e.Error())
		}
	}
	return nil
}

//Parse 不带默认值的解析
func (hr *HTTPRequest) Parse(params ...interface{}) error {
	if len(params)%2 != 0 {
		return errors.New("params count must be odd")
	}
	b, e := hr.Body()
	if e != nil {
		return e
	}
	for i := 0; i < len(params); i += 2 {
		key := utils.ToString(params[i])
		if v, ok := b[key]; ok {
			var e error
			switch ref := params[i+1].(type) {
			case *string:
				*ref = utils.ToString(v)
			case *float64:
				*ref, e = utils.ToFloat64(v)
			case *int:
				*ref, e = utils.ToInt(v)
			case *uint8:
				*ref, e = utils.ToUint8(v)
			case *uint16:
				*ref, e = utils.ToUint16(v)
			case *uint32:
				*ref, e = utils.ToUint32(v)
			case *uint64:
				*ref, e = utils.ToUint64(v)
			case *int64:
				*ref, e = utils.ToInt64(v)
			case *int16:
				*ref, e = utils.ToInt16(v)
			case *int8:
				*ref, e = utils.ToInt8(v)
			case *bool:
				*ref, e = utils.ToBool(v)
			case *uint:
				*ref, e = utils.ToUint(v)
			case *map[string]interface{}:
				switch m := v.(type) {
				case map[string]interface{}:
					*ref = m
				default:
					e = errors.New("value is not map[string]iterface{}")
				}
			case *[]string:
				*ref, e = utils.ToStringSlice(v)
			case *interface{}:
				*ref = v
			default:
				return fmt.Errorf("unknown type %v ", reflect.TypeOf(ref))
			}
			if e != nil {
				return fmt.Errorf("parse [%v] error:%v", key, e.Error())
			}
			if key == "ps" {
				ps, e := utils.ToUint64(v)
				if e == nil && ps > MAX_PS {
					return errors.New("ps too large")
				}
			}
		} else {
			return fmt.Errorf("%v not provided", key)
		}
	}
	return nil
}

//ParseGet 获取Get方式请求的参数
func (hr *HTTPRequest) ParseGet(params ...interface{}) error {
	if len(params)%2 != 0 {
		return errors.New("params count must be odd")
	}
	urlValues := hr.request.URL.Query()
	for i := 0; i < len(params); i += 2 {
		key := utils.ToString(params[i])
		if v := urlValues.Get(key); v != "" {
			var e error
			switch ref := params[i+1].(type) {
			case *string:
				*ref = v
			case *float64:
				*ref, e = utils.ToFloat64(v)
			case *int:
				*ref, e = utils.ToInt(v)
			case *uint8:
				*ref, e = utils.ToUint8(v)
			case *uint16:
				*ref, e = utils.ToUint16(v)
			case *uint32:
				*ref, e = utils.ToUint32(v)
			case *uint64:
				*ref, e = utils.ToUint64(v)
			case *int64:
				*ref, e = utils.ToInt64(v)
			case *int16:
				*ref, e = utils.ToInt16(v)
			case *int8:
				*ref, e = utils.ToInt8(v)
			case *bool:
				*ref, e = utils.ToBool(v)
			case *uint:
				*ref, e = utils.ToUint(v)
			case *interface{}:
				*ref = v
			default:
				return fmt.Errorf("unknown type %v ", reflect.TypeOf(ref))
			}
			if e != nil {
				return fmt.Errorf("parse [%v] error:%v", key, e.Error())
			}
			if key == "ps" {
				ps, e := utils.ToUint64(v)
				if e == nil && ps > MAX_PS {
					return errors.New("ps too large")
				}
			}
		} else {
			return fmt.Errorf("%v not provided", key)
		}
	}
	return nil
}

//ParseGetOpt 带有默认值的参数解析
func (hr *HTTPRequest) ParseGetOpt(params ...interface{}) error {
	if len(params)%3 != 0 {
		return errors.New("params count invalid")
	}
	urlValues := hr.request.URL.Query()
	for i := 0; i < len(params); i += 3 {
		key := utils.ToString(params[i])
		v := urlValues.Get(key)
		ok := v == ""
		var e error
		switch ref := params[i+1].(type) {
		case *string:
			fmt.Println("string", v)
			if ok {
				*ref = utils.ToString(v)
			} else {
				*ref = utils.ToString(params[i+2])
			}

		case *float64:
			if ok {
				*ref, e = utils.ToFloat64(v)
			} else {
				*ref, e = utils.ToFloat64(params[i+2])
			}
		case *int:
			if ok {
				*ref, e = utils.ToInt(v)
			} else {
				*ref, e = utils.ToInt(params[i+2])
			}
		case *uint32:
			if ok {
				*ref, e = utils.ToUint32(v)
			} else {
				*ref, e = utils.ToUint32(params[i+2])
			}
		case *uint64:
			if ok {
				*ref, e = utils.ToUint64(v)
			} else {
				*ref, e = utils.ToUint64(params[i+2])
			}
		case *int64:
			fmt.Println("int64", v)
			if ok {
				*ref, e = utils.ToInt64(v)
			} else {
				*ref, e = utils.ToInt64(params[i+2])
			}
		case *int8:
			if ok {
				*ref, e = utils.ToInt8(v)
			} else {
				*ref, e = utils.ToInt8(params[i+2])
			}
		case *uint:
			if ok {
				*ref, e = utils.ToUint(v)
			} else {
				*ref, e = utils.ToUint(params[i+2])
			}
		case *bool:
			if ok {
				*ref, e = utils.ToBool(v)
			} else {
				*ref, e = utils.ToBool(params[i+2])
			}
		case *[]string:
			if ok {
				*ref, e = utils.ToStringSlice(v)
			} else {
				*ref = params[i+2].([]string)
			}
		case *map[string]interface{}:
			e = fmt.Errorf("%v is not map[string]iterface{}, but is %v", key, reflect.TypeOf(v))
			//	switch m := params[i+1].(type) {
			//	case map[string]interface{}:
			//		*ref = m
			//	default:
			//		e = errors.New(fmt.Sprintf("%v is not map[string]iterface{}", key, reflect.TypeOf(v)))
			//	}
			//} else {
			//	*ref = params[i+2].(map[string]interface{})
			//}
		case *interface{}:
			if ok {
				*ref = v
			} else {
				*ref = params[i+2]
			}
		default:
			return fmt.Errorf("key [%v] unknown type %v ", key, reflect.TypeOf(ref))
		}
		if e != nil {
			return fmt.Errorf("parse [%v] error:%v", key, e.Error())
		}
	}
	return nil
}

//IsWeChatUserAgent 是否是微信客户端返回
func (hr *HTTPRequest) IsWeChatUserAgent() bool {
	return strings.Contains(strings.ToLower(hr.request.UserAgent()), "micromessenger")
}

//IsAliPayUserAgent 是否是微信客户端返回
func (hr *HTTPRequest) IsAliPayUserAgent() bool {
	return strings.Contains(strings.ToLower(hr.request.UserAgent()), "alipay")
}
