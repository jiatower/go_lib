package push

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"yh_pkg/encrypt/md5"
	"yh_pkg/log"
	"yh_pkg/net/http"
	"yh_pkg/service"
	tm "yh_pkg/time"
	"yh_pkg/utils"
)

type UserTags struct {
	Status string   `json:"status"`
	Code   int      `json:"code"`
	Msg    string   `json:"msg"`
	Tags   []string `json:"tags"`
}

type TargetUser struct {
	Appid  string
	Appuid string
	Devid  string
}

var host string
var sys func(string) int
var logger *log.MLogger
var mode string
var appleDomain string

func DefaultSys(to string) int {
	return SYSTEM_XIAOMI
}

func Init(addr string, l *log.MLogger, system func(string) int, app_mode string) {
	host = addr
	if system == nil {
		sys = DefaultSys
	} else {
		sys = system
	}
	logger = l
	mode = app_mode
	if mode == "production" {
		appleDomain = "api.xmpush.xiaomi.com"
	} else {
		appleDomain = "sandbox.xmpush.xiaomi.com"
	}
}

func GetEndpoint(uid string) (address string, key string, e error) {
	body, e := http.HttpSend(host, "push/GetEndpoint", map[string]string{"uid": uid}, nil, nil)
	if e != nil {
		return "", "", e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return "", "", e
	}
	if m.Status != "ok" {
		return "", "", errors.New(fmt.Sprintf("<%v,%v>", m.Code, m.Detail))
	}
	address, _ = m.GetString("address")
	key, _ = m.GetString("key")
	return
}

func Kick(uid string) error {
	body, e := http.HttpSend(host, "push/Kick", map[string]string{"uid": uid}, nil, nil)
	if e != nil {
		return e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return e
	}
	if m.Status != "ok" {
		return errors.New(fmt.Sprintf("<%v,%v>", m.Code, m.Detail))
	}
	return nil
}

func SendMsg(to string, content map[string]interface{}, ring, shake bool) (msgid uint64, e error) {
	msgid, _, e = SendMsgWithOnline(to, content, true, ring, shake)
	return msgid, e
}

func SendMsgNoPush(to string, content map[string]interface{}) (msgid uint64, e error) {
	msgid, _, e = SendMsgWithOnline(to, content, false, false, false)
	return msgid, e
}

//通过消息Id发送消息,不在保存该消息到message表中,用于获取离线消息
func SendMsgById(to string, msgid uint64, content map[string]interface{}) (e error) {
	_, _, e = SendMsgOnlineById(to, msgid, content)
	return
}

