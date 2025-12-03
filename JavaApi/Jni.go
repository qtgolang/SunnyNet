package JavaJni

import "C"

// #cgo CFLAGS: -Wno-incompatible-pointer-types
// #include "jni.h"
// #include <stdlib.h>
// #include <unistd.h>
// #define LOG_ERROR( tag, ... ) ( (void) __android_log_print( ANDROID_LOG_ERROR, tag, __VA_ARGS__ ) )
//
// static inline jobject GetModule(JNIEnv * env, jclass clazz) {
//     return (*env)->GetModule(env, clazz);
// }
//
// static inline jint AttachCurrentThread(JavaVM *vm, JNIEnv **p_env) {
//     return (*vm)->AttachCurrentThread(vm, (void **) p_env, NULL);
// }
//
// static inline jint AttachCurrentThreadAsDaemon(JavaVM *vm, JNIEnv **p_env) {
//     return (*vm)->AttachCurrentThreadAsDaemon(vm, (void **) p_env, NULL);
// }
//
// static inline jint GetEnv(JavaVM *vm, JNIEnv **penv, jint version) {
//     return (*vm)->GetEnv(vm, (void **) penv, version);
// }
//
// static inline jint GetJavaVM(JNIEnv * env, JavaVM **vm) {
//     return (*env)->GetJavaVM(env, vm);
// }
//
// static inline int GetObjectRefType(JNIEnv * env, jobject obj) {
//     return (int) (*env)->GetObjectRefType(env, obj);
// }
//
// static inline jint DestroyJavaVM(JavaVM * vm) {
//     return (*vm)->DestroyJavaVM(vm);
// }
//
// static inline jint DetachCurrentThread(JavaVM * vm) {
//     return (*vm)->DetachCurrentThread(vm);
// }
//
// static inline jclass FindClass(JNIEnv * env, char * name) {
////		while(1){
////			(*env)->ExceptionClear(env);
////			jclass res = (*env)->FindClass(env, name);
////      	if (res != NULL) {
////				return res;
////     		}
////			LOG_ERROR("SunnyFindClass","FindClass[%s]",name);
////			sleep(1);
////	   }
////      return NULL;
//	(*env)->ExceptionClear(env);
//  return (*env)->FindClass(env, name);
// }
//
// static inline jint GetVersion(JNIEnv * env) {
//     return (*env)->GetVersion(env);
// }
//
// static inline jmethodID FromReflectedMethod(JNIEnv * env, jobject method) {
//     return (*env)->FromReflectedMethod(env, method);
// }
//
// static inline jfieldID FromReflectedField(JNIEnv * env, jobject field) {
//     return (*env)->FromReflectedField(env, field);
// }
//
// static inline jobject ToReflectedMethod(JNIEnv * env, jclass cls, jmethodID methodID, jboolean isStatic) {
//     return (*env)->ToReflectedMethod(env, cls, methodID, isStatic);
// }
//
// static inline jclass GetSuperclass(JNIEnv * env, jclass sub) {
//     return (*env)->GetSuperclass(env, sub);
// }
//
// static inline jboolean IsAssignableFrom(JNIEnv * env, jclass sub, jclass sup) {
//     return (*env)->IsAssignableFrom(env, sub, sup);
// }
//
// static inline jobject ToReflectedField(JNIEnv * env, jclass cls, jfieldID fieldID, jboolean isStatic) {
//     return (*env)->ToReflectedField(env, cls, fieldID, isStatic);
// }
//
// static inline jint Throw(JNIEnv * env, jthrowable obj) {
//     return (*env)->Throw(env, obj);
// }
//
// static inline jint ThrowNew(JNIEnv * env, jclass clazz, char * msg) {
//     return (*env)->ThrowNew(env, clazz, msg);
// }
//
// static inline jthrowable ExceptionOccurred(JNIEnv * env) {
//     return (*env)->ExceptionOccurred(env);
// }
//
// static inline void ExceptionDescribe(JNIEnv * env) {
//     (*env)->ExceptionDescribe(env);
// }
//
// static inline void ExceptionClear(JNIEnv * env) {
//     (*env)->ExceptionClear(env);
// }
//
// static inline void FatalError(JNIEnv * env, char * msg) {
//     (*env)->FatalError(env, msg);
// }
//
// static inline jint PushLocalFrame(JNIEnv * env, jint capacity) {
//     return (*env)->PushLocalFrame(env, capacity);
// }
//
// static inline jobject PopLocalFrame(JNIEnv * env, jobject result) {
//     return (*env)->PopLocalFrame(env, result);
// }
//
// static inline jobject NewGlobalRef(JNIEnv * env, jobject lobj) {
//     return (*env)->NewGlobalRef(env, lobj);
// }
//
// static inline void DeleteGlobalRef(JNIEnv * env, jobject gref) {
//     (*env)->DeleteGlobalRef(env, gref);
// }
//
// static inline void DeleteLocalRef(JNIEnv * env, jobject obj) {
//     (*env)->DeleteLocalRef(env, obj);
// }
//
// static inline jboolean IsSameObject(JNIEnv * env, jobject obj1, jobject obj2) {
//     return (*env)->IsSameObject(env, obj1, obj2);
// }
//
// static inline jobject NewLocalRef(JNIEnv * env, jobject ref) {
//     return (*env)->NewLocalRef(env, ref);
// }
//
// static inline jint EnsureLocalCapacity(JNIEnv * env, jint capacity) {
//     return (*env)->EnsureLocalCapacity(env, capacity);
// }
//
// static inline jobject AllocObject(JNIEnv * env, jclass clazz) {
//     return (*env)->AllocObject(env, clazz);
// }
//
// static inline jobject NewObjectA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->NewObjectA(env, clazz, methodID, args);
// }
//
// static inline jclass GetObjectClass(JNIEnv * env, jobject obj) {
//     return (*env)->GetObjectClass(env, obj);
// }
//
// static inline jboolean IsInstanceOf(JNIEnv * env, jobject obj, jclass clazz) {
//     return (*env)->IsInstanceOf(env, obj, clazz);
// }
//
// static inline jmethodID GetMethodID(JNIEnv * env, jclass clazz, char * name, char * sig) {
//     return (*env)->GetMethodID(env, clazz, name, sig);
// }
//
// static inline jobject CallObjectMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallObjectMethodA(env, obj, methodID, args);
// }
//
// static inline jboolean CallBooleanMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallBooleanMethodA(env, obj, methodID, args);
// }
//
// static inline jbyte CallByteMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallByteMethodA(env, obj, methodID, args);
// }
//
// static inline jchar CallCharMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallCharMethodA(env, obj, methodID, args);
// }
//
// static inline jshort CallShortMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallShortMethodA(env, obj, methodID, args);
// }
//
// static inline jint CallIntMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallIntMethodA(env, obj, methodID, args);
// }
//
// static inline jlong CallLongMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallLongMethodA(env, obj, methodID, args);
// }
//
// static inline jfloat CallFloatMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallFloatMethodA(env, obj, methodID, args);
// }
//
// static inline jdouble CallDoubleMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     return (*env)->CallDoubleMethodA(env, obj, methodID, args);
// }
//
// static inline void CallVoidMethodA(JNIEnv * env, jobject obj, jmethodID methodID, jvalue * args) {
//     (*env)->CallVoidMethodA(env, obj, methodID, args);
// }
//
// static inline jobject CallNonvirtualObjectMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualObjectMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jboolean CallNonvirtualBooleanMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualBooleanMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jbyte CallNonvirtualByteMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualByteMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jchar CallNonvirtualCharMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualCharMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jshort CallNonvirtualShortMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualShortMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jint CallNonvirtualIntMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualIntMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jlong CallNonvirtualLongMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualLongMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jfloat CallNonvirtualFloatMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualFloatMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jdouble CallNonvirtualDoubleMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallNonvirtualDoubleMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline void CallNonvirtualVoidMethodA(JNIEnv * env, jobject obj, jclass clazz, jmethodID methodID, jvalue * args) {
//     (*env)->CallNonvirtualVoidMethodA(env, obj, clazz, methodID, args);
// }
//
// static inline jfieldID GetFieldID(JNIEnv * env, jclass clazz, char * name, char * sig) {
//     return (*env)->GetFieldID(env, clazz, name, sig);
// }
//
// static inline jobject GetObjectField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetObjectField(env, obj, fieldID);
// }
//
// static inline jboolean GetBooleanField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetBooleanField(env, obj, fieldID);
// }
//
// static inline jbyte GetByteField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetByteField(env, obj, fieldID);
// }
//
// static inline jchar GetCharField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetCharField(env, obj, fieldID);
// }
//
// static inline jshort GetShortField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetShortField(env, obj, fieldID);
// }
//
// static inline jint GetIntField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetIntField(env, obj, fieldID);
// }
//
// static inline jlong GetLongField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetLongField(env, obj, fieldID);
// }
//
// static inline jfloat GetFloatField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetFloatField(env, obj, fieldID);
// }
//
// static inline jdouble GetDoubleField(JNIEnv * env, jobject obj, jfieldID fieldID) {
//     return (*env)->GetDoubleField(env, obj, fieldID);
// }
//
// static inline void SetObjectField(JNIEnv * env, jobject obj, jfieldID fieldID, jobject val) {
//     (*env)->SetObjectField(env, obj, fieldID, val);
// }
//
// static inline void SetBooleanField(JNIEnv * env, jobject obj, jfieldID fieldID, jboolean val) {
//     (*env)->SetBooleanField(env, obj, fieldID, val);
// }
//
// static inline void SetByteField(JNIEnv * env, jobject obj, jfieldID fieldID, jbyte val) {
//     (*env)->SetByteField(env, obj, fieldID, val);
// }
//
// static inline void SetCharField(JNIEnv * env, jobject obj, jfieldID fieldID, jchar val) {
//     (*env)->SetCharField(env, obj, fieldID, val);
// }
//
// static inline void SetShortField(JNIEnv * env, jobject obj, jfieldID fieldID, jshort val) {
//     (*env)->SetShortField(env, obj, fieldID, val);
// }
//
// static inline void SetIntField(JNIEnv * env, jobject obj, jfieldID fieldID, jint val) {
//     (*env)->SetIntField(env, obj, fieldID, val);
// }
//
// static inline void SetLongField(JNIEnv * env, jobject obj, jfieldID fieldID, jlong val) {
//     (*env)->SetLongField(env, obj, fieldID, val);
// }
//
// static inline void SetFloatField(JNIEnv * env, jobject obj, jfieldID fieldID, jfloat val) {
//     (*env)->SetFloatField(env, obj, fieldID, val);
// }
//
// static inline void SetDoubleField(JNIEnv * env, jobject obj, jfieldID fieldID, jdouble val) {
//     (*env)->SetDoubleField(env, obj, fieldID, val);
// }
//
// static inline jmethodID GetStaticMethodID(JNIEnv * env, jclass clazz, char * name, char * sig) {
//     return (*env)->GetStaticMethodID(env, clazz, name, sig);
// }
//
// static inline jobject CallStaticObjectMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticObjectMethodA(env, clazz, methodID, args);
// }
//
// static inline jboolean CallStaticBooleanMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticBooleanMethodA(env, clazz, methodID, args);
// }
//
// static inline jbyte CallStaticByteMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticByteMethodA(env, clazz, methodID, args);
// }
//
// static inline jchar CallStaticCharMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticCharMethodA(env, clazz, methodID, args);
// }
//
// static inline jshort CallStaticShortMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticShortMethodA(env, clazz, methodID, args);
// }
//
// static inline jint CallStaticIntMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticIntMethodA(env, clazz, methodID, args);
// }
//
// static inline jlong CallStaticLongMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticLongMethodA(env, clazz, methodID, args);
// }
//
// static inline jfloat CallStaticFloatMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticFloatMethodA(env, clazz, methodID, args);
// }
//
// static inline jdouble CallStaticDoubleMethodA(JNIEnv * env, jclass clazz, jmethodID methodID, jvalue * args) {
//     return (*env)->CallStaticDoubleMethodA(env, clazz, methodID, args);
// }
//
// static inline void CallStaticVoidMethodA(JNIEnv * env, jclass cls, jmethodID methodID, jvalue * args) {
//     (*env)->CallStaticVoidMethodA(env, cls, methodID, args);
// }
//
// static inline jfieldID GetStaticFieldID(JNIEnv * env, jclass clazz, char * name, char * sig) {
//     return (*env)->GetStaticFieldID(env, clazz, name, sig);
// }
//
// static inline jobject GetStaticObjectField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticObjectField(env, clazz, fieldID);
// }
//
// static inline jboolean GetStaticBooleanField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticBooleanField(env, clazz, fieldID);
// }
//
// static inline jbyte GetStaticByteField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticByteField(env, clazz, fieldID);
// }
//
// static inline jchar GetStaticCharField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticCharField(env, clazz, fieldID);
// }
//
// static inline jshort GetStaticShortField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticShortField(env, clazz, fieldID);
// }
//
// static inline jint GetStaticIntField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticIntField(env, clazz, fieldID);
// }
//
// static inline jlong GetStaticLongField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticLongField(env, clazz, fieldID);
// }
//
// static inline jfloat GetStaticFloatField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticFloatField(env, clazz, fieldID);
// }
//
// static inline jdouble GetStaticDoubleField(JNIEnv * env, jclass clazz, jfieldID fieldID) {
//     return (*env)->GetStaticDoubleField(env, clazz, fieldID);
// }
//
// static inline void SetStaticObjectField(JNIEnv * env, jclass clazz, jfieldID fieldID, jobject value) {
//     (*env)->SetStaticObjectField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticBooleanField(JNIEnv * env, jclass clazz, jfieldID fieldID, jboolean value) {
//     (*env)->SetStaticBooleanField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticByteField(JNIEnv * env, jclass clazz, jfieldID fieldID, jbyte value) {
//     (*env)->SetStaticByteField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticCharField(JNIEnv * env, jclass clazz, jfieldID fieldID, jchar value) {
//     (*env)->SetStaticCharField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticShortField(JNIEnv * env, jclass clazz, jfieldID fieldID, jshort value) {
//     (*env)->SetStaticShortField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticIntField(JNIEnv * env, jclass clazz, jfieldID fieldID, jint value) {
//     (*env)->SetStaticIntField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticLongField(JNIEnv * env, jclass clazz, jfieldID fieldID, jlong value) {
//     (*env)->SetStaticLongField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticFloatField(JNIEnv * env, jclass clazz, jfieldID fieldID, jfloat value) {
//     (*env)->SetStaticFloatField(env, clazz, fieldID, value);
// }
//
// static inline void SetStaticDoubleField(JNIEnv * env, jclass clazz, jfieldID fieldID, jdouble value) {
//     (*env)->SetStaticDoubleField(env, clazz, fieldID, value);
// }
//
// static inline jstring NewString(JNIEnv * env, jchar * unicode, jsize len) {
//     return (*env)->NewString(env, unicode, len);
// }
//
// static inline jsize GetStringLength(JNIEnv * env, jstring str) {
//     return (*env)->GetStringLength(env, str);
// }
//
// static inline jsize GetStringUTFLength(JNIEnv * env, jstring str) {
//     return (*env)->GetStringUTFLength(env, str);
// }
//
// static inline jsize GetArrayLength(JNIEnv * env, jarray array) {
//     return (*env)->GetArrayLength(env, array);
// }
//
// static inline jobjectArray NewObjectArray(JNIEnv * env, jsize len, jclass clazz, jobject init) {
//     return (*env)->NewObjectArray(env, len, clazz, init);
// }
//
// static inline jobject GetObjectArrayElement(JNIEnv * env, jobjectArray array, jsize index) {
//     return (*env)->GetObjectArrayElement(env, array, index);
// }
//
// static inline void SetObjectArrayElement(JNIEnv * env, jobjectArray array, jsize index, jobject val) {
//     (*env)->SetObjectArrayElement(env, array, index, val);
// }
//
// static inline jbooleanArray NewBooleanArray(JNIEnv * env, jsize len) {
//     return (*env)->NewBooleanArray(env, len);
// }
//
// static inline jbyteArray NewByteArray(JNIEnv * env, jsize len) {
//     return (*env)->NewByteArray(env, len);
// }
//
// static inline jcharArray NewCharArray(JNIEnv * env, jsize len) {
//     return (*env)->NewCharArray(env, len);
// }
//
// static inline jshortArray NewShortArray(JNIEnv * env, jsize len) {
//     return (*env)->NewShortArray(env, len);
// }
//
// static inline jintArray NewIntArray(JNIEnv * env, jsize len) {
//     return (*env)->NewIntArray(env, len);
// }
//
// static inline jlongArray NewLongArray(JNIEnv * env, jsize len) {
//     return (*env)->NewLongArray(env, len);
// }
//
// static inline jfloatArray NewFloatArray(JNIEnv * env, jsize len) {
//     return (*env)->NewFloatArray(env, len);
// }
//
// static inline jdoubleArray NewDoubleArray(JNIEnv * env, jsize len) {
//     return (*env)->NewDoubleArray(env, len);
// }
//
// static inline void GetBooleanArrayRegion(JNIEnv * env, jbooleanArray array, jsize start, jsize l, jboolean * buf) {
//     (*env)->GetBooleanArrayRegion(env, array, start, l, buf);
// }
//
// static inline void GetByteArrayRegion(JNIEnv * env, jbyteArray array, jsize start, jsize len, jbyte * buf) {
//     (*env)->GetByteArrayRegion(env, array, start, len, buf);
// }
//
// static inline void GetCharArrayRegion(JNIEnv * env, jcharArray array, jsize start, jsize len, jchar * buf) {
//     (*env)->GetCharArrayRegion(env, array, start, len, buf);
// }
//
// static inline void GetShortArrayRegion(JNIEnv * env, jshortArray array, jsize start, jsize len, jshort * buf) {
//     (*env)->GetShortArrayRegion(env, array, start, len, buf);
// }
//
// static inline void GetIntArrayRegion(JNIEnv * env, jintArray array, jsize start, jsize len, jint * buf) {
//     (*env)->GetIntArrayRegion(env, array, start, len, buf);
// }
//
// static inline void GetLongArrayRegion(JNIEnv * env, jlongArray array, jsize start, jsize len, jlong * buf) {
//     (*env)->GetLongArrayRegion(env, array, start, len, buf);
// }
//
// static inline void GetFloatArrayRegion(JNIEnv * env, jfloatArray array, jsize start, jsize len, jfloat * buf) {
//     (*env)->GetFloatArrayRegion(env, array, start, len, buf);
// }
//
// static inline void GetDoubleArrayRegion(JNIEnv * env, jdoubleArray array, jsize start, jsize len, jdouble * buf) {
//     (*env)->GetDoubleArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetBooleanArrayRegion(JNIEnv * env, jbooleanArray array, jsize start, jsize l, jboolean * buf) {
//     (*env)->SetBooleanArrayRegion(env, array, start, l, buf);
// }
//
// static inline void SetByteArrayRegion(JNIEnv * env, jbyteArray array, jsize start, jsize len, jbyte * buf) {
//     (*env)->SetByteArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetCharArrayRegion(JNIEnv * env, jcharArray array, jsize start, jsize len, jchar * buf) {
//     (*env)->SetCharArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetShortArrayRegion(JNIEnv * env, jshortArray array, jsize start, jsize len, jshort * buf) {
//     (*env)->SetShortArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetIntArrayRegion(JNIEnv * env, jintArray array, jsize start, jsize len, jint * buf) {
//     (*env)->SetIntArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetLongArrayRegion(JNIEnv * env, jlongArray array, jsize start, jsize len, jlong * buf) {
//     (*env)->SetLongArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetFloatArrayRegion(JNIEnv * env, jfloatArray array, jsize start, jsize len, jfloat * buf) {
//     (*env)->SetFloatArrayRegion(env, array, start, len, buf);
// }
//
// static inline void SetDoubleArrayRegion(JNIEnv * env, jdoubleArray array, jsize start, jsize len, jdouble * buf) {
//     (*env)->SetDoubleArrayRegion(env, array, start, len, buf);
// }
//
// static inline jint MonitorEnter(JNIEnv * env, jobject obj) {
//     return (*env)->MonitorEnter(env, obj);
// }
//
// static inline jint MonitorExit(JNIEnv * env, jobject obj) {
//     return (*env)->MonitorExit(env, obj);
// }
//
// static inline void GetStringUTFRegion(JNIEnv * env, jstring str, jsize start, jsize len, char * buf) {
//     (*env)->GetStringUTFRegion(env, str, start, len, buf);
// }
//
// static inline void * GetPrimitiveArrayCritical(JNIEnv * env, jarray array) {
//     return (*env)->GetPrimitiveArrayCritical(env, array, NULL);
// }
//
// static inline void ReleasePrimitiveArrayCritical(JNIEnv * env, jarray array, void * carray, jint mode) {
//     (*env)->ReleasePrimitiveArrayCritical(env, array, carray, mode);
// }
//
// static inline jweak NewWeakGlobalRef(JNIEnv * env, jobject obj) {
//     return (*env)->NewWeakGlobalRef(env, obj);
// }
//
// static inline void DeleteWeakGlobalRef(JNIEnv * env, jweak ref) {
//     (*env)->DeleteWeakGlobalRef(env, ref);
// }
//
// static inline jboolean ExceptionCheck(JNIEnv * env) {
//     return (*env)->ExceptionCheck(env);
// }
//
// static inline jobject NewDirectByteBuffer(JNIEnv * env, void * address, jlong capacity) {
//     return (*env)->NewDirectByteBuffer(env, address, capacity);
// }
//
// static inline void * GetDirectBufferAddress(JNIEnv * env, jobject buf) {
//     return (*env)->GetDirectBufferAddress(env, buf);
// }
//
// static inline jlong GetDirectBufferCapacity(JNIEnv * env, jobject buf) {
//     return (*env)->GetDirectBufferCapacity(env, buf);
// }
//
//
// static inline int GetByteHash(JNIEnv * env, jbyteArray array) {
//		jbyte* bytes = (*env)->GetByteArrayElements(env, array, NULL);
//		jsize length = (*env)->GetArrayLength(env, array);
//		size_t longValue = 2166136261;
//		int hash = (int)longValue;
//		for (int i = 0; i < length; i++) {
//			hash = (hash ^ bytes[i]) * 16777619;
// 		}
//		return hash;
// }
import "C"
import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unsafe"

	"github.com/qtgolang/SunnyNet/JavaApi/sig"
)

