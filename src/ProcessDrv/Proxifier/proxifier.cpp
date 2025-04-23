#include "Proxifier.hpp"

#include <iostream>
#include <string>
#include <cstring>
#include <locale>
#include <codecvt>
#include <iomanip>

#pragma comment(lib, "ws2_32.lib") // 链接 Winsock 库
#define BUFFER_SIZE 0x534          // 1324 bytes
#define RESPONSE_SIZE 0x3FC        // 1020 bytes

int ____pid = 0;
extern "C"
{
    void Call(HANDLE hPipe, char *raw);
}
SECURITY_DESCRIPTOR Attributes;
SECURITY_ATTRIBUTES SecurityAttributes;
static HANDLE WAIT_init_Descriptor()
{
    InitializeSecurityDescriptor(&Attributes, SECURITY_DESCRIPTOR_REVISION);
    SetSecurityDescriptorDacl(&Attributes, TRUE, NULL, FALSE);
    SecurityAttributes.nLength = sizeof(SECURITY_ATTRIBUTES);
    SecurityAttributes.lpSecurityDescriptor = &Attributes; // 使用自定义安全描述符
    SecurityAttributes.bInheritHandle = FALSE;             // 不允许继承句柄
    return 0;
}

HANDLE hMStop = WAIT_init_Descriptor();

static void WAIT()
{
    // 创建命名管道
    HANDLE hPipe = CreateNamedPipeW(
        L"\\\\.\\pipe\\proxifier", // 管道名称
        PIPE_ACCESS_DUPLEX,        // 双向管道
        6,                         // 字节流模式
        1,                         // 最大实例数
        0,                         // 输出缓冲区大小
        0,                         // 输入缓冲区大小
        0,                         // 默认超时
        &SecurityAttributes);      // 默认安全属性

    if (hPipe == INVALID_HANDLE_VALUE)
    {
        return;
    }

    // 等待客户端连接
    BOOL connected = ConnectNamedPipe(hPipe, NULL);
    if (!connected)
    {
        DWORD error = GetLastError();
        if (error != ERROR_PIPE_CONNECTED)
        {
            CloseHandle(hPipe);
            return;
        }
    }
    // 处理数据
    char buffer[BUFFER_SIZE] = {0};
    DWORD bytesRead = 0;
    if (ReadFile(hPipe, buffer, BUFFER_SIZE, &bytesRead, NULL))
    {
        int receivedValue = *reinterpret_cast<int *>(buffer);
        if (receivedValue == bytesRead)
        {
            Call(hPipe, buffer);
        }
    }
    CloseHandle(hPipe);
    return;
}
void ProxifierWriteFile(HANDLE hPipe, char *lpBuffer, DWORD len)
{
    DWORD bytesWritten = 0;
    WriteFile(hPipe, lpBuffer, len, &bytesWritten, NULL);
    return;
}

// 始终尝试占用这个锁
void ProxifierCreateMutex()
{
    // 尝试创建获取这个锁
    HANDLE hProxifier = CreateMutexW(0, 0, L"Global\\ProxifierStd300Mutex");
    if (hProxifier == NULL)
    {
        return;
    }
    // 判断是获取成功还是创建成功
    if (GetLastError() == ERROR_ALREADY_EXISTS)
    {
        // 如果是被其他进程创建的,则释放句柄
        ReleaseMutex(hProxifier);
        CloseHandle(hProxifier);
        return;
    }
    // 如果是创建成功,则丢弃句柄,不管了,让锁始终处于占用状态
    return;
}
HANDLE hMutexProxifier = 0;
// ProxifierStd300Mutex
// Proxifier32Mutex1040

int StartProxifier()
{
    ProxifierCreateMutex();
    if (hMStop != 0)
    {
        return 1;
    }
    if (hMutexProxifier != 0)
    {
        hMStop = hMutexProxifier;
        return 1;
    }
    hMutexProxifier = CreateMutexW(NULL, FALSE, L"Global\\Proxifier32Mutex1040");
    if (hMutexProxifier == NULL)
    {
        hMutexProxifier = 0;
        hMStop = 0;
        return 0;
    }
    if (GetLastError() == ERROR_ALREADY_EXISTS)
    {
        hMutexProxifier = 0;
        hMStop = 0;
        return 0;
    }
    hMStop = hMutexProxifier;
    return 1;
}
int StopProxifier()
{
    if (hMStop != 0)
    {
        // ReleaseMutex(hMutexProxifier);
        // CloseHandle(hMutexProxifier);
        hMStop = 0;
        return 1;
    }
    return 0;
}

void ProxifierInit(int myPid)
{
    ____pid = myPid;
    while (true)
    {
        ProxifierCreateMutex();
        if (hMStop == 0)
        {
            Sleep(200);
            continue;
        }
        WAIT();
    }
    return;
}
int ProxifierIsInit()
{
    ProxifierCreateMutex();
    HANDLE hMutex = OpenMutexW(MUTEX_ALL_ACCESS, FALSE, L"Global\\Proxifier32Mutex1040");
    if (hMutex != NULL)
    {
        ReleaseMutex(hMutex);
        CloseHandle(hMutex);
        return 1;
    }
    return 0;
}