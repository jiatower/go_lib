package p2p_storage

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"yh_pkg/log"
	"yh_pkg/random"
	"yh_pkg/service"
	"yh_pkg/time"
	"yunhui/redis_db"
)

type GroupPieceInfo struct {
	PieceSize     uint32 `json:"piece_size"`     //碎片大小
	MinPieces     uint32 `json:"min_pieces"`     //最小碎片数（能恢复出原始数据的最小碎片数）
	SafePieces    uint32 `json:"safe_pieces"`    //安全碎片数（小于该数量就会出发再扩散流程）
	PerfectPieces uint32 `json:"perfect_pieces"` //再扩散后要达到的碎片数
}

var GROUP_CONFIG []GroupPieceInfo = []GroupPieceInfo{{1024, 32, 48, 64}, {1024, 64, 96, 128}, {1024, 128, 160, 208}}

//客户端汇报上来的节点信息
type Group struct {
	ID             string `json:"id"`
	Size           uint64 `json:"size"`
	FileSize       uint32 `json:"file_size"`        //分组文件大小范围，1: 1-10M，10: 10-100M, 100: 100-1000M，以此类推
	PieceSize      uint32 `json:"piece_size"`       //碎片大小
	MinPieces      uint32 `json:"min_pieces"`       //最小碎片数（能恢复出原始数据的最小碎片数）
	SafePieces     uint32 `json:"safe_pieces"`      //安全碎片数（小于该数量就会出发再扩散流程）
	PerfectPieces  uint32 `json:"perfect_pieces"`   //再扩散后要达到的碎片数
	FirstFinishVer uint64 `json:"first_finish_ver"` //组中首次扩散完成的文件版本
	DeletedVer     uint64 `json:"deleted_ver"`      //组中删除的文件版本号
}

type NodeGroupDetail struct {
	Group   `json:"group"`
	FileVer uint64 `json:"file_ver"` //分组当前版本号
	NodeVer uint64 `json:"node_ver"` //节点当前版本号
	State   int    `json:"state"`    //ONLINE/OFFLINE
	MaxVer  uint64 `json:"max_ver"`  //组所在节点历史最大版本
	AddVer  uint64 `json:"add_ver"`  // 分组当前最新 add_ver 版本
}

/*
	创建一个可以用的分组
*/
func createGroup(idx int, file_size uint32, node string) (group *Group, e error) {
	logger.AppendObj(nil, "createGroup--1", idx, file_size, node)
	file_size = 0
	nodes, e := dataSource.Raw.GetAvailableNodesCount(GROUP_NODE_CAPACITY, time.Now.Unix()-NODE_VALID_TIME, time.Now.Unix()-NODE_VALID_AFTER_REGTM, NODE_MIN_ACTIVE_GROUPS, NODE_EXPAND_MIN_ONLINE_CNT)
	if e != nil {
		return
	}
	if idx == -1 {
		idx = 0
		if nodes > 500 {
			idx = 2
		} else if nodes > 200 {
			idx = 1
		}
	}

	g := GROUP_CONFIG[idx]
	if nodes < g.PerfectPieces {
		return nil, errors.New(fmt.Sprintf("no enough online nodes for create group(%v < %v)", nodes, g.PerfectPieces))
	}
	group = &Group{random.RandomAlphanumeric(GID_LEN), 0, file_size, g.PieceSize, g.MinPieces, g.SafePieces, g.PerfectPieces, 0, 0}
	if e = dataSource.Raw.AddGroup(group); e != nil {
		return
	}
	return group, group.ExpandNodesToPerfectSize(NODE_MIN_ACTIVE_GROUPS, node)
}

/*
	使用一些节点创建一个分组
*/
func createGroupByNodes(idx int, file_size uint32, nodes map[string]bool) (group *Group, e error) {
	file_size = 0
	if idx == -1 {
		if len(nodes) > int(GROUP_CONFIG[1].PerfectPieces) {
			idx = 2
		} else if len(nodes) > int(GROUP_CONFIG[0].PerfectPieces) {
			idx = 1
		} else {
			idx = 0
		}
	}

	if idx >= len(GROUP_CONFIG) {
		return nil, errors.New("createGroupByNodes idx error")
	}
	g := GROUP_CONFIG[idx]

	group = &Group{random.RandomAlphanumeric(GID_LEN), 0, file_size, g.PieceSize, g.MinPieces, g.SafePieces, g.PerfectPieces, 0, 0}
	if e = dataSource.Raw.AddGroup(group); e != nil {
		logger.AppendObj(e, "createGroupByNodes addGroup error")
		return
	}
	if e = group.ExpandNodesToGroup(nodes); e != nil {
		return
	}
	if len(nodes) < int(group.PerfectPieces) {
		return group, group.ExpandNodesToPerfectSize(NODE_MAX_ACTIVE_GROUPS, "")
	}
	return group, nil
}

