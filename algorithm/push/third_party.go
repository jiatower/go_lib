package push

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"yh_pkg/log"
	"yh_pkg/net/http"
	tm "yh_pkg/time"
	"yh_pkg/utils"
	"yunhui/yh_service/cls/common"
)

const (
	SYSTEM_OTHER  = 0
	SYSTEM_APPLE  = 1
	SYSTEM_XIAOMI = 2
	SYSTEM_XINGE  = 3
	SYSTEM_IPAD   = 4
)
const (
	XIAOMI_MODE_NOTIFICATION = 0
	XIAOMI_MODE_MESSAGE      = 1
	XINGE_MODE_NOTIFICATION  = 1
	XINGE_MODE_MESSAGE       = 2
	MODE_MESSAGE             = XIAOMI_MODE_MESSAGE
	MODE_NOTIFICATION        = XIAOMI_MODE_NOTIFICATION
)

const (
	BY_ALIAS = "alias"
	BY_TOPIC = "topic"
	BY_REGID = "regid"
	BY_ALL   = "all"
)

type Message struct {
	Provider int
	By       string
	Mode     int
	Shake    bool //是否震动
	Ring     bool //是否响铃
	To       []string
	Title    string
	Content  string
	Desc     string
	Channel  int //渠道，即store的model
}

func (m *Message) ToString() string {
	var provider, mode string
	switch m.Provider {
	case SYSTEM_XIAOMI:
		provider = "XiaoMi"
	case SYSTEM_XINGE, SYSTEM_OTHER:
		provider = "Xinge"
	case SYSTEM_APPLE:
		provider = "Apple"
	case SYSTEM_IPAD:
		provider = "ipad"
	default:
		provider = "Unknown"
	}
	switch m.Mode {
	case MODE_MESSAGE:
		mode = "message"
	case MODE_NOTIFICATION:
		mode = "notification"
	default:
		mode = "none"
	}
	return fmt.Sprintf("%v %v %v %v %v %v %v", provider, m.By, mode, m.To, m.Title, m.Desc, m.Content)
}

var MsgChan chan *Message = make(chan *Message, 1000)

func push() {
	for msg := range MsgChan {
		if msg.By == BY_ALL && mode != "production" {
			logger.Append("only production env can send to all users", log.ERROR)
			continue
		}

		fmt.Println(msg.ToString(), msg.Provider)
		switch msg.Provider {
		case SYSTEM_XIAOMI, SYSTEM_APPLE, SYSTEM_IPAD:
			if e := XiaoMi(mapppingKey(msg.Provider, msg.Channel), msg); e != nil {
				logger.Append(fmt.Sprintf("push to xiaomi failed : %v. Provider: %v", e.Error(), msg.Provider))
			}

			/*
				if e := Apple("X7s5vejPxOT471H/nYh/TQ==", msg); e != nil {
					logger.Append(fmt.Sprintf("push to apple failed : %v", e.Error()))
				}
			*/
		case SYSTEM_XINGE, SYSTEM_OTHER:
			if msg.Mode == MODE_NOTIFICATION {
				msg.Mode = XINGE_MODE_NOTIFICATION
			} else if msg.Mode == MODE_MESSAGE {
				msg.Mode = XINGE_MODE_MESSAGE
			}
			if e := XinGe("2100118793", "1938ca2b83eb73d80c48831bfef138cc", msg); e != nil {
				logger.Append(fmt.Sprintf("push to xinge failed : %v", e.Error()))
			}
		default:
		}
		logger.Append(msg.ToString(), log.DEBUG)
	}
}

