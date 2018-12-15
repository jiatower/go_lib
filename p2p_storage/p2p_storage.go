package p2p_storage

import (
	"errors"
	"fmt"
	"math/rand"
	"yh_pkg/log"
	"yh_pkg/service"
	"yh_pkg/time"
	"yh_pkg/utils"
	"yunhui/redis_db"
)

var logger *log.MLogger

var dataSource *DataSource

var ConfigMap *ConfigSet

var P2pLockExpireSec int64 = 5  //同步锁到期时间5秒(单位秒)
var P2pGetLockTimeOut int64 = 1 //所有使用同步锁的地方，超过该值还未获取到时，这直接放弃，业务需要根据实际情况来处理(单位秒)

func Init(ds IDataSource, lg *log.MLogger, open_check bool) (e error) {
	logger = lg
	dataSource = newDataSource(ds)
	ConfigMap = NewConfigSet()
	rand.Seed(time.Now.Unix())
	if open_check {
		go checkTimeoutNodes()
		go checkExpandTaskTimeout()
		go checkDelLongTimeOutNode()
		go updateNodeOnlineTime()
		//go checkExpandGroup()
		go clearNewAddGroupFileTimeOut()
		go updateConfigMap()
	}
	return
}

func AddNode(id string) (e error) {
	exist, e := dataSource.Raw.IsNodeExist(id)
	if e != nil || exist {
		return
	}
	return dataSource.Raw.AddNode(newNodeDetail(id))
}

func DeleteNode(id string) (e error) {
	groups, e := dataSource.Raw.GetNodeGroups(id)
	if e != nil {
		return
	}

	for _, group := range groups {
		if e = dataSource.Raw.DeleteGroupNode(group.ID, id); e != nil {
			return
		}
		e = group.ExpandNodesToPerfectSize(NODE_MAX_ACTIVE_GROUPS, "")
		logger.AppendObj(e, "--DeleteNode-DeleteGroupNode--", group.ID, id)
	}
	return dataSource.Raw.DeleteNode(id)
}

/*func AddFile(md5 string, size uint64, src_node string) (e error) {
	if len(md5) != 32 {
		return errors.New("md5 " + md5 + " is invalid")
	}
	//如果文件存在或出错，则直接返回
	if ok, e := dataSource.IsFileExists(md5); e != nil || ok {
		return e
	}
	file_size := CalculateFileSize(size)
	group, e := dataSource.Raw.GetAvailableGroup(file_size)
	if e != nil {
		return
	}
	if group == nil {
		//没有可以容纳该文件的分组了，创建新的分组
		fmt.Printf("no available group(%v), create new one...\n", file_size)
		group, e = createGroup(2, file_size, "")
		if e != nil {
			return
		}
	}
	//计算组的最大空间，每个节点在组上规定2G，该组的最大容量为minPieces*2G
	groupCapacity := uint64(group.MinPieces) * GROUP_NODE_CAPACITY
	//目前暂定如果文件大于了组总大小的三分之一，则先报错
	if size > groupCapacity/3 {
		return errors.New(fmt.Sprintf("file %s size %ld too large", md5, size))
	}
	if group.Size >= groupCapacity {
		//没有可以容纳该文件的分组了，创建新的分组
		fmt.Printf("all groups(%v) are full, create new one...group.Size=%v, groupCapacity=%v\n", file_size, group.Size, groupCapacity)
		//没有可以容纳该文件的分组了，创建新的分组
		group, e = createGroup(2, file_size, "")
		if e != nil {
			return
		}
	}
	return group.AddFile(md5, src_node, size)
}*/

func DeleteFile(md5 string) (e error) {
	files, e := dataSource.Raw.GetFileGroups(md5, NORMAL)
	if e != nil {
		return
	}
	for gid, file := range files {
		group, e := dataSource.Raw.GetGroup(gid)
		if e != nil || group == nil {
			return e
		}
		if e := group.DeleteFile(&file); e != nil {
			return e
		}
	}
	return
}

/*
获取文件的下载节点

参数：
	md5: 文件的md5
返回值:
	nodes: 可用的节点列表
	group: 这些节点所在的分组信息
	sources: 源文件所在的节点（都是在线的）
*/
func Download(md5 string) (nodes []Peer, group *Group, sources []Peer, e error) {
	//如果所有用户的魔盒中都没有该文件了，则也从p2p系统删除
	/*
		count, e := dataSource.Raw.GetSourceFileCount(md5)
		if e != nil {
			return
		}
		if count == 0 {
			e = DeleteFile(md5)
			return
		}
	*/
	return DownloadMore(md5, nil)
}

/*
获取更多文件的下载节点，排除掉已经用过的分组

参数：
	md5: 文件的md5
	usedGroups: 已经用过的分组
返回值:
	nodes: 可用的节点列表，最多返回拼回原始数据所需要最小节点数的1倍，客户端如果还是下载不成功，
		   则认为下载失败。
	group: 节点所属分组信息
	sources: 源文件所在的节点（都是在线的）
*/
func DownloadMore(md5 string, usedGroups []string) (nodes []Peer, group *Group, sources []Peer, e error) {
	nodes, group, e = getPeers(md5, usedGroups, false)
	if e != nil {
		return
	}
	sources, e = dataSource.GetOnlineSourceFileNodes(md5, 5)
	return
}

