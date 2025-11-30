#include "c_iphlpapi_tcp.h"
#include <windows.h>
#include <tlhelp32.h>
#include <stdio.h>
#include <stdlib.h>
/* 定义一个结构体用于保存进程信息 */
typedef struct ProcessInfo {
	DWORD			pid;
	char			name[MAX_PATH];
	struct ProcessInfo	* next;
} ProcessInfo;
/* 全局链表头 */
ProcessInfo* processListHead = NULL;

/* 初始化进程列表 */
void InitializeProcessList()
{
	HANDLE		hSnapshot;
	PROCESSENTRY32	pe32;

	/* 创建一个进程快照 */
	hSnapshot = CreateToolhelp32Snapshot( TH32CS_SNAPPROCESS, 0 );
	if ( hSnapshot == INVALID_HANDLE_VALUE )
	{
		return;
	}

	/* 初始化 PROCESSENTRY32 结构 */
	pe32.dwSize = sizeof(PROCESSENTRY32);

	/* 枚举进程列表并存入链表 */
	if ( Process32First( hSnapshot, &pe32 ) )
	{
		do
		{
			/* 创建新的链表节点 */
			ProcessInfo* newNode = (ProcessInfo *) malloc( sizeof(ProcessInfo) );
			if ( newNode == NULL )
			{
				CloseHandle( hSnapshot );
				return;
			}

			newNode->pid = pe32.th32ProcessID;
			snprintf( newNode->name, sizeof(newNode->name), "%s", pe32.szExeFile );
			newNode->next	= processListHead;
			processListHead = newNode;
		}
		while ( Process32Next( hSnapshot, &pe32 ) );
	}

	CloseHandle( hSnapshot );
}


/* 根据 PID 查找进程名称 */
char* GetProcessNameByPID( DWORD pid )
{
	ProcessInfo* current = processListHead;

	while ( current )
	{
		if ( current->pid == pid )
		{
			return(current->name);
		}
		current = current->next;
	}

	return(NULL); /* 如果找不到，返回 NULL */
}


/* 清理链表 */
void CleanupProcessList()
{
	ProcessInfo* current = processListHead;
	while ( current )
	{
		ProcessInfo* temp = current;
		current = current->next;
		free( temp );
	}
	processListHead = NULL;
}


/* 定义指向 GetTcpTable2 函数的指针变量 */
GetTcpTable2 pGetTcpTable2;

/* 定义指向 GetExtendedTcpTable 函数的指针变量 */
GETEXTENDEDTABLE pGetExtendedTcpTable;

/* 定义指向 SetTcpEntry 函数的指针变量 */
SETTCPENTRY pSetTcpEntry;

/* 初始化关闭 TCP 连接函数 */
void closeTcpConnectionInit()
{
	/* 加载 iphlpapi.dll 动态库 */
	HMODULE hModule = LoadLibrary( "iphlpapi.dll" );
	if ( hModule == NULL )
	{
		return;
		/* 加载失败，退出函数 */
	}

	/* 获取 GetExtendedTcpTable 函数地址 */
	pGetExtendedTcpTable = (GETEXTENDEDTABLE) GetProcAddress( hModule, "GetExtendedTcpTable" );

	/* 获取 SetTcpEntry 函数地址 */
	pSetTcpEntry = (SETTCPENTRY) GetProcAddress( hModule, "SetTcpEntry" );

	/* 获取 GetTcpTable2 函数地址 */
	pGetTcpTable2 = (GetTcpTable2) GetProcAddress( hModule, "GetTcpTable2" );

	if ( pGetExtendedTcpTable == NULL || pSetTcpEntry == NULL || pGetTcpTable2 == NULL )
	{
		/* 获取失败，释放动态库并退出函数 */
		FreeLibrary( hModule );
		return;
	}
}