/*
	向分组中添加文件

	参数：
		md5: 文件md5
		size: 文件大小
*/
func (group *Group) AddFile(md5 string, src_node string, size uint64) (e error) {

	/*ver, e := dataSource.Raw.AtomicIncrID(group.ID)
	if e != nil {
		return errors.New("redis error : " + e.Error())
	}
	if e = dataSource.AddFileToGroup(group.ID, group, newGroupFile(md5, size, ver, GROUPFILE_TYPE_SPRAND_FIRST, 0, 0, src_node), ver); e != nil {
		return
	}
	*/

	//获取锁
	if !dataSource.Raw.GetLock(redis_db.CACHE_THUNDER_REQUEST_POOL, group.ID, P2pLockExpireSec, P2pGetLockTimeOut) {
		e = service.NewSimpleError(service.ERR_PERMISSION_DENIED, "get lock is errror")
		logger.AppendObj(e, "P2pLock-AddFile has no lock", group.ID, md5)
		return
	}

	e = doAddFileToGroup(group, md5, size, src_node)

	//释放锁
	if err := dataSource.Raw.UnLock(redis_db.CACHE_THUNDER_REQUEST_POOL, group.ID); err != nil {
		logger.AppendObj(err, "P2pLock-AddFile unlock is error", group.ID, md5)
	}

	if e != nil {
		logger.AppendObj(e, "AddFile  is error", group.ID, md5)
		return
	}

	if err := group.ExpandNodesToPerfectSize(NODE_MAX_ACTIVE_GROUPS, ""); err != nil {
		logger.Append("ExpandNodesToPerfectSize error: "+err.Error(), log.ERROR)
	}

	go GenPiece(group.ID, "", md5)
	return
}

func getAtomicIncrKey(gid string) (key string) {
	return "add_" + gid
}

func doAddFileToGroup(group *Group, md5 string, size uint64, srcNode string) (e error) {
	ver, e := dataSource.Raw.AtomicIncrID(group.ID)
	if e != nil {
		return errors.New("redis error : " + e.Error())
	}
	if e = dataSource.AddFileToGroup(group.ID, group, newGroupFile(md5, size, ver, GROUPFILE_TYPE_SPRAND_FIRST, 0, 0, srcNode), ver); e != nil {
		return
	}
	return
}

/*
	向分组中添加新增文件

	参数：
		md5: 文件md5
		size: 文件大小
*/
func (group *Group) AddP2PFile(md5 string, size uint64, src_node string, fileVer uint64) (e error) {

	/*ver, e := dataSource.Raw.AtomicIncrID(getAtomicIncrKey(group.ID))
	if e != nil {
		return errors.New("redis error : " + e.Error())
	}
	if e = dataSource.AddFileToGroup(group.ID, group, newGroupFile(md5, size, 0, GROUPFILE_TYPE_NEW_ADD, ver, uint64(time.Now.Unix()), src_node), fileVer); e != nil {
		return
	}*/

	//获取锁
	if !dataSource.Raw.GetLock(redis_db.CACHE_THUNDER_REQUEST_POOL, getAtomicIncrKey(group.ID), P2pLockExpireSec, P2pGetLockTimeOut) {
		e = service.NewSimpleError(service.ERR_PERMISSION_DENIED, "get lock is errror")
		logger.AppendObj(e, "P2pLock-AddP2PFile has no lock", group.ID, md5)
		return
	}

	e = doAddP2pFileToGroup(group, md5, size, src_node, fileVer)
	//释放锁
	if err := dataSource.Raw.UnLock(redis_db.CACHE_THUNDER_REQUEST_POOL, getAtomicIncrKey(group.ID)); err != nil {
		logger.AppendObj(err, "P2pLock-AddP2PFile unlock is error", group.ID, md5)
	}

	return
}

func doAddP2pFileToGroup(group *Group, md5 string, size uint64, src_node string, fileVer uint64) (e error) {
	ver, e := dataSource.Raw.AtomicIncrID(getAtomicIncrKey(group.ID))
	if e != nil {
		return errors.New("redis error : " + e.Error())
	}
	if e = dataSource.AddFileToGroup(group.ID, group, newGroupFile(md5, size, 0, GROUPFILE_TYPE_NEW_ADD, ver, uint64(time.Now.Unix()), src_node), fileVer); e != nil {
		return
	}
	return
}