/* 请求生成碎片

参数：
	gid: 分组ID
	nid: 节点ID
	md5: 文件md5
返回值：
*/
func GenPiece(gid, nid, md5 string) (e error) {
	key := CHECKER_GEN_PIECETM_PRIFIX + gid + md5
	//获取检测时间，并判断是否需要执行检测,间隔时间去检测周期的倍
	if !checkCanRunService(key) {
		//logger.AppendObj(nil, "GenPiece-contine, gid: ", gid, "md5:", md5)
		return
	}
	//logger.AppendObj(nil, "GenPiece-ok, gid: ", gid, "md5:", md5)
	if e = dataSource.Raw.SetAtomicGetLastCheckerTm(key, time.Now.Unix(), getCheckExpireTm(key)); e != nil {
		logger.Append("SetAtomicGetLastCheckerTm setTm error: "+e.Error(), log.ERROR)
		return
	}

	//对节点版本分布进行控制, 为了避免不必要的扩散，节省流量
	nodes, e := dataSource.GetNoFileNodes(gid, md5)
	if e != nil {
		logger.AppendObj(e, "GenPiece---GetNoFileNodes---group:", gid, "md5:", md5)
		return
	}
	group, e := dataSource.Raw.GetGroup(gid)
	if e != nil || group == nil || len(nodes) <= int(float32(group.PerfectPieces-group.SafePieces)*0.1) {
		logger.AppendObj(e, "GenPiece---node-count:", len(nodes), "gid", gid, "md5", md5, "node", nid)
		return
	}

	//获取当前正在扩散的任务jww
	exNodes, e := dataSource.Raw.GetValidExpandNodes(gid, md5)
	if e != nil {
		return e
	}
	//logger.AppendObj(nil, "GenPiece--gid-1", gid, "md5: ", md5, "exNodes:", exNodes)
	if len(exNodes) < int(MAX_EXPAND_NODE_NUM) {
		peers := make([]Peer, 0)

		//首先找源节点
		gfs, e := dataSource.Raw.GetFileByMd5AndState(md5, NORMAL)
		if e != nil || len(gfs) <= 0 {
			return e
		}
		var ids = make([]string, 0, 1)
		for _, gf := range gfs {
			if gf.SrcNode != "" {
				ids = append(ids, gf.SrcNode)
			}
		}
		if len(ids) > 0 {
			if peers, e = dataSource.Raw.GetOnlinePeers(ids, time.Now.Unix()-NODE_VALID_TIME); e != nil {
				return e
			}
		}

		//源节点不在线或没有源节点
		if len(peers) <= 0 {
			//获取资源所在节点并且在的线存储节点
			peers, e = dataSource.GetOnlineSourceFileNodes(md5, int(MAX_EXPAND_NODE_NUM))
			if e != nil {
				return e
			}
		}
		expanding := false
		for _, peer := range peers {
			//是否有节点正在扩散
			for _, exNode := range exNodes {
				if peer.ID == exNode.Node {
					expanding = true
					break
				}
			}

			if !expanding {
				//先判断是否存在任务记录，并判断任务失败次数
				/*ex, e := dataSource.Raw.GetExpandNode(gid, peer.ID, md5)
				if e != nil {
					return e
				}*/
				file, e := dataSource.Raw.GetGroupFile(gid, md5)
				if e != nil {
					return e
				}
				if file == nil {
					return errors.New("file " + md5 + " not found in group " + gid)
				}

				level, e := GetExpandTaskLevel(gid, file.Ver)
				if e != nil {
					return e
				}

				//首次扩散完成的任务在新分配时，获取节点任务数,如果已分配任务数超过配置数量，则不在往该节点添加任务了
				if level > 0 {

					//logger.AppendObj(e, "GenPiece--gid-2-0 GetExpandTaskCount random node continue", gid, "md5: ", md5, "nid: ", peer.ID, ex)
					continue
					/*
						//失败次数
						if ex != nil && ex.FailedTimes >= uint32(MAX_EXPAND_TAKS_FAIL_NUMS) {
							logger.AppendObj(nil, "GenPiece--failed_times is more than ", MAX_EXPAND_TAKS_FAIL_NUMS, "failed_times: ", ex.FailedTimes, " gid : ", gid, "md5: ", md5, "peers:", peers)
							//添加概率过滤,80仍旧添加到源节点上
							//if rand.Intn(100) > 80 {
							if rand.Intn(100) > 0 {
								logger.AppendObj(nil, "GenPiece--failed_times is more than and contine ", MAX_EXPAND_TAKS_FAIL_NUMS, "failed_times: ", ex.FailedTimes, " gid : ", gid, "md5: ", md5, "peers:", peers)
								continue
							}
						}

						task_cnt, e := dataSource.Raw.GetExpandTaskCount(peer.ID)
						if e != nil {
							logger.AppendObj(e, "GenPiece--gid-2-1 GetExpandTaskCount is errror", gid, "md5: ", md5, "nid: ", peer.ID)
							return e
						}
						logger.AppendObj(nil, "GenPiece--gid-2-2 - node task cnt is more than ", MAX_NODE_EXPANDTASK_CNT, "now: ", task_cnt, "md5: ", md5, "nid: ", peer.ID, "level:", level)
						if task_cnt >= MAX_NODE_EXPANDTASK_CNT {
							logger.AppendObj(nil, "GenPiece--gid-2-3 - node task cnt is more than ", MAX_NODE_EXPANDTASK_CNT, "now: ", task_cnt, "md5: ", md5, "nid: ", peer.ID, "level:", level)
							continue
						}
					*/
				}
				logger.AppendObj(nil, "GenPiece--gid-3", gid, "md5: ", md5, "nid: ", peer.ID, "level:", level)
				return addOrUpdateExpandNode(createExpandNode(gid, peer.ID, md5, file.Size, level))
			}
		}
		//没有正在扩散的节点任务,并且没有在线的源节点
		if !expanding {
			available, e := IsAvailable(md5)
			if e != nil {
				return e
			}
			if available {
				if nid != "" {
					file, e := dataSource.Raw.GetGroupFile(gid, md5)
					if e != nil {
						return e
					}
					if file == nil {
						return errors.New("file " + md5 + " not found in group " + gid)
					}

					level, e := GetExpandTaskLevel(gid, file.Ver)
					if e != nil {
						return e
					}

					levelConfig, e := ConfigMap.GetInt64Value(GEN_PIECE_LEVEL_CONFIG_KEY)
					if e != nil {
						return e
					}
					if int64(level) < levelConfig {
						return e
					}

					node, e := dataSource.Raw.GetRandomGroupNode(gid)
					if e != nil {
						return e
					}
					if node != nil {
						/*file, e := dataSource.Raw.GetGroupFile(gid, md5)
						if e != nil {
							return e
						}
						if file == nil {
							return errors.New("file " + md5 + " not found in group " + gid)
						}

						level, e := GetExpandTaskLevel(gid, file.Ver)
						if e != nil {
							return e
						}

						levelConfig, e := ConfigMap.GetInt64Value(GEN_PIECE_LEVEL_CONFIG_KEY)
						if e != nil {
							return e
						}
						if int64(level) < levelConfig {
							return e
						}
						*/
						logger.AppendObj(nil, "GenPiece--gid-randNode", gid, "md5: ", md5, "nid: ", node.Node, "level:", level)
						return addOrUpdateExpandNode(createExpandNode(gid, node.Node, md5, file.Size, level))
					}
				}
			} else {
				logger.AppendObj(nil, "-GenPiece-source not online", md5, gid)
				ids, e := dataSource.Raw.GetSourceFileNodes(md5, 1)
				if e != nil {
					return e
				}

				file, e := dataSource.Raw.GetGroupFile(gid, md5)
				if e != nil {
					return e
				}
				if file == nil || len(ids) <= 0 {
					//如果文件不存在或者没有源节点（不论是否在线）时，则将其移除
					logger.AppendObj(nil, "-GenPiece-source not online file "+md5+" not found in group "+gid)
					//e = DeleteFile(md5)
					return e
				}

				level, e := GetExpandTaskLevel(gid, file.Ver)
				if e != nil {
					return e
				}

				levelConfig, e := ConfigMap.GetInt64Value(GEN_PIECE_LEVEL_CONFIG_KEY)
				if e != nil {
					return e
				}

				if int64(level) < levelConfig {
					//logger.AppendObj(nil, "-GenPiece--gid-5-source not online add sourceTask continue", md5, gid, "sourceNode:", ids, level, levelConfig)
					return e
				}

				logger.AppendObj(nil, "-GenPiece--gid-5-source not online add sourceTask", md5, gid, "sourceNode:", ids, level, levelConfig)
				return addOrUpdateExpandNode(createExpandNode(gid, ids[0], md5, file.Size, level))

			}
		}
	}
	return
}