const (
	JNI_VERSION_1_1 = 0x00010001
	JNI_VERSION_1_2 = 0x00010002
	JNI_VERSION_1_4 = 0x00010004
	JNI_VERSION_1_6 = 0x00010006
	JNI_VERSION_1_8 = 0x00010008
	JNI_VERSION_9   = 0x00090000
	JNI_VERSION_10  = 0x000a0000

	JNI_FALSE = 0
	JNI_TRUE  = 1

	JNI_OK        = 0    /* success */
	JNI_ERR       = (-1) /* unknown error */
	JNI_EDETACHED = (-2) /* maps detached from the VM */
	JNI_EVERSION  = (-3) /* JNI version error */
	JNI_ENOMEM    = (-4) /* not enough memory */
	JNI_EEXIST    = (-5) /* VM already created */
	JNI_EINVAL    = (-6) /* invalid arguments */

	JNI_COMMIT = 1
	JNI_ABORT  = 2
)

type RefType int

const (
	Invalid RefType = iota
	Local
	Global
	WeakGlobal
)

type Jobject = uintptr
type Jclass = uintptr
type Jthrowable = uintptr
type Jstring = uintptr
type Jarray = uintptr
type JbooleanArray = uintptr
type JbyteArray = uintptr
type JcharArray = uintptr
type JshortArray = uintptr
type JintArray = uintptr
type JlongArray = uintptr
type JfloatArray = uintptr
type JdoubleArray = uintptr
type JobjectArray = uintptr
type Jweak = uintptr
type Jvalue = uint64
type JmethodID = uintptr
type JfieldID = uintptr

