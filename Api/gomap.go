package Api

import (
	"github.com/qtgolang/SunnyNet/src/public"
	"strconv"
	"sync"
	"time"
)

type KeysType struct {
	Handle map[string]interface{}
}

var Keys = make(map[int]*KeysType)
var KeysW sync.Mutex

func CreateKeys() int {
	_Keys := &KeysType{Handle: make(map[string]interface{})}
	KeysContext := newMessageId()
	Keys[KeysContext] = _Keys
	return KeysContext
}

func RemoveKeys(KeysHandle int) {
	KeysW.Lock()
	defer KeysW.Unlock()
	v := Keys[KeysHandle]
	if v != nil {
		for one := range v.Handle {
			delete(v.Handle, one)
		}
	}
	delete(Keys, KeysHandle)
}

func KeysDelete(KeysHandle int, name string) {
	KeysW.Lock()
	defer KeysW.Unlock()
	v := Keys[KeysHandle]
	if v != nil {
		if v.Handle != nil {
			delete(v.Handle, name)
		}
	}
}

func KeysRead(KeysHandle int, name string) uintptr {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return 0
	}
	s := k.Handle[name]
	if s == nil {
		return 0
	}
	switch v := s.(type) {
	case []byte:
		if len(v) < 1 {
			return 0
		}
		return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(v)), v))
	case string:
		if len(v) < 1 {
			return 0
		}
		return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(v)), []byte(v)))
	case int64:
		sb := public.Int64ToBytes(v)
		//k.Handle[(name)] = sb
		return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(sb)), sb))
	case float64:
		sb := public.Float64ToBytes(v)
		//k.Handle[(name)] = sb
		return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(sb)), sb))
	case float32:
		sb := public.Float64ToBytes(float64(v))
		//k.Handle[(name)] = sb
		return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(sb)), sb))
	case int:
		sb := public.IntToBytes(v)
		//k.Handle[(name)] = sb
		return public.PointerPtr(public.BytesCombine(public.IntToBytes(len(sb)), sb))
	}
	return 0
}

func KeysWrite(KeysHandle int, name string, val uintptr, length int) {
	data := public.CStringToBytes(val, length)
	KeysW.Lock()
	defer KeysW.Unlock()
	v := Keys[KeysHandle]
	if v == nil {
		return
	}
	v.Handle[(name)] = data
}

func KeysWriteFloat(KeysHandle int, name string, val float64) {
	KeysW.Lock()
	defer KeysW.Unlock()
	v := Keys[KeysHandle]
	if v == nil {
		return
	}
	v.Handle[(name)] = val
}

func KeysReadFloat(KeysHandle int, name string) float64 {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return 0
	}
	s := k.Handle[(name)]
	if s == nil {
		return 0
	}
	switch r := s.(type) {
	case float64:
		return r
	case int:
		return float64(r)
	case int64:
		return float64(r)
	default:
		return 0
	}
}

func KeysWriteLong(KeysHandle int, name string, val int64) {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return
	}
	k.Handle[name] = val
}

func KeysReadLong(KeysHandle int, name string) int64 {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return 0
	}
	s := k.Handle[name]
	if s == nil {
		return 0
	}
	switch r := s.(type) {
	case int64:
		return r
	case float64:
		return int64(r)
	case int:
		return int64(r)
	default:
		return 0
	}
}

func KeysWriteInt(KeysHandle int, name string, val int) {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return
	}
	k.Handle[name] = val
}

func KeysReadInt(KeysHandle int, name string) int {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return 0
	}
	s := k.Handle[name]
	if s == nil {
		return 0
	}
	switch r := s.(type) {
	case int64:
		return int(r)
	case float64:
		return int(r)
	case int:
		return r
	default:
		return 0
	}
}

func KeysEmpty(KeysHandle int) {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return
	}
	if k.Handle != nil {
		for s := range k.Handle {
			delete(k.Handle, s)
		}
	}
}

func KeysGetCount(KeysHandle int) int {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return 0
	}
	return len(k.Handle)
}

func KeysGetJson(KeysHandle int) uintptr {
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return 0
	}
	var get = func(cc interface{}) string {
		switch cv := cc.(type) {
		case string:
			return "\"" + cv + "\""
		case time.Time:
			return "\"" + cv.Format("2006-01-02 15:04:05") + "\""
		case bool:
			if cv {
				return "true"
			}
			return "false"
		case []byte:
			r := "["
			for _, v := range cv {
				r += strconv.Itoa(int(v)) + ","
			}
			r = r[0:len(r)-1] + "]"
			return r
		case int:
			return strconv.Itoa(cv)
		case int8:
			return strconv.Itoa(int(cv))
		case int16:
			return strconv.Itoa(int(cv))
		case int32:
			return strconv.Itoa(int(cv))
		case int64:
			return strconv.FormatInt(cv, 10)
		case byte:
			return strconv.Itoa(int(cv))
		case uintptr:
			return strconv.Itoa(int(cv))
		case uint:
			return strconv.Itoa(int(cv))
		case uint16:
			return strconv.Itoa(int(cv))
		case uint32:
			return strconv.Itoa(int(cv))
		case uint64:
			return strconv.Itoa(int(cv))
		case float32:
			return strconv.FormatFloat(float64(cv), 'f', 6, 64)
		case float64:
			return strconv.FormatFloat(cv, 'f', 6, 64)
		default:
			return "\"类似不支持\""
		}
	}
	if k.Handle != nil {
		record := ""
		for kc, v := range k.Handle {
			record += "\"" + kc + "\":" + get(v) + ","
		}
		if len(record) > 0 {
			record = "{" + record[0:len(record)-1] + "}"
		} else {
			record = "{}"
		}
		return public.PointerPtr(record)
	}
	return 0
}

func KeysWriteStr(KeysHandle int, name string, val uintptr, len int) {
	data := public.CStringToBytes(val, len)
	KeysW.Lock()
	defer KeysW.Unlock()
	k := Keys[KeysHandle]
	if k == nil {
		return
	}
	k.Handle[(name)] = string(data)
}
