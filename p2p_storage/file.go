package p2p_storage

import "fmt"

//文件信息
type File struct {
	MD5  string `json:"md5"`
	Size uint64 `json:"size"`
}

//分组文件信息
type GroupFile struct {
	File      `json:"file"`
	Ver       uint64 `json:"ver"`
	State     int    `json:"state"`       //1-NORMAL,0-DELETED
	Group     string `json:"group"`       //文件所在组
	Type      int    `json:"type"`        //文件类型 0-首次扩散 1-新增
	AddVer    uint64 `json:"add_ver"`     // 新增文件版本
	LastAddTm uint64 `json:"last_add_tm"` // 新增文件版本
	SrcNode   string `json:"src_node"`    // 文件源节点
}

func newGroupFile(md5 string, size, ver uint64, tp int, add_ver, last_add_tm uint64, src_node string) (file *GroupFile) {
	return &GroupFile{File{md5, size}, ver, NORMAL, "", tp, add_ver, last_add_tm, src_node}
}

func (gf *GroupFile) ToString() (val string) {
	return fmt.Sprintf("md5=%s, size=%v, ver=%v, state=%v,tp=%v,add_ver=%v", gf.MD5, gf.Size, gf.Ver, gf.State, gf.Type, gf.AddVer)
}

func (gf *GroupFile) IsNewAdd() (ok bool) {
	if gf == nil || gf.Type == GROUPFILE_TYPE_NEW_ADD {
		ok = true
	}
	return
}
