package p2p_storage

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

type Peer struct {
	ID            string `json:"id"`
	IP            string `json:"ip"`
	Port          int32  `json:"port"`
	UPNPIP        string `json:"upnp_ip"`
	UPNPPort      int32  `json:"upnp_port"`
	NATType       int8   `json:"nat_type"`
	UPNPAvailable int8   `json:"upnp_available"`
}

var r *rand.Rand

func (p *Peer) FillUPNPAvailable() {
	if p.NATType == 2 || p.NATType == 1 {
		p.UPNPIP = p.IP
		p.UPNPPort = p.Port
		p.UPNPAvailable = int8(YES)
	} else {
		p.UPNPAvailable = int8(NO)
	}
}

type Node struct {
	Peer       `json:"peer"`
	TotalSpace uint64 `json:"total_space"` //魔盒的总存储空间
	LeftSpace  int64  `json:"left_space"`  //剩余空闲空间
	State      int    `json:"state"`       // 超级硬盘状态
	UpSpeed    int64  `json:"up_speed"`    //上行带宽字节
	Upload     int64  `json:"upload"`      //上传速度
	Download   int64  `json:"download"`    //下载速度
}

type NodeDetail struct {
	Peer         `json:"peer"`
	TotalSpace   uint64  `json:"total_space"`    //魔盒的总存储空间
	LeftP2pSpace int64   `json:"left_p2p_space"` //P2P剩余空闲空间（total_space*percent/100-各分组容量之和）
	Percent      int8    `json:"percent"`        //节点空间占用比例
	UpdateTm     int64   `json:"update_tm"`      //上次活跃时间
	RegTm        int64   `json:"reg_tm"`         //节点注册时间
	ActiveGroups int     `json:"activ_groups"`   //还未满的分组数量
	OnlineTm     int64   `json:"online_tm"`      //上次在线时间
	Weight       float64 `json:"weight"`         //节点权重
	OnlineCount  int     `json:"online_cnt"`     //节点在线时间计数
	UpSpeed      int64   `json:"up_speed"`       //上行带宽字节
	Upload       int64   `json:"upload"`         //上传速度
	Download     int64   `json:"download"`       //下载速度

}

func newNodeDetail(id string) *NodeDetail {
	return &NodeDetail{Peer{id, "", 0, "", 0, 0, int8(NO)}, 0, 0, NODE_OCCUPY_PERCENT, 0, time.Now().Unix(), 0, 0, 0, 0, 0, 0, 0}
}

/*
	更新节点详情

	参数：
		node: 节点基本信息
		groups: 节点所在的分组列表
*/
func (detail *NodeDetail) Update(node *Node) (e error) {
	if detail.ID != node.ID {
		return errors.New("not same node")
	}
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	detail.Peer = node.Peer
	detail.FillUPNPAvailable()
	detail.TotalSpace = node.TotalSpace
	count, e := dataSource.Raw.GetNodeGroupCount(node.ID)
	if e != nil {
		return errors.New("GetNodeGroupCount error: " + e.Error())
	}
	detail.UpSpeed = node.UpSpeed

	usedSpace := int64(GROUP_NODE_CAPACITY * uint64(count))
	detail.LeftP2pSpace = int64(detail.TotalSpace*uint64(detail.Percent)/100) - usedSpace
	if node.LeftSpace < detail.LeftP2pSpace {
		detail.LeftP2pSpace = node.LeftSpace
	}
	if detail.LeftP2pSpace < 0 {
		detail.LeftP2pSpace = 0
	}
	detail.OnlineTm = time.Now().UnixNano()

	var nodeWeight float64
	//当节点在线更新时间戳和权重
	if node.State == YES {
		detail.UpdateTm = time.Now().UnixNano()
		nodeWeight = detail.GetWeight()
	}

	detail.Weight = nodeWeight
	detail.Download = node.Download
	detail.Upload = node.Upload
	return dataSource.Raw.UpdateNode(detail)
}

func (detail *NodeDetail) GetWeight() (weight float64) {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	if detail.OnlineCount >= 144 {
		//2018-11-06 计算权重规则变更：
		weight = (1 / (math.Log2(1+float64(detail.ActiveGroups)) + 1)) * (r.Float64() + 1e-15)
		/*
			activaGroups := rand.Intn(100)
			weight = (1 / (math.Log2(1+float64(activaGroups)) + 1)) * (r.Float64() + 1e-15)
		*/
	}
	return
}

type GroupNode struct {
	Node   string `json:"node"`
	Ver    uint64 `json:"ver"`
	State  int    `json:"state"` //ONLINE/OFFLINE
	MaxVer uint64 `json:"max_ver"`
}

func newGroupNode(nid string, state int) (node *GroupNode) {
	return &GroupNode{nid, 0, state, 0}
}