/*
发送实时消息给用户

参数：
	withPush: 是否使用第三方推送
*/
func SendMsgWithOnline(to string, content map[string]interface{}, withPush, ring, shake bool) (msgid uint64, online bool, e error) {
	content["tm"] = tm.Now
	j, e := json.Marshal(content)
	if e != nil {
		return 0, false, e
	}
	params := make(map[string]string)
	params["to"] = to
	body, e := http.HttpSend(host, "push/Send", params, nil, j)
	logger.AppendObj(e, "--Send--", to, content, " result: ", string(body))
	if e != nil {
		return 0, false, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return 0, false, e
	}
	if m.Status != "ok" {
		return 0, false, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	online, found, e := m.GetBool("online")
	if e != nil {
		return 0, false, e
	}
	if !found {
		return 0, false, nil
	}
	msgid, found, e = m.GetUint64("msgid")
	if e != nil {
		return 0, false, e
	}
	if !found {
		return 0, false, errors.New("no msgid in result")
	}
	if withPush && !online {
		//	sendThirdParty(to, msgid, common.PUSH_MSG, content, ring, shake)
	}
	return msgid, online, nil
}

/*
通过消息id发送消息，不在保存消息到数据库

参数：
	withPush: 是否使用第三方推送
*/
func SendMsgOnlineById(to string, id uint64, content map[string]interface{}) (msgid uint64, online bool, e error) {
	j, e := json.Marshal(content)
	if e != nil {
		return 0, false, e
	}
	params := make(map[string]string)
	params["to"] = to
	params["msgid"] = utils.ToString(id)
	body, e := http.HttpSend(host, "push/SendById", params, nil, j)
	logger.AppendObj(e, "--SendById--", to, content, id, " result: ", string(body))
	if e != nil {
		return 0, false, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return 0, false, e
	}
	if m.Status != "ok" {
		return 0, false, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	online, found, e := m.GetBool("online")
	if e != nil {
		return 0, false, e
	}
	if !found {
		return 0, false, nil
	}
	msgid, found, e = m.GetUint64("msgid")
	if e != nil {
		return 0, false, e
	}
	if !found {
		return 0, false, errors.New("no msgid in result")
	}
	return msgid, online, nil
}

//准备发送消息，用于预先得到消息的msgid
func PrepareSendMsg(to string, content map[string]interface{}) (msgid uint64, e error) {
	content["tm"] = tm.Now
	j, e := json.Marshal(content)
	if e != nil {
		return 0, e
	}
	params := make(map[string]string)
	params["to"] = to
	body, e := http.HttpSend(host, "push/PrepareSend", params, nil, j)
	if e != nil {
		return 0, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return 0, e
	}
	if m.Status != "ok" {
		return 0, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	msgid, found, e := m.GetUint64("msgid")
	if e != nil {
		return 0, e
	}
	if !found {
		return 0, errors.New("no msgid in result")
	}
	return msgid, nil
}

//实际发送消息
func ExecSend(msgid uint64, to string, content map[string]interface{}, withPush, ring, shake bool) (online bool, e error) {
	content["tm"] = tm.Now
	j, e := json.Marshal(content)
	if e != nil {
		return false, e
	}
	params := make(map[string]string)
	params["msgid"] = utils.ToString(msgid)
	params["to"] = to
	body, e := http.HttpSend(host, "push/ExecSend", params, nil, j)
	if e != nil {
		return false, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return false, e
	}
	if m.Status != "ok" {
		return false, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	online, found, e := m.GetBool("online")
	if e != nil {
		return false, e
	}
	if !found {
		return false, nil
	}
	if withPush && !online {
		//sendThirdParty(to, msgid, common.PUSH_MSG, content, ring, shake)
	}
	return online, nil
}

//向多个用户发送消息
func SendMsgM(to []string, content map[string]interface{}, withPush, ring, shake bool) (msgid map[string]uint64, e error) {
	if len(to) <= 0 {
		return
	}
	content["tm"] = tm.Now
	j, e := json.Marshal(content)
	if e != nil {
		return nil, e
	}
	params := make(map[string]string)
	v, e := utils.Join(to, ",")
	if e != nil {
		return nil, e
	}
	params["to"] = v
	body, e := http.HttpSend(host, "push/SendM", params, nil, j)
	logger.AppendObj(e, "--SendMsgByM--", to, string(j), " result: ", string(body))
	if e != nil {
		return nil, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return nil, e
	}
	if m.Status != "ok" {
		return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	msgid = map[string]uint64{}
	switch users := m.Res.(type) {
	case map[string]interface{}:
		for uid, ret := range users {
			switch tmp := ret.(type) {
			case map[string]interface{}:
				mid, _ := utils.ToUint64(tmp["msgid"])
				online, _ := utils.ToBool(tmp["online"])
				msgid[uid] = mid
				if withPush && !online {
					//sendThirdParty(uid, mid, common.PUSH_MSG, content, ring, shake)
				}
			}
		}
	}
	return msgid, nil
}

//向多个用户发送消息(该消息可以删除)
func SendMsgMCanDel(to []string, content map[string]interface{}, withPush, ring, shake bool) (msgid map[string]uint64, e error) {
	if len(to) <= 0 {
		return
	}
	content["tm"] = tm.Now
	content["type"] = "del"
	j, e := json.Marshal(content)
	if e != nil {
		return nil, e
	}
	params := make(map[string]string)
	v, e := utils.Join(to, ",")
	if e != nil {
		return nil, e
	}
	params["to"] = v
	body, e := http.HttpSend(host, "push/SendM", params, nil, j)
	logger.AppendObj(e, "--SendMsgByM-del-", to, string(j), " result: ", string(body))
	if e != nil {
		return nil, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return nil, e
	}
	if m.Status != "ok" {
		return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	msgid = map[string]uint64{}
	switch users := m.Res.(type) {
	case map[string]interface{}:
		for uid, ret := range users {
			switch tmp := ret.(type) {
			case map[string]interface{}:
				mid, _ := utils.ToUint64(tmp["msgid"])
				online, _ := utils.ToBool(tmp["online"])
				msgid[uid] = mid
				if withPush && !online {
					//sendThirdParty(uid, mid, common.PUSH_MSG, content, ring, shake)
				}
			}
		}
	}
	return msgid, nil
}

func sendThirdParty(to string, msgid uint64, group int8, content map[string]interface{}, ring, shake bool) {
	/*	var by string
		switch to {
		case "":
			by = BY_ALL
		default:
			by = BY_ALIAS
		}
		data := map[string]interface{}{}
		data["msgid"] = msgid
		data["group"] = group
		data["content"] = content
		j, e := json.Marshal(data)
		if e != nil {
			return
		}
		//管道中积压消息过多，删掉老消息
		if len(MsgChan) > 500 {
			<-MsgChan
			logger.Append("too many push message in channel MsgChan > 500")
		}
		platform := sys(to)
			switch platform {
			case SYSTEM_APPLE:
				tp := utils.ToString(content["type"])
				switch tp {
				case ycm.MSG_TYPE_TEXT:
					MsgChan <- &Message{platform, by, MODE_NOTIFICATION, shake, ring, []string{to}, "私聊消息", utils.ToString(msgid), "你收到一条私聊消息"}
				case ycm.MSG_TYPE_VOICE:
					MsgChan <- &Message{platform, by, MODE_NOTIFICATION, shake, ring, []string{to}, "私聊消息", utils.ToString(msgid), "你收到一条语音消息"}
				case ycm.MSG_TYPE_PIC:
					MsgChan <- &Message{platform, by, MODE_NOTIFICATION, shake, ring, []string{to}, "私聊消息", utils.ToString(msgid), "你收到一张图片"}
				}
			default:
				MsgChan <- &Message{platform, by, MODE_NOTIFICATION, shake, ring, []string{to}, "私聊消息", string(j), "点击查看详情"}
			}
	*/
}

func SendThirdPartyDirect(mode int, by string, users []TargetUser, title string, content map[string]interface{}, desc string, channel int) error {
	data := map[string]interface{}{}
	data["msgid"] = -1
	data["content"] = content
	j, e := json.Marshal(data)
	if e != nil {
		return e
	}

	//管道中积压消息过多，删掉老消息
	if len(MsgChan) > 500 {
		<-MsgChan
		err_desc := "too many push message in channel MsgChan > 500"
		logger.Append(err_desc)
		return errors.New(err_desc)
	}
	switch by {
	case BY_ALL, BY_TOPIC:
		MsgChan <- &Message{SYSTEM_XIAOMI, by, mode, true, true, []string{}, title, string(j), desc, channel}
		MsgChan <- &Message{SYSTEM_APPLE, by, mode, true, true, []string{}, title, string(j), desc, channel}
	case BY_ALIAS:
		for _, v := range users {
			MsgChan <- &Message{sys(v.Devid), by, mode, true, true, []string{v.GetThirdPushUid()}, title, string(j), desc, channel}
		}
	default:
		err_desc := "only support BY_ALIAS, BY_TOPICS and BY_ALL"
		logger.Append(err_desc)
		return errors.New(err_desc)
	}
	return nil
}

/*
同步调用
*/
func Call(to string, content map[string]interface{}) (result service.Result, e error) {
	j, e := json.Marshal(content)
	if e != nil {
		return
	}
	params := make(map[string]string)
	params["to"] = to
	body, e := http.HttpSend(host, "push/Call", params, nil, j)
	if e != nil {
		logger.AppendObj(e, "--Call error--", to, content, " result: ", string(body))
		return
	}
	if e := json.Unmarshal(body, &result); e != nil {
		logger.AppendObj(e, "--Call--error", to, content, " result: ", string(body))
		return result, e
	}
	logger.AppendObj(e, "--Call--", to, content["uri"], " result: ", string(body))
	return
}

//检测用户是否在线
func IsOnline(to []string) (res map[string]bool, e error) {
	if len(to) <= 0 {
		return
	}
	params := make(map[string]string)
	v, e := utils.Join(to, ",")
	if e != nil {
		return nil, e
	}
	params["to"] = v
	body, e := http.HttpSend(host, "push/IsOnline", params, nil, nil)
	logger.AppendObj(e, "--IsOnline--", to, " result: ", string(body))
	if e != nil {
		return nil, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		return nil, e
	}
	if m.Status != "ok" {
		return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}
	res = make(map[string]bool)
	switch users := m.Res.(type) {
	case map[string]interface{}:
		for uid, ret := range users {
			switch tmp := ret.(type) {
			case map[string]interface{}:
				online, _ := utils.ToBool(tmp["online"])
				res[uid] = online
			}
		}
	}
	return res, nil
}

//检测是否在线
func CheckIsOnline(to string) (online bool) {
	m, e := IsOnline([]string{to})
	if e != nil {
		return false
	}
	if v, ok := m[to]; ok && v {
		content := map[string]interface{}{"uri": "storage/GetID", "body": map[string]interface{}{}}
		//发送探测请求
		res, e := Call(to, content)
		if e != nil {
			return false
		}
		if res.Status == "ok" {
			return true
		}
	}
	return false
}

//检测用户是否在线
func OnlineUsers() (res map[string]int, e error) {
	res = make(map[string]int)
	body, e := http.HttpSend(host, "push/OnlineUsers", nil, nil, nil)
	logger.AppendObj(e, "--OnlineUsers--", " result: ", string(body))
	if e != nil {
		return nil, e
	}
	var m service.Result
	if e := json.Unmarshal(body, &m); e != nil {
		logger.AppendObj(e, "--OnlineUsers--1")
		return nil, e
	}
	if m.Status != "ok" {
		logger.AppendObj(e, "--OnlineUsers--2")
		return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
	}

	if rm, ok := m.Res.(map[string]interface{}); ok {

		for _, r := range rm {
			b, e := json.Marshal(r)
			if e != nil {
				return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
			}

			var rr service.Result
			if e := json.Unmarshal(b, &rr); e != nil {
				return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
			}

			if rr.Status != "ok" {
				return nil, service.NewError(service.ERR_INTERNAL, m.Detail, m.Msg)
			}

			if user_res, ok := rr.Res.(map[string]interface{}); ok {
				if v, ok := user_res["users"]; ok {
					logger.AppendObj(nil, "OnlineUser", v)
				}
			}

		}
	}
	return res, nil
}

//获取客户端client
func GetTargetUserByClient(appuid string, devid string) TargetUser {
	return GetTargetUser("manager.chainedbox", appuid, devid)
}

//获取TargetUser
func GetTargetUser(appid string, appuid string, devid string) (u TargetUser) {
	return TargetUser{appid, appuid, devid}
}

//获取推送id
func (user TargetUser) GetThirdPushUid() string {
	key := fmt.Sprintf("%v:%v:%v", user.Appid, user.Appuid, user.Devid)
	ar := sha256.Sum256([]byte(key))
	return md5.MD5Sum(base64.StdEncoding.EncodeToString(ar[:]))
}