//添加或者修改扩散节点
func addOrUpdateExpandNode(exNode *ExpandNode) (e error) {
	//检测是否存在任务
	ex, e := dataSource.Raw.GetExpandNode(exNode.Group, exNode.Node, exNode.MD5)
	if e != nil {
		return
	}
	if ex != nil {
		logger.AppendObj(nil, "AddOrUpdateExpandNode--existExpand ", exNode.Group, "md5: ", exNode.MD5, "node: ", exNode.Node, ex.ID)
		//增加文件版本
		if e = IncrGroupFileVer(exNode.Group, exNode.MD5); e != nil {
			logger.Append("IncrGroupFileVer error: "+e.Error(), log.ERROR)
			return
		}

	}
	_, e = dataSource.Raw.AddOrUpdateExpandNode(exNode)
	return
}

/*
	获取扩充节点任务的相关信息

	参数：
		id: Id
	返回值：
		nodes: 需要碎片的节点ID
*/
func GetExpandTaskById(id uint64) (nodes []string, file *GroupFile, group *Group, exNode *ExpandNode, e error) {
	exNode, e = dataSource.Raw.GetExpandNodeById(id)
	if e != nil {
		return
	}
	if exNode == nil {
		return nil, nil, nil, nil, errors.New(fmt.Sprintf("expand node [%v] not found", id))
	}
	if exNode.State != EXPAND_STATE_NOTIFIED {
		return nil, nil, nil, nil, errors.New(fmt.Sprintf("invalid expand node state: %d", exNode.State))
	}
	if exNode.Timeout < time.Now.Unix() {
		return nil, nil, nil, nil, errors.New("invalid expand timeout")
	}

	gid, nid, md5 := exNode.Group, exNode.Node, exNode.MD5
	nodes, e = dataSource.GetNoFileNodes(gid, md5)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	if e = dataSource.UpdateExpandNodeState(gid, nid, md5, EXPAND_STATE_STARTED); e != nil {
		return nil, nil, nil, nil, e
	}
	group, e = dataSource.Raw.GetGroup(gid)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	if group == nil {
		return nil, nil, nil, nil, errors.New("can't find group")
	}
	file, e = dataSource.Raw.GetGroupFile(gid, md5)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	if file == nil {
		return nil, nil, nil, nil, errors.New("can't find file")
	}

	return nodes, file, group, exNode, nil
}

