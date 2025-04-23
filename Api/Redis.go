package Api

import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/qtgolang/SunnyNet/src/Call"
	redis "github.com/qtgolang/SunnyNet/src/Redis"
	"github.com/qtgolang/SunnyNet/src/public"
	"strings"
	"sync"
)

var RedisMap = make(map[int]interface{})
var RedisL sync.Mutex

const nbsp = "++&nbsp&++"

func DelRedisContext(Context int) {
	RedisL.Lock()
	delete(RedisMap, Context)
	RedisL.Unlock()
}
func LoadRedisContext(Context int) *redis.Redis {
	RedisL.Lock()
	s := RedisMap[Context]
	RedisL.Unlock()
	if s == nil {
		return nil
	}
	return s.(*redis.Redis)
}

func SubCall(msg string, call int, nc bool) {
	if call > 0 {
		if nc {
			go Call.Call(call, msg)
		} else {
			Call.Call(call, msg)
		}
	}
}

// CreateRedis 创建 Redis 对象
func CreateRedis() int {
	w := redis.NewRedis()
	Context := newMessageId()
	w.Context = Context
	RedisL.Lock()
	RedisMap[Context] = w
	RedisL.Unlock()
	return Context
}

// RemoveRedis 释放 Redis 对象
func RemoveRedis(Context int) {
	k := LoadRedisContext(Context)
	if k != nil {
		k.Close()
	}
	DelRedisContext(Context)
}

// RedisDial Redis 连接
func RedisDial(Context int, host, pass string, db, PoolSize, MinIdleCons, DialTimeout, ReadTimeout, WriteTimeout, PoolTimeout, IdleCheckFrequency, IdleTimeout int, error uintptr) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		public.WriteErr(errors.New("Context 未创建 "), error)
		return false
	}
	ex := w.Open(
		host,
		pass,
		db,
		PoolSize, MinIdleCons, DialTimeout, ReadTimeout, WriteTimeout, PoolTimeout, IdleCheckFrequency, IdleTimeout)
	if ex != nil {
		public.WriteErr(ex, error)
	}
	return ex == nil
}

// RedisSet Redis 设置值
func RedisSet(Context int, key, val string, expr int) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		return false
	}
	return w.Set(key, val, expr)
}

// RedisSetBytes Redis 设置Bytes值
func RedisSetBytes(Context int, key string, val []byte, expr int) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		return false
	}
	return w.Set(key, val, expr)
}

// RedisSetNx Redis 设置NX 【如果键名存在返回假】
func RedisSetNx(Context int, key, val string, expr int) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		return false
	}
	return w.SetNX(key, val, expr)
}

// RedisExists Redis 检查指定 key 是否存在
func RedisExists(Context int, key string) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		return false
	}
	return w.Exists(key)
}

// RedisGetStr Redis 取文本值
func RedisGetStr(Context int, key string) string {
	w := LoadRedisContext(Context)
	if w == nil {
		return ""
	}
	s := w.GetStr(key)
	return s
}

// RedisGetBytes Redis 取文本值
func RedisGetBytes(Context int, key string) []byte {
	w := LoadRedisContext(Context)
	if w == nil {
		return nil
	}
	s := w.GetBytes(key)
	if len(s) < 1 {
		return nil
	}
	return s
}

// RedisDo Redis 自定义 执行和查询命令 返回操作结果可能是值 也可能是JSON文本
func RedisDo(Context int, args string) ([]byte, error) {

	w := LoadRedisContext(Context)
	if w == nil {
		return nil, errors.New("Redis no create 0x002 ")
	}
	arr := strings.Split(strings.ReplaceAll(args, "\\ ", nbsp), " ")
	var InterFaceArr = make([]interface{}, 0)
	for _, v := range arr {
		if len(v) > 0 {
			InterFaceArr = append(InterFaceArr, strings.ReplaceAll(v, nbsp, " "))
		}
	}
	if len(InterFaceArr) < 1 {
		return nil, errors.New("Parameter error ")
	}
	Val, er := w.Client.Do(InterFaceArr...).Result()
	if er != nil {
		return nil, er
	}
	b, er := json.Marshal(Val)
	if er != nil {
		return nil, er
	}
	if len(b) < 1 {
		return nil, errors.New("The execution succeeds but no data is returned ")
	}
	return b, nil
}

// RedisGetKeys Redis 取指定条件键名
func RedisGetKeys(Context int, key string) []byte {
	w := LoadRedisContext(Context)
	if w == nil {
		return nil
	}
	var b bytes.Buffer
	keys, _ := w.Client.Keys(key).Result()
	for _, v := range keys {
		b.WriteString(v)
		b.WriteByte(0)
	}
	return b.Bytes()
}

// RedisGetInt Redis 取整数值
func RedisGetInt(Context int, key string) int64 {
	w := LoadRedisContext(Context)
	if w == nil {
		return 0
	}
	return w.GetInt(key)
}

// RedisClose Redis 关闭
func RedisClose(Context int) {
	w := LoadRedisContext(Context)
	if w == nil {
		return
	}
	w.Close()
}

// RedisFlushAll Redis 清空redis服务器
func RedisFlushAll(Context int) {
	//用于清空整个 redis 服务器的数据(删除所有数据库的所有 key )。
	w := LoadRedisContext(Context)
	if w == nil {
		return
	}
	w.FlushAll()
}

// RedisFlushDB Redis 清空当前数据库
func RedisFlushDB(Context int) {
	//用于清空当前数据库中的所有 key。
	w := LoadRedisContext(Context)
	if w == nil {
		return
	}
	w.FlushDB()
}

// RedisDelete Redis 删除
func RedisDelete(Context int, key string) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		return false
	}
	return w.Delete(key)
}

// RedisSubscribe Redis 订阅消息
func RedisSubscribe(Context int, scribe string, call int, nc bool) bool {
	w := LoadRedisContext(Context)
	if w == nil {
		return false
	}
	w.Sub(scribe, call, nc, SubCall)
	return true
}

// RedisSubscribeGo Redis 订阅消息
func RedisSubscribeGo(Context int, scribe string, call func(msg string)) {
	w := LoadRedisContext(Context)
	if w == nil {
		return
	}
	w.Sub(scribe, 0, false, func(str string, _ int, _ bool) {
		call(str)
	})
}
