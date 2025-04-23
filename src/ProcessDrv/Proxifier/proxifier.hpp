

#ifndef PROXIFIER_HPP
#define PROXIFIER_HPP
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#ifdef __cplusplus
extern "C" {
#endif
    void ProxifierWriteFile(HANDLE hFile, char* lpBuffer,DWORD len);
    void ProxifierInit(int myPid);
    int ProxifierIsInit();
    int StartProxifier();
    int StartProxifier();
    int StopProxifier ();

#ifdef __cplusplus
}
#endif

#endif // PROXIFIER_HPP