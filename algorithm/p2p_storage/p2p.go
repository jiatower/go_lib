package p2p_storage

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"yh_pkg/service"
	"yh_pkg/time"
)

var remainGroupIdKey string = "remian_group"

var remainGroupCntKey string = "remian_group_cnt"

const (
	maxGroupAddFileNum = 100
	maxSize            = uint64(5 * 1024 * 1024)
)

var groupId string

var ops uint32 = 0

func incryRemainAddCnt() (cnt uint32) {
	return atomic.AddUint32(&ops, 1)
}

//新版逻辑独立添加文件逻辑，通过查找符合条件的节点，然后确定分组，然后生成任务，并返回
func AddP2PFile(md5, src_node string, size uint64, times int, add_no_source_file bool) (task_id int64, e error) {
	if len(md5) != 32 {
		return 0, errors.New("md5 " + md5 + " is invalid")
	}

	//判断该文件是否已经存在
	ok, file, e := dataSource.IsP2PFileExists(md5)
	logger.AppendObj(e, "-AddP2PFile-IsP2PFileExists-:md5: ", md5, "-file: ", file, "ok: ", ok)
	if e != nil {
		return 0, e
	}
	var target_group string
	var fileVer uint64
	if ok {
		//如果存在md5文件记录，需要在进行判断,是否已经完成首次扩散，否则需要判断该任务是否有生成任务
		if file.State == NORMAL && !file.IsNewAdd() {
			if !add_no_source_file {
				//如果是add_no_source_file==false,则表示有正常节点获取了该文件，则判断是否src_node为空，如果为空，则需要其存储ID设置为原始节点
				if file.SrcNode == "" {
					file.SrcNode = src_node
					if e = dataSource.Raw.UpdateGroupFile(file.Group, file); e != nil {
						return
					}
				}
				//2018-10-24日取消，时no_srcNode和 srcNode保持一致
				/*	e = service.NewSimpleError(service.ERR_P2P_FILE_ALREADY_EXIST, "file exist")
					return*/
			}

			group, e := dataSource.Raw.GetGroup(file.Group)
			if e != nil || group == nil {
				return 0, errors.New("AddP2PFile-GetGroup error groupid: " + file.Group)
			}

			file, e = dataSource.Raw.GetGroupFile(file.Group, file.MD5)
			if e != nil {
				return 0, e
			}

			//2018-11-05：修改判断file_exist条件： 文件ver<=first_finish_ver . 去掉原有严格条件：在线节点数>165并且版本号>当前文件
			if file.Ver <= group.FirstFinishVer {
				return 0, service.NewSimpleError(service.ERR_P2P_FILE_ALREADY_EXIST, "file exist")
			}

			/*
				//添加无源节点文件到p2p系统时，文件存在时先不返回“文件已存在错误”，还需要在进一步较验是否满足节点数,为解决 file的tp由1变为0时，节点版本汇报不及时，导致在GenPiece时认为未首次扩散完成，就会增加版本号，导致该文件一直不可用,但是在源节点P2pAddFile确告诉文件已存在(客户端也需要添加比较严格ExpandFinish 条件)
					expandCount := getGroupLimitCount(group.SafePieces, EXPAND_TASK_FINISH_COUNT_PART)
					nodeCount, e := dataSource.Raw.GetNodeCountByVerAndState(group.ID, file.Ver, ONLINE)
					if e != nil {
						return 0, errors.New("AddP2PFile-GetNodeCountByVerAndState error groupid: " + group.ID)
					}
					logger.AppendObj(nil, "AddP2PFile check_file_exist add_no_source_file,md5: ", md5, file.Group, "need: ", expandCount, " now_has: ", nodeCount)
					if nodeCount > uint32(expandCount) {
						return 0, service.NewSimpleError(service.ERR_P2P_FILE_ALREADY_EXIST, "file exist")
					}
			*/

		}
		fileVer = file.Ver
		target_group = file.Group
	}

	var g *Group
	var node string
	// 根据target_group 情况确定是否需要生成生成新的分组
	if target_group == "" {
		if size <= maxSize {
			if groupId != "" {
				cnt := incryRemainAddCnt()
				if cnt <= maxGroupAddFileNum {
					g, e = dataSource.Raw.GetGroup(groupId)
					if e != nil {
						logger.AppendObj(e, "-AddP2PFile-groupId-is error-: ", src_node, g, target_group)
						return 0, e
					}
				} else {
					ops = 0
					groupId = ""
				}
			}

			if g == nil {
				node, g, e = getAvailableNodeAndGroup(md5, false)
				if e != nil {
					return
				}
				//更新groupId
				groupId = g.ID
				logger.AppendObj(e, "-AddP2PFile-notExistTarGetGroup--:md5: ", md5, " node: ", node, g)
			}

		} else {

			node, g, e = getAvailableNodeAndGroup(md5, false)
			if e != nil {
				return
			}
			logger.AppendObj(e, "-AddP2PFile-notExistTarGetGroup--:md5: ", md5, " node: ", node, g)
		}

	} else {
		g, e = dataSource.Raw.GetGroup(target_group)
		if e != nil {
			logger.AppendObj(e, "-AddP2PFile-ExistTarGetGroup-is error-: ", src_node, g, target_group)
			return 0, e
		}

		addFileCount := getGroupLimitCount(g.SafePieces, ADD_FILE_COUNT_PART)
		nodeCount, e := dataSource.Raw.GetNodeCountByVerAndState(g.ID, g.FirstFinishVer, ONLINE)
		if e != nil {
			logger.AppendObj(e, "-AddP2PFile-GetNodeCountByVerAndState-error-groupid: "+g.ID)
			return 0, e
		}

		if int(nodeCount) < addFileCount {
			logger.AppendObj(nil, "-AddP2PFile-group-firstver-node-count:", nodeCount, "group:", g.ID, "md5:", md5)
			return 0, service.NewSimpleError(service.ERR_INTERNAL, "group online_num is less than safe_piece num")
		}
		logger.AppendObj(e, "-AddP2PFile-ExistTarGetGroup-get exist group: md5:", md5, g, target_group)
	}

	//为获取到可用分组,需要创建分组
	if g == nil || g.ID == "" {
		logger.AppendObj(e, "-AddP2PFile-GetNodeGroupCount-g is null-: ", node)
		groupCount, e := dataSource.Raw.GetNodeGroupCount(node)
		if e != nil {
			logger.AppendObj(e, "-AddP2PFile-GetNodeGroupCount-error-: ", node)
			return 0, e
		}
		if groupCount != 0 { // 老节点 直接创建分组
			g, e = createGroup(2, 0, node)
			if e != nil {
				logger.AppendObj(e, "-AddP2PFile-doGetAavialbeGroup-is error-: ", md5, node, g)
				return 0, e
			}
			logger.AppendObj(nil, "-AddP2PFile-groupCount old node=src_node:", src_node, "groupCount:", groupCount, node, g)
		} else { // 新节点 检测新节点数量
			g, e = CheckNewNodeAndCreateGroup(node)
			if e != nil {
				return 0, e
			}
		}
	}

	if g == nil {
		e = service.NewSimpleError(service.ERR_INTERNAL, "no group can use")
		return
	}

	file_src_node := src_node
	if add_no_source_file {
		file_src_node = ""
	}
	//获取到分组后，需要生成对应的扩散任务
	if e = g.AddP2PFile(md5, size, file_src_node, fileVer); e != nil {
		logger.AppendObj(e, "-AddP2PFile-GetP2PFile is error--: ", md5)
		return
	}
	//生成扩散任务
	exNodes, e := dataSource.Raw.GetValidExpandNodes(g.ID, md5)
	if e != nil {
		return task_id, e
	}
	if len(exNodes) > 0 {
		e = service.NewSimpleError(service.ERR_P2P_TASK_OTHER_NODE_DOING, "task is doing")
		return
	}
	//判断任务状态
	exNode := createP2PExpandNode(g.ID, src_node, md5, size, 0)
	if exNode == nil {
		logger.AppendObj(e, "-AddP2PFile-CreateExpand is error--: ", md5, g.ID, src_node)
		e = service.NewSimpleError(service.ERR_INTERNAL, "createExpand is error")
		return
	}
	task_id, e = addOrUpdateP2PExpandNode(exNode)
	logger.AppendObj(e, "-AddP2PFile-AddP2PFile is over--: ", md5, exNode, task_id)
	if e != nil {
		return
	}

	//如果未完成文件被其他用户再次添加则更新其src_node 和 lastAddTm
	if file != nil && file.State != DELETED && file.SrcNode != src_node {
		if !add_no_source_file {
			file.SrcNode = src_node
		}
		file.LastAddTm = uint64(time.Now.Unix())
		e = dataSource.Raw.UpdateGroupFile(file.Group, file)
		logger.AppendObj(e, "-AddP2PFile-UpdateGroupFile-md5:", file.MD5, "group:", file.Group, "src_node:", file.SrcNode)
	}
	return
}

