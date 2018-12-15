package p2p_storage

import "yh_pkg/time"

//危险文件上传任务对象
type UnSafeExpandNode struct {
	ID    uint64 `json:"id"`
	Group string `json:"group"`
	Node  string `json:"node"`
	MD5   string `json:"md5"`
	State int8   `json:"state"`
	Tm    int64  `json:"tm"` //创建时间，秒数
}

func createUnSafeExpandNode(gid, nid, md5 string) (exNode *UnSafeExpandNode) {
	return &UnSafeExpandNode{0, gid, nid, md5, 0, time.Now.Unix()}
}

func (en *UnSafeExpandNode) IsFinished() bool {
	return en.State == UNSAFE_EXPAND_STATE_FINISHED
}