/*
	删除分组中文件

	参数：
		md5: 文件md5
*/
func (group *Group) DeleteFile(file *GroupFile) (e error) {

	//获取锁
	if !dataSource.Raw.GetLock(redis_db.CACHE_THUNDER_REQUEST_POOL, group.ID, P2pLockExpireSec, P2pGetLockTimeOut) {
		e = service.NewSimpleError(service.ERR_PERMISSION_DENIED, "get lock is errror")
		logger.AppendObj(e, "P2pLock-DeleteFile has no lock", group.ID, file.MD5)
		return
	}

	e = doUpdateGroupFile(group.ID, file)

	//释放锁
	if err := dataSource.Raw.UnLock(redis_db.CACHE_THUNDER_REQUEST_POOL, group.ID); err != nil {
		logger.AppendObj(err, "P2pLock-DeleteFile unlock is error", group.ID, file.MD5)
	}

	if e != nil {
		return
	}

	return dataSource.UpdateGroupSize(false, group, file.Size)
}

func doUpdateGroupFile(gid string, file *GroupFile) (e error) {
	ver, e := dataSource.Raw.AtomicIncrID(gid)
	if e != nil {
		return
	}
	file.Ver, file.State = ver, DELETED
	if e = dataSource.Raw.UpdateGroupFile(gid, file); e != nil {
		logger.AppendObj(e, "deleteFile doUpdateGroupFile is error", gid, file.MD5, ver)
		return
	}
	return
}

/*
	扩充分组中的节点到PerfectSize

	参数：
*/
func (group *Group) ExpandNodesToPerfectSize(active_groups int8, nid string) (e error) {
	//获取检测时间，并判断是否需要执行检测,间隔时间去检测
	if !checkCanRunService(CHECKER_EXPAND_NODE_PREFIX + group.ID) {
		return
	}
	if e = dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_EXPAND_NODE_PREFIX+group.ID, time.Now.Unix(), getCheckExpireTm(CHECKER_EXPAND_NODE_PREFIX+group.ID)); e != nil {
		logger.Append("GetNodeCheckedTime setTm error: "+e.Error(), log.ERROR)
		return
	}

	//检测活跃节点占比小于45时则不再执行文件扩散了
	//获取活跃节点占比

	onlineNodes, e := dataSource.Raw.GetGroupOnlineNodesCount(group.ID)
	if e != nil {
		return e
	}
	logger.AppendObj(nil, fmt.Sprintf("createGroup-2-%v\t ExpandNodesToPerfectSize group=%v, online=%v PerfectPieces=%v,nid=%v\n", time.Now.Format(time.TIME_LAYOUT_1), group.ID, onlineNodes, group.PerfectPieces, nid))
	if onlineNodes < group.SafePieces+(group.MinPieces/EXPAND_GROUP_ADDRATIO) {
		logger.AppendObj(e, fmt.Sprintf("ExpandNodesToPerfectSize group=%v, expand=%v\n", group.ID, group.PerfectPieces-onlineNodes))
		if e := group.ExpandNodes(group.PerfectPieces-onlineNodes, active_groups, nid); e != nil {
			return e
		}
		onlineNodes, e = dataSource.Raw.GetGroupOnlineNodesCount(group.ID)
		if e != nil {
			return e
		}
		if onlineNodes < group.PerfectPieces {
			return errors.New("no enough online nodes for expand group nodes")
		}
	}
	return
}

/*
	将节点扩充至分组

	参数：
		nodes: 要扩充至分组的节点(须保证不与组内现有节点重复)
*/
func (group *Group) ExpandNodesToGroup(nodes map[string]bool) (e error) {
	add_nids := make([]string, 0, len(nodes))
	for id, ok := range nodes {
		if ok {
			if e = group.addNodeToGroup(id); e != nil {
				return
			}
			add_nids = append(add_nids, id)
		}
	}
	go UpdateNodeWeight(add_nids)
	return
}

