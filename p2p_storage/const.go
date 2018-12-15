package p2p_storage

//分组ID的长度
const GID_LEN uint = 32

//节点的有效期（秒数）
const NODE_VALID_TIME int64 = 600

//组扩散节点选取节点有效时间（2个汇报周期)
const NODE_EXPAND_GROUP_VALID_TIME int64 = 120

//检测超时检点删除间隔时间（小时）
const NODE_CHECKDELETE_TM int64 = 6

//节点超时则将其从p2p系统移除（天）
const NODE_DELETE_TIMEOUT int64 = 30

//节点所属的未满分组的最大数量
const NODE_MAX_ACTIVE_GROUPS int8 = 11
const NODE_MIN_ACTIVE_GROUPS int8 = 9

//新节点所属未满分组最大数量
const NEW_NODE_ACTIVE_GROUP_COUNT int8 = 0

//每次分配给节点扩充任务的数量上限
const MAX_EXPAND_TASK_NUM int = 30

//节点老化时长
const NODE_VALID_AFTER_REGTM int64 = 7 * 86400

//节点空间占用比例（%）
const NODE_OCCUPY_PERCENT int8 = 50

//普通节点(非超级硬盘)空间占用比例(%)
const NORMAL_NODE_OCCUPY_PERCENT int8 = 1

//分组在每个节点上所占用的空间（字节）
const GROUP_NODE_CAPACITY uint64 = 2 * 1024 * 1024 * 1024

//文件尺寸上限
const MAX_FILE_SIZE uint64 = 50 * 1024 * 1024 * 1024

//并行生成碎片的最大节点数量
const MAX_EXPAND_NODE_NUM uint8 = 1

//节点最大的分配任务数
const MAX_NODE_EXPANDTASK_CNT uint32 = 100

//扩散任务最大失败次数
const MAX_EXPAND_TAKS_FAIL_NUMS uint8 = 3

const PIECE_SIZE uint32 = 1024
const PIECE_MIN_NUM uint32 = 32
const PIECE_SAFE_NUM uint32 = 48
const PIECE_PERFECT_NUM uint32 = 64

const YES, NO int = 1, 0
const ONLINE, OFFLINE int = 1, 0
const ALL, NORMAL, DELETED int = 2, 1, 0

const EXPAND_STATE_INIT int8 = 0
const EXPAND_STATE_NOTIFIED int8 = 1
const EXPAND_STATE_STARTED int8 = 2
const EXPAND_STATE_FINISHED int8 = 3
const EXPAND_STATE_FAILED int8 = 4

//危险任务状态 0-初始化 1-完成
const (
	UNSAFE_EXPAND_STATE_INIT     = 0
	UNSAFE_EXPAND_STATE_FINISHED = 1
)

//分组文件状态 0-完成首次扩散 1- 新增文件
const (
	GROUPFILE_TYPE_SPRAND_FIRST = 0
	GROUPFILE_TYPE_NEW_ADD      = 1
)

const EXPAND_MAX_FAIL_TIMES_EACH_NODE uint32 = 1

//node所能同时执行的接收任务数
const NODE_TAKS_MAX_NUM int8 = 100

//扩散任务删除时间
const EXPAND_TASK_DELETE_TIME int64 = 5 * 24 * 3600

//扩散传输文件方式
const EXPAND_TRANS_TYPE_OSS int8 = 0
const EXPAND_TRANS_TYPE_NODE int8 = 1
const EXPAND_TRANS_TYPE_BOTH int8 = 2

//p2p自启动检测服务执行时间戳
const CHECKER_TIMEOUT_LAST_TM string = "checker_last_tm"
const CHECKER_CREATEGROUP_LAST_TM string = "create_group_last_tm"
const CHECKER_EXPAND_NODE_PREFIX string = "expand_prefix_"
const CHECKER_NODE_ONLINETM_CHECKER string = "node_online_checker"
const CHECKER_EXPAND_TASK_TIME string = "checker_expand_task_time"
const CHECKER_GEN_PIECETM_PRIFIX string = "gen_piece_"

// 检测 group_file 中不活跃节点(上次 AddP2pFile时间距离现在超过1周但还未完成的文件)
const CHECKER_GROUP_FILE_NEW_ADD_TIMEOUT string = "new_gf_timeout"

