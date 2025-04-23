package redis

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/qtgolang/SunnyNet/src/public"
	"sync"
	"time"
)

type Redis struct {
	Client  *redis.Client
	Mutex   sync.Mutex
	db      int
	Context int
}

func NewRedis() *Redis {
	R := &Redis{}
	return R
}

func (t *Redis) Open(host, pass string, db int, PoolSize_, MinIdleCons_, DialTimeout_, ReadTimeout_, WriteTimeout_, PoolTimeout_, IdleCheckFrequency_, IdleTimeout_ int) error {
	t.db = db
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	PoolSize := PoolSize_ //连接池数量
	if PoolSize < 1 {
		PoolSize = 15
	}
	Min := MinIdleCons_ //最小连接数
	if Min < 1 {
		Min = 10
	}
	DialTimeout := DialTimeout_ //连接超时时间
	if DialTimeout < 1 {
		DialTimeout = 5
	}
	ReadTimeout := ReadTimeout_ //读取超时
	if ReadTimeout < 1 {
		ReadTimeout = 5
	}
	WriteTimeout := WriteTimeout_ //写入超时
	if WriteTimeout < 1 {
		WriteTimeout = 5
	}
	PoolTimeout := PoolTimeout_ //当所有连接都在繁忙状态时,客户端等待可用连接的最大等待时间
	if PoolTimeout < 1 {
		PoolTimeout = 5
	}
	IdleCheckFrequency := IdleCheckFrequency_ //闲置连接检查周期
	if IdleCheckFrequency < 1 {
		IdleCheckFrequency = 60
	}
	IdleTimeout := IdleTimeout_ //闲置超时
	if IdleTimeout < 1 {
		IdleTimeout = 5
	}
	t.Client = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     host, // "127.0.0.1:6379",
		Password: pass, //"",   //密码
		DB:       db,   //Redis数据库

		PoolSize:     PoolSize, //连接池数量
		MinIdleConns: Min,      //好比最小连接数

		DialTimeout:  time.Duration(DialTimeout) * time.Second,  //连接超时时间
		ReadTimeout:  time.Duration(ReadTimeout) * time.Second,  //读取超时
		WriteTimeout: time.Duration(WriteTimeout) * time.Second, //写入超时
		PoolTimeout:  time.Duration(PoolTimeout) * time.Second,  //当所有连接都在繁忙状态时,客户端等待可用连接的最大等待时间

		IdleCheckFrequency: time.Duration(IdleCheckFrequency) * time.Second, //闲置连接检查周期
		IdleTimeout:        time.Duration(IdleTimeout) * time.Second,        //闲置超时
		MaxConnAge:         0 * time.Second,                                 //连接存活时长，从创建开始记时，超过指定时长则关闭

		MaxRetries:      0,                      //命令执行失败时，最多重试多少次，默认为0不重试
		MinRetryBackoff: 8 * time.Microsecond,   //每次计算重试间隔时间的下限,默认8毫秒
		MaxRetryBackoff: 512 * time.Microsecond, //每次计算重试间隔时间的上限,默认512毫秒
	})
	_, err := t.Client.Ping().Result()
	return err
}
func (t *Redis) Set(key string, val interface{}, expr int) bool {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if t.Client == nil {
		return false
	}
	_, e := t.Client.Set(key, val, time.Duration(expr)*time.Second).Result()
	return e == nil
}
func (t *Redis) SetNX(key string, val interface{}, expr int) bool {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if t.Client == nil {
		return false
	}
	ok, e := t.Client.SetNX(key, val, time.Duration(expr)*time.Second).Result()
	return e == nil && ok
}
func (t *Redis) Exists(key string) bool {
	if t.Client == nil {
		return false
	}
	z, _ := t.Client.Do("EXISTS", key).Result()
	return z.(int64) != 0
}
func (t *Redis) GetStr(key string) string {
	if t.Client == nil {
		return ""
	}
	s, _ := t.Client.Get(key).Result()
	return s
}
func (t *Redis) GetBytes(key string) []byte {
	if t.Client == nil {
		return []byte{}
	}
	s, _ := t.Client.Get(key).Bytes()
	s = public.BytesCombine(public.IntToBytes(len(s)), s)
	return s
}
func (t *Redis) GetInt(key string) int64 {
	if t.Client == nil {
		return 0
	}
	x, _ := t.Client.Get(key).Int()
	return int64(x)
}
func (t *Redis) Close() {
	if t.Client == nil {
		return
	}
	_ = t.Client.Close()
}
func (t *Redis) FlushAll() {
	if t.Client == nil {
		return
	}
	t.Client.FlushAll() //用于清空整个 redis 服务器的数据(删除所有数据库的所有 key )。

}
func (t *Redis) FlushDB() {
	if t.Client == nil {
		return
	}
	t.Client.FlushDB() //用于清空当前数据库中的所有 key。
}
func (t *Redis) Delete(key string) bool {
	if t.Client == nil {
		return false
	}
	i, e := t.Client.Del(key).Result()
	if e != nil {
		return false
	}
	return i != 0
}
func (t *Redis) Sub(Msg string, call int, nc bool, callFunc func(str string, call int, nc bool)) {
	if t.Client == nil {
		return
	}
	go func() {
		sub := t.Client.Subscribe(Msg) //"__keyevent@0__:expired"
		for {
			msg := <-sub.Channel()
			if msg == nil {
				return
			}
			b, e := json.Marshal(msg)
			if e == nil {
				t.Mutex.Lock()
				callFunc(string(b), call, nc)
				//fmt.Println("=============================")
				//fmt.Println("过期的键名：", string(b))
				t.Mutex.Unlock()
			}
		}
	}()
}