var NULL = uint64(0)

type VM uintptr

func (vm VM) AttachCurrentThread() (Env, int) {
	var env *C.JNIEnv
	ret := int(C.AttachCurrentThread((*C.JavaVM)(unsafe.Pointer(vm)), &env))
	return Env(unsafe.Pointer(env)), ret
}

func (vm VM) AttachCurrentThreadAsDaemon() (Env, int) {
	var env *C.JNIEnv
	ret := int(C.AttachCurrentThreadAsDaemon((*C.JavaVM)(unsafe.Pointer(vm)), &env))
	return Env(unsafe.Pointer(env)), ret
}

func (vm VM) GetEnv(version int) (Env, int) {
	var env *C.JNIEnv
	ret := int(C.GetEnv((*C.JavaVM)(unsafe.Pointer(vm)), &env, C.jint(version)))
	return Env(unsafe.Pointer(env)), ret
}

type Env uintptr

func (env Env) GetJavaVM() (VM, int) {
	var vm *C.JavaVM
	ret := int(C.GetJavaVM((*C.JNIEnv)(unsafe.Pointer(env)), &vm))
	return VM(unsafe.Pointer(vm)), ret
}

func (env Env) GetObjectRefType(obj Jobject) RefType {
	return RefType(C.GetObjectRefType((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj)))
}

