package redis

import (
	"errors"
	"time"
)

func (rp *RedisPool) Get(db int, key interface{}) (value interface{}, e error) {
	scon := rp.GetReadConnection(db)
	defer scon.Close()
	return scon.Do("GET", key)
}

func (rp *RedisPool) Set(db int, key interface{}, value interface{}) (e error) {
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	_, e = scon.Do("SET", key, value)
	return
}

func (rp *RedisPool) SetNX(db int, key interface{}, value interface{}) (int, error) {
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	return Int(scon.Do("SETNX", key, value))
}

func (rp *RedisPool) TTL(db int, key interface{}) (int64, error) {
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	return Int64(scon.Do("TTL", key))
}

func (rp *RedisPool) Lock(db int, key_suffix string, expire_sec int64, timeout int64) (getLock bool) {
	now := time.Now().Unix()
	if expire_sec == 0 {
		expire_sec = 10
	}
	expireAt := now + expire_sec
	key := "yh_dlock_" + key_suffix
	timeoutAt := now + timeout
	for {
		i, e := rp.SetNX(db, key, expireAt)
		if e != nil {
			return
		}
		if i == 1 {
			//为lock设置过期时间
			if e = rp.ExpireAt(db, expireAt, key); e != nil {
				return
			}
			getLock = true
			return
		}
		//未能获取lock，验证lock是否成功设置过期时间
		ttl, e := rp.TTL(db, key)
		if e != nil {
			return
		}
		if ttl < 0 { //-2:没有该key，-1:没有过期时间 则将该lock据为己有
			if e = rp.SetEx(db, key, int(expire_sec), expireAt); e != nil {
				return
			}
			getLock = true
			return

		}

		if timeout == 0 || timeoutAt <= time.Now().Unix() {
			break
		}
		time.Sleep(time.Millisecond * 20)

	}
	return
}

func (rp *RedisPool) UnLock(db int, key_suffix string) (e error) {
	key := "yh_dlock_" + key_suffix
	e = rp.Del(db, key)
	return
}

func (rp *RedisPool) SetEx(db int, key interface{}, seconds int, value interface{}) (e error) {
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	_, e = scon.Do("SETEX", key, seconds, value)
	return
}

func (rp *RedisPool) Incr(db int, key interface{}) (uint64, error) {
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	return Uint64(scon.Do("INCR", key))
}

func (rp *RedisPool) IncrBy(db int, key interface{}, value interface{}) (uint64, error) {
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	return Uint64(scon.Do("INCRBY", key, value))
}

// 批量获取
func (rp *RedisPool) MGet(db int, keys ...interface{}) (value interface{}, e error) {
	scon := rp.GetReadConnection(db)
	defer scon.Close()
	return scon.Do("MGET", keys...)
}

/*
批量设置
< key value > 序列
*/
func (rp *RedisPool) MSet(db int, kvs ...interface{}) (value interface{}, e error) {
	if len(kvs)%2 != 0 {
		return nil, errors.New("invalid arguments number")
	}
	scon := rp.GetWriteConnection(db)
	defer scon.Close()
	return scon.Do("MSET", kvs...)
}