/*
	获取扩充节点任务的相关信息

	参数：
		gid: 分组ID
		nid: 节点ID
		md5: 文件md5
	返回值：
		nodes: 需要碎片的节点ID
*/
func GetExpandTask(gid, nid, md5 string) (nodes []string, file *GroupFile, group *Group, e error) {
	exNode, e := dataSource.Raw.GetExpandNode(gid, nid, md5)
	if e != nil {
		return
	}
	if exNode == nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("expand node [%v:%v:%v] not found", gid, nid, md5))
	}
	if exNode.State != EXPAND_STATE_NOTIFIED {
		return nil, nil, nil, errors.New(fmt.Sprintf("invalid expand node state: %d", exNode.State))
	}
	nodes, e = dataSource.GetNoFileNodes(gid, md5)
	if e != nil {
		return nil, nil, nil, e
	}
	if e = dataSource.UpdateExpandNodeState(gid, nid, md5, EXPAND_STATE_STARTED); e != nil {
		return nil, nil, nil, e
	}
	group, e = dataSource.Raw.GetGroup(gid)
	if e != nil {
		return nil, nil, nil, e
	}
	file, e = dataSource.Raw.GetGroupFile(gid, md5)
	if e != nil {
		return nil, nil, nil, e
	}

	return nodes, file, group, nil
}

/*
	扩散结束

	参数：
		id:12321
		state: 0-失败，1-成功
	返回值：
*/
func ExpandFinished(id uint64, state int8) (e error) {
	exNode, e := dataSource.Raw.GetExpandNodeById(id)
	if e != nil {
		return e
	}
	if exNode == nil {
		logger.AppendObj(e, "ExpandFinished-  expand  is not found", id)
		return errors.New(utils.ToString(id) + " not found")
	}
	//logger.AppendObj(nil, "ExpandFinished--GetExpandNodeById: exNode", exNode.MD5, exNode.State, exNode.Timeout, time.Now.Unix())
	if exNode.IsFinished() {
		//logger.AppendObj(nil, "ExpandFinished-  exnode is finished", id, exNode)
		return errors.New("expand state is invalid")
	}
	gid := exNode.Group
	md5 := exNode.MD5
	nid := exNode.Node
	switch state {
	case int8(YES):
		exNode.State = EXPAND_STATE_FINISHED
		if e = dataSource.Raw.DeleteExpandNodeByMd5(exNode.MD5); e != nil {
			logger.AppendObj(e, "ExpandFinished-  DeleteExpandNodeByMd5 is error", md5)
			return e
		}
	case int8(NO):
		exNode.State = EXPAND_STATE_FAILED
		/*total, e := dataSource.Raw.GetExpandTaskTotalFailedTimes(gid, md5)
		if e != nil {
			logger.Append("GetExpandTaskTotalFailedTimes error: "+e.Error(), log.ERROR)
		} else {
			if total >= uint32(EXPAND_MAX_FAIL_TIMES_EACH_NODE)*uint32(MAX_EXPAND_NODE_NUM) {
				if e = dataSource.UpdateExpandNodeState(gid, nid, md5, exNode.State); e != nil {
					logger.Append("UpdateExpandNodeStat error: "+e.Error(), log.ERROR)
				}
				return nil
			}
		}
		*/
	}
	return dataSource.UpdateExpandNodeState(gid, nid, md5, exNode.State)
}

/*
	扩散结束

	参数：
		id:12321
		state: 0-失败，1-成功
	返回值：
*/
func P2PExpandFinished(id uint64, state int8) (e error) {
	exNode, e := dataSource.Raw.GetExpandNodeById(id)
	if e != nil {
		return e
	}
	if exNode == nil {
		return errors.New(utils.ToString(id) + " not found")
	}
	logger.AppendObj(nil, "ExpandFinished--GetExpandNodeById: exNode", exNode.MD5, exNode.State, exNode.Timeout, time.Now.Unix())
	if exNode.State == EXPAND_STATE_FINISHED {
		return errors.New("expand state is invalid")
	}
	if int8(YES) == state {
		//判断当前实际扩散情况
		var expandCount int
		group, e := dataSource.Raw.GetGroup(exNode.Group)
		if e != nil || group == nil {
			return errors.New("GetGroup error groupid: " + exNode.Group)
		}
		expandCount = getGroupLimitCount(group.SafePieces, EXPAND_TASK_FINISH_COUNT_PART)
		nodeCount, e := dataSource.Raw.GetNodeCountByVerAndState(exNode.Group, group.FirstFinishVer, ONLINE)
		if e != nil {
			return errors.New("GetNodeCountByVerAndState error groupid: " + exNode.Group)
		}

		gf, e := dataSource.Raw.GetGroupFile(exNode.Group, exNode.MD5)
		if e != nil {
			return e
		}

		if int(nodeCount) >= expandCount && gf.Type == GROUPFILE_TYPE_NEW_ADD {

			//获取锁
			if !dataSource.Raw.GetLock(redis_db.CACHE_THUNDER_REQUEST_POOL, exNode.Group, P2pLockExpireSec, P2pGetLockTimeOut) {
				e = service.NewSimpleError(service.ERR_PERMISSION_DENIED, "get lock is errror")
				logger.AppendObj(e, "P2pLock-P2PExpandFinished has no lock", exNode.Group)
				return e
			}

			e = doUpdateGroupFileTpAndVer(exNode.Group, exNode.MD5)

			//释放锁
			if err := dataSource.Raw.UnLock(redis_db.CACHE_THUNDER_REQUEST_POOL, exNode.Group); err != nil {
				logger.AppendObj(err, "P2pLock-P2PExpandFinished unlock is error", exNode.Group, exNode.MD5)
			}

			if e != nil {
				logger.AppendObj(e, "ExpandFinished-  doUpdateGroupFileTpAndVer is error", exNode)
				return e
			}

		}
	}

	if e = dataSource.Raw.DeleteExpandNodeByMd5(exNode.MD5); e != nil {
		logger.AppendObj(e, "ExpandFinished-  DeleteExpandNodeByMd5 is error", exNode)
		return e
	}

	return
}