func (env Env) NewString(s string) Jstring {
	codes := utf16.Encode([]rune(s))
	size := len(codes)
	if size <= 0 {
		codes = utf16.Encode([]rune{0})
		return Jstring(C.NewString((*C.JNIEnv)(unsafe.Pointer(env)), (*C.jchar)(unsafe.Pointer(&codes[0])), C.jsize(size)))
	} else {
		return Jstring(C.NewString((*C.JNIEnv)(unsafe.Pointer(env)), (*C.jchar)(unsafe.Pointer(&codes[0])), C.jsize(size)))
	}
}
func (env Env) GetException() (string, bool) {
	if !env.ExceptionCheck() {
		return "", false
	}
	exception := env.ExceptionOccurred()
	env.ExceptionClear()
	exceptionClass := env.GetObjectClass(exception)
	if exceptionClass == 0 {
		return "", false
	}
	toStringMethod := env.GetMethodID(exceptionClass, "toString", fmt.Sprintf("()%s", sig.String))
	if toStringMethod == 0 {
		return "", false
	}
	exceptionString := env.CallObjectMethodA(exception, toStringMethod, 0)
	if exceptionString == 0 {
		return "", false
	}
	defer env.DeleteLocalRef(exceptionString)
	str := env.GetStringUTF(exceptionString)
	return string(str), true
}
func (env Env) GetStringUTF(ptr Jstring) []byte {
	jstr := C.jstring(ptr)
	size := C.GetStringUTFLength((*C.JNIEnv)(unsafe.Pointer(env)), jstr)
	ret := make([]byte, int(size))
	C.GetStringUTFRegion((*C.JNIEnv)(unsafe.Pointer(env)), jstr, C.jsize(0), C.GetStringLength((*C.JNIEnv)(unsafe.Pointer(env)), jstr), cmem(ret))
	return ret
}
func (env Env) GetString(ptr Jstring) string {
	return string(env.GetStringUTF(ptr))
}
func (env Env) GetBytes(ptr JbyteArray) []byte {
	jbs := C.jarray(ptr)
	size := C.GetArrayLength((*C.JNIEnv)(unsafe.Pointer(env)), jbs)
	buf := make([]byte, int(size))
	C.GetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(jbs), C.jsize(0), C.jsize(len(buf)), cByteArray(buf))
	return buf
}
func (env Env) NewDirectByteBuffer(address unsafe.Pointer, capacity int) Jobject {
	return Jobject(C.NewDirectByteBuffer((*C.JNIEnv)(unsafe.Pointer(env)), address, C.jlong(capacity)))
}