//获取某节点可用分组，如果没有则根据情况创建
func doGetAavialbeGroup(node string, create bool) (g *Group, e error) {
	//进入选择负载节点并获取该节点的可用分组逻辑
	g, e = GetNodeAvailableGroup(node)
	if e != nil {
		logger.AppendObj(e, "-AddP2PFile-GetNodeAvailableGroup-is error-: ", node)
		return
	}
	if g != nil && g.ID != "" {
		return
	}
	if !create {
		return
	}
	logger.AppendObj(nil, "doGetAavialbeGroup createGroup", node, create)
	g, e = createGroup(2, 0, node)
	return
}

/**
获取可用节点，并获取可用组，如果没有这尝试更换节点
*/
func getAvailableNodeAndGroup(md5 string, createGroup bool) (node string, g *Group, e error) {
	tryNum := 3
	nodes, e := GetAvailableNode(tryNum)
	logger.AppendObj(e, "-AddP2PFile-GetAvailableNode-new a node :md5: ", md5, "-node: ", node)
	if e != nil {
		logger.AppendObj(e, "-AddP2PFile-GetAvailableNode-is error-: ", node, md5)
		return
	}

	if len(nodes) <= 0 {
		e = service.NewSimpleError(service.ERR_INTERNAL, "not available node can use")
		return
	}

	for _, node = range nodes {
		g, e = doGetAavialbeGroup(node, createGroup)
		if e != nil {
			logger.AppendObj(e, "-AddP2PFile-doGetAavialbeGroup-is error-: ", md5, node, g)
			return
		}

		if g != nil && g.ID != "" {
			logger.AppendObj(e, "-AddP2PFile-GetAvailableNode-new a node and group md5: ", md5, "-node: ", node, g.ID)
			return
		}
	}
	return
}

