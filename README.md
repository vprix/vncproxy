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

## 使用方式

### Proxy
### Recorder
### Player
### Video
### Screenshot

## 项目参考

* [vncproxy](https://github.com/amitbet/vncproxy)
* [vnc2video](https://github.com/amitbet/vnc2video)
* [rfbproto](https://github.com/rfbproto/rfbproto)





