#include <windows.h>
#include <iphlpapi.h>
#include <stdio.h>

// 定义 MIB_TCPROW2 结构体
typedef struct _MIB_TCPROW2 {
	DWORD dwState;
	DWORD dwLocalAddr;
	DWORD dwLocalPort;
	DWORD dwRemoteAddr;
	DWORD dwRemotePort;
	DWORD dwOwningPid;
	DWORD dwOffloadState;
} MIB_TCPROW2, *PMIB_TCPROW2;

// 定义 MIB_TCPTABLE2 结构体
typedef struct _MIB_TCPTABLE2 {
	DWORD dwNumEntries;
	MIB_TCPROW2 table[ANY_SIZE];
} MIB_TCPTABLE2, *PMIB_TCPTABLE2;

// 定义 GETEXTENDEDTABLE 函数指针类型
typedef DWORD(WINAPI* GETEXTENDEDTABLE)(PVOID, PDWORD, BOOL, ULONG, TCP_TABLE_CLASS, ULONG);

// 定义 SETTCPENTRY 函数指针类型
typedef DWORD(WINAPI* SETTCPENTRY)(PMIB_TCPROW);

// 定义 GetTcpTable2 函数指针类型
typedef DWORD (WINAPI * GetTcpTable2)(PMIB_TCPTABLE2 TcpTable, PULONG SizePointer, BOOL Order);

// 关闭 TCP 连接初始化
void closeTcpConnectionInit();

// 根据 PID 关闭 TCP 连接
void closeTcpConnectionByPid(DWORD pid, DWORD ulAf);

// 获取指定 TCP 地址和端口的 PID
int getTcpInfoPID(char* Addr, int SunnyProt);

/* IsPortListening 判断指定 TCP 端口是否在当前机器上处于 LISTEN 状态
 * 返回 1 表示有 LISTEN 套接字
 * 返回 0 表示未监听或查询失败
 */
int IsPortListening(int port);