func doUpdateGroupFileTpAndVer(gid, md5 string) (e error) {
	ver, e := dataSource.Raw.AtomicIncrID(gid)
	if e != nil {
		return e
	}
	//修改group_file 中tp和ver
	return dataSource.Raw.UpdateGroupFileTpAndVer(gid, md5, ver)
}

/*
获取可以用作代理的节点

参数：
返回值:
	nodes: 可用的节点列表
		   则认为下载失败。
*/
func GetDelegates(num int) (peers []Peer, e error) {
	return dataSource.Raw.GetUPNPAvailableNodes(num, time.Now.Unix()-NODE_VALID_TIME)
}

/*
	文件是否可用

	参数：
		md5: 文件的md5
*/
func IsAvailable(md5 string) (ok bool, e error) {
	/*
		nodes, _, e := getPeers(md5, nil, true)
		if e != nil || len(nodes) == 0 {
			return false, nil
		} else {
			return true, nil
		}
	*/

	nodes, group, e := getPeers(md5, nil, false)
	if e != nil {
		return
	}

	if group == nil {
		return
	}

	if uint32(len(nodes)) >= group.MinPieces {
		ok = true
		return
	}

	//本身节点不够MinPiece,查询unsafe+node
	ex_nids := make([]string, 0, 10)
	for _, n := range nodes {
		ex_nids = append(ex_nids, n.ID)
	}

	num := group.MinPieces - uint32(len(ex_nids))

	nids, e := GetHasUnSafeFileNode(group.ID, md5, num, ex_nids)
	if e != nil {
		logger.AppendObj(e, "IsAvailable is error", md5, "ok_num ", len(nodes), "need_num: ", num)
		return
	}

	logger.AppendObj(nil, "IsAvailable", md5, "ok_num ", len(nodes), "need_num: ", num, "Query_num:", len(nids))
	if uint32(len(nids)) >= num {
		ok = true
	}
	return
}

/*
	文件是否存在

	参数：
		md5: 文件的md5
*/
func IsExists(md5 string) (exist bool, e error) {
	return dataSource.IsFileExists(md5)
}

/*
	批量检测文件是否存在

	参数：
		md5: 文件的md5
*/
func IsExistsMore(md5s []string) (m map[string]bool, e error) {
	return dataSource.IsMoreFileExists(md5s)
}

/*
	获取分组中版本号大于ver的文件列表

	参数：
		gid: 分组ID
		ver: 起始版本号（不包括本版本号）
		num: 获取文件数量
	返回值：
		files: 文件列表
*/
func ListUpdatedFiles(gid string, ver uint64, num int, tp int) (files []GroupFile, e error) {
	return dataSource.Raw.ListUpdatedFiles(gid, ver, num, tp)
}

/*
	节点是否有权下载此文件。节点所属的分组中必须含有此文件，或者节点自身就含有此文件。
*/
func CanDownloadFile(nid, md5 string) (yes bool, e error) {
	fgroups, e := dataSource.Raw.GetFileGroups(md5, NORMAL)
	if e != nil {
		return
	}
	ngroups, e := dataSource.Raw.GetNodeGroupState(nid)
	for gid, _ := range fgroups {
		if _, yes = ngroups[gid]; yes {
			return
		}
	}
	yes, e = dataSource.Raw.IsNodeHasFile(nid, md5)
	if e != nil {
		return
	}
	return
}