/* 将指定的 TCP 连接关闭 */
void closeTcpConnectionByPid( DWORD pid, DWORD ulAf )
{
	/* 如果函数指针变量未初始化，退出函数 */
	if ( pGetExtendedTcpTable == NULL || pSetTcpEntry == NULL )
	{
		return;
	}

	/* 定义指向 TCP 连接表的指针变量 */
	MIB_TCPTABLE_OWNER_PID* tcpTable = NULL;

	/* TCP 连接表的大小 */
	DWORD tcpTableSize = 0;

	/* Windows API 函数调用结果 */
	DWORD result = 0;


	if ( pid == -1 )
	{
		InitializeProcessList();
	}

	/* 获取 TCP 连接列表 */
	result = pGetExtendedTcpTable( NULL, &tcpTableSize, TRUE, ulAf, TCP_TABLE_OWNER_PID_ALL, 0 );
	if ( result == ERROR_INSUFFICIENT_BUFFER )
	{
		tcpTable = (MIB_TCPTABLE_OWNER_PID *) malloc( tcpTableSize );
		/* 分配内存空间 */
		result = pGetExtendedTcpTable( tcpTable, &tcpTableSize, TRUE, ulAf, TCP_TABLE_OWNER_PID_ALL, 0 );
		/* 获取 TCP 连接列表 */
		if ( result == NO_ERROR )
		{
			/* 遍历 TCP 连接列表，查找指定 PID 的连接 */
			for ( DWORD i = 0; i < tcpTable->dwNumEntries; i++ )
			{
				MIB_TCPROW_OWNER_PID* tcpRow = &tcpTable->table[i];
				if ( (pid == -1 && tcpRow->dwState == MIB_TCP_STATE_ESTAB) || (tcpRow->dwOwningPid == pid && tcpRow->dwState == MIB_TCP_STATE_ESTAB) )
				{
					if ( pid == -1 )
					{
						char* name = GetProcessNameByPID( tcpRow->dwOwningPid );
						//之所以排除掉 dlv.exe 因为如果使用goland 动态调试 需要用到网络连接 ,如果这里不排除会导致无法动态调试
						if ( strcmp( name, "dlv.exe" ) == 0 )
						{
							continue;
						}
						//之所以排除掉 msedgewebview2.exe 因为如果使用 webview2 框架开发的程序 可能需要用到网络连接 ,如果这里不排除会导致程序崩溃
						if ( strcmp( name, "msedgewebview2.exe" ) == 0 )
						{
							continue;
						}
						//排除掉 自己进程
						if ( tcpRow->dwOwningPid == GetCurrentProcessId() )
						{
							continue;
						}
					}
					/* 关闭指定的 TCP 连接 */
					MIB_TCPROW tcpRow2;
					tcpRow2.dwState		= MIB_TCP_STATE_DELETE_TCB;
					tcpRow2.dwLocalAddr	= tcpRow->dwLocalAddr;
					tcpRow2.dwLocalPort	= tcpRow->dwLocalPort;
					tcpRow2.dwRemoteAddr	= tcpRow->dwRemoteAddr;
					tcpRow2.dwRemotePort	= tcpRow->dwRemotePort;
					pSetTcpEntry(&tcpRow2);
				}
			}
		}
		free( tcpTable );
		/* 释放内存空间 */
	}
	if ( pid == -1 )
	{
		CleanupProcessList();
	}
}


/* 将网络字节序转换为主机字节序 */
int ntohs2( u_short v )
{
	return( (int) ( (u_short) (v >> 8) | (u_short) (v << 8) ) );
}