func (env Env) GetDirectBufferAddress(buf Jobject) unsafe.Pointer {
	return C.GetDirectBufferAddress((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(buf))
}

func (env Env) GetDirectBufferCapacity(buf Jobject) int {
	return int(C.GetDirectBufferCapacity((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(buf)))
}

func (env Env) GetBooleanArrayElement(array JbooleanArray, index int) bool {
	var ret C.jboolean
	C.GetBooleanArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbooleanArray(array), C.jsize(index), C.jsize(1), &ret)
	return ret != C.JNI_FALSE
}

func (env Env) GetByteArrayElement(array JbyteArray, index int) byte {
	var ret C.jbyte
	C.GetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array), C.jsize(index), C.jsize(1), &ret)
	return byte(ret)
}
func (env Env) GetByteArrayHash(array JbyteArray) int {
	return int(C.GetByteHash((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array)))
}
func (env Env) GetCharArrayElement(array JcharArray, index int) uint16 {
	var ret C.jchar
	C.GetCharArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jcharArray(array), C.jsize(index), C.jsize(1), &ret)
	return uint16(ret)
}

func (env Env) GetShortArrayElement(array JshortArray, index int) int16 {
	var ret C.jshort
	C.GetShortArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jshortArray(array), C.jsize(index), C.jsize(1), &ret)
	return int16(ret)
}

func (env Env) GetIntArrayElement(array JintArray, index int) int {
	var ret C.jint
	C.GetIntArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jintArray(array), C.jsize(index), C.jsize(1), &ret)
	return int(ret)
}

func (env Env) GetLongArrayElement(array JlongArray, index int) int64 {
	var ret C.jlong
	C.GetLongArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jlongArray(array), C.jsize(index), C.jsize(1), &ret)
	return int64(ret)
}

func (env Env) GetFloatArrayElement(array JfloatArray, index int) float32 {
	var ret C.jfloat
	C.GetFloatArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jfloatArray(array), C.jsize(index), C.jsize(1), &ret)
	return float32(ret)
}

func (env Env) GetDoubleArrayElement(array JdoubleArray, index int) float64 {
	var ret C.jdouble
	C.GetDoubleArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jdoubleArray(array), C.jsize(index), C.jsize(1), &ret)
	return float64(ret)
}

