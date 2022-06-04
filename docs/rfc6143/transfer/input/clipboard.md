# 剪贴板

复制粘贴是双向事件，可以由客户端向服务端复制，也可以由服务端向客户端复制。

## 协议报文

复制剪贴板的报文如下：

```
  +--------------+--------------+--------------+
  | No. of bytes | Type [Value] | Description  |
  +--------------+--------------+--------------+
  | 1            | U8 [3/6]     | message-type |
  | 3            |              | padding      |
  | 4            | U32          | length       |
  | length       | U8 array     | text         |
  +--------------+--------------+--------------+
```

- message-type: 消息类型，客户端是 `0x6`，服务端是 `0x3`
- length: 文本长度
- text: 复制粘贴的文本，长度由 length 限制

协议有几点限制

- 只支持 ISO 8859-1 (Latin-1) 字符集
- 使用单独换行符 `0x0a`，不应该使用回车符 `0x0d`

## 拓展剪贴板伪协议

> RFB 3.8 协议限制，剪贴板只能传输 Latin-1 字符集。
> 2016年，Cendio Ossman 将 [Extended Clipboard Pseudo-Encoding](https://github.com/rfbproto/rfbproto/commit/08018f655acd52970680b34021159924357efb5d) 合入协议主分支，支持在剪贴板消息中传输 unicode 字符集。
> UltraVNC/TigerVNC/RealVNC 服务端都支持此拓展协议，x11vnc 尚未提供支持（2021/8/11）。

拓展剪贴板伪协议需要客户端和服务端软件同时支持。报文拓展了 `ServerCutText` 和 `ClientCutText`， 如下：

```
  +--------------+--------------+--------------+
  | No. of bytes | Type [Value] | Description  |
  +--------------+--------------+--------------+
  | 1            | U8 [3/6]     | message-type |
  | 3            |              | padding      |
  | 4            | S32          | length       |
  | 4            | U32          | text-type    |
  | length-4     | U8 array     | text         |
  +--------------+--------------+--------------+
```

- length: 数据由 U32 改为 S32。首bit是标志位，0 表示传递原始 Latin-1 消息，1 表示传递拓展信息。abs(length) 是实际的消息长度。
- text-type: 消息头部，指示消息类型
- text: 消息内容

### 消息类型

消息用 4 字节的 text-type 作为头部。

```
  +--------------+--------------+--------------+
  | No. of bytes | Type [Value] | Description  |
  +--------------+--------------+--------------+
  | 1            | U32          | message-type |
  +--------------+--------------+--------------+
```

text-type 分为指令（`action`）和格式（`formats`）两类。指令传输操作命令，格式传输剪贴板内容。

text-type 标记的含义如下：

| Bit	| Name | Description |
|-|-|---|
| 0	| [text](#文本内容) | 文本内容 |
| 1| rtf | 微软富文本格式 |
| 2	| html | 微软 HTML 格式 |
| 3	| dib | Microsoft Device Independent Bitmap |
| 4	| files | 文件，暂未实现 |
| 5-15| fotmats 保留位 |
| 16-23	| 保留 |
| 24	| [caps](#能力声明) | 指示支持的 text-type 和最大长度 |
| 25	| request | 强制对端传递剪贴板内容 |
| 26	| peek | 强制对端提供支持的 text-type |
| 27	| notify | peek 回包，返回支持的 text-type |
| 28	| [provide](#粘贴板内容) | request 回包，返回粘贴板内容 |
| 29-31	| actions 保留位 | |

### 文本内容 text

纯文本，unicode 编码的无格式文本。以 `\r\n` 作为行的结尾，原始协议的换行符是 `\n`。

文本应该以 `\0` 结尾，即使在 ClientCutText/ServerCutText 中都声明了文本长度。

### 能力声明 caps

Caps 指示期望收到的文本类型，发送的结构是长度数组。数组大小跟格式数量相等（0-15），数组的每个条目，指示格式支持的最大长度。

#### 数据结构

```
  +--------------+--------------+--------------+
  | No. of bytes | Type [Value] | Description  |
  +--------------+--------------+--------------+
  | formats*4    | U32 array    | sizes        |
  +--------------+--------------+--------------+
```

例如：

```
[1024,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]
```

表示只接受 1024 byte 以内的纯文本信息。

#### 行为约束

服务端收到支持拓展剪贴板协议的 SetEncodings 报文时，必须主动发送类型为 caps 的 ServerCutText 消息。

客户端收到 caps 消息时，应该发送类型为 caps 的 ClientCutText 消息作为回应。
否则，客户端默认会接受 text/rtf/html/request/notify/provide 消息，其中 text 默认长度为 20 Mib，其他为 0 字节。

当最大长度限制为 0 时，认为长度没有限制，如果内容长度大于声明的长度限制，则剪贴板变动的消息不会被发送。建议将所有的 caps 设置为 0，以便接受所有的剪贴板消息变动。

> 某些实现的默认行为与协议描述不一致，例如：
> - dib 也是默认支持的格式
> - text 的默认限制是 10Mid
> - rft/html 默认限制为 2Mib
> - dib 默认限制为 0 字节
> - 客户端忽略 caps 消息建议的格式和长度限制

在发送 caps 之前，只能发送指令，不能发送格式和内容。

### 粘贴板内容

粘贴板内容的 text-type 是 provide。在剪贴板变化，或对端发送 request 后发送。
在 text-type 后面，是 Zlib 压缩的字节流。对于每种支持的 text-type，会发送 size + data 数据对。

```
  +--------------+--------------+--------------+
  | No. of bytes | Type [Value] | Description  |
  +--------------+--------------+--------------+
  | 4            | U32          | size         |
  | size         | U8 array     | data         |
  +--------------+--------------+--------------+
```
