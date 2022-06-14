# VncProxy [![GitHub release](https://img.shields.io/github/v/release/vprix/vncproxy.svg?style=flat-square)](https://github.com/vprix/vncproxy/releases) [![report card](https://goreportcard.com/badge/github.com/vprix/vncproxy?style=flat-square)](http://goreportcard.com/report/vprix/vncproxy) [![github issues](https://img.shields.io/github/issues/vprix/vncproxy.svg?style=flat-square)](https://github.com/vprix/vncproxy/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/vprix/vncproxy.svg?style=flat-square)](https://github.com/vprix/vncproxy/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/vprix/vncproxy) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/vprix/vncproxy/tree/main/examples)
## VncProxy简介

`VncProxy` 是使用`Golang`实现的`Vnc`远程桌面代理组件，完全解析`rfb`协议，支持远程桌面代理，rbs文件录屏，rbs文件回放，截图，录制视频.

* 全协议支持的vnc proxy。
  * 支持Tcp代理
  * 支持Websocket代理
* 屏幕录像，保存为`RBS`文件
* 重播服务器，支持vnc客户端链接，播放`RBS`文件。
* 支持实时录制视频
* 支持通过`RBS`文件录制视频。
* 支持屏幕截图

## 支持的编码格式

- [x] Raw
- [x] CopyRect
- [x] CoRRE
- [x] rre
- [x] Hextile
- [x] Tight
- [x] TightPng
- [x] ZLib
- [x] Zrle
- [x] CursorPseudo
- [x] CursorWithAlphaPseudo
- [x] DesktopNamePseudo
- [x] DesktopSizePseudo
- [x] ExtendedDesktopSizePseudo
- [x] LedStatePseudo
- [x] CursorPosPseudo
- [x] XCursorPseudo
- [ ] jpeg
- [ ] jrle
- [ ] trle

## 组件说明

### Proxy

1. 启动`server`接受`vnc viewer`的链接.
2. 启动`client`连接到指定的`vnc server`.
3. 为`vnc viewer`和`vnc server`之间建立起消息转发通道。
4. 因为`rfb`协议被完全解析，可以针对通信的消息进行转发处理，产生了后续的功能。

### Recorder

1. 启动`client`连接到指定的`vnc server`.
2. 发送帧缓冲区更新消息`FramebufferUpdateRequest`到`vnc server`。
3. 处理`vnc server`回复的界面更新消息`FramebufferUpdate`。
4. 把这一过程以`rbs`文件格式记录下来。

### Player

1. 启动`server`接受`vnc viewer`的链接.
2. 读取`rbs`文件，并按格式生成`FramebufferUpdate`消息发送给`vnc viewer`。
3. `vnc viewer`的界面就会回放动作。

### Video

1. 支持`Proxy`,`Recorder`和`rbs`文件作为输入源。
2. 把`FramebufferUpdate`消息转换为视频文件。

### Screenshot

1. 支持`Proxy`,`Recorder`和`rbs`文件作为输入源。
2. 把当前的界面视图转换为图片文件。

## 使用说明

`vncProxy`项目有多种应用场景。
可以作为单独的应用程序编译，也可以作为库被其他应用程序引用。
接下来，分别介绍各种场景下的使用方式。
### 编译

```shell
# 使用方式:
# build.sh [-s app_name] [-v version] [-g go_bin]
# app_name 需要编译的应用名称
#          选项: proxy,player,recorder,video,screenshot.
#          默认是所有应用,多个应用可以逗号分割.
# version  编译后的文件版本号,默认为当前git的commit id.
# go_bin   使用的golang程序

# 编译所有应用
$ ./build 

# 编译proxy
$ ./build -s proxy -v v0.1.0

# 编译player,recorder
$ ./build -s player,recorder -v v0.1.0
```

编译后的二进制文件在`./bin/`目录

### Proxy

代码路径在`./cmd/proxy`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译

#### 获取帮助信息
```shell
# 查看帮助信息
$ ./proxy --help

# 查看版本信息
$ ./proxy version

```

#### 启动tcp服务
```shell

# 启动tcp server接受vnc viewer的连接
# vncHost  vnc服务器host
# vncPort  vnc服务器port
# vncPassword  vnc服务器密码
# tcpHost  本地监听的地址
# tcpPort  本地监听的端口
# proxyPassword  vnc连接的密码
# debug  使用debug模式启动服务

$ ./proxy start tcpServer --vncHost=192.168.1.2 \     
                          --vncPort=5901 \           
                          --vncPassword=vprix \       
                          --tcpHost=0.0.0.0 \        
                          --tcpPort=8989 \           
                          --proxyPassword=12345612 \  
                          --debug                    
```

