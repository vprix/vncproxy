# 编码

”编码“ 是像素数据的表达方式。在发往客户端前，服务器经常将发送的图形数据进行编码、压缩，提高图像传输效率。

```
+--------+-----------------------------+
| Number | Name                        |
+--------+-----------------------------+
| 0      | Raw                         |
| 1      | CopyRect                    |
| 2      | RRE                         |
| 5      | Hextile                     |
| 15     | TRLE                        |
| 16     | ZRLE                        |
| -239   | Cursor pseudo-encoding      |
| -223   | DesktopSize pseudo-encoding |
+--------+-----------------------------+
```

- [Raw](/rfc6143/transfer/encoding/raw.md): 原始位图编码，即不编码
- [CopyRect](/rfc6143/transfer/encoding/copy-rect.md): 从帧缓冲复制
- [RRE](/rfc6143/transfer/encoding/rise-and-run-length.md): rise-and-run-length 二维游程编码
- Hextile: RRE 的变种，图块游程编码
- [TRLE](/rfc6143/transfer/encoding/tiled-run-length.md): 图块游程编码
- [ZRLE](/rfc6143/transfer/encoding/zlib-run-length.md): Zlib Run-Length Encoding，zlib 压缩的游程编码
- Cursor pseudo-encoding: 鼠标指针伪编码
- DesktopSize pseudo-encoding: 桌面分辨率伪编码
