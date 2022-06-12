# VncProxy [![GitHub release](https://img.shields.io/github/v/release/vprix/vncproxy.svg?style=flat-square)](https://github.com/vprix/vncproxy/releases) [![report card](https://goreportcard.com/badge/github.com/vprix/vncproxy?style=flat-square)](http://goreportcard.com/report/vprix/vncproxy) [![github issues](https://img.shields.io/github/issues/vprix/vncproxy.svg?style=flat-square)](https://github.com/vprix/vncproxy/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/vprix/vncproxy.svg?style=flat-square)](https://github.com/vprix/vncproxy/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/vprix/vncproxy) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/vprix/vncproxy/tree/main/examples)
## VncProxy简介

`VncProxy` 是使用`golang`实现的rfb协议解析库，支持rfb协议解析，在其上实现了很多好用的功能。

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

## 项目参考

* [vncproxy](https://github.com/amitbet/vncproxy)
* [vnc2video](https://github.com/amitbet/vnc2video)
* [rfbproto](https://github.com/rfbproto/rfbproto)

## 交流

我在做该项目的过程中碰到了很多问题,。
我建立了一个交流的微信群，大家在使用的过程中如果有疑问可以加我微信，我会给大家解答问题。

![微信二维码](/docs/images/5bb8dbe702ce04b0bdde8c26583b152.jpg)