/*
	扩充分组中的节点

	参数：
		num: 要扩充的节点数量
*/
func (group *Group) ExpandNodes(num uint32, active_groups int8, nid string) (e error) {
	nodes, e := dataSource.Raw.GetGroupNodes(group.ID)
	if e != nil {
		return
	}
	set := make(map[string]bool, len(nodes))
	ipSet := make(map[string]bool, len(nodes))
	detailSet := make(map[string]NodeDetail, len(nodes))

	ids := make([]string, 0)
	for _, node := range nodes {
		set[node.Node] = true
		ids = append(ids, node.Node)
	}

	//获取原有节点IP
	if e = getDetailsByIds(ids, detailSet); e != nil {
		return
	}
	for _, v := range detailSet {
		if ip, e := getIpv4First2Part(v.IP); e == nil {
			ipSet[ip] = true
		}
	}

	var offset, added, expandTime, tryNum, queryRatio uint32 = 0, 0, 0, 0, 10
	for added < num {
		new_nodes, e := dataSource.Raw.GetAvailableNodes(GROUP_NODE_CAPACITY, time.Now.Unix()-NODE_EXPAND_GROUP_VALID_TIME, time.Now.Unix()-NODE_VALID_AFTER_REGTM, offset, num*queryRatio, active_groups, NODE_EXPAND_MIN_ONLINE_CNT)
		if e != nil {
			return e
		}
		if len(new_nodes) == 0 {
			//没有足够的在线节点,则设置offset=0，重新nil查询. 重试3次
			offset = 0
			tryNum++
			logger.AppendObj(nil, "ExpandNodes offset set", tryNum)
			if tryNum > 3 {
				return nil
			}
		}
		newNodes := make([]string, 0, len(new_nodes)+1)
		if nid != "" {
			newNodes = append(newNodes, nid)
			newNodes = append(newNodes, new_nodes...)
		} else {
			newNodes = new_nodes
		}
		if e = getDetailsByIds(newNodes, detailSet); e != nil {
			return e
		}

		add_nids := make([]string, 0, len(new_nodes)+1)
		for _, id := range newNodes {
			if _, ok := set[id]; ok {
				//已经在分组中
				continue
			}
			var detail NodeDetail
			if v, ok := detailSet[id]; ok {
				detail = v
			} else {
				logger.AppendObj(nil, "ExpandNodes no detail node", id)
				continue
			}

			var ip string
			if ip, e = getIpv4First2Part(detail.IP); e != nil {
				logger.AppendObj(nil, "ExpandNodes ip format error", id)
				continue
			}
			if _, ok := ipSet[ip]; ok {
				logger.AppendObj(nil, "ExpandNodes filter ip:", detail.IP, "group:", group.ID, "node:", id)
				//过滤IP前两段重复节点
				continue
			}
			add_nids = append(add_nids, id)

			if e := group.addNodeToGroup(id); e != nil {
				continue
			}

			set[id] = true
			ipSet[ip] = true
			added += 1
			if added >= num {
				//添加完成
				break
			}
		}
		expandTime += 1
		logger.AppendObj(nil, "ExpandNodes UpdateNodeWeight count:", len(add_nids), offset, "group:", group.ID, "expandTime:", expandTime)
		//添加完分组后，需要刷新节点的权重
		go UpdateNodeWeight(add_nids)
		offset = offset + num*queryRatio
	}
	return
}

//将节点添加到组内
func (group *Group) addNodeToGroup(id string) (e error) {
	if e = dataSource.Raw.AddNodeToGroup(group.ID, newGroupNode(id, ONLINE)); e != nil {
		logger.Append("AddNodeToGroup error: "+e.Error(), log.ERROR)
		return
	}
	if e = dataSource.Raw.IncrementActiveGroups(id); e != nil {
		logger.Append("IncrementActiveGroups error: "+e.Error(), log.ERROR)
		return
	}

	//往分组中添加了新节点后需要及时的为该节点所需要的文件生成扩散任务
	go genNewNodeExpandTask(group.ID, id)
	return
}

//创建分组或者扩容后需要修改节点权重
func UpdateNodeWeight(nids []string) (e error) {
	if len(nids) <= 0 {
		return
	}
	//执行更新节点权重
	nodes, e := dataSource.Raw.GetNodesByIds(nids)
	if e != nil {
		return
	}
	for _, n := range nodes {
		weight := n.GetWeight()
		logger.AppendObj(e, "AddNodeToGroup-- UpdateNodeWeight--do old_w", n.ID, n.Weight, " old_actives", n.ActiveGroups, n.OnlineCount, "new: ", weight)
		if e = dataSource.Raw.UpdateNodeWeight(n.ID, weight); e != nil {
			logger.AppendObj(e, "AddNodeToGroup-- UpdateNodeWeight--is error", n.ID)
		}
	}
	return
}

func CalculateFileSize(size uint64) (file_size uint32) {
	return 0
	/*MB := size / 1024 / 1024
	switch {
	case MB < 10:
		return 1
	case MB >= 10 && MB < 300:
		return 2
	case MB >= 300 && MB < 1200:
		return 3
	default:
		return 4
	}
	*/
}

