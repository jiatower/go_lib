package p2p_storage

import (
	"strings"
	"time"
	"yh_pkg/log"
	tm "yh_pkg/time"
)

//检测当前是否可以启动检测服务
func checkCanRunService(key string) (can bool) {
	last_tm, e := dataSource.Raw.GetAtomicLastCheckerTm(key)
	if e != nil {
		logger.AppendObj(e, "checkCanRunService-is error--", key)
		return
	}
	if last_tm <= 0 {
		can = true
	}
	//	logger.AppendObj(e, "checkCanRunService---key", key, last_tm, "-res:", can)
	return
}

func getCheckExpireTm(key string) (expire_second int) {
	if key == CHECKER_TIMEOUT_LAST_TM {
		expire_second = 60
	} else if key == CHECKER_CREATEGROUP_LAST_TM {
		expire_second = 60
	} else if strings.Contains(key, CHECKER_EXPAND_NODE_PREFIX) {
		expire_second = 30
	} else if strings.Contains(key, CHECKER_NODE_ONLINETM_CHECKER) {
		expire_second = NODE_ONLINETM_INTERVAL_TM * 3600
	} else if key == CHECKER_EXPAND_TASK_TIME {
		expire_second = 10
	} else if strings.Contains(key, CHECKER_GEN_PIECETM_PRIFIX) {
		expire_second = 60
	} else if key == CHECKER_GROUP_FILE_NEW_ADD_TIMEOUT {
		expire_second = 300
	} else if key == CHECKER_GROUP_EXPAND {
		expire_second = 300
	} else if key == CHECKER_ONLINE_NODE {
		expire_second = 300
	}
	return

}

func checkTimeoutNodes() {
	time.Sleep(time.Second * time.Duration(NODE_VALID_TIME))
	for {
		time.Sleep(time.Minute * 1)

		//获取检测时间，并判断是否需要执行检测,间隔时间去检测
		if !checkCanRunService(CHECKER_TIMEOUT_LAST_TM) {
			continue
		}
		if e := dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_TIMEOUT_LAST_TM, tm.Now.Unix(), getCheckExpireTm(CHECKER_TIMEOUT_LAST_TM)); e != nil {
			logger.Append("GetNodeCheckedTime setTm error: "+e.Error(), log.ERROR)
			continue
		}

		lastCkTm, e := dataSource.Raw.GetTimeoutNodeCheckedTime()
		if e != nil {
			logger.Append("GetNodeCheckedTime error: "+e.Error(), log.ERROR)
			continue
		}
		nodes, e := dataSource.Raw.GetTimeoutNodes(lastCkTm, tm.Now.UnixNano()-NODE_VALID_TIME*int64(time.Second), 100)
		if e != nil {
			logger.Append("GetTimeoutNodes error: "+e.Error(), log.ERROR)
			continue
		}
		success := true
		for _, node := range nodes {
			groupNodes, e := dataSource.Raw.GetNodeGroupState(node.ID)
			if e != nil {
				logger.Append("GetNodeGroupState error: "+e.Error(), log.ERROR)
				success = false
				break
			}
			for gid, node := range groupNodes {
				if node.State == ONLINE {
					node.State = OFFLINE
					if e = dataSource.Raw.UpdateGroupNode(gid, &node, false); e != nil {
						logger.Append("UpdateGroupNode error: "+e.Error(), log.ERROR)
						success = false
						break
					}
					group, e := dataSource.Raw.GetGroup(gid)
					if e != nil {
						logger.Append("GetGroup error: "+e.Error(), log.ERROR)
						success = false
						break
					}
					if group == nil {
						logger.Append("GetGroup error: "+gid+"not found", log.ERROR)
						success = false
						break
					}
					if e = group.ExpandNodesToPerfectSize(NODE_MAX_ACTIVE_GROUPS, ""); e != nil {
						logger.Append("ExpandNodesToPerfectSiz error: "+e.Error(), log.ERROR)
						success = false
						break
					}
				}
			}
			if e = dataSource.Raw.UpdateExpandNodesState(node.ID, EXPAND_STATE_FAILED, CalculateExpandNodeTimeout(EXPAND_STATE_FAILED)); e != nil {
				logger.Append("UpdateExpandNodeStat error: "+e.Error(), log.ERROR)
			}

			if !success {
				break
			}
			if lastCkTm < node.UpdateTm {
				lastCkTm = node.UpdateTm
			}
		}
		if success {
			e = dataSource.Raw.UpdateTimeoutNodeCheckedTime(lastCkTm)
		}
	}
}

