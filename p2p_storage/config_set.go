package p2p_storage

import (
	"errors"
	"yh_pkg/thread_safe/safe_map"
	"yh_pkg/utils"
)

//ConfigSet set
type ConfigSet struct {
	ConfigValue *safe_map.SafeMap
}

//NewConfigSet 新建一个ConfigSet
func NewConfigSet() *ConfigSet {
	var cs ConfigSet
	cs.ConfigValue = safe_map.New()
	cs.InitConfigValue()
	cs.FlushConfigValue()
	return &cs
}

func (cs *ConfigSet) InitConfigValue() {
	cs.ConfigValue.Set(DELEGATES_MIN_SPEED_CONFIG_KEY, DEFAULT_DELEGATES_NODE_SPEED)
	cs.ConfigValue.Set(SPREAD_MIN_SPEED_CONFIG_KEY, DEFAULT_SECOND_EXPAND_SPEED)
	cs.ConfigValue.Set(MAX_HOUR_CONFIG_KEY, DEFAULT_MAX_HOUR)
	cs.ConfigValue.Set(OSS_SPLIT_SIZE_CONFIG_KEY, DEFAULT_OSS_SPLIT_SIZE)
	cs.ConfigValue.Set(CON_CONFIG_KEY, DEFAULT_CON)
	cs.ConfigValue.Set(TRANS_NODE_CONFIG_KEY, DEFAULT_TRANS_NODE)
	cs.ConfigValue.Set(ADD_P2P_FILE_CONFIG_KEY, DEFAULT_ADD_P2P_FILE)
	cs.ConfigValue.Set(GEN_PIECE_LEVEL_CONFIG_KEY, DEFAULT_GEN_PIECE_LEVEL)
	cs.ConfigValue.Set(P2P_UPSPEED_LIMIT_KEY, DEFAULT_P2P_UPSPEED_LIMIT)
	cs.ConfigValue.Set(P2P_MERGE_PIECE, 0)
	cs.ConfigValue.Set(P2P_DOWNLOAD_CACHE, 0)
}

func (cs *ConfigSet) FlushConfigValue() (err error) {
	tempConfigMap := make(map[interface{}]interface{})
	if err = dataSource.Raw.GetMapFromConfig(tempConfigMap); err != nil {
		return err
	}
	for k, v := range tempConfigMap {
		cs.ConfigValue.Set(k, v)
	}
	return nil
}

func (cs *ConfigSet) GetInt64Value(key string) (value int64, err error) {
	if v, exist := cs.ConfigValue.Get(key); exist {
		return utils.ToInt64(v)
	} else {
		return -1, errors.New("GetValue value not exists")
	}
}

func (cs *ConfigSet) GetFloat64Value(key string) (value float64, err error) {
	if v, exist := cs.ConfigValue.Get(key); exist {
		return utils.ToFloat64(v)
	} else {
		return .0, errors.New("GetValue value not exists")
	}
}

func (cs *ConfigSet) GetStringValue(key string) (value string, err error) {
	if v, exist := cs.ConfigValue.Get(key); exist {
		return utils.ToString(v), nil
	} else {
		return "", errors.New("GetValue value not exists")
	}
}