func (env Env) SetBooleanArrayElement(array JbooleanArray, index int, v bool) {
	cv := cbool(v)
	C.SetBooleanArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbooleanArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetByteArrayElement(array JbyteArray, index int, v byte) {
	cv := C.jbyte(v)
	C.SetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetCharArrayElement(array JcharArray, index int, v uint16) {
	cv := C.jchar(v)
	C.SetCharArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jcharArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetShortArrayElement(array JshortArray, index int, v int16) {
	cv := C.jshort(v)
	C.SetShortArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jshortArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetIntArrayElement(array JintArray, index int, v int) {
	cv := C.jint(v)
	C.SetIntArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jintArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetLongArrayElement(array JlongArray, index int, v int64) {
	cv := C.jlong(v)
	C.SetLongArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jlongArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetFloatArrayElement(array JfloatArray, index int, v float32) {
	cv := C.jfloat(v)
	C.SetFloatArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jfloatArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetDoubleArrayElement(array JdoubleArray, index int, v float64) {
	cv := C.jdouble(v)
	C.SetDoubleArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jdoubleArray(array), C.jsize(index), C.jsize(1), &cv)
}
func (vm VM) DestroyJavaVM() int {
	return int(C.DestroyJavaVM((*C.JavaVM)(unsafe.Pointer(vm))))
}

func (vm VM) DetachCurrentThread() int {
	return int(C.DetachCurrentThread((*C.JavaVM)(unsafe.Pointer(vm))))
}

func (env Env) FindClass(name string) Jclass {
	cstr_name := C.CString(strings.ReplaceAll(name, ".", "/"))
	defer C.free(unsafe.Pointer(cstr_name))
	return Jclass(C.FindClass((*C.JNIEnv)(unsafe.Pointer(env)), cstr_name))
}
func (env Env) GetVersion() int {
	return int(C.GetVersion((*C.JNIEnv)(unsafe.Pointer(env))))
}

func (env Env) FromReflectedMethod(method Jobject) JmethodID {
	return JmethodID(unsafe.Pointer(C.FromReflectedMethod((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(method))))
}

func (env Env) FromReflectedField(field Jobject) JfieldID {
	return JfieldID(unsafe.Pointer(C.FromReflectedField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(field))))
}

func (env Env) ToReflectedMethod(cls Jclass, methodID JmethodID, isStatic bool) Jobject {
	return Jobject(C.ToReflectedMethod((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(cls), C.jmethodID(unsafe.Pointer(methodID)), cbool(isStatic)))
}

func (env Env) GetSuperclass(sub Jclass) Jclass {
	return Jclass(C.GetSuperclass((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(sub)))
}

func (env Env) IsAssignableFrom(sub Jclass, sup Jclass) bool {
	return C.IsAssignableFrom((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(sub), C.jclass(sup)) != C.JNI_FALSE
}

func (env Env) ToReflectedField(cls Jclass, fieldID JfieldID, isStatic bool) Jobject {
	return Jobject(C.ToReflectedField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(cls), C.jfieldID(unsafe.Pointer(fieldID)), cbool(isStatic)))
}

func (env Env) Throw(obj Jthrowable) int {
	return int(C.Throw((*C.JNIEnv)(unsafe.Pointer(env)), C.jthrowable(obj)))
}

func (env Env) ThrowNew(clazz Jclass, msg string) int {
	cstr_msg := C.CString(msg)
	defer C.free(unsafe.Pointer(cstr_msg))
	return int(C.ThrowNew((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), cstr_msg))
}

func (env Env) ExceptionOccurred() Jthrowable {
	return Jthrowable(C.ExceptionOccurred((*C.JNIEnv)(unsafe.Pointer(env))))
}

func (env Env) ExceptionDescribe() {
	C.ExceptionDescribe((*C.JNIEnv)(unsafe.Pointer(env)))
}

func (env Env) ExceptionClear() {
	C.ExceptionClear((*C.JNIEnv)(unsafe.Pointer(env)))
}

func (env Env) FatalError(msg string) {
	cstr_msg := C.CString(msg)
	defer C.free(unsafe.Pointer(cstr_msg))
	C.FatalError((*C.JNIEnv)(unsafe.Pointer(env)), cstr_msg)
}

func (env Env) PushLocalFrame(capacity int) int {
	return int(C.PushLocalFrame((*C.JNIEnv)(unsafe.Pointer(env)), C.jint(capacity)))
}

func (env Env) PopLocalFrame(result Jobject) Jobject {
	return Jobject(C.PopLocalFrame((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(result)))
}

func (env Env) NewGlobalRef(lobj Jobject) Jobject {
	return Jobject(C.NewGlobalRef((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(lobj)))
}

func (env Env) DeleteGlobalRef(gref Jobject) {
	C.DeleteGlobalRef((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(gref))
}

func (env Env) DeleteLocalRef(obj Jobject) {
	if obj == 0 {
		return
	}
	C.DeleteLocalRef((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj))
}

func (env Env) IsSameObject(obj1 Jobject, obj2 Jobject) bool {
	return C.IsSameObject((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj1), C.jobject(obj2)) != C.JNI_FALSE
}

func (env Env) NewLocalRef(ref Jobject) Jobject {
	return Jobject(C.NewLocalRef((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(ref)))
}

func (env Env) EnsureLocalCapacity(capacity int) int {
	return int(C.EnsureLocalCapacity((*C.JNIEnv)(unsafe.Pointer(env)), C.jint(capacity)))
}

func (env Env) AllocObject(clazz Jclass) Jobject {
	return Jobject(C.AllocObject((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz)))
}

func (env Env) NewObjectA(clazz Jclass, methodID JmethodID, args ...Jvalue) Jobject {
	return Jobject(C.NewObjectA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) GetObjectClass(obj Jobject) Jclass {
	return Jclass(C.GetObjectClass((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj)))
}

func (env Env) IsInstanceOf(obj Jobject, clazz Jclass) bool {
	return C.IsInstanceOf((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz)) != C.JNI_FALSE
}

func (env Env) GetMethodID(clazz Jclass, name string, sig string) JmethodID {
	cstr_name := C.CString(name)
	defer C.free(unsafe.Pointer(cstr_name))
	cstr_sig := C.CString(sig)
	defer C.free(unsafe.Pointer(cstr_sig))
	return JmethodID(unsafe.Pointer(C.GetMethodID((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), cstr_name, cstr_sig)))
}

func (env Env) CallObjectMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) Jobject {
	return Jobject(C.CallObjectMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallBooleanMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) bool {
	return C.CallBooleanMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)) != C.JNI_FALSE
}

func (env Env) CallByteMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) byte {
	return byte(C.CallByteMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallCharMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) uint16 {
	return uint16(C.CallCharMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallShortMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) int16 {
	return int16(C.CallShortMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallIntMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) int {
	return int(C.CallIntMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallLongMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) int64 {
	return int64(C.CallLongMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallFloatMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) float32 {
	return float32(C.CallFloatMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallDoubleMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) float64 {
	return float64(C.CallDoubleMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallVoidMethodA(obj Jobject, methodID JmethodID, args ...Jvalue) {
	C.CallVoidMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jmethodID(unsafe.Pointer(methodID)), cvals(args))
}

func (env Env) CallNonvirtualObjectMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) Jobject {
	return Jobject(C.CallNonvirtualObjectMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualBooleanMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) bool {
	return C.CallNonvirtualBooleanMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)) != C.JNI_FALSE
}

func (env Env) CallNonvirtualByteMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) byte {
	return byte(C.CallNonvirtualByteMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualCharMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) uint16 {
	return uint16(C.CallNonvirtualCharMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualShortMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) int16 {
	return int16(C.CallNonvirtualShortMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualIntMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) int {
	return int(C.CallNonvirtualIntMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualLongMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) int64 {
	return int64(C.CallNonvirtualLongMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualFloatMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) float32 {
	return float32(C.CallNonvirtualFloatMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualDoubleMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) float64 {
	return float64(C.CallNonvirtualDoubleMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallNonvirtualVoidMethodA(obj Jobject, clazz Jclass, methodID JmethodID, args ...Jvalue) {
	C.CallNonvirtualVoidMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args))
}

func (env Env) GetFieldID(clazz Jclass, name string, sig string) JfieldID {
	cstr_name := C.CString(name)
	defer C.free(unsafe.Pointer(cstr_name))
	cstr_sig := C.CString(sig)
	defer C.free(unsafe.Pointer(cstr_sig))
	return JfieldID(unsafe.Pointer(C.GetFieldID((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), cstr_name, cstr_sig)))
}

func (env Env) GetObjectField(obj Jobject, fieldID JfieldID) Jobject {
	return Jobject(C.GetObjectField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetBooleanField(obj Jobject, fieldID JfieldID) bool {
	return C.GetBooleanField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))) != C.JNI_FALSE
}

func (env Env) GetByteField(obj Jobject, fieldID JfieldID) byte {
	return byte(C.GetByteField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetCharField(obj Jobject, fieldID JfieldID) uint16 {
	return uint16(C.GetCharField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetShortField(obj Jobject, fieldID JfieldID) int16 {
	return int16(C.GetShortField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetIntField(obj Jobject, fieldID JfieldID) int {
	return int(C.GetIntField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetLongField(obj Jobject, fieldID JfieldID) int64 {
	return int64(C.GetLongField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetFloatField(obj Jobject, fieldID JfieldID) float32 {
	return float32(C.GetFloatField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetDoubleField(obj Jobject, fieldID JfieldID) float64 {
	return float64(C.GetDoubleField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) SetObjectField(obj Jobject, fieldID JfieldID, val Jobject) {
	C.SetObjectField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jobject(val))
}

func (env Env) SetBooleanField(obj Jobject, fieldID JfieldID, val bool) {
	C.SetBooleanField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), cbool(val))
}

func (env Env) SetByteField(obj Jobject, fieldID JfieldID, val byte) {
	C.SetByteField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jbyte(val))
}

func (env Env) SetCharField(obj Jobject, fieldID JfieldID, val uint16) {
	C.SetCharField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jchar(val))
}

func (env Env) SetShortField(obj Jobject, fieldID JfieldID, val int16) {
	C.SetShortField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jshort(val))
}

func (env Env) SetIntField(obj Jobject, fieldID JfieldID, val int) {
	C.SetIntField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jint(val))
}

func (env Env) SetLongField(obj Jobject, fieldID JfieldID, val int64) {
	C.SetLongField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jlong(val))
}

func (env Env) SetFloatField(obj Jobject, fieldID JfieldID, val float32) {
	C.SetFloatField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jfloat(val))
}

func (env Env) SetDoubleField(obj Jobject, fieldID JfieldID, val float64) {
	C.SetDoubleField((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj), C.jfieldID(unsafe.Pointer(fieldID)), C.jdouble(val))
}

func (env Env) GetStaticMethodID(clazz Jclass, name string, sig string) JmethodID {
	cstr_name := C.CString(name)
	defer C.free(unsafe.Pointer(cstr_name))
	cstr_sig := C.CString(sig)
	defer C.free(unsafe.Pointer(cstr_sig))
	return JmethodID(unsafe.Pointer(C.GetStaticMethodID((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), cstr_name, cstr_sig)))
}

func (env Env) CallStaticObjectMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) Jobject {
	return Jobject(C.CallStaticObjectMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticBooleanMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) bool {
	return C.CallStaticBooleanMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)) != C.JNI_FALSE
}

func (env Env) CallStaticByteMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) byte {
	return byte(C.CallStaticByteMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticCharMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) uint16 {
	return uint16(C.CallStaticCharMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticShortMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) int16 {
	return int16(C.CallStaticShortMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticIntMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) int {
	return int(C.CallStaticIntMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticLongMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) int64 {
	return int64(C.CallStaticLongMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticFloatMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) float32 {
	return float32(C.CallStaticFloatMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticDoubleMethodA(clazz Jclass, methodID JmethodID, args ...Jvalue) float64 {
	return float64(C.CallStaticDoubleMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jmethodID(unsafe.Pointer(methodID)), cvals(args)))
}

func (env Env) CallStaticVoidMethodA(cls Jclass, methodID JmethodID, args ...Jvalue) {
	C.CallStaticVoidMethodA((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(cls), C.jmethodID(unsafe.Pointer(methodID)), cvals(args))
}

func (env Env) GetStaticFieldID(clazz Jclass, name string, sig string) JfieldID {
	cstr_name := C.CString(name)
	defer C.free(unsafe.Pointer(cstr_name))
	cstr_sig := C.CString(sig)
	defer C.free(unsafe.Pointer(cstr_sig))
	return JfieldID(unsafe.Pointer(C.GetStaticFieldID((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), cstr_name, cstr_sig)))
}

func (env Env) GetStaticObjectField(clazz Jclass, fieldID JfieldID) Jobject {
	return Jobject(C.GetStaticObjectField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticBooleanField(clazz Jclass, fieldID JfieldID) bool {
	return C.GetStaticBooleanField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))) != C.JNI_FALSE
}

func (env Env) GetStaticByteField(clazz Jclass, fieldID JfieldID) byte {
	return byte(C.GetStaticByteField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticCharField(clazz Jclass, fieldID JfieldID) uint16 {
	return uint16(C.GetStaticCharField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticShortField(clazz Jclass, fieldID JfieldID) int16 {
	return int16(C.GetStaticShortField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticIntField(clazz Jclass, fieldID JfieldID) int {
	return int(C.GetStaticIntField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticLongField(clazz Jclass, fieldID JfieldID) int64 {
	return int64(C.GetStaticLongField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticFloatField(clazz Jclass, fieldID JfieldID) float32 {
	return float32(C.GetStaticFloatField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) GetStaticDoubleField(clazz Jclass, fieldID JfieldID) float64 {
	return float64(C.GetStaticDoubleField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID))))
}

func (env Env) SetStaticObjectField(clazz Jclass, fieldID JfieldID, value Jobject) {
	C.SetStaticObjectField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jobject(value))
}

func (env Env) SetStaticBooleanField(clazz Jclass, fieldID JfieldID, value bool) {
	C.SetStaticBooleanField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), cbool(value))
}

func (env Env) SetStaticByteField(clazz Jclass, fieldID JfieldID, value byte) {
	C.SetStaticByteField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jbyte(value))
}

func (env Env) SetStaticCharField(clazz Jclass, fieldID JfieldID, value uint16) {
	C.SetStaticCharField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jchar(value))
}

func (env Env) SetStaticShortField(clazz Jclass, fieldID JfieldID, value int16) {
	C.SetStaticShortField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jshort(value))
}

func (env Env) SetStaticIntField(clazz Jclass, fieldID JfieldID, value int) {
	C.SetStaticIntField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jint(value))
}

func (env Env) SetStaticLongField(clazz Jclass, fieldID JfieldID, value int64) {
	C.SetStaticLongField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jlong(value))
}

func (env Env) SetStaticFloatField(clazz Jclass, fieldID JfieldID, value float32) {
	C.SetStaticFloatField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jfloat(value))
}

func (env Env) SetStaticDoubleField(clazz Jclass, fieldID JfieldID, value float64) {
	C.SetStaticDoubleField((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz), C.jfieldID(unsafe.Pointer(fieldID)), C.jdouble(value))
}

func (env Env) GetStringLength(str Jstring) int {
	return int(C.GetStringLength((*C.JNIEnv)(unsafe.Pointer(env)), C.jstring(str)))
}

func (env Env) GetArrayLength(array Jarray) int {
	return int(C.GetArrayLength((*C.JNIEnv)(unsafe.Pointer(env)), C.jarray(array)))
}

func (env Env) NewObjectArray(len int, clazz Jclass, init Jobject) JobjectArray {
	return JobjectArray(C.NewObjectArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len), C.jclass(clazz), C.jobject(init)))
}

func (env Env) GetObjectArrayElement(array JobjectArray, index int) Jobject {
	return Jobject(C.GetObjectArrayElement((*C.JNIEnv)(unsafe.Pointer(env)), C.jobjectArray(array), C.jsize(index)))
}

func (env Env) SetObjectArrayElement(array JobjectArray, index int, val Jobject) {
	C.SetObjectArrayElement((*C.JNIEnv)(unsafe.Pointer(env)), C.jobjectArray(array), C.jsize(index), C.jobject(val))
}

func (env Env) NewBooleanArray(len int) JbooleanArray {
	return JbooleanArray(C.NewBooleanArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) NewByteArray(data []byte) JbyteArray {
	mdata := data
	l := len(mdata)

	if l == 0 {
		mdata = make([]byte, 1)
	}
	array := C.NewByteArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(l))
	C.SetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), array, C.jsize(0), C.jsize(l), cByteArray(data))
	return JbyteArray(array)
}

func (env Env) NewCharArray(len int) JcharArray {
	return JcharArray(C.NewCharArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) NewShortArray(len int) JshortArray {
	return JshortArray(C.NewShortArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) NewIntArray(len int) JintArray {
	return JintArray(C.NewIntArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) NewLongArray(len int) JlongArray {
	return JlongArray(C.NewLongArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) NewFloatArray(len int) JfloatArray {
	return JfloatArray(C.NewFloatArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) NewDoubleArray(len int) JdoubleArray {
	return JdoubleArray(C.NewDoubleArray((*C.JNIEnv)(unsafe.Pointer(env)), C.jsize(len)))
}

func (env Env) GetBooleanArrayRegion(array JbooleanArray, start int, buf []bool) {
	C.GetBooleanArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbooleanArray(array), C.jsize(start), C.jsize(len(buf)), cBooleanArray(buf))
}

func (env Env) GetByteArrayRegion(array JbyteArray, start int, buf []byte) {
	C.GetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array), C.jsize(start), C.jsize(len(buf)), cByteArray(buf))
}

func (env Env) GetCharArrayRegion(array JcharArray, start int, buf []uint16) {
	C.GetCharArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jcharArray(array), C.jsize(start), C.jsize(len(buf)), cCharArray(buf))
}

func (env Env) GetShortArrayRegion(array JshortArray, start int, buf []int16) {
	C.GetShortArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jshortArray(array), C.jsize(start), C.jsize(len(buf)), cShortArray(buf))
}

func (env Env) GetIntArrayRegion(array JintArray, start int, buf []int32) {
	C.GetIntArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jintArray(array), C.jsize(start), C.jsize(len(buf)), cIntArray(buf))
}

func (env Env) GetLongArrayRegion(array JlongArray, start int, buf []int64) {
	C.GetLongArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jlongArray(array), C.jsize(start), C.jsize(len(buf)), cLongArray(buf))
}

func (env Env) GetFloatArrayRegion(array JfloatArray, start int, buf []float32) {
	C.GetFloatArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jfloatArray(array), C.jsize(start), C.jsize(len(buf)), cFloatArray(buf))
}

func (env Env) GetDoubleArrayRegion(array JdoubleArray, start int, buf []float64) {
	C.GetDoubleArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jdoubleArray(array), C.jsize(start), C.jsize(len(buf)), cDoubleArray(buf))
}

func (env Env) SetBooleanArrayRegion(array JbooleanArray, start int, buf []bool) {
	C.SetBooleanArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbooleanArray(array), C.jsize(start), C.jsize(len(buf)), cBooleanArray(buf))
}

func (env Env) SetByteArrayRegion(array JbyteArray, start int, buf []byte) {
	C.SetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array), C.jsize(start), C.jsize(len(buf)), cByteArray(buf))
}

func (env Env) SetCharArrayRegion(array JcharArray, start int, buf []uint16) {
	C.SetCharArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jcharArray(array), C.jsize(start), C.jsize(len(buf)), cCharArray(buf))
}

func (env Env) SetShortArrayRegion(array JshortArray, start int, buf []int16) {
	C.SetShortArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jshortArray(array), C.jsize(start), C.jsize(len(buf)), cShortArray(buf))
}

func (env Env) SetIntArrayRegion(array JintArray, start int, buf []int32) {
	C.SetIntArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jintArray(array), C.jsize(start), C.jsize(len(buf)), cIntArray(buf))
}

func (env Env) SetLongArrayRegion(array JlongArray, start int, buf []int64) {
	C.SetLongArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jlongArray(array), C.jsize(start), C.jsize(len(buf)), cLongArray(buf))
}

func (env Env) SetFloatArrayRegion(array JfloatArray, start int, buf []float32) {
	C.SetFloatArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jfloatArray(array), C.jsize(start), C.jsize(len(buf)), cFloatArray(buf))
}

func (env Env) SetDoubleArrayRegion(array JdoubleArray, start int, buf []float64) {
	C.SetDoubleArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jdoubleArray(array), C.jsize(start), C.jsize(len(buf)), cDoubleArray(buf))
}

func (env Env) MonitorEnter(obj Jobject) int {
	return int(C.MonitorEnter((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj)))
}

func (env Env) MonitorExit(obj Jobject) int {
	return int(C.MonitorExit((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj)))
}

func (env Env) GetPrimitiveArrayCritical(array Jarray) unsafe.Pointer {
	return C.GetPrimitiveArrayCritical((*C.JNIEnv)(unsafe.Pointer(env)), C.jarray(array))
}

func (env Env) ReleasePrimitiveArrayCritical(array Jarray, carray unsafe.Pointer, mode int) {
	C.ReleasePrimitiveArrayCritical((*C.JNIEnv)(unsafe.Pointer(env)), C.jarray(array), carray, C.jint(mode))
}

func (env Env) NewWeakGlobalRef(obj Jobject) Jweak {
	return Jweak(C.NewWeakGlobalRef((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj)))
}

func (env Env) DeleteWeakGlobalRef(ref Jweak) {
	C.DeleteWeakGlobalRef((*C.JNIEnv)(unsafe.Pointer(env)), C.jweak(ref))
}

func (env Env) ExceptionCheck() bool {
	return C.ExceptionCheck((*C.JNIEnv)(unsafe.Pointer(env))) != C.JNI_FALSE
}

func (env Env) GetModule(clazz Jclass) Jobject {
	return Jobject(C.GetModule((*C.JNIEnv)(unsafe.Pointer(env)), C.jclass(clazz)))
}

func DoubleValue(f float64) Jvalue {
	return *(*Jvalue)(unsafe.Pointer(&f))
}

func FloatValue(f float32) Jvalue {
	return Jvalue(*(*uint32)(unsafe.Pointer(&f)))
}

func Int8Value(i int8) Jvalue {
	return Jvalue(*(*uint8)(unsafe.Pointer(&i)))
}

func Int16Value(i int16) Jvalue {
	return Jvalue(*(*uint16)(unsafe.Pointer(&i)))
}

func Int32Value(i int32) Jvalue {
	return Jvalue(*(*uint32)(unsafe.Pointer(&i)))
}

func IntValue(i int) Jvalue {
	return Jvalue(*(*uint)(unsafe.Pointer(&i)))
}

func BooleanValue(b bool) Jvalue {
	return Jvalue(cbool(b))
}

func Bool(b uint8) bool {
	return b != 0
}

func CMalloc(capacity int) unsafe.Pointer {
	return C.malloc(C.size_t(capacity))
}

func CFree(p unsafe.Pointer) {
	C.free(p)
}
func GetPid() int {
	return int(C.getpid())
}

func ByteSlicePtr(b []byte) unsafe.Pointer {
	return unsafe.Pointer(unsafe.SliceData(b))
}

func cmem(b []byte) *C.char {
	return (*C.char)(ByteSlicePtr(b))
}

func cbool(b bool) C.jboolean {
	if b {
		return C.JNI_TRUE
	} else {
		return C.JNI_FALSE
	}
}

func cvals(v []Jvalue) *C.jvalue {
	if len(v) == 0 {
		return nil
	}
	return (*C.jvalue)(unsafe.Pointer(unsafe.SliceData(v)))
}

func cBooleanArray(a []bool) *C.jboolean {
	return (*C.jboolean)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cByteArray(a []byte) *C.jbyte {
	return (*C.jbyte)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cShortArray(a []int16) *C.jshort {
	return (*C.jshort)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cCharArray(a []uint16) *C.jchar {
	return (*C.jchar)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cIntArray(a []int32) *C.jint {
	return (*C.jint)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cLongArray(a []int64) *C.jlong {
	return (*C.jlong)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cFloatArray(a []float32) *C.jfloat {
	return (*C.jfloat)(unsafe.Pointer(unsafe.SliceData(a)))
}

func cDoubleArray(a []float64) *C.jdouble {
	return (*C.jdouble)(unsafe.Pointer(unsafe.SliceData(a)))
}

var GlobalVM VM