func createNewGroups() {
	time.Sleep(time.Second * time.Duration(NODE_VALID_TIME))

	idx := 2
	g := GROUP_CONFIG[idx]
	groupCapacity := uint64(g.MinPieces) * GROUP_NODE_CAPACITY
	for {
		time.Sleep(time.Minute * 1)
		//获取检测时间，并判断是否需要执行检测,间隔时间去检测周期的1.5倍，所以设置为90秒
		if !checkCanRunService(CHECKER_CREATEGROUP_LAST_TM) {
			continue
		}
		if e := dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_CREATEGROUP_LAST_TM, tm.Now.Unix(), getCheckExpireTm(CHECKER_CREATEGROUP_LAST_TM)); e != nil {
			logger.Append("SetAtomicGetLastCheckerTm setTm error: "+e.Error(), log.ERROR)
			continue
		}

		var e error = nil
		var group *Group
		spaces, e := dataSource.Raw.GetActiveGroupsLeftSpace(groupCapacity)
		file_size := MinFileSize(spaces)

		var ratio int = 0
		if v, ok := spaces[file_size]; ok {
			ratio = int(v)
		}

		//需要获取对应容量的分组空闲占比时，才添加新的分组
		if CREATE_NEWGROUP_BALANCE_RATIO <= ratio {
			logger.AppendObj(e, "-createNewGroups-has enouth space", spaces, file_size, ratio, "-ratio_config-", CREATE_NEWGROUP_BALANCE_RATIO)
			continue
		}
		logger.AppendObj(e, "-createNewGroups-:", spaces, file_size, "bal_ratio:", ratio)

		group, e = createGroup(idx, file_size, "")
		if e == nil {
			logger.Append("-createNewGroups- create group "+group.ID, log.DEBUG)
		} else {
			logger.Append("-createNewGroups-no enough available nodes to create new group: "+e.Error(), log.DEBUG)
		}
	}

}

