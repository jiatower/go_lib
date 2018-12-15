package p2p_storage

import (
	"errors"
	"yh_pkg/log"
	"yh_pkg/service"
	"yh_pkg/time"
	"yunhui/redis_db"
)

type DataSource struct {
	//数据接口的实现
	Raw IDataSource
}

func newDataSource(imp IDataSource) (ds *DataSource) {
	return &DataSource{imp}
}

func (ds *DataSource) AddFileToGroup(gid string, g *Group, file *GroupFile, fileVer uint64) (e error) {
	f, e := ds.Raw.GetGroupFile(gid, file.MD5)
	if e != nil {
		return
	}
	if f != nil {
		if f.State == DELETED {
			if e := ds.Raw.UpdateGroupFile(gid, file); e != nil {
				return e
			}
			if fileVer >= g.FirstFinishVer {
				logger.AppendObj(nil, "AddFileToGroup-FillEmptyGroupFile--gid:", gid, "fileVer:", fileVer, "groupFirstFinishVer:", g.FirstFinishVer)
				if e := ds.FillEmptyGroupFile(gid); e != nil {
					return e
				}
			}
			return ds.UpdateGroupSize(true, g, file.Size)
		}
		return
	} else {
		if e = ds.Raw.AddFileToGroup(gid, file); e != nil {
			return
		}
		return ds.UpdateGroupSize(true, g, file.Size)
	}
	return
}

func (ds *DataSource) IsFileExists(md5 string) (ok bool, e error) {
	c, e := ds.Raw.GetFileGroupsCount(md5)
	if e != nil {
		return
	}
	return c != 0, nil
}

func (ds *DataSource) IsMoreFileExists(md5s []string) (m map[string]bool, e error) {
	return ds.Raw.GetMoreFileGroupsCount(md5s)

}

//判断是否需要添加该文件
func (ds *DataSource) IsP2PFileExists(md5 string) (ok bool, file *GroupFile, e error) {
	files, e := ds.Raw.GetFileByMd5AndState(md5, NORMAL)
	if e != nil {
		return
	}
	if len(files) > 0 {
		ok = true
		file = &files[0]
	}
	return
}

/*
	更新分组大小

	参数：
		isAdd: 是否是添加文件，false表示是删除文件
		gid: 分组的ID
		size: 文件大小
	返回值：
*/
func (ds *DataSource) UpdateGroupSize(isAdd bool, group *Group, size uint64) (e error) {
	logger.AppendObj(e, "UpdateGroupSize-before_size:", group.ID, group.Size, size)
	var filesize int64
	m := size % 10
	if m == 0 || m == 1 || m == 2 {
		e = dataSource.Raw.CalculateGroupSize(group.ID)
		return
	} else {
		if isAdd {
			filesize = int64(size)
		} else {
			filesize = -int64(size)
		}
	}
	logger.AppendObj(e, "UpdateGroupSize-size:", group.ID, group.Size, filesize)
	return ds.Raw.UpdateGroupSize(group, filesize)
}

/*
	   随机获取源文件所在的节点列表

	   参数：
	   		md5: 文件的md5
			num: 需要的数量
	   返回值：
	   		peers: 拥有原始文件的节点ID列表
*/
func (ds *DataSource) GetOnlineSourceFileNodes(md5 string, num int) (peers []Peer, e error) {
	ids, e := ds.Raw.GetSourceFileNodes(md5, num)
	if e != nil {
		return
	}
	return ds.Raw.GetOnlinePeers(ids, time.Now.Unix()-NODE_VALID_TIME)
}

func (ds *DataSource) GetOnlineNodesByIds(ids []string, min_update_tm int64) (peers []Peer, e error) {
	return ds.Raw.GetOnlinePeers(ids, min_update_tm)
}

/*
   获取分组中没有该文件碎片的节点

   参数：
   		gid: 分组ID
   		md5: 文件的md5
   返回值：
   		nodes: 没有该文件的节点列表
*/
func (ds *DataSource) GetNoFileNodes(gid, md5 string) (nodes []string, e error) {
	file, e := ds.Raw.GetGroupFile(gid, md5)
	if e != nil {
		return nil, e
	}
	if file == nil {
		return nil, errors.New("file not exists")
	}
	if file.State == DELETED {
		return nil, errors.New("file deleted")
	}
	if file.Type == GROUPFILE_TYPE_SPRAND_FIRST {
		return ds.Raw.GetNoFileNodes(gid, file.Ver)
	} else if file.Type == GROUPFILE_TYPE_NEW_ADD {
		return ds.Raw.GetAllFileNodes(gid)
	} else {
		e = errors.New("file Type error")
		return
	}

}

