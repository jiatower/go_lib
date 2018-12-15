package p2p_storage

type IDataSource interface {
	/*
		自增ID
	*/
	AtomicIncrID(key string) (uint64, error)
	/*
		获取自增ID，如果key不存在，返回0
	*/
	GetIncrID(key string) (uint64, error)
	/*
	   获取节点表已完成超时检查的时间点（秒数）。

	   参数：
	   返回值：
	   		tm: 上次检查到的时间点（秒数）
	*/
	GetTimeoutNodeCheckedTime() (tm int64, e error)
	/*
	   更新节点表已完成超时检查的时间点（秒数）。

	   参数：
	   		tm: 新的时间点（秒数）
	   返回值：
	*/
	UpdateTimeoutNodeCheckedTime(tm int64) (e error)
	/*
	   获取state状态下文件在各分组的详情

	   参数：
	   		md5: 文件的md5
	   		state: 文件的状态，ALL/NORMAL/DELETED
	   返回值：
	   		files: 文件分组详情，key-分组ID
	*/
	GetFileGroups(md5 string, state int) (files map[string]GroupFile, e error)
	//获取新增超时文件
	GetNewAddTimeOutGroupFile(tm int64, num int) (files []GroupFile, e error)
	/*
		   随机获取源文件所在的节点列表

		   参数：
		   		md5: 文件的md5
				num: 需要的数量
		   返回值：
		   		ids: 拥有原始文件的节点ID列表
	*/
	GetSourceFileNodes(md5 string, num int) (ids []string, e error)
	/*
		   节点自身是否有这个文件

		   参数：
		   		md5: 文件的md5
				nid: 节点ID
		   返回值：
		   		yes: 是否存在
	*/
	IsNodeHasFile(nid string, md5 string) (yes bool, e error)
	/*
	   获取文件在系统中的数量

	   参数：
	   		md5: 文件的md5
	   返回值：
	   		count: 数量
	*/
	GetSourceFileCount(md5 string) (count int, e error)
	/*
		   根据节点ID筛选其中在线的，并返回Peer信息

		   参数：
		   		ids: 候选节点
				timeout: 超时的时间点
		   返回值：
		   		peers: 当前在线的节点信息
	*/
	GetOnlinePeers(ids []string, timeout int64) (peers []Peer, e error)
	/*
		   获取文件详情

		   参数：
		   		md5: 文件的md5
				state: 状态
		   返回值：
		   		file: 文件详情，nil-在e==nil时表示未找到
	*/
	GetFileByMd5AndState(md5 string, state int) (files []GroupFile, e error)
	/*
	   获取分组文件详情

	   参数：
	   		gid: 分组ID
	   		md5: 文件的md5
	   返回值：
	   		file: 文件详情，nil-在e==nil时表示未找到
	*/
	GetGroupFile(gid, md5 string) (file *GroupFile, e error)
	/*
		   获取版本号大于ver的文件详情

		   参数：
		   		gid: 分组ID
		   		ver: 起始版本号
				num: 要获取的数量
		   返回值：
		   		files: 文件分组详情
	*/
	ListUpdatedFiles(gid string, ver uint64, num int, tp int) (files []GroupFile, e error)
	/*
		根据分组中的文件计算分组的大小，要剔除掉已删除的文件
	*/
	CalculateGroupSize(gid string) (e error)
	/*
		获取包含该文件的分组数量，不包括文件状态是已删除的分组
	*/
	GetFileGroupsCount(md5 string) (count int, e error)

	/*
		获取包含该文件的分组数量，不包括文件状态是已删除的分组
	*/
	GetMoreFileGroupsCount(md5s []string) (m map[string]bool, e error)

	/*
		向分组添加文件

		参数：
			gid: 分组ID
			file: 文件详情
		返回值：
	*/
	AddFileToGroup(gid string, file *GroupFile) (e error)
	UpdateGroupFile(gid string, file *GroupFile) (e error)
	UpdateGroupFileTpAndVer(gid, md5 string, ver uint64) (e error)
	DeleteGroupFile(gid string, md5 string) (e error)
	//添加文件版本
	IncrFileVer(gid string, md5 string, ver uint64) (e error)

	/*
	   获取超时的节点

	   参数：
	   		from: 起始时间（秒数）
	   		to: 截止时间（秒数）
	   		num: 最多获取多少个节点
	   返回值：
	   		nodes: 超时节点列表
	*/
	GetTimeoutNodes(from int64, to int64, num int) (nodes []NodeDetail, e error)
	AddNode(node *NodeDetail) (e error)
	DeleteNode(id string) (e error)
	IsNodeExist(nid string) (exist bool, e error)
	UpdateNode(node *NodeDetail) (e error)
	UpdateNodeWeight(nid string, weight float64) (e error)
	/*
		获取剩余空间足够的在线节点，注册时间也要超过一定时间

		参数：
			groupCapacity: 分组的容量（字节）
			updateTm: 上次更新时间必须晚于此时刻
			regTm: 注册时间必须早于此时间点
			num: 需要的节点数量
		返回值：
			nodes: 节点ID列表
	*/
	GetAvailableNodes(groupCapacity uint64, updateTm, regTm int64, offset, num uint32, active_groups int8, online_cnt int) (nodes []string, e error)
	GetAvailableNodesCount(groupCapacity uint64, updateTm, regTm int64, activ_groups int8, online_cnt int) (num uint32, e error)
	//获取最近更新且符合添加到组的节点
	GetAvailableNode(groupCapacity uint64, updateTm, regTm int64, online_cnt, num int) (nodes []string, e error)
	/*
		获取占用空间最小的分组

		参数：
			groupCapacity: 分组的容量（字节）
			fileSize: 分组文件大小范围
		返回值：
			group: 分组详情
	*/
	GetAvailableGroup(fileSize uint32) (group *Group, e error)
	GetGroup(gid string) (group *Group, e error)
	GetAllGroup() (groups map[string]Group, e error)
	AddGroup(group *Group) (e error)
	UpdateGroupSize(group *Group, filesize int64) (e error)
	/*
		获取每种文件尺寸范围的可用分组数量

		参数：
			groupCapacity: 分组的容量（字节）
		返回值：
			groups: 分组数量，key-file_size，value-数量
	*/
	GetActiveGroupsCount(groupCapacity uint64) (groups map[uint32]uint32, e error)
	/*
		获取每种文件尺寸范围的可用分组剩余空间

		参数：
			groupCapacity: 分组的容量（字节）
		返回值：
			groups: 分组剩余空间，key-file_size，value-字节
	*/
	GetActiveGroupsLeftSpace(groupCapacity uint64) (groups map[uint32]uint64, e error)
	/*
		向分组添加节点

		参数：
			gid: 分组ID
			node: 节点信息
		返回值：
	*/
	AddNodeToGroup(gid string, node *GroupNode) (e error)
	UpdateGroupNode(gid string, node *GroupNode, isVerChange bool) (e error)
	DeleteGroupNode(gid, nid string) (e error)
	/*
		获取分组中节点版本号>=ver的在线节点
	*/
	GetFileNodes(gid string, ver uint64) (nodes []Peer, e error)
	/*
		获取分组中节点版本号<ver的所有节点，包括不在线的
	*/
	GetNoFileNodes(gid string, ver uint64) (nodes []string, e error)
	/*
		获取分组中所有节点的MD5值，包括不在线的
	*/
	GetAllFileNodes(gid string) (nodes []string, e error)
	GetGroupNodes(gid string) (nodes []GroupNode, e error)
	/*
		随机获取分组中一个在线的节点
	*/
	GetRandomGroupNode(gid string) (node *GroupNode, e error)
	/*
	   获取节点详情

	   参数：
	   		nid: 节点ID
	   返回值：
	   		detail: 节点详情，nil-在e==nil时表示未找到
	*/
	GetNodeDetail(nid string) (detail *NodeDetail, e error)
	/*
		获取节点所属的group集合

		参数：
			nid: 节点ID
	*/
	GetNodeGroups(nid string) (groups []Group, e error)

	/*
		随机获取 一个可用分组
	*/
	GetRandomNodeGroup(nid string) (group Group, e error)

	/*
		获取节点所属的group数量

		参数：
			nid: 节点ID
	*/
	GetNodeGroupCount(nid string) (num uint32, e error)
	/*
		获取节点所在分组详情

		参数：
			nid: 节点ID
	*/
	GetNodeGroupDetail(nid string) (groups []NodeGroupDetail, e error)
	GetNodeGroupState(nid string) (groups map[string]GroupNode, e error)
	/*
		获取分组中在线节点的数量

		参数：
			gid: 分组ID
	*/
	GetGroupOnlineNodesCount(gid string) (num uint32, e error)
	/*
		获取UPNP可用的节点列表，按节点更新时间升序排列返回
	*/
	GetUPNPAvailableNodes(num int, updateTm int64) (nodes []Peer, e error)

	AddToInvalidFile(nid, gid, md5 string, tm int64) (e error)

	UpdateChecksum(md5, checksum string) (e error)
	/*
			获取文件的checksum

		   参数：
		   		md5: 文件的md5
		   返回值：
		   		checksum: 校验和，"" - 在e==nil时表示未找到
	*/
	GetChecksum(md5 string) (checksum string, e error)

	IncrementActiveGroups(nid string) (e error)

	/*
			获取当前正在扩充的节点列表

		   参数：
		   		gid: 分组ID
		   		md5: 文件的md5
		   返回值：
		   		exNodes: 正在扩充节点列表
	*/
	GetValidExpandNodes(gid, md5 string) (exNodes []ExpandNode, e error)
	/*
			获取扩充节点信息

		   参数：
		   		gid: 分组ID
				nid: 节点ID
		   		md5: 文件的md5
		   返回值：
		   		exNode: 扩充节点详情，nil - 在e==nil时表示未找到
	*/
	GetExpandNode(gid, nid, md5 string) (exNode *ExpandNode, e error)

	/*
			获取扩充节点信息

		   参数：
		   		id: ID
		   返回值：
		   		exNode: 扩充节点详情，nil - 在e==nil时表示未找到
	*/
	GetExpandNodeById(id uint64) (exNode *ExpandNode, e error)

	/*
			获取某个节点的扩充任务列表

		   参数：
				nid: 节点ID
				state: 任务状态
				num: 返回的数量上限
		   返回值：
		   		exNodes: 扩充任务列表
	*/
	GetExpandTasks(nid string, state int8, num int) (exNodes []ExpandNode, e error)

	/*
			添加或更新扩充节点信息

		   参数：
		   		exNodes: 扩充节点详情
		   返回值：
	*/
	AddOrUpdateExpandNode(exNode *ExpandNode) (task_id int64, e error)

	/*
			更新扩充节点状态

		   参数：
		   		gid: 扩充节点所在分组
		   		nid: 扩充节点ID
				md5: 要扩充的文件
				state: 扩充节点状态
		   返回值：
	*/
	UpdateExpandNodeState(gid, nid, md5 string, state int8, timouet int64, increment_failed_times bool) (e error)

	/*
			批量更新扩充节点状态

		   参数：
		   		nid: 扩充节点ID
				state: 扩充节点状态
		   返回值：
	*/
	UpdateExpandNodesState(nid string, state int8, timouet int64) (e error)
	/*
			更新扩充节点超时时间

		   参数：
		   		gid: 扩充节点所在分组
		   		nid: 扩充节点ID
				md5: 要扩充的文件
		   返回值：
	*/
	UpdateExpandNodeTimeout(id uint64, timouet int64) (e error)
	/*
			删除扩充节点

		   参数：
		   		gid: 扩充节点所在分组
		   		nid: 扩充节点ID
				md5: 要扩充的文件
		   返回值：
	*/
	DeleteExpandNode(gid, nid, md5 string) (e error)

	/*
			获取分组文件总失败次数

			参数：
				gid: 扩充节点所在分组
				md5: 要扩充的文件
		 	返回值：
				times: 总失败次数
	*/
	GetExpandTaskTotalFailedTimes(gid, md5 string) (times uint32, e error)

	/*
		根据node获取该节点等待做的任务数

		参数：
			node:节点id
		返回值：
			cnt: 任务数
	*/
	GetExpandTaskCount(node string) (cnt uint32, e error)

	/*
	   获取扩散任务表中已完成超时检查的时间点（秒数）。

	   参数：
	   返回值：
	   		tm: 上次检查到的时间点（秒数）
	*/
	GetTimeoutExpandTaskCheckedTime() (tm, id int64, e error)

	/*
	   更新扩散表已完成超时检查的时间点（秒数）。

	   参数：
	   		tm: 新的时间点（秒数）
	   返回值：
	*/
	UpdateTimeoutExpandTaskCheckedTime(tm, id int64) (e error)

	/*
	   获取超时扩展任务

	   参数：
	   		from: 起始时间（秒数）
	   		to: 截止时间（秒数）
	   		num: 最多获取多少个节点
	   返回值：
	   		tasks: 超时节点列表
	*/
	GetTimeoutExpandTask(from int64, to int64, lastId int64, num int) (nodes []ExpandNode, e error)
	/*
		根据ID批量获取节点

			参数：
				gid: 扩充节点所在分组
				md5: 要扩充的文件
		 	返回值：
				times: 总失败次数
	*/
	GetNodesByIds(ids []string) (nodes []NodeDetail, e error)

	/*
		添加任务节点关系表
	*/
	AddTaskNode(task_id uint64, nids []string) (e error)

	/*
		根据扩散任务id删除所有任务节点值
	*/
	DeleteTaskNodeByTask(id uint64) (e error)

	/*
			删除扩充节点

		   参数：
		   		id: 任务id
		   返回值：
	*/

	DeleteExpandNodeById(id uint64) (e error)
	DeleteExpandNodeByMd5(md5 string) (e error)

	/*
		根据超时时间删除任务记录

		参数：
			tm:查询超时时间
	*/
	DeleteExpandNodeByTimeOut(tm uint64) (e error)

	/*
		   参数：
		   		gid: 组id
				nid: 节点id
		   返回值：
		   		ver:版本
	*/
	GetGroupFileVer(gid, nid string) (ver uint64, e error)

	/*
		   参数：
		   		gid: 组id
				nid: 节点id
				ver: 版本
				num: 数量
		   返回值：
		   	 	files 返回需要文件数组
	*/
	GetGroupFileByVer(gid, nid string, ver uint64, num int) (files []GroupFile, e error)

	/*
		获取某组中已经同步某文件的节点数
			   参数：
				   		gid: 组id
						ver: 版本
			   返回值：
				   		cnt:  符合条件的节点数
	*/
	GetFileNodesCountByVer(gid string, ver uint64) (cnt uint32, e error)

	/*
		获取超时并可以删除的节点
		参数：
				tm:时间
				num:个数
		返回值：
				获取节点列表
	*/
	GetCanDelTimeoutNodes(tm uint64, num int) (nodes []string, e error)

	/*
		检测是否完成了首次扩散
		参数：
			gid:组id
			ver:文件版本
		返回值：
		    返回bool
	*/
	CheckIsFinishFirstExpand(gid string, ver uint64) (finish bool, e error)

	/*
		根据文件版本和节点状态，获取符合条件的节点数
		参数：
			gid:组id
			ver:文件版本
			state: 节点状态
		返回值：
		    返回num
	*/
	GetNodeCountByVerAndState(gid string, ver uint64, state int) (num uint32, e error)

	/*
		checker文件中各种携程检测是否启动依据，需要获取改时间戳
	*/
	GetAtomicLastCheckerTm(key string) (tm int64, e error)

	/*
		checker文件中各种携程检测是否启动依据，需要更新改时间戳
	*/
	SetAtomicGetLastCheckerTm(key string, tm int64, expire_second int) (e error)

	/*
		修改组的首次扩散完成的版本号
	*/
	UpdateGroupFirstFinishVer(gid string, ver uint64) (e error)

	/*
		获取某个组首次扩散完成的文件版本
	*/
	GetGroupFirstFinishExpandVer(gid string) (finish_ver uint64, e error)

	/*
		获取节点的在线时长
	*/
	GetNodeOnlineTm(nids []string) (node_online_map map[string]int, e error)

	/*
		分页获取节点id
	*/
	GetAllNode(begin string) (nodes []string, e error)

	/*
		更新节点在线时长
	*/
	UpdateNodeOnlineCnt(nodeMap map[string]int) (e error)

	/*
			获取某个节点的危险扩充任务列表

		   参数：
				nid: 节点ID
				state: 任务状态
				num: 返回的数量上限
		   返回值：
		   		exNodes: 扩充任务列表
	*/
	GetUnSafeExpandTasks(nid string, state int8, num int) (exNodes []UnSafeExpandNode, e error)

	/*
		修改危险任务状态
	*/
	UpdateUnSafeExpandNodeState(id uint64, state int) (e error)

	/*
			获取危险上传信息

		   参数：
		   		id: ID
		   返回值：
		   		exNode: 扩充节点详情，nil - 在e==nil时表示未找到
	*/
	GetUnSafeExpandNodeById(id uint64) (exNode *UnSafeExpandNode, e error)

	/*
		根据版本号获取文件

		参数：
			gid: 组
			start,end: 开始和终止版本
	*/
	//	GetGroupFileByVerRange(gid string, start, end int64) (files []GroupFile, e error)

	//添加文件到unsafe_group_file中
	AddOrUpdateUnSafeFile(gid, md5 string) (e error)

	/*
		批量获取危险扩散节点

		参数：
			exNodes : 任务数组
	*/
	AddOrUpdateUnSafeExpandNodes(exNodes []UnSafeExpandNode) (e error)

	/*
		检测危险文件是否已经添加上传任务

		参数：
			gid:
			md5:
	*/
	/*CheckUnSafeFileExist(gid, md5 string) (exist bool, e error)*/

	/*
		删除危险文件

		参数：
			gid:
			md5:
	*/
	DeleteUnSafeFile(gid, md5 string) (e error)

	/*
	 删除危险文件ExpandNode

	*/
	DeleteUnSafeFileExpandNode(gid, node, md5 string) (e error)

	/*
		获取危险文件任务集合
	*/
	GetUnSafeFileExpandNode() (exNodes []UnSafeExpandNode, e error)

	/*
		获取拥有某文件危险 piece 的节点
	*/
	GetHasUnSafeFileNode(gid, md5 string, num uint32, ex_nids []string) (nids []string, e error)

	/*
		更改groupfile的state和add_ver
	*/
	UpdateGroupFileStateAndAddVer(gid string, md5 string, state int, add_ver uint64) (e error)
	/*
		获取各组在线节点或离线节点数量
	*/
	GetGroupNodeCountByState(state int) (countMap map[string]int, e error)
	/*
		从config表中获取配置表
	*/
	GetMapFromConfig(configMap map[interface{}]interface{}) (e error)
	/*
		获取active_group<=0的节点
	*/
	GetNodesAGZero(groupCapacity uint64, updateTm, regTm int64, active_groups int8, online_cnt int) (nodes map[string]bool, e error)
	/*
		获取新加入没有任何分组的节点
	*/
	GetNewNodes(groupCapacity uint64, updateTm, regTm int64, online_cnt int) (nodes map[string]bool, e error)
	/*
		获取任务卡主的组和节点
	*/
	GetGroupNodesTaskProcessSlow(nowTm int64) (groupNodesMap map[string]GroupNode, e error)

	/*
		重置未超时且未完成、未失败的任务节点的状态
	*/
	SetExpandNodeStateFailed(node string) (e error)

	/**
	分布式锁，通过redis实现
	*/
	GetLock(db int, key string, expireSec int64, timeout int64) (getLock bool)

	/**
	释放锁
	*/
	UnLock(db int, key string) (e error)
}
