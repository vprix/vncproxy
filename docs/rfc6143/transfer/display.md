# 显示协议

客户端可能处于弱网络环境，或只有较低性能的渲染设备。如果服务端不加限制的向客户端发送像素画面，很容易造成客户端卡死或网络堵塞。

在 RFB 协议中，当且仅当客户端主动请求显示数据时，服务端才会将 [FramebufferUpdate](#FramebufferUpdate) 发往客户端。响应 [FramebufferUpdateRequest](#FramebufferUpdateRequest) 往往需要返回多条 FramebufferUpdate。


```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: FramebufferUpdateRequest
    loop no change
      Client->>Server: FramebufferUpdateRequest
    
    end
    Server->>Client: FramebufferUpdate
    Server->>Client: FramebufferUpdate
    Server->>Client: FramebufferUpdate
```

## FramebufferUpdateRequest

FramebufferUpdateRequest 告知服务端，客户端希望得到指定区域的内容。

```
+--------------+--------------+--------------+
| No. of bytes | Type [Value] | Description  |
+--------------+--------------+--------------+
| 1            | U8 [3]       | message-type |
| 1            | U8           | incremental  |
| 2            | U16          | x-position   |
| 2            | U16          | y-position   |
| 2            | U16          | width        |
| 2            | U16          | height       |
+--------------+--------------+--------------+
```

- message-type: 消息类型，固定 `3`
- incremental: 是否是增量请求。
- x-position/y-position: 区域的起始坐标
- width/height: 区域的长度和宽度

incremental 通常为非 0 值，服务器只需要发有变化的图像信息。当客户端丢失了缓存的帧缓冲信息，或者刚建立连接，需要完整的图像信息时，将 incremental 置为 0，获取全量信息。

## FramebufferUpdate

FramebufferUpdate 由一组矩形图像(rectangles of pixel)组成，客户端收到 FramebufferUpdate 消息后，将消息内的矩形填充到帧缓冲对应区域，完成图像展示。

```
+--------------+--------------+----------------------+
| No. of bytes | Type [Value] | Description          |
+--------------+--------------+----------------------+
| 1            | U8 [0]       | message-type         |
| 1            |              | padding              |
| 2            | U16          | number-of-rectangles |
+--------------+--------------+----------------------+
```

- message-type: 消息类型，固定 0
- number-of-rectangles: 矩形的数量

### FramebufferUpdateRectangle

FramebufferUpdate 携带 `number-of-rectangles` 数量的矩形信息，每个矩形都有头部信息

```
+--------------+--------------+---------------+
| No. of bytes | Type [Value] | Description   |
+--------------+--------------+---------------+
| 2            | U16          | x-position    |
| 2            | U16          | y-position    |
| 2            | U16          | width         |
| 2            | U16          | height        |
| 4            | S32          | encoding-type |
+--------------+--------------+---------------+
```

- x-position/y-position: 矩形起始坐标
- width/height: 矩形宽度和高度
- encoding-type: 编码类型