func MinFileSize(nums map[uint32]uint64) (file_size uint32) {
	min := uint64(math.MaxUint64)
	var i uint32 = 1
	for ; i <= 4; i++ {
		if num, found := nums[i]; !found {
			return i
		} else {
			if num < min {
				file_size = i
				min = num
			}
		}
	}
	return
}

/*
	修改文件版本号,修改版本前需要查看该版本(节点历史最高版本)是否有足够节点，没有才需要添加版本

	参数：
		md5: 文件md5
*/
func IncrGroupFileVer(gid, md5 string) (e error) {
	//获取文件版本
	gf, e := dataSource.Raw.GetGroupFile(gid, md5)
	if e != nil || gf == nil {
		logger.AppendObj(nil, "group file node has more minPieces,do not IncrFileVer", gid, md5)
		return errors.New("IncrGroupFileVer is error")
	}

	//2018-11-05:修改加版本号逻辑，判断只要ver>first_finish_ver则可添加版本号
	g, e := dataSource.Raw.GetGroup(gid)
	if e != nil {
		return
	}

	if gf.Ver <= g.FirstFinishVer {
		logger.AppendObj(nil, "group file ver <=first_finish_ver", gid, md5, gf.Ver, g.FirstFinishVer)
		return
	}

	/*
		cnt, e := dataSource.Raw.GetFileNodesCountByVer(gid, gf.Ver)
		if e != nil {
			logger.AppendObj(e, "group file node has more minPieces,do not IncrFileVer", gid, md5)
			return
		}
		if cnt >= uint32(FIRST_EXPAND_FINISH_NUM) {
			logger.AppendObj(nil, "group file node has more minPieces, cnt > FIRST_EXPAND_FINISH_NUM", gid, md5)
			return
		}
	*/

	//获取锁
	if !dataSource.Raw.GetLock(redis_db.CACHE_THUNDER_REQUEST_POOL, gid, P2pLockExpireSec, P2pGetLockTimeOut) {
		e = service.NewSimpleError(service.ERR_PERMISSION_DENIED, "get lock is errror")
		logger.AppendObj(e, "P2pLock-IncrGroupFileVer has no lock", gid)
		return
	}

	e = doIncrFileVer(gid, md5, gf.Ver)

	//释放锁
	if err := dataSource.Raw.UnLock(redis_db.CACHE_THUNDER_REQUEST_POOL, gid); err != nil {
		logger.AppendObj(err, "P2pLock-IncrGroupFileVer unlock is error", gid, md5)
	}

	if e != nil {
		logger.AppendObj(e, "IncrGroupFileVer is error ", gid, md5, gf.Ver)
		return
	}

	return
}

func doIncrFileVer(gid, md5 string, oldVer uint64) (e error) {
	ver, e := dataSource.Raw.AtomicIncrID(gid)
	if e != nil {
		logger.AppendObj(e, "group file node has more minPieces,AtomicIncrID is error", gid, md5)
		return
	}
	if e = dataSource.Raw.IncrFileVer(gid, md5, ver); e != nil {
		logger.AppendObj(e, "group file node has more minPieces,do not IncrFileVer ", gid, md5)
		return
	}
	logger.AppendInfo(nil, " IncrGroupFileVer - add group_file", gid, md5, "old_ver", oldVer, " new_ver:", ver)
	return
}

/*
	为新加入节点添加扩散任务

	参数：
		md5: 文件md5
*/
func genNewNodeExpandTask(gid, nid string) (e error) {
	ver, e := dataSource.Raw.GetGroupFileVer(gid, nid)
	if e != nil {
		return
	}
	//获取需要扩散文件件
	files, e := dataSource.Raw.GetGroupFileByVer(gid, nid, ver, 500)
	if e != nil {
		return
	}
	//循环生成扩散任务
	for _, f := range files {
		GenPiece(gid, nid, f.MD5)
	}
	return
}

/*
	获取IPV4地址前两个段

	参数:
		ip: ipv4字符串地址
*/
func getIpv4First2Part(ip string) (string, error) {
	slice := strings.Split(ip, ".")
	if len(slice) != 4 {
		return "", errors.New("ip format error")
	}

	return strings.Join(slice[0:2], "."), nil
}

/*
	使用id获取nodeDetail

	参数：
		ids：node的id
*/
func getDetailsByIds(ids []string, detailSet map[string]NodeDetail) (e error) {
	nodeDetails, e := dataSource.Raw.GetNodesByIds(ids)
	if e != nil {
		return
	}

	for _, detail := range nodeDetails {
		detailSet[detail.ID] = detail
	}
	return
}