func mapppingKey(provider, channel int) (key string) {
	switch provider {
	case SYSTEM_XIAOMI:
		if channel == common.GOODS_MODEL_M1_NUM || channel == common.GOODS_MODEL_L1_NUM || channel == common.GOODS_MODEL_L1PRO_NUM {
			key = "0PJUwOzx+DQrQZ5EI5tUsQ=="
		} else if channel == common.GOODS_MODEL_L1PRO_INTERNATIONAL_NUM {
			key = "+wJ9TexF2hQQYvYoRtSMuw=="
		} else if channel == common.GOODS_MODEL_L1PRO_LENOVO_NUM || channel == common.GOODS_MODEL_L1PRO_LENOVO2_NUM {
			key = "bWm1+3sEHZMqL5DnSurjbQ=="
		}

	case SYSTEM_APPLE:
		if channel == common.GOODS_MODEL_M1_NUM || channel == common.GOODS_MODEL_L1_NUM || channel == common.GOODS_MODEL_L1PRO_NUM {
			key = "9DMJoshQf3nHAHuw/8JYhQ=="
		} else if channel == common.GOODS_MODEL_L1PRO_INTERNATIONAL_NUM {
			key = "JGtET/aUy1cH1ZmkIvGh9A=="
		} else if channel == common.GOODS_MODEL_L1PRO_LENOVO_NUM || channel == common.GOODS_MODEL_L1PRO_LENOVO2_NUM {
			key = "yifDZxx9uxlvRwRsimBeCw=="
		}

	case SYSTEM_IPAD:
		if channel == common.GOODS_MODEL_M1_NUM || channel == common.GOODS_MODEL_L1_NUM || channel == common.GOODS_MODEL_L1PRO_NUM {
			key = "rm4cxtL9cgyaty1Znrp3eA=="
		} else if channel == common.GOODS_MODEL_L1PRO_INTERNATIONAL_NUM {
			key = ""
		} else if channel == common.GOODS_MODEL_L1PRO_LENOVO_NUM || channel == common.GOODS_MODEL_L1PRO_LENOVO2_NUM {
			key = ""
		}
	default:

	}
	return
}

func init() {
	go push()
}

//by: alias/regid/topic/all
//mode: just notification
func Apple(key string, msg *Message) error {
	fmt.Println("key=", key, "msg=", msg.ToString())
	header := make(map[string]string)
	header["Authorization"] = "key=" + key
	params := make(map[string]string)
	params["description"] = msg.Desc
	if msg.Ring {
		params["extra.sound_url"] = "default"
	}
	params["time_to_live"] = "3600000"
	params["extra.msgid"] = msg.Content
	switch msg.By {
	case BY_ALIAS:
		params["alias"] = strings.Join(msg.To, ",")
	case BY_REGID:
		params["registration_id"] = strings.Join(msg.To, ",")
	case BY_TOPIC:
		for _, tag := range msg.To {
			params["topic"] = tag
			body, e := http.Send("https", appleDomain, "/v2/message/"+msg.By, params, header, nil, []byte(""))
			if e != nil {
				return e
			} else {
				var result map[string]interface{}
				if e := json.Unmarshal(body, &result); e != nil {
					return e
				}
				if result["result"] != "ok" {
					return errors.New("error :" + string(body))
				}
			}
		}
		return nil
	case BY_ALL:
	default:
		return errors.New(msg.By + " not supported")
	}

	body, e := http.Send("https", appleDomain, "/v2/message/"+msg.By, params, header, nil, []byte(""))
	if e != nil {
		return e
	} else {
		var result map[string]interface{}
		if e := json.Unmarshal(body, &result); e != nil {
			return e
		}
		if result["result"] != "ok" {
			return errors.New("error :" + string(body))
		}
		return nil
	}
}

//by: alias/regid/topic/all
//mode: 0-notification, 1-message
func XiaoMi(key string, msg *Message) error {
	header := make(map[string]string)
	header["Authorization"] = "key=" + key
	params := make(map[string]string)
	params["pass_through"] = utils.ToString(msg.Mode)

	// TODO payload需要做urlencode
	// 参考小米文档 https://dev.mi.com/console/doc/detail?pId=1163#_5
	params["payload"] = msg.Content

	params["aps_proper_fields.mutable-content"] = msg.Content
	params["title"] = msg.Title
	params["description"] = msg.Desc
	ntype := 4
	if msg.Ring {
		ntype += 1
	}
	if msg.Shake {
		ntype += 2
	}
	params["notify_type"] = utils.ToString(ntype)
	params["time_to_live"] = "3600000"
	if msg.Mode == XIAOMI_MODE_NOTIFICATION {
		params["extra.notify_id"] = "1"
		params["extra.notify_effect"] = "1"
		params["extra.ticker"] = "1"
		params["extra.notify_foreground"] = "0"
	}
	switch msg.By {
	case BY_ALIAS:
		params["alias"] = strings.Join(msg.To, ",")
	case BY_REGID:
		params["registration_id"] = strings.Join(msg.To, ",")
	case BY_TOPIC:
		for _, tag := range msg.To {
			params["topic"] = tag
			body, e := http.Send("https", "api.xmpush.xiaomi.com", "/v2/message/"+msg.By, params, header, nil, []byte(""))
			if e != nil {
				return e
			} else {
				var result map[string]interface{}
				if e := json.Unmarshal(body, &result); e != nil {
					return e
				}
				fmt.Println(result)
				if result["result"] != "ok" {
					return errors.New("error :" + string(body))
				}
			}
		}
		return nil
	case BY_ALL:
	default:
		return errors.New(msg.By + " not supported")
	}

	base_url := appleDomain

	body, e := http.Send("https", base_url, "/v2/message/"+msg.By, params, header, nil, []byte(""))
	if e != nil {
		fmt.Println("Http Send error", e.Error())
		return e
	} else {
		var result map[string]interface{}
		if e := json.Unmarshal(body, &result); e != nil {
			return e
		}
		fmt.Println("send res:", base_url, result, params)
		if result["result"] != "ok" {
			return errors.New("error :" + string(body))
		}
		return nil
	}
}

