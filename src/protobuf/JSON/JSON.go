package JSON

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

type SyJson struct {
	_Data interface{}
}

func NewSyJson() *SyJson {
	s := new(SyJson)
	_ = json.Unmarshal([]byte("{}"), &s._Data)
	return s
}

func (s *SyJson) Parse(xstr string) bool {
	_Data := make(map[string]any)
	if xstr == "" {
		_Data["/"] = ""
		s._Data = _Data
		return false
	}
	return json.Unmarshal([]byte("{\"/\":"+xstr+"}"), &s._Data) == nil
}
func (s *SyJson) GetMap() map[string]any {
	if s._Data == nil {
		return make(map[string]any)
	}
	b := s._Data.(map[string]any)
	if b == nil {
		return make(map[string]any)
	}
	c := b["/"]
	if c == nil {
		return make(map[string]any)
	}
	return c.(map[string]any)
}
func (s *SyJson) ToString() string {
	return s.GetData("")
}
func parsingPath(b string) []string {
	if b == "" {
		return make([]string, 0)
	}
	s := b
	if s[0:1] != "[" {
		s = strings.ReplaceAll(s, "[", ".[")
	}
	return strings.Split(s, ".")
}
func parsePathArray(b string, c []string) ([]string, []string) {
	p := strings.ReplaceAll(b, "\\.", "\\。")
	p = strings.ReplaceAll(p, "\\[", "\\【")
	p = strings.ReplaceAll(p, "\\]", "\\】")
	arr := strings.Split(p, ".")
	var patharr []string
	for _, v := range arr {
		for _, vv := range parsingPath(v) {
			s := strings.ReplaceAll(vv, "\\。", "\\.")
			s = strings.ReplaceAll(s, "\\【", "[")
			s = strings.ReplaceAll(s, "\\】", "]")
			patharr = append(patharr, s)
			if c != nil {
				c = append(c, s)
			}
		}
	}
	return patharr, c
}
func (s *SyJson) SetData(path string, value any) bool {
	//解析传入的路径
	var _path []string
	_path = append(_path, "/")
	_, _path = parsePathArray(path, _path)
	var val any
	if _v, ok := value.(*SyJson); ok {
		val = _v.GetMap()
	} else {
		val = value
	}
	//如果传入的是空，则重新解析
	if len(_path) < 2 {
		v, ok := val.(string)
		if ok {
			var _json interface{}
			IsOK := json.Unmarshal([]byte("{\"/\":"+v+"}"), &_json) == nil
			if IsOK {
				s._Data = _json
			}
			return IsOK
		}
		return false
	}
	type _type struct {
		object interface{}
		Name   string
	}
	data := s._Data.(map[string]interface{})["/"]
	v := make([]*_type, len(_path))
	max := len(v)
	var SetFunc = func(object, upper *_type, directory, lowerdirectory string) {
		if upper == nil {
			object.object = data
			object.Name = "/"
			return
		}
		object.Name = directory
		//当前指定目录 是否是数组格式
		index := _IsPathArray(directory)
		if index > -1 {
			//是数组格式
			//上级目录对象是否是数组
			if _IsArray(upper.object) {
				var _object []interface{}
				switch obj := upper.object.(type) {
				case []int:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []int64:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []int32:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []byte:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []float64:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []bool:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []string:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				case []interface{}:
					for i := 0; i < len(obj); i++ {
						_object = append(_object, obj[i])
					}
				default:
					panic(_object)
				}
				for i := len(_object); i < index+1; i++ {
					var ts interface{}
					_object = append(_object, ts)
				}
				if lowerdirectory == "" {
					_object[index] = val
					upper.object = _object
					return
				}
				object.object = _object[index]
				//是数组格式
			} else {
				//不是是数组
				_object := make([]interface{}, index+1)
				if lowerdirectory == "" {
					_object[index] = val
					upper.object = _object
					return
				}
				upper.object = _object
				object.object = _object[index]
			}
		} else {
			//不是是数组
			//上级目录对象是否是数组
			if _IsArray(upper) {
				//是数组格式
				_object := make(map[string]interface{})
				upper.object = _object
				object.object = _object[directory]
			} else {
				//不是是数组
				if !_IsMap(upper.object) {
					upper.object = make(map[string]interface{})
				}
				_object, ok := upper.object.(map[string]interface{})
				if ok {
					if lowerdirectory == "" {
						_object[directory] = val
						upper.object = _object
						return
					}
					object.object = _object[directory]
				} else {
					upper.object = make(map[string]interface{})
				}
			}
		}
	}
	for i := 0; i < len(v); i++ {
		v[i] = new(_type)
		if i == max-1 {
			SetFunc(v[i], v[i-1], _path[i], "")
			continue
		}
		if i == 0 {
			SetFunc(v[i], nil, _path[i], _path[i+1])
		} else {
			SetFunc(v[i], v[i-1], _path[i], _path[i+1])
		}
	}
	v = append(v[0:len(v)-1], v[len(v):]...)
	for i := len(v) - 1; i > 0; i-- {
		index := _IsPathArray(v[i].Name)
		if index > -1 {
			var _object []interface{}
			switch obj := v[i-1].object.(type) {
			case []int:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []int64:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []int32:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []byte:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []float64:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []bool:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []string:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			case []interface{}:
				for i := 0; i < len(obj); i++ {
					_object = append(_object, obj[i])
				}
			default:
				panic(_object)
			}
			for i1 := len(_object); i1 < index+1; i1++ {
				var n interface{}
				_object = append(_object, n)
			}
			_object[index] = v[i].object
			v[i-1].object = _object
		} else {
			_object := v[i-1].object.(map[string]interface{})
			_object[v[i].Name] = v[i].object
		}
	}
	if ss, ok := s._Data.(map[string]interface{}); ok {
		ss["/"] = v[0].object
	}
	return true
}
func (s *SyJson) GetData(path string) string {
	//解析传入的路径

	PathToTheArray, _ := parsePathArray(path, nil)
	//先获取顶级
	ss, ok := s._Data.(map[string]interface{})
	if !ok {
		return ""
	}
	data := ss["/"]

	//如果传入的是空，返回全部
	if len(PathToTheArray) == 0 {
		return tostring(data)
	}
	//从传入路径中依次解析
	for i := 0; i < len(PathToTheArray)-1; i++ {
		if _DataIsArray(data) {
			ind := _IsPathArray(PathToTheArray[i])
			if ind >= 0 {
				if !_IsArray(data) {
					return ""
				}
				switch obj := data.(type) {
				case []int:
					if i < len(PathToTheArray)-1 {
						if ind < 0 || ind >= len(obj) {
							return ""
						}
						data = obj[ind]
						continue
					}
					return tostring(obj[ind])
				case []string:
					if i < len(PathToTheArray)-1 {
						if ind < 0 || ind >= len(obj) {
							return ""
						}
						data = obj[ind]
						continue
					}
					return tostring(obj[ind])
				case []float64:
					if i < len(PathToTheArray)-1 {
						if ind < 0 || ind >= len(obj) {
							return ""
						}
						data = obj[ind]
						continue
					}
					return tostring(obj[ind])
				default:
					break
				}
				obj, o := data.([]interface{})
				if obj == nil || !o {
					return ""
				}
				if i < len(PathToTheArray)-1 {
					if ind < 0 || ind >= len(obj) {
						return ""
					}
					data = obj[ind]
					continue
				}
				if ind < 0 || ind >= len(obj) {
					return ""
				}
				return tostring(obj[ind])
			} else {
				return ""
			}
		} else {
			ind := _IsPathArray(PathToTheArray[i])
			if ind >= 0 {
				return ""
			}
			if !_IsMap(data) {
				return ""
			}
			mm, o := data.(map[string]interface{})
			if o {
				data = mm[PathToTheArray[i]]
			}

		}
	}
	if _DataIsArray(data) {
		ind := _IsPathArray(PathToTheArray[len(PathToTheArray)-1])
		if ind >= 0 {
			if !_IsArray(data) {
				return ""
			}
			switch obj := data.(type) {
			case []int:
				if ind < 0 || ind >= len(obj) {
					return ""
				}
				return tostring(obj[ind])
			case []string:
				if ind < 0 || ind >= len(obj) {
					return ""
				}
				return tostring(obj[ind])
			case []float64:
				if ind < 0 || ind >= len(obj) {
					return ""
				}
				return tostring(obj[ind])
			default:
				break
			}
			obj, o := data.([]interface{})
			if obj == nil || !o {
				return ""
			}
			if ind < 0 || ind >= len(obj) {
				return ""
			}
			return tostring(obj[ind])
		}
		return ""
	}
	ind := _IsPathArray(PathToTheArray[len(PathToTheArray)-1])
	if ind >= 0 {
		return ""
	}
	if !_IsMap(data) {
		return ""
	}
	obj, o := data.(map[string]interface{})
	if obj == nil || !o {
		return ""
	}
	ind = len(PathToTheArray) - 1
	if ind < 0 || ind >= len(PathToTheArray) {
		return ""
	}
	ssx := PathToTheArray[ind]
	return tostring(obj[ssx])
}
func (s *SyJson) GetCount(path string) int {
	return s.GetNum(path)
}
func (s *SyJson) GetNum(path string) int {
	//解析传入的路径

	PathToTheArray, _ := parsePathArray(path, nil)
	//先获取顶级
	ss, ok := s._Data.(map[string]interface{})
	if !ok {
		return 0
	}
	data := ss["/"]
	GetDtaNum := func(d interface{}) int {
		switch obj := d.(type) {
		case []int:
			return len(obj)
		case []int32:
			return len(obj)
		case []int64:
			return len(obj)
		case []byte:
			return len(obj)
		case []string:
			return len(obj)
		case []interface{}:
			return len(obj)
		case []bool:
			return len(obj)
		case []float64:
			return len(obj)
		case map[string]interface{}:
			return len(obj)
		case map[int]interface{}:
			return len(obj)
		}
		return 0
	}
	//如果传入的是空，返回全部
	if len(PathToTheArray) == 0 {
		return GetDtaNum(data)
	}
	//从传入路径中依次解析
	for i := 0; i < len(PathToTheArray)-1; i++ {
		if _DataIsArray(data) {
			ind := _IsPathArray(PathToTheArray[i])
			if ind >= 0 {
				if !_IsArray(data) {
					return 0
				}
				switch obj := data.(type) {
				case []int:
					return GetDtaNum(obj[ind])
				case []string:
					if i < len(PathToTheArray)-1 {
						data = obj[ind]
						continue
					}
					return GetDtaNum(obj[ind])
				case []float64:
					if i < len(PathToTheArray)-1 {
						data = obj[ind]
						continue
					}
					return GetDtaNum(obj[ind])
				default:
					break
				}
				obj, o := data.([]interface{})
				if obj == nil || !o {
					return GetDtaNum(obj)
				}
				if i < len(PathToTheArray)-1 {
					data = obj[ind]
					continue
				}
				return GetDtaNum(obj[ind])
			} else {
				return 0
			}
		} else {
			ind := _IsPathArray(PathToTheArray[i])
			if ind >= 0 {
				return 0
			}
			if !_IsMap(data) {
				return 0
			}
			mm, o := data.(map[string]interface{})
			if o {
				data = mm[PathToTheArray[i]]
			}
		}
	}
	if _DataIsArray(data) {
		ind := _IsPathArray(PathToTheArray[len(PathToTheArray)-1])
		if ind >= 0 {
			if !_IsArray(data) {
				return 0
			}
			switch obj := data.(type) {
			case []int:
				return GetDtaNum(obj[ind])
			case []string:
				return GetDtaNum(obj[ind])
			case []float64:
				return GetDtaNum(obj[ind])
			default:
				break
			}
			obj := data.([]interface{})
			if obj == nil {
				return 0
			}
			return GetDtaNum(obj[ind])
		}
		return 0
	}
	ind := _IsPathArray(PathToTheArray[len(PathToTheArray)-1])
	if ind >= 0 {
		return 0
	}
	if !_IsMap(data) {
		return 0
	}
	//return GetDtaNum(data.(map[string]interface{})[PathToTheArray[len(PathToTheArray)-1]])
	obj, o := data.(map[string]interface{})
	if obj == nil || !o {
		return 0
	}
	ind = len(PathToTheArray) - 1
	if ind < 0 || ind >= len(PathToTheArray) {
		return 0
	}
	sxs := PathToTheArray[ind]
	return GetDtaNum(obj[sxs])
}
func _IsArray(interf interface{}) bool {
	if interf == nil {
		return false
	}
	switch interf.(type) {
	case []interface{}:
		return true
	default:
		s := reflect.TypeOf(interf).String()
		if len(s) > 2 {
			if s[0:1] == "[" && s[1:2] == "]" {
				return true
			}
		}
		return false
	}
	return false
}
func _IsMap(interf interface{}) bool {
	if interf == nil {
		return false
	}
	switch interf.(type) {
	case map[string]interface{}:
		return true
	default:
		return false
	}
	return false
}
func _IsPathArray(s string) int {
	if len(s) < 3 {
		return -1
	}
	if s[0:1] == "[" && s[len(s)-1:] == "]" {
		i, e := strconv.Atoi(s[1 : len(s)-1])
		if e != nil {
			return -1
		}
		return i
	}
	return -1
}
func _DataIsArray(interf interface{}) bool {
	if interf == nil {
		return false
	}
	s := reflect.TypeOf(interf).String()
	if len(s) < 2 {
		return false
	}
	return s[0:2] == "[]"
}
func tostring(interf interface{}) string {
	data, err := json.Marshal(interf)
	if err != nil {
		return ""
	}
	content := string(data)
	content = strings.Replace(content, "\\u003c", "<", -1)
	content = strings.Replace(content, "\\u003e", ">", -1)
	content = strings.Replace(content, "\\u0026", "&", -1)
	if content[0:1] == "\"" && content[len(content)-1:] == "\"" {
		content = content[1 : len(content)-1]
	}
	return content
}