/*
	更新节点及其分组信息

	参数：
		node: 节点信息
		groupVersions: 节点所属分组的文件同步版本号
		tasks: 正在扩散的任务的心跳
	返回值：
		groups: 节点所属分组的最新信息
		exNodes: 需要此节点执行扩充任务的文件列表
*/
func UpdateNode(node *Node, groupVersions map[string]uint64, tasks []uint64, is_super int) (groups []NodeGroupDetail, exNodes []ExpandNode, e error) {
	detail, e := dataSource.Raw.GetNodeDetail(node.ID)
	if e != nil {
		return nil, nil, errors.New("GetNodeDetail error: " + e.Error())
	}
	if detail == nil {
		return nil, nil, errors.New("node " + node.ID + " not found")
	}
	groups, e = dataSource.Raw.GetNodeGroupDetail(node.ID)
	if e != nil {
		return nil, nil, errors.New("GetNodeGroupDetail error: " + e.Error())
	}

	//todo-临时修改UpdateNode中非超级硬盘修改为不在线
	if is_super == NO {
		node.State = NO
	}

	//必须要在线
	if node.State == YES {
		//更新节点容量比例
		if is_super == NO {
			detail.Percent = NORMAL_NODE_OCCUPY_PERCENT
		} else {
			detail.Percent = NODE_OCCUPY_PERCENT
		}
		detail.ActiveGroups = 0
		for idx, _ := range groups {
			if groups[idx].Size < (uint64(groups[idx].MinPieces) * GROUP_NODE_CAPACITY) {
				detail.ActiveGroups++
			}
			if ver, ok := groupVersions[groups[idx].ID]; ok {
				if ver != groups[idx].NodeVer || groups[idx].State != ONLINE || ver < groups[idx].MaxVer {
					//groups[idx].NodeVer = ver
					max_ver := groups[idx].MaxVer

					var update_finish_ver bool
					state := ONLINE
					//如果新版本比历史版本大，则更新最新值，否则不更新,用于在任务失败时修改文件版本时获取已有文件节点数,防止掉盘时节点汇报版本为0时,将文件版本添加了
					if ver > max_ver {
						max_ver = ver
						//只有在修改节点max_ver时，才触发获取当前组首次扩散完毕的文件版本，并修改到group中
						update_finish_ver = true
					}

					//如果ver+max_ver*10%<max_ver 则设置组内不在线
					if ver < max_ver*90/100 {
						state = OFFLINE
						//logger.AppendObj(e, "---DoUpdateGroupFirstExpandVer--set group offline-", node.ID, groups[idx].ID, ver, groups[idx].MaxVer, max_ver)
					}

					//更新，并比较ver，如果改变则修改last_update_tm
					if e := dataSource.Raw.UpdateGroupNode(groups[idx].ID, &GroupNode{node.ID, ver, state, max_ver}, ver != groups[idx].NodeVer); e != nil {
						return nil, nil, errors.New("UpdateGroupNode error: " + e.Error())
					}

					//先修改，在获取
					if update_finish_ver {
						//logger.AppendObj(e, "---DoUpdateGroupFirstExpandVer---", groups[idx].ID, ver, groups[idx].MaxVer, update_finish_ver)
						if e = DoUpdateGroupFirstExpandVer(groups[idx].ID, groups[idx].FirstFinishVer); e != nil {
							return
						}
					}
				}
			}
		}
	}

	for _, id := range tasks {
		if e := dataSource.UpdateExpandNodeTimeout(id); e != nil {
			logger.Append("UpdateExpandNodeTimout error: "+e.Error(), log.ERROR)
		}
	}
	if e = detail.Update(node); e != nil {
		return nil, nil, errors.New("detail.Update error: " + e.Error())
	}
	exNodes, e = dataSource.FetchExpandTasks(node.ID, MAX_EXPAND_TASK_NUM)
	return
}

/*
	更新节点及其分组信息

	参数：
		node: 节点信息
		groupVersions: 节点所属分组的文件同步版本号
		tasks: 正在扩散的任务的心跳
	返回值：
		groups: 节点所属分组的最新信息
		exNodes: 需要此节点执行扩充任务的文件列表
*/
func UpdateNode2(node *Node, groupVersions map[string]uint64, tasks []uint64, is_super int) (returnGroups []NodeGroupDetail, exNodes []ExpandNode, deleteGids []string, e error) {
	deleteGids = make([]string, 0, 1)
	returnGroups = make([]NodeGroupDetail, 0, len(groupVersions))

	detail, e := dataSource.Raw.GetNodeDetail(node.ID)
	if e != nil {
		return nil, nil, nil, errors.New("GetNodeDetail error: " + e.Error())
	}
	if detail == nil {
		return nil, nil, nil, errors.New("node " + node.ID + " not found")
	}

	//获取节点现有组信息（可能有新的组）
	groups, e := dataSource.Raw.GetNodeGroupDetail(node.ID)
	if e != nil {
		return nil, nil, nil, errors.New("GetNodeGroupDetail error: " + e.Error())
	}

	if is_super == NO {
		node.State = NO
	}

	allNowGroupIdMap := make(map[string]int)

	//必须要在线

	//更新节点容量比例
	if is_super == NO {
		detail.Percent = NORMAL_NODE_OCCUPY_PERCENT
	} else {
		detail.Percent = NODE_OCCUPY_PERCENT
	}
	detail.ActiveGroups = 0

	for idx, _ := range groups {

		allNowGroupIdMap[groups[idx].ID] = 1

		if groups[idx].Size < (uint64(groups[idx].MinPieces) * GROUP_NODE_CAPACITY) {
			detail.ActiveGroups++
		}

		if ver, ok := groupVersions[groups[idx].ID]; ok {

			//如果现有组在汇报map中时，需要判断版本号是否一致，如果一致，则不返回，否则返回改组最新信息
			if ver != groups[idx].FileVer {
				returnGroups = append(returnGroups, groups[idx])
			} else {
				//logger.AppendObj(e, "---UpdateNode2-nodeVer is eq updateVer---", groups[idx].ID, ver, groups[idx].MaxVer)
			}

			if node.State == YES {
				if ver != groups[idx].NodeVer || groups[idx].State != ONLINE || ver < groups[idx].MaxVer {
					maxVer := groups[idx].MaxVer

					var update_finish_ver bool
					state := ONLINE
					//如果新版本比历史版本大，则更新最新值，否则不更新,用于在任务失败时修改文件版本时获取已有文件节点数,防止掉盘时节点汇报版本为0时,将文件版本添加了
					if ver > maxVer {
						maxVer = ver
						//只有在修改节点max_ver时，才触发获取当前组首次扩散完毕的文件版本，并修改到group中
						update_finish_ver = true
					}

					//如果ver+max_ver*10%<max_ver 则设置组内不在线
					if ver < maxVer*90/100 {
						state = OFFLINE
						//logger.AppendObj(e, "---DoUpdateGroupFirstExpandVer--set group offline-", node.ID, groups[idx].ID, ver, groups[idx].MaxVer, max_ver)
					}

					//更新，并比较ver，如果改变则修改last_update_tm
					if e := dataSource.Raw.UpdateGroupNode(groups[idx].ID, &GroupNode{node.ID, ver, state, maxVer}, ver != groups[idx].NodeVer); e != nil {
						return nil, nil, nil, errors.New("UpdateGroupNode error: " + e.Error())
					}

					//先修改，在获取
					if update_finish_ver {
						//logger.AppendObj(e, "---DoUpdateGroupFirstExpandVer---", groups[idx].ID, ver, groups[idx].MaxVer, update_finish_ver)
						if e = DoUpdateGroupFirstExpandVer(groups[idx].ID, groups[idx].FirstFinishVer); e != nil {
							return
						}
					}
				}
			}

		} else {
			//如果现有组不在汇报map中时，这直接将新加组返回
			returnGroups = append(returnGroups, groups[idx])
		}
	}

	for _, id := range tasks {
		if e := dataSource.UpdateExpandNodeTimeout(id); e != nil {
			logger.Append("UpdateExpandNodeTimout error: "+e.Error(), log.ERROR)
		}
	}
	if e = detail.Update(node); e != nil {
		return nil, nil, nil, errors.New("detail.Update error: " + e.Error())
	}

	//对比汇报上的组，是否在新的组中，如果不在则放入到deleteGid中，通知客户端删除该组（比较重要，需要保证获取组正确）
	for uGid, _ := range groupVersions {
		if _, ok := allNowGroupIdMap[uGid]; !ok {
			deleteGids = append(deleteGids, uGid)
		}
	}

	exNodes, e = dataSource.FetchExpandTasks(node.ID, MAX_EXPAND_TASK_NUM)
	return
}