//by: alias/regid/topic/all
//mode: 1-notification, 2-message
func XinGe(access_id string, key string, msg *Message) (e error) {
	params := make([]string, 0, 10)
	params = append(params, "POST", "openapi.xg.qq.com")
	kv := make(map[string]string)
	switch msg.By {
	case BY_ALIAS:
		v, e := json.Marshal(msg.To)
		if e != nil {
			return e
		}
		params = append(params, "/v2/push/account_list")
		params = append(params, "account_list="+string(v))
		kv["account_list"] = url.QueryEscape(string(v))
	case BY_REGID:
		if len(msg.To) != 1 {
			return errors.New("can only send to 1 device")
		}
		params = append(params, "/v2/push/single_device")
		params = append(params, "device_token="+msg.To[0])
		kv["device_token"] = url.QueryEscape(msg.To[0])
	case BY_TOPIC:
		v, e := json.Marshal(msg.To)
		if e != nil {
			return e
		}
		params = append(params, "/v2/push/tags_device")
		params = append(params, "tags_list="+string(v), "tags_op=OR")
		kv["tags_list"] = url.QueryEscape(string(v))
		kv["tags_op"] = "OR"
	case BY_ALL:
		params = append(params, "/v2/push/all_device")
	default:
		return errors.New(msg.By + " not supported")
	}
	params = append(params, "message_type="+utils.ToString(msg.Mode))
	kv["message_type"] = utils.ToString(msg.Mode)
	params = append(params, "expire_time=3600")
	kv["expire_time"] = "3600"
	var msgByte []byte
	if msg.Mode == XINGE_MODE_MESSAGE {
		msgByte, e = json.Marshal(map[string]interface{}{"title": msg.Title, "content": msg.Content})
	} else {
		msgByte, e = json.Marshal(map[string]interface{}{"title": msg.Title, "content": msg.Desc, "custom_content": msg.Content})
	}

	if e != nil {
		return e
	}
	kv["message"] = url.QueryEscape(string(msgByte))
	params = append(params, "message="+string(msgByte))
	kv["access_id"] = url.QueryEscape(access_id)
	params = append(params, "access_id="+access_id)
	if msg.Ring {
		kv["ring"] = "1"
		params = append(params, "ring=1")
	}
	kv["timestamp"] = url.QueryEscape(utils.ToString(tm.Now.Unix()))
	params = append(params, fmt.Sprintf("timestamp=%v", tm.Now.Unix()))
	params = append(params, key)
	sort.Strings(params[3 : len(params)-1])
	sign := fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(params, ""))))
	kv["sign"] = sign
	encoded := make([]string, 0, 10)
	for k, v := range kv {
		encoded = append(encoded, k+"="+v)
	}
	qstr := strings.Join(encoded, "&")
	header := make(map[string]string)
	header["Content-type"] = "application/x-www-form-urlencoded"
	body, e := http.Send("http", params[1], params[2], nil, header, nil, []byte(qstr))
	if e != nil {
		return e
	} else {
		var result map[string]interface{}
		if e := json.Unmarshal(body, &result); e != nil {
			return e
		}
		//		fmt.Println(result)
		r, e := utils.ToInt(result["ret_code"])
		if e != nil {
			return e
		}
		if r != 0 {
			return errors.New("error : " + string(body))
		}
		return nil
	}
}