/*
   取出还未通知节点的任务列表

   参数：
   		nid: 节点ID
   		num: 最多返回的个数
   返回值：
   		exNodes: 任务列表
*/
func (ds *DataSource) FetchExpandTasks(nid string, num int) (exnodes []ExpandNode, e error) {
	exNodes, e := ds.Raw.GetExpandTasks(nid, EXPAND_STATE_INIT, num)
	if e != nil {
		return nil, e
	}
	exnodes = make([]ExpandNode, 0, len(exNodes))
	for _, exNode := range exNodes {
		if e = ds.UpdateExpandNodeState(exNode.Group, nid, exNode.MD5, EXPAND_STATE_NOTIFIED); e != nil {
			logger.Append("UpdateExpandNodeStat error: "+e.Error(), log.ERROR)
		}

		//需要给节点返回任务的文件对应的版本号
		f, e := ds.Raw.GetGroupFile(exNode.Group, exNode.MD5)
		if e != nil {
			return nil, e
		}
		if f != nil {
			exNode.Ver = f.Ver
		}
		exnodes = append(exnodes, exNode)
	}
	return
}

/*
   更新任务状态

   参数：
   		gid: 分组ID
   		nid: 节点ID
		md5: 文件
   		state: 任务状态
		increment_failed_times: 是否增加失败次数
   返回值：
   		exNodes: 任务列表
*/
func (ds *DataSource) UpdateExpandNodeState(gid, nid, md5 string, state int8) (e error) {
	increment := false
	if state == EXPAND_STATE_FAILED {
		increment = true
		//如果任务失败并且获取该文节点数小于min_perrs时则则需要将group_file的版本添加
		if e = IncrGroupFileVer(gid, md5); e != nil {
			return e
		}
	}
	return dataSource.Raw.UpdateExpandNodeState(gid, nid, md5, state, CalculateExpandNodeTimeout(state), increment)
}

/*
   更新任务超时时间

   参数：
   		gid: 分组ID
   		nid: 节点ID
		md5: 文件
   返回值：
   		exNodes: 任务列表
*/
func (ds *DataSource) UpdateExpandNodeTimeout(id uint64) (e error) {
	return dataSource.Raw.UpdateExpandNodeTimeout(id, CalculateExpandNodeTimeout(EXPAND_STATE_STARTED))
}

/*
   批量获取节点ids

   参数：
		ids:节点ids
   返回值：
   		exNodes: 节点列表
*/
func (ds *DataSource) GetNodesByIds(ids []string) (nodes []NodeDetail, e error) {
	return dataSource.Raw.GetNodesByIds(ids)
}

/*
	填充空白文件版本记录

*/

func (ds *DataSource) FillEmptyGroupFile(gid string) (e error) {
	var emptyFile = "00000000000000000000000000000000"
	f, e := ds.Raw.GetGroupFile(gid, emptyFile)
	if e != nil {
		return
	}

	/*ver, e := dataSource.Raw.AtomicIncrID(gid)
	if e != nil {
		return errors.New("redis error : " + e.Error())
	}

	if f == nil {
		if e = ds.Raw.AddFileToGroup(gid, &GroupFile{File{emptyFile, 0}, ver, DELETED, gid, 0, 0, uint64(time.Now.Unix()), ""}); e != nil {
			return
		}
		logger.AppendObj(nil, "FillEmptyGroupFile-add-new-empty-file--gid:", gid, "ver:", ver)
	} else { // 已有空白文件记录则更新原记录
		verRecord := f.Ver
		f.Ver, f.LastAddTm = ver, uint64(time.Now.Unix())
		if e = ds.Raw.UpdateGroupFile(gid, f); e != nil {
			return e
		}
		logger.AppendObj(nil, "FillEmptyGroupFile-update-empty-file--gid:", gid, "ver from:", verRecord, "to:", f.Ver)
	}*/

	//获取锁
	if !dataSource.Raw.GetLock(redis_db.CACHE_THUNDER_REQUEST_POOL, gid, P2pLockExpireSec, P2pGetLockTimeOut) {
		e = service.NewSimpleError(service.ERR_PERMISSION_DENIED, "get lock is errror")
		logger.AppendObj(e, "P2pLock-FillEmptyGroupFile has no lock", gid)
		return
	}
	e = doFillEmptyGroupFile(ds, gid, f, emptyFile)

	if err := dataSource.Raw.UnLock(redis_db.CACHE_THUNDER_REQUEST_POOL, gid); err != nil {
		logger.AppendObj(err, "P2pLock-FillEmptyGroupFile unlock is error", gid)
	}

	return
}

func doFillEmptyGroupFile(ds *DataSource, gid string, f *GroupFile, emptyFile string) (e error) {
	ver, e := dataSource.Raw.AtomicIncrID(gid)
	if e != nil {
		return errors.New("redis error : " + e.Error())
	}

	if f == nil {
		if e = ds.Raw.AddFileToGroup(gid, &GroupFile{File{emptyFile, 0}, ver, DELETED, gid, 0, 0, uint64(time.Now.Unix()), ""}); e != nil {
			return
		}
		logger.AppendObj(nil, "FillEmptyGroupFile-add-new-empty-file--gid:", gid, "ver:", ver)
	} else { // 已有空白文件记录则更新原记录
		verRecord := f.Ver
		f.Ver, f.LastAddTm = ver, uint64(time.Now.Unix())
		if e = ds.Raw.UpdateGroupFile(gid, f); e != nil {
			return e
		}
		logger.AppendObj(nil, "FillEmptyGroupFile-update-empty-file--gid:", gid, "ver from:", verRecord, "to:", f.Ver)
	}
	return
}