/* 获取指定 TCP 地址和端口的 PID */
int getTcpInfoPID( char* Addr, int SunnyProt )
{
	if ( pGetTcpTable2 == NULL )
	{
		return(-1);
	}
	ULONG bufferSize = 0;

	/* Windows API 函数调用结果 */
	DWORD result = pGetTcpTable2( NULL, &bufferSize, TRUE );
	if ( result != ERROR_INSUFFICIENT_BUFFER )
	{
		/* 获取 TCP 连接列表失败，退出函数 */
		return(-2);
	}

	/* 分配内存空间 */
	PMIB_TCPTABLE2 tcpTable = (PMIB_TCPTABLE2) malloc( bufferSize );
	result = pGetTcpTable2( tcpTable, &bufferSize, TRUE );
	if ( result != NO_ERROR )
	{
		free( tcpTable );
		/* 获取 TCP 连接列表失败，释放内存空间后退出函数 */
		return(-3);
	}

	/* 定义缓存字符串 */
	char buf[64];
	/* 遍历 TCP 连接列表，查找指定地址和端口的连接的 PID */
	for ( DWORD i = 0; i < tcpTable->dwNumEntries; i++ )
	{
		sprintf( buf, "%d.%d.%d.%d:%d",
			 (tcpTable->table[i].dwLocalAddr >> 0) & 0xff,
			 (tcpTable->table[i].dwLocalAddr >> 8) & 0xff,
			 (tcpTable->table[i].dwLocalAddr >> 16) & 0xff,
			 (tcpTable->table[i].dwLocalAddr >> 24) & 0xff,
			 ntohs2( (u_short) tcpTable->table[i].dwLocalPort ) );
		int cmpResult = strcmp( buf, Addr );
		if ( cmpResult == 0 )
		{
			/* 找到指定连接，返回 PID */
			int r = (int) (tcpTable->table[i].dwOwningPid);
			free( tcpTable );
			return(r);
		}
	}

	free( tcpTable );
	/* 没有找到指定连接，释放内存空间后返回错误代码 */
	return(-4);
}
/* IsPortListening 判断指定 TCP 端口是否在当前机器上处于 LISTEN 状态
 * 返回 1 表示有 LISTEN 套接字
 * 返回 0 表示未监听或查询失败
 */
int IsPortListening(int port)
{
	/* 如果 GetTcpTable2 函数指针没有初始化，直接认为未监听 */
	if (pGetTcpTable2 == NULL) {
		return 0;	// 函数未就绪，返回未监听
	}

	DWORD bufferSize = 0;	// 保存所需缓冲区大小
	DWORD result;		// Windows API 返回值

	/* 第一次调用只为了获取所需缓冲区大小 */
	result = pGetTcpTable2(NULL, &bufferSize, TRUE);	// TRUE 表示按地址排序
	if (result != ERROR_INSUFFICIENT_BUFFER) {
		// 如果不是缓冲区不足错误，说明调用失败，直接返回未监听
		return 0;
	}

	/* 分配保存 TCP 表的内存 */
	PMIB_TCPTABLE2 tcpTable = (PMIB_TCPTABLE2)malloc(bufferSize);	// 为 TCP 表分配内存
	if (tcpTable == NULL) {
		// 分配失败，直接返回未监听
		return 0;
	}

	/* 真正获取 TCP 连接表 */
	result = pGetTcpTable2(tcpTable, &bufferSize, TRUE);	// 再次调用获取数据
	if (result != NO_ERROR) {
		// 获取失败，释放内存后返回未监听
		free(tcpTable);	// 释放内存避免泄漏
		return 0;
	}

	int listening = 0;	// 标志位：是否找到 LISTEN 记录

	/* 遍历所有 TCP 条目 */
	for (DWORD i = 0; i < tcpTable->dwNumEntries; i++) {
		MIB_TCPROW2 *row = &tcpTable->table[i];	// 当前行指针

		/* 将网络字节序的端口转换为主机字节序 */
		int localPort = ntohs2((u_short)row->dwLocalPort);	// 提取本地端口

		/* 判断是否是目标端口且处于 LISTEN 状态 */
		if (localPort == port && row->dwState == MIB_TCP_STATE_LISTEN) {
			// 找到至少一个处于 LISTEN 状态的套接字
			listening = 1;	// 标记为已监听
			break;		// 可以提前结束循环
		}
	}

	/* 用完 TCP 表需要释放内存 */
	free(tcpTable);	// 释放 TCP 表内存

	/* 返回是否监听的结果 */
	return listening;	// 1：监听中，0：未监听或失败
}