/*
检测扩散任务超时，修改任务对应节点的任务记录, 成功和失败的情况下，在ExpandFinish接口中已经处理过了，现在需要在处理已经获取任务任务或者失败的情况的超时
*/
func checkExpandTaskTimeout() {
	time.Sleep(time.Second * time.Duration(NODE_VALID_TIME))
	for {
		time.Sleep(time.Second * 10)
		//获取检测时间，并判断是否需要执行检测,间隔时间去检测周期的1.5倍，所以设置为90秒
		if !checkCanRunService(CHECKER_EXPAND_TASK_TIME) {
			continue
		}
		if e := dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_EXPAND_TASK_TIME, tm.Now.Unix(), getCheckExpireTm(CHECKER_EXPAND_TASK_TIME)); e != nil {
			logger.Append("GetNodeCheckedTime setTm error: "+e.Error(), log.ERROR)
			continue
		}

		lastCkTm, last_id, e := dataSource.Raw.GetTimeoutExpandTaskCheckedTime()
		if e != nil {
			logger.Append("checkExpandTaskTimeout error: "+e.Error(), log.ERROR)
			continue
		}
		tasks, e := dataSource.Raw.GetTimeoutExpandTask(lastCkTm, tm.Now.Unix()-NODE_VALID_TIME, last_id, 1000)
		if e != nil {
			logger.Append("GetTimeoutExpandTask error: "+e.Error(), log.ERROR)
			continue
		}
		logger.AppendObj(e, "--GetTimeoutExpandTask-lastCkTm：", lastCkTm, tm.Now.Unix(), tasks)

		success := true
		for _, t := range tasks {
			/*
				if e = dataSource.Raw.DeleteTaskNodeByTask(t.ID); e != nil {
					logger.Append("GetTimeoutExpandTask error: "+e.Error(), log.ERROR)
					success = false
					continue
				}
			*/
			if t.State == EXPAND_STATE_FINISHED {
				logger.AppendObj(nil, "DeleteExpandNodeById", t)
				if e = dataSource.Raw.DeleteExpandNodeById(t.ID); e != nil {
					logger.Append("DeleteExpandNodeById error: "+e.Error(), log.ERROR)
					success = false
					continue
				}
			} else {
				//如果任务失败，则需要将group_file的版本添加
				go IncrGroupFileVer(t.Group, t.MD5)
			}

			if !success {
				break
			}
			if lastCkTm < t.Timeout {
				lastCkTm = t.Timeout
			}
			last_id = int64(t.ID)
		}
		if success {
			e = dataSource.Raw.UpdateTimeoutExpandTaskCheckedTime(lastCkTm, last_id)
		}
	}
}

//检测并将长时间不在线的node
func checkDelLongTimeOutNode() {
	time.Sleep(time.Second * 2 * time.Duration(NODE_VALID_TIME))
	for {
		for {
			time.Sleep(time.Minute * 1)
			//纳秒
			tm := (time.Now().Unix() - NODE_DELETE_TIMEOUT*24*3600) * 1000000000
			if tm < 0 {
				tm = 0
			}
			nodes, e := dataSource.Raw.GetCanDelTimeoutNodes(uint64(tm), 1000)
			logger.AppendObj(nil, "checkDelLongTimeOutNode--GetCanDelTimeoutNodes", tm, nodes)
			if e != nil {
				logger.Append("checkDelLongTimeOutNode--GetCanDelTimeoutNodes error: "+e.Error(), log.ERROR)
				continue
			}
			if len(nodes) <= 0 {
				break
			}
			for _, id := range nodes {
				logger.AppendObj(e, "checkDelLongTimeOutNode deleteId: ", id)
				if e = DeleteNode(id); e != nil {
					logger.AppendObj(e, "checkDelLongTimeOutNode error: ", id, e.Error())
					continue
				}
			}
		}
		//删除超过7天的任务
		go dataSource.Raw.DeleteExpandNodeByTimeOut(uint64(tm.Now.Unix() - EXPAND_TASK_DELETE_TIME))
		// 间隔12小时执行一次
		time.Sleep(time.Hour * time.Duration(NODE_CHECKDELETE_TM))
	}
}