#### 启动WebSocket服务
```shell

# 启动ws server接受novnc的连接
# vncHost  vnc服务器host
# vncPort  vnc服务器port
# vncPassword  vnc服务器密码
# wsHost  本地监听的地址
# wsPort  本地监听的端口
# wsPath  websocket连接的地址
# proxyPassword  vnc连接的密码
# debug  使用debug模式启动服务

$ ./proxy start wsServer  --vncHost=192.168.1.2 \      
                          --vncPort=5901         \     
                          --vncPassword=vprix \       
                          --wsHost=0.0.0.0 \          
                          --wsPort=8988    \           
                          --wsPath=/websockify \         
                          --proxyPassword=12345612 \    
                          --debug              
```

### Recorder

代码路径在`./cmd/recorder`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译
#### 获取帮助信息
```shell
# 查看帮助信息
$ ./recorder --help

# 查看版本信息
$ ./recorder version

```

#### 启动Recorder服务
```shell

# rbsFile  要保存的rbs文件路径(必填)
# vncHost  vnc服务器host
# vncPort  vnc服务器port
# vncPassword  vnc服务器密码
# debug  使用debug模式启动服务

$ ./recorder start --rbsFile=/path/to/foo.rbs
							--vncHost=192.168.1.2 
							--vncPort=5901
							--vncPassword=vprix
							--debug             
```

### Player

代码路径在`./cmd/player`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译
#### 获取帮助信息
```shell
# 查看帮助信息
$ ./player --help

# 查看版本信息
$ ./player version
```

#### 启动Player Tcp服务
```shell

# rbsFile  要保存的rbs文件路径(必填)
# tcpHost   本地监听的tcp协议地址 默认0.0.0.0
# tcpPort  本地监听的tcp协议端口 默认8989
# proxyPassword  连接到proxy的密码   不传入密码则使用auth none
# debug  使用debug模式启动服务

$ ./player start tcpServer  --rbsFile=/path/to/foo.rbs
                            --tcpHost=0.0.0.0
                            --tcpPort=8989
                            --proxyPassword=12345612
                            --debug             
```

#### 启动Player WS服务
```shell

# rbsFile  要保存的rbs文件路径(必填)
# wsHost   启动websocket服务的本地地址  默认 0.0.0.0
# wsPort   启动websocket服务的本地端口 默认8988
# wsPath   启动websocket服务的url path 默认'/'
# proxyPassword  连接到proxy的密码   不传入密码则使用auth none
# debug  使用debug模式启动服务

$ ./player start wsServer --rbsFile=/path/to/foo.rbs
                          --wsHost=0.0.0.0
                          --wsPort=8989
                          --wsPath=/
                          --proxyPassword=12345612
                          --debug             
```
### Screenshot
代码路径在`./cmd/screenshot`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译
#### 获取帮助信息
```shell
# 查看帮助信息
$ ./screenshot --help

# 查看版本信息
$ ./screenshot version
```

#### 启动Screenshot 获取vnc服务器的屏幕截图
```shell

# imageFile  要生成的截图地址,暂时只支持jpeg格式(必填)
# vncHost   要连接的vnc服务端地址(必填)
# vncPort   要连接的vnc服务端端口(必填)
# vncPassword  要连接的vnc服务端密码，不传则使用auth none

$ ./screenshot --imageFile=./screen.jpeg --vncHost=127.0.0.1 --vncPort=5900 --vncPassword=12345612       
```

## 项目参考

本项目参考了以下项目完成。
* [vncproxy](https://github.com/amitbet/vncproxy)
* [vnc2video](https://github.com/amitbet/vnc2video)
* [rfbproto](https://github.com/rfbproto/rfbproto)

## 交流

我在做这个项目的过程中碰到了很多问题，查遍了互联网，缺少中文资料，大部分信息都是雷同的。
所以我萌生了开源的想法，帮助更多有需要的人。

我建立了一个可供交流的微信群，以便大家在使用的过程中碰到疑问，能有解答的地方。
当然，如果你对vnc有兴趣，也可以加我微信，多多交流。
欢迎各位贡献代码。

![微信二维码](/docs/images/5bb8dbe702ce04b0bdde8c26583b152.jpg)



