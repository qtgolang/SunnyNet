export const builtCmdWords = [
    {
        name: ['Log', 'log', 'print', "Println", "日志输出", "打印日志"],
        insertText: 'Log(${1:str}$0)',
        detail: "打印日志",
        contents: [
            {value: '**打印日志**'},
            {value: '打印日志,触发日志回调,传递到软件中 (无返回值)'},
            {value: '**示例代码**'},
            {value: '```go\nLog("Hello","SunnyNet",2024)\n```'}
        ]
    },
    {
        name: ['Sprintf', '格式化字符串', '格式化文本'],
        insertText: 'fmt.Sprintf(${1:format}$0,${2:value})',
        detail: "格式化字符串,支持任意参数",
        contents: [
            {value: '**格式化字符串**'},
            {value: '用于格式化字符串：根据格式说明符格式化并返回结果字符串。 (无返回值)'},
            {value: '**示例代码**'},
            {value: '```go\nconst name, age = "Kim", 22\ns := fmt.Sprintf("%s is %d years old.", name, age)\n```'}
        ]
    },
    {
        name: ['GetPidName', 'PID获取进程名', 'pid获取进程名'],
        insertText: 'GetPidName(Conn.PID())$0',
        detail: "获取指定PID对应的进程名称",
        contents: [
            {value: '**获取指定PID对应的进程名称 (返回值:字符串)**'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nGetPidName(Conn.PID())\n```'}
        ]
    },
    {
        name: ['计次循环', 'for'],
        insertText: 'for i := 0; i < ${1:10}; i++ {\n\t$0\n}',
        detail: "计次循环",
        contents: [
            {value: '**计次循环**'},
            {value: '**示例代码**'},
            {value: '```go\nfor i := 0; i < 10; i++ {\n\tLog(i)\n}\n```'}
        ]
    },
    {
        name: ['跳出循环', 'break'],
        insertText: 'break$0',
        detail: "跳出循环",
        contents: [
            {value: '**跳出当前循环**'},
            {value: '**示例代码**'},
            {value: '```go\nfor i := 0; i < 10; i++ {\n\tLog(i)\n\tbreak\n}\n```'}
        ]
    },
    {
        name: ['到循环尾', 'continue'],
        insertText: 'continue$0',
        detail: "到循环尾",
        contents: [
            {value: '**跳处当前循环**'},
            {value: '**示例代码**'},
            {value: '```go\nfor i := 0; i < 10; i++ {\n\tLog(i)\n\tcontinue\n\tLog("--",i)\n}\n```'}
        ]
    },
    {
        name: ['如果', 'if'],
        insertText: '\tif ${1:Conn.Type() == 1}$0 {\n\t\t\n\t}else{\n\n\t}',
        detail: "如果",
        contents: [
            {value: '**如果**'},
            {value: '**示例代码**'},
            {value: '```go\nif 1 < 2 {\n\tlog("1<2")\n}else{\n\tLog(1>2)\n}\n```'}
        ]
    },
    {
        name: ['如果真'],
        insertText: '\tif ${1:Conn.Type() == 1}$0 {\n\t\t\n\t}',
        detail: "如果真",
        contents: [
            {value: '**如果真**'},
            {value: '**示例代码**'},
            {value: '```go\nif 1 < 2 {\n\tlog("1<2")\n}\n```'}
        ]
    },
    {
        name: ['否则', 'else'],
        insertText: 'else{\n$0\n}',
        detail: "否则",
        contents: [
            {value: '**否则**'},
            {value: '**示例代码**'},
            {value: '```go\nif 1 < 2 {\n\tlog("1<2")\n}else{\n\tLog(1>2)\n}\n```'}
        ]
    },
    {
        name: ['声明变量'],
        insertText: 'value :=${1:"format"}$0',
        detail: "声明变量",
        contents: [
            {value: '**声明变量**'},
            {value: '自动根据值类型,声明对应变量类型'},
            {value: '**示例代码**'},
            {value: '```go\n//字符串类型变量\nvalue := "format"\n//int类型变量\nvalue := 123\n//int64类型变量\nvalue := int64(123)\n```'}
        ]
    },
    {
        name: ['启动协程', "go"],
        insertText: 'go func(){\n\t$0\n}()',
        detail: "启动协程",
        contents: [
            {value: '**启动协程**'},
            {value: '启动协程,可以理解为启动线程'},
            {value: '**示例代码**'},
            {value: '```go\ngo func(){\n\tLog(""协程中执行的代码)\n}()\n```'}
        ]
    },
    {
        name: ['GoHexEncode', 'HexEncode', '字符串到十六进制', '字符串转十六进制', '文本到十六进制', '字节集到十六进制', 'bytes到十六进制'],
        insertText: 'GoHexEncode(${1:bs}$0)',
        detail: "字符串、字节集到十六进制",
        contents: [
            {value: '**字符串、字节集到十六进制**'},
            {value: '可以传入字符串或字节集,返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nGoHexEncode("123456")\n```'},
            {value: '```go\nGoHexEncode([]byte("123456")\n```'}
        ]
    },
    {
        name: ['GoHexDecode', 'HexDecode', '十六进制到字节集', '十六进制转字节集'],
        insertText: 'GoHexDecode(${1:hexStr}$0)',
        detail: "十六进制到字节集",
        contents: [
            {value: '**十六进制到字节集**'},
            {value: '将给定的十六进制编码解码为字节集'},
            {value: '需传入参数类型字符串,返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nGoHexDecode("123456")\n```'}
        ]
    },
    {
        name: ['Base64编码', 'GoBase64Encode'],
        insertText: 'GoBase64Encode(${1:bs}$0)',
        detail: "Base64编码",
        contents: [
            {value: '**Base64编码**'},
            {value: '将给定的字符串或字节集编码为Base64字符串'},
            {value: '可以传入字符串或字节集,返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nGoBase64Encode("123456")\n```'},
            {value: '```go\nGoBase64Encode([]byte("123456"))\n```'}
        ]
    },
    {
        name: ['Base64解码', 'GoBase64Decode'],
        insertText: 'GoBase64Decode(${1:bs}$0)',
        detail: "Base64解码",
        contents: [
            {value: '**Base64编码**'},
            {value: '将给定的Base64字符串解码为字节集'},
            {value: '可以传入字符串或字节集,返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nGoBase64Decode("MTIzNA==")\n```'},
            {value: '```go\nGoBase64Decode([]byte("MTIzNA=="))\n```'}
        ]
    },
    {
        name: ['Base64解码到十六进制', 'Base64ToHex'],
        insertText: 'Base64ToHex(${1:bs}$0)',
        detail: "Base64解码到十六进制",
        contents: [
            {value: '**Base64解码到十六进制**'},
            {value: '将给定的Base64字符串解码为十六进制'},
            {value: '可以传入字符串或字节集,返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nBase64ToHex("MTIzNA==")\n```'},
            {value: '```go\nBase64ToHex([]byte("MTIzNA=="))\n```'}
        ]
    },
    {
        name: ['十六进制到解码Base64', 'HexToBase64'],
        insertText: 'HexToBase64(${1:str}$0)',
        detail: "十六进制到解码Base64",
        contents: [
            {value: '**十六进制到解码Base64**'},
            {value: '将给定的十六进制字符串解码为Base64字符串'},
            {value: '可以传入字符串,返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nHexToBase64("323232")\n```'},
        ]
    },
    {
        name: ['响应请求404', '返回404', '拦截响应404', '修改响应为404', "HTTPResponse404"],
        insertText: 'HTTPResponse404(Conn)$0',
        detail: "修改响应为404",
        contents: [
            {value: '**修改响应为404**'},
            {value: '直接响应请求 状态码404,没有内容 返回值:无'},
            {value: '仅支持在HTTP请求回调函数中使用！'},
            {value: '**示例代码**'},
            {value: '```go\nHTTPResponse404(Conn)\n```'},
        ]
    },
    {
        name: ['响应请求200空Json', '返回200空Json', '拦截响应200空Json', '修改响应为200空Json', "HTTPResponse200JSon"],
        insertText: 'HTTPResponse200JSon(Conn)$0',
        detail: "修改响应为200Json",
        contents: [
            {value: '**修改响应为200Json**'},
            {value: '直接响应请求 状态码200,响应空的JSON对象 返回值:无'},
            {value: '仅支持在HTTP请求回调函数中使用！'},
            {value: '**示例代码**'},
            {value: '```go\nHTTPResponse200JSon(Conn)\n```'},
        ]
    },
    {
        name: ['响应请求200空数组', '返回200空数组', '拦截响应200空数组', '修改响应为200空数组', "HTTPResponse200Array"],
        insertText: 'HTTPResponse200Array(Conn)$0',
        detail: "修改响应为200空数组",
        contents: [
            {value: '**修改响应为200空数组**'},
            {value: '直接响应请求 状态码200,响应空的JSON数组 返回值:无'},
            {value: '仅支持在HTTP请求回调函数中使用！'},
            {value: '**示例代码**'},
            {value: '```go\nHTTPResponse200Array(Conn)\n```'},
        ]
    },
    {
        name: ['响应请求200空内容', '返回200空内容', '拦截响应200空内容', '修改响应为200空内容', "HTTPResponse200"],
        insertText: 'HTTPResponse200(Conn)$0',
        detail: "修改响应为200空内容",
        contents: [
            {value: '**修改响应为200空内容**'},
            {value: '直接响应请求 状态码200,没有内容 返回值:无'},
            {value: '仅支持在HTTP请求回调函数中使用！'},
            {value: '**示例代码**'},
            {value: '```go\nHTTPResponse200(Conn)\n```'},
        ]
    },
    {
        name: ['响应请求200图片', '返回200图片', '拦截响应200图片', '修改响应为200图片', "HTTPResponse200IMG"],
        insertText: 'HTTPResponse200IMG(Conn)$0',
        detail: "修改响应为200图片",
        contents: [
            {value: '**修改响应为200图片**'},
            {value: '直接响应请求 状态码200,内容为1像素的图片 返回值:无'},
            {value: '仅支持在HTTP请求回调函数中使用！'},
            {value: '**示例代码**'},
            {value: '```go\nHTTPResponse200IMG(Conn)\n```'},
        ]
    },
    {
        name: ['取数据摘要', '取数据MD5', "GoMD5"],
        insertText: 'GoMD5(${1:value}$0)',
        detail: "取数据MD5",
        contents: [
            {value: '**取数据MD5**'},
            {value: '可以传入字符串或字节集,返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nGoMD5("123456")\nGoMD5([]byte("123456"))\nGoMD5(Conn.GetRequestBody())\n```'},
            {value: '**如果需要HMAC-SHA1**'},
            {value: '```go\nGoMD5("123456","key")\nGoMD5([]byte("123456","key"))\nGoMD5(Conn.GetRequestBody(),"key")\n```'},
        ]
    },
    {
        name: ['取数据SHA1', "GoSHA1"],
        insertText: 'GoSHA1(${1:value}$0)',
        detail: "取数据SHA1",
        contents: [
            {value: '**取数据SHA1**'},
            {value: '可以传入字符串或字节集,返回值:字节集\u3000\u3000'},
            {value: '**示例代码**'},
            {value: '```go\nGoSHA1("123456")\nGoSHA1([]byte("123456"))\nGoSHA1(Conn.GetRequestBody())\n```'},
            {value: '**如果需要HMAC-SHA1**'},
            {value: '```go\nGoSHA1("123456","key")\nGoSHA1([]byte("123456","key"))\nGoSHA1(Conn.GetRequestBody(),"key")\n```'},
        ]
    },
    {
        name: ['取数据SHA256', "GoSHA256"],
        insertText: 'GoSHA256(${1:value}$0)',
        detail: "取数据SHA256",
        contents: [
            {value: '**取数据SHA256**'},
            {value: '可以传入字符串或字节集,返回值:字节集\u3000\u3000'},
            {value: '**示例代码**'},
            {value: '```go\nGoSHA256("123456")\nGoSHA256([]byte("123456"))\nGoSHA256(Conn.GetRequestBody())\n```'},
            {value: '**如果需要HMAC-SHA1**'},
            {value: '```go\nGoSHA256("123456","key")\nGoSHA256([]byte("123456","key"))\nGoSHA256(Conn.GetRequestBody(),"key")\n```'},
        ]
    },
    {
        name: ['取数据SHA512', "GoSHA512"],
        insertText: 'GoSHA512(${1:value}$0)',
        detail: "取数据SHA512",
        contents: [
            {value: '**取数据SHA512**'},
            {value: '可以传入字符串或字节集,返回值:字节集\u3000\u3000'},
            {value: '**示例代码**'},
            {value: '```go\nGoSHA512("123456")\nGoSHA512([]byte("123456"))\nGoSHA512(Conn.GetRequestBody())\n```'},
            {value: '**如果需要HMAC-SHA512**'},
            {value: '```go\nGoSHA512("123456","key")\nGoSHA512([]byte("123456","key"))\nGoSHA512(Conn.GetRequestBody(),"key")\n```'},
        ]
    },
    {
        name: ['RSA私钥解密', 'GoRsaPrivateDecrypt'],
        insertText: 'GoRsaPrivateDecrypt(${1:key}$0,${2:cipher})',
        detail: "RSA私钥解密",
        contents: [
            {value: '**RSA私钥解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    字符串  [PEM格式Base64字符串]'},
            {value: '参数2:cipher 字节集  [要解密的数据]'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`-----BEGIN RSA PRIVATE KEY-----\nMIIUHJcGVydGllcw..........-----END RSA PRIVATE KEY-----\n`\ndata,err := GoRsaPrivateDecrypt(key,Conn.GetRequestBody())\nif err!=nil {\n\tLog("RSA解密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['RSA公钥加密', 'GoRsaPublicEncrypt'],
        insertText: 'GoRsaPublicEncrypt(${1:key}$0,${2:cipher})',
        detail: "RSA公钥加密",
        contents: [
            {value: '**RSA公钥加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    字符串  [PEM格式Base64字符串]'},
            {value: '参数2:cipher 字节集  [要解密的数据]'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`-----BEGIN RSA PUBLIC KEY-----\nMIGUHJcGVydGllcw..........-----END RSA PUBLIC KEY-----\n`\ndata,err := GoRsaPublicEncrypt(key,Conn.GetRequestBody())\nif err!=nil {\n\tLog("RSA加密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['GoAESCBCEncode', 'AES_CBC_加密'],
        insertText: 'GoAESCBCEncode(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "AES CBC 加密",
        contents: [
            {value: '**AES CBC 加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:i v    可以是字符串或字节集,长度需16'},
            {value: '参数3:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数4:cipher  要加密的内容,可以是字符串或字节集'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`1234567890123456`\niv:=`6543210123456789`\ndata,err := GoAESCBCEncode(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("AES CBC 加密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['GoDESCBCEncode', 'DES_CBC_加密'],
        insertText: 'GoDESCBCEncode(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "DES CBC 加密",
        contents: [
            {value: '**DES CBC 加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:i v    可以是字符串或字节集,长度需8'},
            {value: '参数3:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数4:cipher  要加密的内容,可以是字符串或字节集'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`123456789012345678901234`\niv:=`90123456`\ndata,err := GoDESCBCEncode(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("DES CBC 加密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['Go3DESCBCEncode', 'DESEDE_CBC_加密', '3DES_CBC_加密'],
        insertText: 'Go3DESCBCEncode(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "3DES CBC 加密",
        contents: [
            {value: '**3DES CBC 加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:i v    可以是字符串或字节集,长度需8'},
            {value: '参数3:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数4:cipher  要加密的内容,可以是字符串或字节集'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`123456789012345678901234`\niv:=`90123456`\ndata,err := Go3DESCBCEncode(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("3DES CBC 加密错误",err)\n}\n```'},
        ]
    },

    {
        name: ['GoAESCBCDecrypt', 'AES_CBC_解密'],
        insertText: 'GoAESCBCDecrypt(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "AES CBC 解密",
        contents: [
            {value: '**AES CBC 解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:i v    可以是字符串或字节集,长度需16'},
            {value: '参数3:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数4:cipher  要解密的内容,需传入字节集类型'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`1234567890123456`\niv:=`6543210123456789`\ndata,err := GoAESCBCDecrypt(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("AES CBC 解密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['GoDESCBCDecrypt', 'DES_CBC_解密'],
        insertText: 'GoDESCBCDecrypt(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "DES CBC 解密",
        contents: [
            {value: '**DES CBC 解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:i v    可以是字符串或字节集,长度需8'},
            {value: '参数3:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数4:cipher  要解密的内容,需传入字节集类型'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`11223344`\niv:=`90123456`\ndata,err := GoDESCBCDecrypt(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("DES CBC 解密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['Go3DESCBCDecrypt', 'DESEDE_CBC_解密', '3DES_CBC_解密'],
        insertText: 'Go3DESCBCDecrypt(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "3DES CBC 解密",
        contents: [
            {value: '**3DES CBC 解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:i v    可以是字符串或字节集,长度需8'},
            {value: '参数3:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数4:cipher  要解密的内容,需传入字节集类型'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`123456789012345678901234`\niv:=`90123456`\ndata,err := Go3DESCBCDecrypt(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("3DES CBC 解密错误",err)\n}\n```'},
        ]
    },

    {
        name: ['GoAESECBEncode', 'AES_ECB_加密'],
        insertText: 'GoAESECBEncode(${1:key}$0,"PKCS7",cipher)',
        detail: "AES ECB 加密",
        contents: [
            {value: '**AES ECB 加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数3:cipher  要加密的内容,可以是字符串或字节集'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`1234567890123456`\ndata,err := GoAESECBEncode(key,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("AES ECB 加密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['GoDESECBEncode', 'DES_ECB_加密'],
        insertText: 'GoDESECBEncode(${1:key}$0,"PKCS7",cipher)',
        detail: "DES ECB 加密",
        contents: [
            {value: '**DES ECB 加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数3:cipher  要加密的内容,可以是字符串或字节集'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`12345678`\ndata,err := GoDESECBEncode(key,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("DES ECB 加密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['Go3DESECBEncode', '3DES_ECB_加密', 'DESEDE_ECB_加密'],
        insertText: 'Go3DESECBEncode(${1:key}$0,"PKCS7",cipher)',
        detail: "3DES ECB 加密",
        contents: [
            {value: '**3DES ECB 加密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数3:cipher  要加密的内容,可以是字符串或字节集'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`12345678`\ndata,err := Go3DESECBEncode(key,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("3DES ECB 加密错误",err)\n}\n```'},
        ]
    },

    {
        name: ['GoAESECBDecrypt', 'AES_ECB_解密'],
        insertText: 'GoAESECBDecrypt(${1:key}$0,"PKCS7",cipher)',
        detail: "AES ECB 解密",
        contents: [
            {value: '**AES ECB 解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数3:cipher  要解密的内容,需传入字节集类型'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`1234567890123456`\niv:=`6543210123456789`\ndata,err := GoAESECBDecrypt(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("AES ECB 解密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['GoDESECBDecrypt', 'DES_ECB_解密'],
        insertText: 'GoDESECBDecrypt(${1:key}$0,${2:iv},"PKCS7",cipher)',
        detail: "DES ECB 解密",
        contents: [
            {value: '**DES ECB 解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数3:cipher  要解密的内容,需传入字节集类型'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`11223344`\niv:=`90123456`\ndata,err := GoDESECBDecrypt(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("DES ECB 解密错误",err)\n}\n```'},
        ]
    },
    {
        name: ['Go3DESECBDecrypt', 'DESEDE_ECB_解密', '3DES_ECB_解密'],
        insertText: 'Go3DESECBDecrypt(${1:key}$0,"PKCS7",cipher)',
        detail: "3DES ECB 解密",
        contents: [
            {value: '**3DES ECB 解密**'},
            {value: '**参数说明**'},
            {value: '参数1:key    可以是字符串或字节集,根据key长度自动选择128/192/256'},
            {value: '参数2:Padding    需要字符串,NOPAD / PKCS5 / PKCS7 / ISO97971 / ANSIX923 / ISO10126 / ZERO\u3000\u3000'},
            {value: '参数3:cipher  要解密的内容,需传入字节集类型'},
            {value: '两个返回值:字节集,error'},
            {value: '**示例代码**'},
            {value: '```go\nkey:=`123456789012345678901234`\niv:=`90123456`\ndata,err := Go3DESECBDecrypt(key,iv,"PKCS7",Conn.GetRequestBody())\nif err!=nil {\n\tLog("3DES ECB 解密错误",err)\n}\n```'},
        ]
    },

    {
        name: ['网页访问对象', 'sendHTTPRequest', 'GoHTTPRequest', "发送HTTP请求"],
        insertText: 'GoHTTPRequest(${1:method}$0,${2:url},${2:header})',
        detail: "发送HTTP请求",
        contents: [
            {value: '**发送HTTP请求**'},
            {value: '**参数说明**'},
            {value: '参数1:method 字符串  [请求方式 GET / POST / PUT 等...]'},
            {value: '参数2:url 字节集  [要请求的地址]'},
            {value: '参数3:header 字节集  [请求时要携带的协议头]'},
            {value: '三个返回值:字节集,协议头,error'},
            {value: '**示例代码**'},
            {value: '```go\nmethod := "GET"\nurl:="https://www.baidu.com"\n//初始化协议头对象\nheader:=make(Header)\n//设置协议头方式1\nheader.Set("User-Agent","Mozilla/5.0 AppleWebKit/537.36 Chrome/129.0.0.0 Safari/537.36")\n//设置协议头方式2-该方式可以设置多个同名协议头\nheader["Token"]=[]string{"123","456"}\nbody,hr,err := GoHTTPRequest(method,url,header)\nif err!=nil {\n\tLog("脚本代码发送HTTP请求失败:错误信息:",err,"URL:"+url)\n} else {\n\t//body为请求得到的结果(字节集类型)\n\t//hr为服务器返回的协议头\n\t//err 为是否请求失败,如果请求成功等于nil\n}\n```'},
        ]
    },

    {
        name: ['DelSpace', '删除所有空格'],
        insertText: 'strings.ReplaceAll(strings.ReplaceAll(${1:str}$0, " ", ""), "\u3000", "")',
        detail: "删除所有空格",
        contents: [
            {value: '**删除所有空格**'},
            {value: '**参数说明**'},
            {value: '参数1:str    字符串  [待删除所有空格的字符串]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nstrings.ReplaceAll(strings.ReplaceAll(str, " ", ""), "\u3000", "")\n```'},
        ]
    },
    {
        name: ['TrimSpace', '删除首尾空格'],
        insertText: 'strings.TrimSpace(${1:str}$0)',
        detail: "删除首尾空格",
        contents: [
            {value: '**删除首尾空格**'},
            {value: '**参数说明**'},
            {value: '参数1:str    字符串  [待删除首尾空格的字符串]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nstrings.TrimSpace(str)\n```'},
        ]
    },
    {
        name: ['ToUpper', '字符串到大写', "到大写"],
        insertText: 'strings.ToUpper(${1:str}$0)',
        detail: "字符串到大写",
        contents: [
            {value: '**字符串到大写**'},
            {value: '**参数说明**'},
            {value: '参数1:str    字符串  [待到大写的字符串]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nstrings.ToUpper(str)\n```'},
        ]
    },
    {
        name: ['ToLower', '字符串到小写', "到小写"],
        insertText: 'strings.ToLower(${1:str}$0)',
        detail: "字符串到小写",
        contents: [
            {value: '**字符串到大写**'},
            {value: '**参数说明**'},
            {value: '参数1:str    字符串  [待到小写的字符串]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nstrings.ToLower(str)\n```'},
        ]
    },
    {
        name: ['BytesReplace', "ReplaceAll", '字节集替换', "替换字节集"],
        insertText: 'BytesReplace(${1:bs}$0, ${2:old}, ${3:new})',
        detail: "替换字节集",
        contents: [
            {value: '**替换字节集**'},
            {value: '**参数说明**'},
            {value: '参数1:bs    字节集  [原始字节集]\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000'},
            {value: '参数2:old    字节集  [要替换的字节集]'},
            {value: '参数3:new    字节集  [替换为的字节集]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nBytesReplace([]byte("123456654321"), []byte("5665"), []byte("---"))\n//将返回1234---4321的字节集\n```'},
        ]
    },
    {
        name: ['StringReplace', '字符串替换', "替换字符串"],
        insertText: 'StringReplace(${1:str}$0, ${2:old}, ${3:new})',
        detail: "替换字符串",
        contents: [
            {value: '**替换字符串**'},
            {value: '**参数说明**'},
            {value: '参数1:bs    字节集  [原始字符串]\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000'},
            {value: '参数2:old    字节集  [要替换的字符串]'},
            {value: '参数3:new    字节集  [替换为的字符串]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nStringReplace("123456654321","5665", "---")\n//将返回1234---4321\n```'},
        ]
    },
    {
        name: ['Contains', '是否包含字符串', "是否包含字节集"],
        insertText: 'Contains(${1:s1}$0, ${2:s2})',
        detail: "是否包含 字符串/字节集",
        contents: [
            {value: '**是否包含 字符串/字节集**'},
            {value: '**参数说明**'},
            {value: '参数1:s1     字符串/字节集  [原始字符串]'},
            {value: '参数2:s2     字符串/字节集  [原始字符串]\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000'},
            {value: '返回值:bool'},
            {value: '**示例代码**'},
            {value: '```go\nContains("123456654321","5665")\n//将返回true\nContains([]byte("123456654321"),[]byte("56165")\n//将返回true\n```'},
        ]
    },
    {
        name: ['取数组长度', '取长度', '取Map长度', "取Map数量", "取map数量", "len", "取字节集长度", "取协议头数量", "取字符串长度"],
        insertText: 'len(${1:array}$0)',
        detail: "取字符串长度/字节集长度/数组长度/MAP数量/协议头数量",
        contents: [
            {value: '**取字符串长度/字节集长度/数组长度/MAP数量/协议头数量**'},
            {value: '**参数说明**'},
            {value: '参数1:array     任意类型  [待获取的数组/map/Header]\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000\u3000'},
            {value: '返回值:int'},
            {value: '**示例代码**'},
            {value: '```go\nlen("123")\n//将返回 3\nlen([]byte("1230"))\n//将返回4\nlen(Conn.GetRequestHeader())\n```'},
        ]
    },
    {
        name: ['toBytes', '到字节集', '字符串到字节集', "文本到字节集"],
        insertText: '[]byte(${1:str}$0)',
        detail: "字符串到字节集",
        contents: [
            {value: '**字符串到字节集**'},
            {value: '**参数说明**'},
            {value: '参数1:str     字符串类型  [待转换的字符串]'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nbs := []byte("11111")\n```'},
        ]
    },
    {
        name: ['BytesToString', '到字符串', '字节集到字符串'],
        insertText: 'BytesToString(${1:bs}$0)',
        detail: "字符串到字节集",
        contents: [
            {value: '**字符串到字节集**'},
            {value: '**参数说明**'},
            {value: '参数1:bs     字节集类型  [待转换的字节集]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nstr := BytesToString([]byte("11111"))\n```'},
        ]
    },
    {
        name: ['字节集拼接', '拼接字节集', 'BytesAdd'],
        insertText: 'BytesAdd(${1:bs1}$0,${2:bs2})',
        detail: "字节集拼接",
        contents: [
            {value: '**字节集拼接**'},
            {value: '**参数说明**'},
            {value: '参数1:bs1     字节集类型'},
            {value: '参数2:bs2     字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nbs := BytesAdd([]byte("1"),[]byte("2"))\n//将输出12\n```'},
        ]
    },
    {
        name: ['取字节集左边', 'GetBytesLeft'],
        insertText: 'GetBytesLeft(${1:bs}$0,${2:count})',
        detail: "取字节集左边",
        contents: [
            {value: '**取字节集左边**'},
            {value: '**参数说明**'},
            {value: '参数1:bs     字节集类型  要取哪个字节集的左边数据'},
            {value: '参数2:count     int类型 获取左边几个字节'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nbs := GetBytesLeft([]byte("123"),2)\n//将输出"23"的字节集\n```'},
        ]
    },
    {
        name: ['取字符串左边', 'GetStringLeft'],
        insertText: 'GetStringLeft(${1:str}$0,${2:count})',
        detail: "取字符串左边",
        contents: [
            {value: '**取字符串左边**'},
            {value: '**参数说明**'},
            {value: '参数1:str     字符串类型  要取哪个字符串的左边数据'},
            {value: '参数2:count     int类型 获取左边几个字符'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nbs := GetStringLeft("123",2)\n//将输出"12"\n```'},
        ]
    },
    {
        name: ['取字节集右边', 'GetBytesRight'],
        insertText: 'GetBytesRight(${1:bs}$0,${2:count})',
        detail: "取字节集右边",
        contents: [
            {value: '**取字节集右边**'},
            {value: '**参数说明**'},
            {value: '参数1:bs     字节集类型  要取哪个字节集的右边数据'},
            {value: '参数2:count     int类型 获取右边几个字节'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\nbs := GetBytesRight([]byte("123"),2)\n//将输出"23"的字节集\n```'},
        ]
    },
    {
        name: ['取字符串右边', 'GetStringRight'],
        insertText: 'GetStringRight(${1:str}$0,${2:count})',
        detail: "取字符串右边",
        contents: [
            {value: '**取字符串右边**'},
            {value: '**参数说明**'},
            {value: '参数1:str     字符串类型  要取哪个字符串的右边数据'},
            {value: '参数2:count     int类型 获取右边几个字节'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nbs := GetStringRight("123",2)\n//将输出"23"\n```'},
        ]
    },
    {
        name: ['GetTimestamp10', '取10位时间戳', "取时间戳10位"],
        insertText: 'GetTimestamp10()$0',
        detail: "取时间戳10位",
        contents: [
            {value: '**取时间戳10位**'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\n ts10 := GetTimestamp10()\nLog("当前10位时间戳:",ts10)\n```'},
        ]
    },
    {
        name: ['GetTimestamp13', '取13位时间戳', "取时间戳13位"],
        insertText: 'GetTimestamp13()$0',
        detail: "取时间戳13位",
        contents: [
            {value: '**取时间戳13位**'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\n\tts13 := GetTimestamp13()\nLog("当前13位时间戳:",ts13)\n```'},
        ]
    },
    {
        name: ['IntToString', '数值到字符串', "整数到字符串"],
        insertText: 'IntToString(${1:number}$0)',
        detail: "整数到字符串",
        contents: [
            {value: '**整数到字符串**'},
            {value: '**参数说明**'},
            {value: '参数1:number     数值类型'},
            {value: '\t请传入任意数字类型,例如:int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,byte,uintptr'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\n\tiStr := IntToString(123)\n\tLog("整数到字符串:",123,"转换后",iStr)\n```'},
        ]
    },
    {
        name: ['StringToInt', '字符串到整数'],
        insertText: 'StringToInt(${1:iStr}$0)',
        detail: "字符串到整数",
        contents: [
            {value: '**字符串到整数**'},
            {value: '**参数说明**'},
            {value: '参数1:iStr     字符串类型 请传入要转换的字符串'},
            {value: '返回值:int'},
            {value: '**示例代码**'},
            {value: '```go\n\tiInt := StringToInt("123")\n\tLog("字符串到整数:","123","转换后",iInt)\n```'},
        ]
    },
    {
        name: ['WriteFile', '写到文件', "写出文件", "输出到文件"],
        insertText: 'WriteFile(${1:filePath}$0,${2:data})',
        detail: "写出文件",
        contents: [
            {value: '**写出文件**'},
            {value: '**参数说明**'},
            {value: '参数1:filePath     字符串类型  [要储存到本地的全路径]'},
            {value: '参数2:data     可以是字符串 也可以是字节集  [要写出的值]'},
            {value: '返回值:bool'},
            {value: '**示例代码**'},
            {value: '```go\n\tWriteFile("c:\\1.txt",Conn.GetRequestBody())\n```'},
        ]
    },
    {
        name: ['ReadFile', '读入文件', "读取文件"],
        insertText: 'ReadFile(${1:filePath}$0)',
        detail: "读入文件",
        contents: [
            {value: '**读入文件**'},
            {value: '**参数说明**'},
            {value: '参数1:filePath     字符串类型  [要读取本地文件的全路径]'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n\tbs := ReadFile("c:\\1.txt")\n\tLog("读取本地文件->长度:",len(bs))\n```'},
        ]
    },
    {
        name: ['BytesIndex', '查找字节集位置', "寻找字节集位置"],
        insertText: 'BytesIndex(${1:bs1}$0,${2:bs2})',
        detail: "查找字节集位置",
        contents: [
            {value: '**查找字节集位置**'},
            {value: '寻找 bs2在bs1中首次出现的位置，失败返回-1\u3000\u3000\u3000\u3000'},
            {value: '**参数说明**'},
            {value: '参数1:bs1     字节集类型'},
            {value: '参数1:bs2     字节集类型'},
            {value: '返回值:int'},
            {value: '**示例代码**'},
            {value: '```go\n\tbs := BytesIndex([]byte("1112"),[]byte("2"))\n\t//将输出3\n```'},
        ]
    },
    {
        name: ['StringIndex', '查找字符串位置', "寻找字符串位置"],
        insertText: 'StringIndex(${1:str1}$0,${2:str2})',
        detail: "查找字符串位置",
        contents: [
            {value: '**查找字符串位置**'},
            {value: '寻找 str2在str1中首次出现的位置，失败返回-1\u3000\u3000\u3000'},
            {value: '**参数说明**'},
            {value: '参数1:str1     字符串类型'},
            {value: '参数2:str2     字符串类型'},
            {value: '返回值:int'},
            {value: '**示例代码**'},
            {value: '```go\n\tbs := StringIndex("1112","2")\n\t//将输出3\n```'},
        ]
    },
    {
        name: ['取字符串中间', '取出字符串中间', 'SubString'],
        insertText: 'SubString(${1:str}$0,${2:left},${3:Right})',
        detail: "查找字符串位置",
        contents: [
            {value: '**查找字符串位置**'},
            {value: '取出字符串中间部分\u3000\u3000\u3000'},
            {value: '**参数说明**'},
            {value: '参数1:str       字符串类型 [原始字符串]'},
            {value: '参数2:left      字符串类型 [左边的字符串]'},
            {value: '参数3:Right     字符串类型 [右边的字符串]'},
            {value: '返回值:字符串'},
            {value: '**示例代码**'},
            {value: '```go\nstr1 := SubString("11123456","2","5")\n//将输出34\n```'},
        ]
    },
    {
        name: ['DeflateCompress', "Deflate压缩"],
        insertText: 'DeflateCompress(${1:value}$0)',
        detail: "Deflate压缩",
        contents: [
            {value: '**Deflate压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  DeflateCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['ZlibCompress', "Zlib压缩"],
        insertText: 'ZlibCompress(${1:value}$0)',
        detail: "Zlib压缩",
        contents: [
            {value: '**Zlib压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  ZlibCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['GzipCompress', "Gzip压缩"],
        insertText: 'GzipCompress(${1:value}$0)',
        detail: "Gzip压缩",
        contents: [
            {value: '**Gzip压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  GzipCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['ZSTDCompress', 'ZSTD压缩'],
        insertText: 'ZSTDCompress(${1:value}$0)',
        detail: "zstd压缩",
        contents: [
            {value: '**zstd压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  ZSTDCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['BrCompress', "Br压缩"],
        insertText: 'BrCompress(${1:value}$0)',
        detail: "Br压缩",
        contents: [
            {value: '**Br压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  BrCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['DeflateUnCompress', "Deflate解压缩"],
        insertText: 'DeflateCompress(${1:value}$0)',
        detail: "Deflate解压缩",
        contents: [
            {value: '**Deflate解压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  DeflateUnCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['ZlibUnCompress', 'zlib解压缩'],
        insertText: 'ZlibUnCompress(${1:value}$0)',
        detail: "Zlib解压缩",
        contents: [
            {value: '**Zlib解压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  ZlibUnCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['GzipUnCompress', "Gzip解压缩"],
        insertText: 'GzipUnCompress(${1:value}$0)',
        detail: "Gzip解压缩",
        contents: [
            {value: '**Gzip解压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  GzipUnCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['ZSTDUnCompress', 'ZSTD解压缩'],
        insertText: 'ZSTDUnCompress(${1:value}$0)',
        detail: "zstd解压缩",
        contents: [
            {value: '**zstd解压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  ZSTDUnCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['BrUnCompress', "Br解压缩"],
        insertText: 'BrUnCompress(${1:value}$0)',
        detail: "Br解压缩",
        contents: [
            {value: '**Br解压缩**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字节集'},
            {value: '**示例代码**'},
            {value: '```go\n  BrUnCompress([]byte("11123456"))\n```'},
        ]
    },
    {
        name: ['Protobuf转pbJSON', 'Protobuf解析', "pb转PbJSON", 'Pb解析', "PbToJson"],
        insertText: 'PbToJson(${1:value}$0)',
        detail: "Protobuf转JSON",
        contents: [
            {value: '**Protobuf转JSON**'},
            {value: '**参数说明**'},
            {value: '参数1:value       字节集类型'},
            {value: '返回值:字符串 失败返回空文本'},
            {value: '**示例代码**'},
            {value: '```go\nbin := GoHexDecode("0a210a0a64656275675f696e666f1213120871514546346a4a761a05332e332e392003")\njsonText := PbToJson(bin)\nif len(jsonText) > 0 {\n\tLog("pb解析成功:",jsonText)\n}else{\n\tLog("Pb解析失败")\n}\n```'},
        ]
    },
    {
        name: ['PbJSON转pb', 'ProtobufJSON转pb', 'Protobuf还原', "JSON转Protobuf", "JsonToPB"],
        insertText: 'JsonToPB(${1:jsonText}$0)',
        detail: "JSON转Protobuf",
        contents: [
            {value: '**JSON转Protobuf**'},
            {value: '**参数说明**'},
            {value: '参数1:jsonText       文本型类型'},
            {value: '返回值:字节集 失败返回空字节集'},
            {value: '**示例代码**'},
            {value: '```go\nbin:=JsonToPB(jsonText)\nif len(bin) > 0 {\n\tLog("PB还原结果:",GoHexEncode(bin))\n}else{\n\tLog("Pb还原失败")\n}\n```'},
        ]
    },
    {
        name: ['JSON解析', 'JsonParse'],
        insertText: 'JsonParse(${1:jsonText}$0)',
        detail: "JSON解析",
        contents: [
            {value: '**JSON解析**'},
            {value: '**参数说明**'},
            {value: '参数1:jsonText       文本型类型'},
            {value: '返回值:JSON对象'},
            {value: '**示例代码**'},
            {value: '```go\n//解析JSON\nobj := JsonParse(data) \n//设置值\nobj.SetData("data.[0].path", "1111")\nobj.SetData("data.[0].ts", 123456)\n//获取值,无论值是什么类型都是返回字符串\nobj.GetData("data.[0].path")\n//获取成员数量\nobj.GetCount("data")\n```'},
        ]
    },
]