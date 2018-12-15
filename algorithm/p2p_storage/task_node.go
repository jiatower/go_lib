package p2p_storage

//任务节点关系表
type TaskNode struct {
	ID   uint64 `json:"id"`
	Node string `json:"node"`
	Tm   int64  `json:"tm"`
}