//间隔时间段内统计节点在线时间，并更新到node表中,作为添加节点的判断条件
func updateNodeOnlineTime() {
	var i int
	for {
		if i > 0 {
			time.Sleep(time.Hour * time.Duration(NODE_ONLINETM_INTERVAL_TM))
		}
		i++
		if !checkCanRunService(CHECKER_NODE_ONLINETM_CHECKER) {
			logger.AppendObj(nil, "UpdateNodeOnlineTime contine", i)
			continue
		}
		if e := dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_NODE_ONLINETM_CHECKER, tm.Now.Unix(), getCheckExpireTm(CHECKER_NODE_ONLINETM_CHECKER)); e != nil {
			logger.Append("UpdateNodeOnlineTime setTm error: "+e.Error(), log.ERROR)
			continue
		}

		tm1 := tm.Now.Unix()
		logger.AppendObj(nil, "UpdateNodeOnlineTime start : ", tm1)

		var start_node string
		for {
			nodes, e := dataSource.Raw.GetAllNode(start_node)
			if e != nil {
				logger.AppendObj(e, "UpdateNodeOnlineTime GetAllNod is error: ", start_node)
				break
			}
			if len(nodes) <= 0 {
				logger.AppendObj(e, "UpdateNodeOnlineTime  has no nodes need update", start_node)
				break
			}

			//查询 nodes 在 storage_log_id 表中的配置天数内在线的次数
			//统计所有在线的存储，并更新到node节点中
			online_map, e := dataSource.Raw.GetNodeOnlineTm(nodes)
			if e != nil {
				logger.Append("UpdateNodeOnlineTime GetNodeOnlineTm error: "+e.Error(), log.ERROR)
				break
			}

			update_map := make(map[string]int)
			for _, node := range nodes {
				var cnt int
				if v, ok := online_map[node]; ok {
					cnt = v
				}
				update_map[node] = cnt
				start_node = node
			}
			logger.AppendObj(e, "UpdateNodeOnlineTime update: ", update_map)
			if e = dataSource.Raw.UpdateNodeOnlineCnt(update_map); e != nil {
				logger.AppendObj(e, "UpdateNodeOnlineTime UpdateNodeOnlineTime is error: ", update_map)
				break
			}
		}
		tm2 := tm.Now.Unix()
		logger.AppendObj(nil, "UpdateNodeOnlineTime end cost: ", tm1, tm2-tm1)
	}
}

//间隔时间段内删除group_file 中不活跃文件（上次AddP2pFile时间距离现在超过1周但还未完成的文件）
func clearNewAddGroupFileTimeOut() {
	var oneDaySec int64 = 86400
	var i int64
	for {
		if i > 0 {
			time.Sleep(time.Minute * time.Duration(GROUP_FILE_NEW_ADD_DIFF_TIME))
		}
		i++
		if !checkCanRunService(CHECKER_GROUP_FILE_NEW_ADD_TIMEOUT) {
			logger.AppendObj(nil, "clearNewAddGroupFileTimeOut contine")
			continue
		}
		for {
			if e := dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_GROUP_FILE_NEW_ADD_TIMEOUT, tm.Now.Unix(), getCheckExpireTm(CHECKER_GROUP_FILE_NEW_ADD_TIMEOUT)); e != nil {
				logger.AppendObj(e, "clearNewAddGroupFileTimeOut setTm error: ")
				break
			}

			tm1 := tm.Now.Unix()
			files, e := dataSource.Raw.GetNewAddTimeOutGroupFile(tm1-oneDaySec*7, 1000)
			if e != nil {
				logger.AppendObj(e, "clearNewAddGroupFileTimeOut GetNewAddTimeOutGroupFile error:")
				break
			}

			if len(files) <= 0 {
				time.Sleep(time.Minute * time.Duration(GROUP_FILE_NEW_ADD_DIFF_TIME))
				logger.AppendObj(e, "clearNewAddGroupFileTimeOut not files continue:")
				break
			}

			groups := make(map[string]bool)
			for _, f := range files {
				/*
					newAddVer, e := dataSource.Raw.AtomicIncrID(getAtomicIncrKey(f.Group))
					if e != nil {
						logger.AppendObj(e, "clearNewAddGroupFileTimeOut AtomicIncrID error:")
					}
					e = dataSource.Raw.UpdateGroupFileStateAndAddVer(f.Group, f.MD5, DELETED, newAddVer)
					if e != nil {
						logger.AppendObj(e, "clearNewAddGroupFileTimeOut UpdateGroupFileStateAndAddVer error:")
					}
					logger.AppendObj(e, "clearNewAddGroupFileTimeOut UpdateGroupFileStateAndAddVer do_delete ", f)
					groups[f.Group] = true
				*/

				if e = dataSource.Raw.DeleteGroupFile(f.Group, f.MD5); e != nil {
					logger.AppendObj(e, "clearNewAddGroupFileTimeOut deelteGroupFile is error:")
				}

				if e = dataSource.Raw.DeleteExpandNodeByMd5(f.MD5); e != nil {
					logger.AppendObj(e, "clearNewAddGroupFileTimeOut deelteExpandNode is error:")
				}

				logger.AppendObj(e, "clearNewAddGroupFileTimeOut UpdateGroupFileStateAndAddVer do_delete ", f)
				groups[f.Group] = true
			}

			//更新组大小
			for gid := range groups {
				g, e := dataSource.Raw.GetGroup(gid)
				if e != nil {
					logger.AppendObj(e, "clearNewAddGroupFileTimeOut GetGroup error gid:", gid)
					continue
				}
				if g == nil {
					logger.AppendObj(e, "clearNewAddGroupFileTimeOut GetGroup error: can't find group gid:", gid)
					continue
				}
				dataSource.UpdateGroupSize(false, g, 0) // 直接更新group大小
				logger.AppendObj(nil, "clearNewAddGroupFileTimeOut updateGroupSize:", g)
			}
		}
	}
}