//检测各分组在线节点数量，并扩张在线数量不足的分组
const CHECKER_GROUP_EXPAND string = "group_expand"

//更新ConfigMap
const UPDATE_CONFIG_MAP string = "update_config_map"
const GROUP_FILE_NEW_ADD_DIFF_TIME int = 10

//检测在线节点
const CHECKER_ONLINE_NODE = "checker_node_online"
const CHECKER_ONLINE_NODE_MIN = 5

//检测卡住任务间隔(分钟)
const CHECKER_TASK_PROCESS_SLOW_MIN = 10

//首次扩散完成节点数
const FIRST_EXPAND_FINISH_NUM int = 160

//组扩散时，节点扩散界限 g.SafePiece+g.MinPiece/16
const EXPAND_GROUP_ADDRATIO uint32 = 16

//创建新分组标准，当分组总量剩余百分比时，则创建新分组
const CREATE_NEWGROUP_BALANCE_RATIO int = 20

//节点在线检测间隔(小时)
const NODE_ONLINETM_INTERVAL_TM int = 1

//扩散组节点在线时长配置，需要符合online_cnt数
const NODE_EXPAND_MIN_ONLINE_CNT int = 144

//分组扩张检测时间间隔(分钟)
const CHECK_GROUP_EXPAND_TIME int = 5

//更新ConfigMap时间间隔(秒)
const UPDATE_CONFIG_MAP_TIME int = 60

//代理上行速度限制默认值(字节)
const DEFAULT_DELEGATES_NODE_SPEED int64 = 500 * 1024

//二次扩算任务上行速度限制默认值(字节)
const DEFAULT_SECOND_EXPAND_SPEED int64 = 300 * 1024

//max_hour默认值
const DEFAULT_MAX_HOUR int64 = 8

//二次扩散任务上行速度限制Key值
const SPREAD_MIN_SPEED_CONFIG_KEY string = "sprand_min_speed"

//代理上行速度限制Key值
const DELEGATES_MIN_SPEED_CONFIG_KEY string = "delegate_min_speed"

//max_hour设置key值
const MAX_HOUR_CONFIG_KEY string = "max_hour"

//oss_split_size设置key值
const OSS_SPLIT_SIZE_CONFIG_KEY = "oss_split_size"

//con设置key值
const CON_CONFIG_KEY = "con"

//addp2pfile概率key值
const ADD_P2P_FILE_CONFIG_KEY = "add_p2p_file"

//node直传的key值 1 BOTH 0 OSS
const TRANS_NODE_CONFIG_KEY = "trans_node"

//genPieceLevel设置的key值
const GEN_PIECE_LEVEL_CONFIG_KEY = "gen_piece_level"

//p2p系统节点上行速度限制key值
const P2P_UPSPEED_LIMIT_KEY = "p2p_upspeed_limit"

//p2p节点开启小文件合并
const P2P_MERGE_PIECE = "merge_piece"

//p2p节点开启下载piece缓存
const P2P_DOWNLOAD_CACHE = "download_cache"

//新节点创建分组数量阈值(节点的数量)
const NEW_NODE_CREATE_GROUP_COUNT = 208

//添加文件时递归尝试次数
const ADD_FILE_TEST_TIME = 1

//oss_split_size默认值
const DEFAULT_OSS_SPLIT_SIZE int64 = 5 * 1024 * 1024

//节点能承受的最大下载并发数con默认值
const DEFAULT_CON int64 = 20

//addp2pfile概率默认值
const DEFAULT_ADD_P2P_FILE int64 = 0

//node直传默认值
const DEFAULT_TRANS_NODE int64 = 1

//genPieceLevel默认值
const DEFAULT_GEN_PIECE_LEVEL int64 = 0

//p2p系统节点上行速度限制
const DEFAULT_P2P_UPSPEED_LIMIT int64 = 1048576

//定义任务卡住的超时用时
const TASK_PROCESS_SLOW_TM = 3600

//危险任务删除标志
const UNSAFE_EXPAND_DEL_DELETE = 1
const UNSAFE_EXPAND_DEL_ALIVE = 0

//扩散完成判断标准
const EXPAND_TASK_FINISH_COUNT_PART = 32

//添加文件节点数量标准
const ADD_FILE_COUNT_PART = 16
