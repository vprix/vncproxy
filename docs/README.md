## Vnc Proxy 介绍

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


## 项目参考

* [vncproxy](https://github.com/amitbet/vncproxy)
* [vnc2video](https://github.com/amitbet/vnc2video)
* [rfbproto](https://github.com/rfbproto/rfbproto)