//检测各分组在线节点数量，并扩张在线数量不足的分组
func checkExpandGroup() {
	var i int
	for {
		time.Sleep(time.Minute * time.Duration(CHECK_GROUP_EXPAND_TIME))
		i++
		if !checkCanRunService(CHECKER_GROUP_EXPAND) {
			logger.AppendObj(nil, "checkExpandGroup contine", i)
			continue
		}
		if e := dataSource.Raw.SetAtomicGetLastCheckerTm(CHECKER_GROUP_EXPAND, tm.Now.Unix(), getCheckExpireTm(CHECKER_GROUP_EXPAND)); e != nil {
			logger.AppendObj(e, "checkExpandGroup setTm error: ")
			continue
		}

		onlineNodeMap, e := dataSource.Raw.GetGroupNodeCountByState(ONLINE)
		if e != nil {
			logger.AppendObj(e, "checkExpandGroup GetGroupNodeCountByStat: ")
			continue
		}
		//批量查询组信息
		groups, e := dataSource.Raw.GetAllGroup()
		if e != nil {
			logger.AppendObj(e, "checkExpandGroup GetAllGroup ")
			continue
		}

		for groupMd5, count := range onlineNodeMap {
			g, ok := groups[groupMd5]
			if !ok || g.ID == "" {
				logger.AppendObj(nil, "checkExpandGroup GetAllGroupGroup is not exist or g is empty ")
				continue
			}

			logger.AppendObj(e, "checkExpandGroup GetAllGroupGroup ", groupMd5, count, g.SafePieces)
			if uint32(count) < (g.SafePieces + (g.MinPieces / EXPAND_GROUP_ADDRATIO)) {
				group, e := dataSource.Raw.GetGroup(groupMd5)
				if e != nil {
					logger.Append("clearNewAddGroupFileTimeOut GetGroup error: "+e.Error(), log.ERROR)
					continue
				}
				if group == nil {
					logger.AppendObj(nil, "checkExpandGroup group not exists: ", groupMd5)
					continue
				}

				e = group.ExpandNodesToPerfectSize(NODE_MAX_ACTIVE_GROUPS, "")
				if e != nil {
					logger.Append("clearNewAddGroupFileTimeOut ExpandNodesToPerfectSize group: "+groupMd5, log.ERROR)
					continue
				}

				logger.AppendObj(nil, "checkExpandGroup expand group: ", groupMd5)
			}
		}
	}
}

//更新configMap
func updateConfigMap() {
	for {
		if ConfigMap != nil {
			if e := ConfigMap.FlushConfigValue(); e != nil {
				logger.AppendObj(e, "updateConfigMap FlushConfigValue error")
			} else {
				logger.AppendObj(nil, "updateConfigMap success")
			}
		}

		time.Sleep(time.Second * time.Duration(UPDATE_CONFIG_MAP_TIME))
	}

}