//修改分组中完成首次扩散文件的版本号
func DoUpdateGroupFirstExpandVer(gid string, old_finish_ver uint64) (e error) {
	finish_ver, e := dataSource.Raw.GetGroupFirstFinishExpandVer(gid)
	if e != nil {
		return
	}
	if finish_ver > old_finish_ver {
		e = dataSource.Raw.UpdateGroupFirstFinishVer(gid, finish_ver)
		logger.AppendObj(e, "---DoUpdateGroupFirstExpandVer---", gid, " old:", old_finish_ver, " new: ", finish_ver)
	}
	return
}

/*
	无效文件汇报，该文件无法生成piece
	将会从p2p系统中删除，移入问题文件表
*/
func InvalidFile(nid, gid, md5 string) (e error) {
	file, e := dataSource.Raw.GetGroupFile(gid, md5)
	if e != nil {
		return
	}
	if file == nil {
		return
	}
	if file.State == NORMAL {
		group, e := dataSource.Raw.GetGroup(gid)
		if e != nil {
			return e
		}
		if group == nil {
			return errors.New("group " + gid + " not found")
		}
		if e = group.DeleteFile(file); e != nil {
			return e
		}
	}
	if e = dataSource.Raw.AddToInvalidFile(nid, gid, md5, time.Now.Unix()); e != nil {
		return
	}
	return
}

func ExpandGroupToPerfectSize(gid string) (e error) {
	group, e := dataSource.Raw.GetGroup(gid)
	if e != nil {
		return e
	}
	if group == nil {
		return errors.New("group " + gid + " not found")
	}
	return group.ExpandNodesToPerfectSize(NODE_MAX_ACTIVE_GROUPS, "")
}

func UpdateChecksum(md5, checksum string) (e error) {
	return dataSource.Raw.UpdateChecksum(md5, checksum)
}

func GetChecksum(md5 string) (checksum string, e error) {
	return dataSource.Raw.GetChecksum(md5)
}

func getPeers(md5 string, usedGroups []string, skipNum bool) (nodes []Peer, group *Group, e error) {
	set := make(map[string]bool)
	if usedGroups != nil {
		for _, gid := range usedGroups {
			set[gid] = true
		}
	}
	gfiles, e := dataSource.Raw.GetFileGroups(md5, NORMAL)
	if e != nil {
		return
	}
	for gid, gfile := range gfiles {
		if _, ok := set[gid]; ok {
			continue
		}
		group, e = dataSource.Raw.GetGroup(gid)
		if e != nil {
			return
		}
		nodes, e = dataSource.Raw.GetFileNodes(gid, gfile.Ver)
		if e != nil {
			return
		}
		if group != nil && len(nodes) >= int(group.MinPieces) {
			break
		}
	}

	if group == nil || (skipNum && len(nodes) < int(group.MinPieces)) {
		nodes = make([]Peer, 0, 1)
	}
	return
}

func CreateGroup() (group *Group, e error) {
	group, e = createGroup(-1, CalculateFileSize(1024*1024*1024), "")
	return
}

//批量获取节点
func GetNodesByIds(ids []string) (nodes []NodeDetail, e error) {
	return dataSource.Raw.GetNodesByIds(ids)
}

//添加节点任务
func AddTaskNode(task_id uint64, nids []string) (e error) {
	return dataSource.Raw.AddTaskNode(task_id, nids)
}

//根据某任务删除某节点全部数据
func DeleteTaskNodeByTask(id uint64) (e error) {
	return dataSource.Raw.DeleteTaskNodeByTask(id)
}

