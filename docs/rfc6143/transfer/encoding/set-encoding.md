# 设置编码

客户端用 SetEncoding 消息告知服务端，接受哪些像素[编码](/rfc6143/transfer/encoding/README.md)。

除了用于解析像素的编码外，客户端可以发送伪编码，向服务端请求拓展功能。如果服务端不识别此编码，可以直接忽略。客户端在未收到服务端明确的”支持“回复前，应当默认服务端不支持伪编码。

数据结构如下：

```
+-----------------------+--------------+---------------------+
| No. of bytes          | Type [Value] | Description         |
+-----------------------+--------------+---------------------+
| 1                     | U8 [2]       | message-type        |
| 1                     |              | padding             |
| 2                     | U16          | number-of-encodings |
| 4*number-of-encodings | S32 array    | encoding-types      |
+-----------------------+--------------+---------------------+
```

- message-type: 消息类型，固定为 `2`
- number-of-encodings: 编码数量
- encoding-types: [编码标识符](/rfc6143/transfer/encoding/README.md)
