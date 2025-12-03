//go:build android
// +build android

package JavaJni

import (
	"net"     // 解析 IP 字符串为 net.IP
	"runtime" // 锁定线程，保证 JNI 在线程内一致
)

// GetWifiAddr 通过 JNI 反射调用 java.net.NetworkInterface 枚举所有内网 IP
// 不依赖 android.permission.ACCESS_WIFI_STATE
func GetWifiAddr() (ips []net.IP) {
	// JNI 要求同一个 JNIEnv 只能在创建它的线程里使用，所以先锁线程
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 如果虚拟机还没初始化，直接返回
	if GlobalVM == 0 {
		return
	}

	// 把当前线程 attach 到 JavaVM，拿到 Env（JNIEnv 封装）
	env, ret := GlobalVM.AttachCurrentThread()
	if ret != JNI_OK {
		// attach 失败也直接返回
		return
	}
	// 函数结束时 detach 当前线程
	defer GlobalVM.DetachCurrentThread()

	// 找到 java.net.NetworkInterface 类
	niClass := env.FindClass("java.net.NetworkInterface")
	if niClass == 0 {
		// 找不到类直接返回
		return
	}

	// 获取静态方法 NetworkInterface.getNetworkInterfaces()Ljava/util/Enumeration;
	getNIMethod := env.GetStaticMethodID(niClass, "getNetworkInterfaces", "()Ljava/util/Enumeration;")
	if getNIMethod == 0 {
		// 方法 ID 获取失败
		return
	}

	// 调用 NetworkInterface.getNetworkInterfaces()，返回一个 Enumeration
	enumObj := env.CallStaticObjectMethodA(niClass, getNIMethod)
	if enumObj == 0 {
		// 返回为 null
		return
	}
	// 用完枚举对象后释放本地引用
	defer env.DeleteLocalRef(enumObj)

	// 拿到 Enumeration 的 Class，用来找 hasMoreElements / nextElement
	enumClass := env.GetObjectClass(enumObj)
	if enumClass == 0 {
		return
	}

	// Enumeration.hasMoreElements()Z
	hasMoreMethod := env.GetMethodID(enumClass, "hasMoreElements", "()Z")
	// Enumeration.nextElement()Ljava/lang/Object;
	nextElementMethod := env.GetMethodID(enumClass, "nextElement", "()Ljava/lang/Object;")
	if hasMoreMethod == 0 || nextElementMethod == 0 {
		return
	}

	// NetworkInterface.getInetAddresses()Ljava/util/Enumeration;
	getInetAddressesMethod := env.GetMethodID(niClass, "getInetAddresses", "()Ljava/util/Enumeration;")
	if getInetAddressesMethod == 0 {
		return
	}

	// InetAddress 相关方法：isSiteLocalAddress / isLoopbackAddress / getHostAddress
	inetClass := env.FindClass("java.net.InetAddress")
	if inetClass == 0 {
		// 理论上不会失败，防御性判断
		return
	}
	isSiteLocalMethod := env.GetMethodID(inetClass, "isSiteLocalAddress", "()Z")                 // 判断是否内网地址
	isLoopbackMethod := env.GetMethodID(inetClass, "isLoopbackAddress", "()Z")                   // 判断是否回环地址
	getHostAddressMethod := env.GetMethodID(inetClass, "getHostAddress", "()Ljava/lang/String;") // 获取 IP 字符串
	if isSiteLocalMethod == 0 || isLoopbackMethod == 0 || getHostAddressMethod == 0 {
		return
	}

	// 外层循环：遍历所有 NetworkInterface
	for {
		// 调用 Enumeration.hasMoreElements()
		hasMore := env.CallBooleanMethodA(enumObj, hasMoreMethod)
		if !hasMore {
			// 没有更多元素，结束外层循环
			break
		}

		// 调用 Enumeration.nextElement()，得到一个 NetworkInterface 对象
		niObj := env.CallObjectMethodA(enumObj, nextElementMethod)
		if niObj == 0 {
			// 理论上不该为 null，防御性判断
			continue
		}

		// 调用 NetworkInterface.getInetAddresses()，拿到一个 InetAddress 的 Enumeration
		addrEnumObj := env.CallObjectMethodA(niObj, getInetAddressesMethod)
		// NetworkInterface 对象用完，释放本地引用
		env.DeleteLocalRef(niObj)
		if addrEnumObj == 0 {
			// 该网卡没有地址，跳过
			continue
		}

		// 拿到 InetAddress 枚举的 Class
		addrEnumClass := env.GetObjectClass(addrEnumObj)
		if addrEnumClass == 0 {
			env.DeleteLocalRef(addrEnumObj)
			continue
		}

		// InetAddress 枚举同样用 hasMoreElements / nextElement 迭代
		addrHasMoreMethod := env.GetMethodID(addrEnumClass, "hasMoreElements", "()Z")
		addrNextElementMethod := env.GetMethodID(addrEnumClass, "nextElement", "()Ljava/lang/Object;")
		if addrHasMoreMethod == 0 || addrNextElementMethod == 0 {
			env.DeleteLocalRef(addrEnumObj)
			continue
		}

		// 内层循环：遍历当前网卡上的所有 InetAddress
		for {
			// hasMoreElements()
			addrHasMore := env.CallBooleanMethodA(addrEnumObj, addrHasMoreMethod)
			if !addrHasMore {
				// 当前网卡没有更多 IP，退出内层循环
				break
			}

			// nextElement() 得到一个 InetAddress 对象
			ipObj := env.CallObjectMethodA(addrEnumObj, addrNextElementMethod)
			if ipObj == 0 {
				// 防御性判断
				continue
			}

			// 调用 InetAddress.isSiteLocalAddress()，判定是否内网地址
			isSiteLocal := env.CallBooleanMethodA(ipObj, isSiteLocalMethod)
			// 调用 InetAddress.isLoopbackAddress()，排除 127.0.0.1 这类回环地址
			isLoopback := env.CallBooleanMethodA(ipObj, isLoopbackMethod)

			if isSiteLocal && !isLoopback {
				// 满足：site-local 且不是 loopback，认为是内网 IP
				// 调用 InetAddress.getHostAddress() 拿到 IP 字符串
				hostStrObj := env.CallObjectMethodA(ipObj, getHostAddressMethod)
				if hostStrObj != 0 {
					// 将 jstring 转成 Go string
					hostStr := env.GetString(Jstring(hostStrObj))
					// 用完 jstring，释放引用
					env.DeleteLocalRef(hostStrObj)

					if hostStr != "" {
						// 解析为 net.IP，成功则加入结果切片
						if parsedIP := net.ParseIP(hostStr); parsedIP != nil {
							ips = append(ips, parsedIP)
						}
					}
				}
			}

			// InetAddress 对象用完，释放本地引用
			env.DeleteLocalRef(ipObj)
		}

		// InetAddress 枚举对象用完，释放本地引用
		env.DeleteLocalRef(addrEnumObj)
	}

	// 把所有收集到的内网 IP 返回
	return
}