//根据条件确定任务优先级
func GetExpandTaskLevel(gid string, ver uint64) (level int8, e error) {
	g, e := dataSource.Raw.GetGroup(gid)
	if e != nil || g == nil {
		return
	}
	//首次扩散完成
	if ver <= g.FirstFinishVer {
		//只要是首次扩散完成，全部设置1
		level = 1
		//需要判断是否处于危险状态,添加对应优先级
		num, e := dataSource.Raw.GetNodeCountByVerAndState(gid, ver, ONLINE)
		if e != nil {
			return level, e
		}

		//任务优先级，数字越大优先级越高， 0-普通，1-首次扩散完成 2-易险 3-濒危 4-危险   5-极危 6-
		if num >= 160 {
			level = 1
		} else if num < 160 && num >= 152 {
			level = 2
		} else if num < 152 && num >= 144 {
			level = 3
		} else if num < 144 && num >= 136 {
			level = 4
		} else if num < 136 && num >= 128 {
			level = 5
		} else {
			level = 6
		}
		//logger.AppendObj(nil, "GenPiece--addLevel-group", gid, "ver: ", ver, "FirstFinishVer: ", g.FirstFinishVer, " has node num:  ", num, "level: ", level)
	}
	return
}

//根据某任务删除某节点全部数据
func CheckFileOssExist(md5 string) (ossExist int, e error) {
	ossExist = YES
	//获取该md5的文件版本号，同时检测是否在组中完成首次扩散,首次扩散完成，则返回0，否则返回1
	gfs, e := dataSource.Raw.GetFileByMd5AndState(md5, NORMAL)
	if e != nil {
		return
	}

	if len(gfs) <= 0 {
		return
	}

	for _, gf := range gfs {
		g, e := dataSource.Raw.GetGroup(gf.Group)
		if e != nil || g == nil || gf.IsNewAdd() {
			continue
		}
		//如果文件版本大于首次扩散版本号，则表示未扩散完成，则ossExist=1
		if gf.Ver > 0 && gf.Ver <= g.FirstFinishVer {
			ossExist = NO
			logger.AppendObj(nil, "--CheckFileOssExist-exist-", gf.MD5, gf.Ver, g.FirstFinishVer, ossExist)
			return ossExist, nil
		}
	}
	return
}

func GetUnSafeExpandTasks(nid string, state int8, num int) (exNodes []UnSafeExpandNode, e error) {
	return dataSource.Raw.GetUnSafeExpandTasks(nid, state, num)
}

func GetHasUnSafeFileNode(gid, md5 string, num uint32, ex_nids []string) (nids []string, e error) {
	return dataSource.Raw.GetHasUnSafeFileNode(gid, md5, num, ex_nids)
}

/*
	扩散结束

	参数：
		id:12321
		state: 0-失败，1-成功
	返回值：
*/
func UnSafeExpandFinished(id uint64, state int) (e error) {
	expand_state := UNSAFE_EXPAND_STATE_INIT
	if int(YES) == state {
		expand_state = UNSAFE_EXPAND_STATE_FINISHED
	}
	return dataSource.Raw.UpdateUnSafeExpandNodeState(id, expand_state)
}

/*
	获取危险任务

	参数：
		id: Id
	返回值：
*/
func GetUnSafeExpandTaskById(id uint64) (exNode *UnSafeExpandNode, file *GroupFile, group *Group, e error) {
	exNode, e = dataSource.Raw.GetUnSafeExpandNodeById(id)
	if e != nil {
		return
	}
	if exNode == nil || exNode.ID <= 0 {
		return
	}

	if exNode.State != UNSAFE_EXPAND_STATE_INIT {
		e = service.NewSimpleError(service.ERR_INVALID_PARAM, "task state is not init")
		return
	}

	group, e = dataSource.Raw.GetGroup(exNode.Group)
	if e != nil {
		return
	}
	file, e = dataSource.Raw.GetGroupFile(exNode.Group, exNode.MD5)
	if e != nil {
		return
	}
	return
}

/*
	添加unsafe_expand_node
*/
func AddOrUpdateUnSafeExpandNode(gid, md5 string, nodes []GroupNode) (e error) {
	exNodes := make([]UnSafeExpandNode, 0, len(nodes))
	tm := time.GetTimeStamp()
	for _, v := range nodes {
		var n UnSafeExpandNode
		n.Group = gid
		n.MD5 = md5
		n.Node = v.Node
		n.Tm = tm
		n.State = int8(NO)
		exNodes = append(exNodes, n)
	}
	e = dataSource.Raw.AddOrUpdateUnSafeExpandNodes(exNodes)
	return
}

/*
	添加unsafe_file
*/
func AddOrUpdateUnSafeFile(gid, md5 string) (e error) {
	e = dataSource.Raw.AddOrUpdateUnSafeFile(gid, md5)
	return
}

/*
	删除危险文件
*/
func DeleteUnSafeFile(gid, md5 string) (e error) {
	return dataSource.Raw.DeleteUnSafeFile(gid, md5)
}

/*
	获取分组节点
*/
func GetGroupNodes(gid string) ([]GroupNode, error) {
	return dataSource.Raw.GetGroupNodes(gid)
}

/*
	删除危险文件任务
*/
func DeleteUnSafeFileExpandNode(gid, node, md5 string) (e error) {
	return dataSource.Raw.DeleteUnSafeFileExpandNode(gid, node, md5)
}

/*
	删除危险文件任务
*/
func GetUnSafeFileExpandNode() (expandNodes []UnSafeExpandNode, e error) {
	return dataSource.Raw.GetUnSafeFileExpandNode()
}

func GetOnlineNodesByIds(ids []string, min_update_tm int64) (peers []Peer, e error) {
	if len(ids) <= 0 {
		return
	}
	return dataSource.GetOnlineNodesByIds(ids, min_update_tm)
}

/*
	重置重启节点的任务状态
*/
func RestartInitExpandNodeState(node string) (e error) {
	return dataSource.Raw.SetExpandNodeStateFailed(node)
}