//获取可以添加文件的节点(获取当前可以添加组的节点)
func GetAvailableNode(num int) (node []string, e error) {
	return dataSource.Raw.GetAvailableNode(GROUP_NODE_CAPACITY, time.Now.Unix()-NODE_VALID_TIME, time.Now.Unix()-NODE_VALID_AFTER_REGTM, NODE_EXPAND_MIN_ONLINE_CNT, num)
}

func GetNodeAvailableGroup(node string) (group *Group, e error) {
	logger.AppendObj(nil, "--GetNodeAvailableGroup--getNode: ", node)
	//获取该节点的可用分组
	groups, e := dataSource.Raw.GetNodeGroupDetail(node)
	if e != nil || len(groups) == 0 {
		logger.AppendObj(nil, "GetNodeAvailableGroup no group")
		return
	}

	usefulGroup := make([]NodeGroupDetail, 0)
	for _, g := range groups {
		//过滤空间已满的分组
		if g.ID == "" || g.Size >= GROUP_NODE_CAPACITY*uint64(g.MinPieces) {
			logger.AppendObj(nil, "GetNodeAvailableGroup filter group: ", g.ID)
			continue
		}

		addFileCount := getGroupLimitCount(g.SafePieces, ADD_FILE_COUNT_PART)
		nodeCount, e := dataSource.Raw.GetNodeCountByVerAndState(g.ID, g.FirstFinishVer, ONLINE)
		if e != nil {
			logger.AppendObj(e, "GetNodeAvailableGroup GetNodeCountByVerAndState error groupid: "+g.ID)
			continue
		}

		if int(nodeCount) >= addFileCount {
			logger.AppendObj(e, "GetNodeAvailableGroup usefulGroup groupid: "+g.ID)
			usefulGroup = append(usefulGroup, g)
		}
	}

	if len(usefulGroup) > 0 {
		group = &(usefulGroup[rand.Intn(len(usefulGroup))].Group)
		logger.AppendObj(e, "GetNodeAvailableGroup random groupid: "+group.ID)
	}

	return
}

//检测新加入节点数量并创建分组
func CheckNewNodeAndCreateGroup(tar_node string) (group *Group, e error) {
	nodes, e := dataSource.Raw.GetNewNodes(GROUP_NODE_CAPACITY, time.Now.Unix()-NODE_VALID_TIME, time.Now.Unix()-NODE_VALID_AFTER_REGTM, NODE_EXPAND_MIN_ONLINE_CNT)
	if e != nil {
		logger.AppendObj(e, "AddP2PFile-CheckNewNodeAndCreateGroup-GetNewNodes-is error-: ", tar_node)
	}
	if len(nodes) >= NEW_NODE_CREATE_GROUP_COUNT {
		group, e = createGroup(2, 0, tar_node)
		if e != nil {
			logger.AppendObj(e, "AddP2PFile-CheckNewNodeAndCreateGroup-createGroup-is error-: ", tar_node)
			return nil, e
		}
	} else {
		logger.AppendObj(nil, "AddP2PFile-CheckNewNodeAndCreateGroup has no enough nodes ", tar_node, len(nodes), "need: ", NEW_NODE_CREATE_GROUP_COUNT)
		return nil, nil
	}
	logger.AppendObj(nil, "AddP2PFile-CheckNewNodeAndCreateGroup success")
	return
}

//添加或者修改扩散节点
func addOrUpdateP2PExpandNode(exNode *ExpandNode) (task_id int64, e error) {
	//检测是否存在任务
	ex, e := dataSource.Raw.GetExpandNode(exNode.Group, exNode.Node, exNode.MD5)
	if e != nil {
		return
	}
	if ex != nil {
		logger.AppendObj(nil, "AddOrUpdateExpandNode--existExpand ", exNode.Group, "md5: ", exNode.MD5, "node: ", exNode.Node, ex.ID)
	}
	return dataSource.Raw.AddOrUpdateExpandNode(exNode)
}

//
func getGroupLimitCount(SafePieces uint32, part uint32) (count int) {
	if part == 0 {
		return 0
	}

	count = int(SafePieces + (SafePieces / part))
	return
}
