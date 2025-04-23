package JavaJni

import (
	"fmt"
	"github.com/qtgolang/SunnyNet/JavaApi/sig"
	"strconv"
)

func (env Env) Boolean(obj Jobject) bool {
	cls := env.FindClass("java/lang/Boolean")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "booleanValue", fmt.Sprintf("()%s", sig.Boolean))
	i := env.CallBooleanMethodA(obj, Method)
	return i
}
func (env Env) Int(obj Jobject) int {
	cls := env.FindClass("java/lang/Integer")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "intValue", fmt.Sprintf("()%s", sig.Int))
	i := env.CallIntMethodA(obj, Method)
	return i
}
func (env Env) Byte(obj Jobject) byte {
	cls := env.FindClass("java/lang/Byte")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "byteValue", fmt.Sprintf("()%s", sig.Byte))
	i := env.CallByteMethodA(obj, Method)
	return i
}
func (env Env) Char(obj Jobject) uint16 {
	cls := env.FindClass("java/lang/Character")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "charValue", fmt.Sprintf("()%s", sig.Char))
	i := env.CallCharMethodA(obj, Method)
	return i
}
func (env Env) Short(obj Jobject) int16 {
	cls := env.FindClass("java/lang/Short")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "shortValue", fmt.Sprintf("()%s", sig.Short))
	i := env.CallShortMethodA(obj, Method)
	return i
}
func (env Env) Float(obj Jobject) float32 {
	cls := env.FindClass("java/lang/Float")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "floatValue", fmt.Sprintf("()%s", sig.Float))
	i := env.CallFloatMethodA(obj, Method)
	return i
}
func (env Env) Double(obj Jobject) float64 {
	cls := env.FindClass("java/lang/Double")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "doubleValue", fmt.Sprintf("()%s", sig.Double))
	i := env.CallDoubleMethodA(obj, Method)
	return i
}
func (env Env) Long(obj Jobject) int64 {
	cls := env.FindClass("java/lang/Long")
	defer env.DeleteLocalRef(cls)
	Method := env.GetMethodID(cls, "toString", fmt.Sprintf("()%s", sig.String))
	i := env.CallObjectMethodA(obj, Method)
	if i == 0 {
		return 0
	}
	defer env.DeleteLocalRef(i)
	S := string(env.GetStringUTF(i))
	i64, _ := strconv.ParseInt(S, 10, 64)
	return i64
}

func (env Env) NewBoolean(obj bool) Jobject {
	_Class := env.FindClass("java/lang/Boolean")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Boolean, sig.Void))
	var o Jobject
	if obj == true {
		o = env.NewObjectA(_Class, Method, JNI_TRUE)
	} else {
		o = env.NewObjectA(_Class, Method, JNI_FALSE)
	}
	return o
}
func (env Env) NewByte(obj byte) Jobject {
	_Class := env.FindClass("java/lang/Byte")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Byte, sig.Void))
	return env.NewObjectA(_Class, Method, Jvalue(obj))
}
func (env Env) NewChar(obj rune) Jobject {
	_Class := env.FindClass("java/lang/Character")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Char, sig.Void))
	return env.NewObjectA(_Class, Method, Jvalue(obj))
}
func (env Env) NewShort(obj int16) Jobject {
	_Class := env.FindClass("java/lang/Short")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Short, sig.Void))
	return env.NewObjectA(_Class, Method, Jvalue(obj))
}
func (env Env) NewInt(obj int32) Jobject {
	_Class := env.FindClass("java/lang/Integer")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Int, sig.Void))
	return env.NewObjectA(_Class, Method, Jvalue(obj))
}
func (env Env) NewLong(obj int64) Jobject {
	_Class := env.FindClass("java/lang/Long")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Long, sig.Void))
	return env.NewObjectA(_Class, Method, Jvalue(obj))
}
func (env Env) NewDouble(obj int64) Jobject {
	_Class := env.FindClass("java/lang/Double")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetMethodID(_Class, "<init>", fmt.Sprintf("(%s)%s", sig.Double, sig.Void))
	return env.NewObjectA(_Class, Method, Jvalue(obj))
}
func (env Env) NewFloat(obj float32) Jobject {
	js := strconv.FormatFloat(float64(obj), 'f', -1, 32)
	msj := env.NewString(js)
	_Class := env.FindClass("java/lang/Float")
	defer env.DeleteLocalRef(_Class)
	Method := env.GetStaticMethodID(_Class, "valueOf", fmt.Sprintf("(%s)%s", sig.String, sig.FloatClass))
	return env.CallStaticObjectMethodA(_Class, Method, Jvalue(msj))
}

func (env Env) NewObject(obj ...Jobject) Jobject {
	objectClass := env.FindClass("java/lang/Object")
	defer env.DeleteLocalRef(objectClass)
	objectArray := env.NewObjectArray(len(obj), objectClass, Jobject(NULL))
	for i := 0; i < len(obj); i++ {
		env.SetObjectArrayElement(objectArray, i, obj[i])
	}
	return objectArray
}
