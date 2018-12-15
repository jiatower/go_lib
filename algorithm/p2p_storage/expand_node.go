package p2p_storage

import "yh_pkg/time"

type ExpandNode struct {
	ID          uint64 `json:"id"`
	Group       string `json:"group"`
	Node        string `json:"node"`
	MD5         string `json:"md5"`
	State       int8   `json:"state"`
	Tm          int64  `json:"tm"`           //创建时间，秒数
	Timeout     int64  `json:"timeout"`      //超时时间，秒数
	FailedTimes uint32 `json:"failed_times"` //扩散失败次数
	Size        uint64 `json:"size"`         //文件大小
	Level       int8   `json:"level"`        //任务优先级
	Ver         uint64 `json:"ver"`          //文件在组中的版本
}

func createExpandNode(gid, nid, md5 string, size uint64, level int8) (exNode *ExpandNode) {
	return &ExpandNode{0, gid, nid, md5, EXPAND_STATE_INIT, time.Now.Unix(), CalculateExpandNodeTimeout(EXPAND_STATE_INIT), 0, size, level, 0}
}

func createP2PExpandNode(gid, nid, md5 string, size uint64, level int8) (exNode *ExpandNode) {
	return &ExpandNode{0, gid, nid, md5, EXPAND_STATE_NOTIFIED, time.Now.Unix(), CalculateExpandNodeTimeout(EXPAND_STATE_NOTIFIED), 0, size, level, 0}
}

func (en *ExpandNode) IsFinished() bool {
	return en.State == EXPAND_STATE_FINISHED || en.State == EXPAND_STATE_FAILED || en.Timeout <= time.Now.Unix()
}

func CalculateExpandNodeTimeout(state int8) (timeout int64) {
	switch state {
	case EXPAND_STATE_INIT:
		timeout = time.Now.Unix() + NODE_VALID_TIME
	case EXPAND_STATE_STARTED:
		timeout = time.Now.Unix() + NODE_VALID_TIME
	case EXPAND_STATE_NOTIFIED:
		timeout = time.Now.Unix() + 1800
	default:
		timeout = time.Now.Unix()
	}
	return timeout
}
