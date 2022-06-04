# 键盘事件

在正常理解中，键盘事件的处理应该是简单明了的。参考以下协议报文

```
  +--------------+--------------+--------------+
  | No. of bytes | Type [Value] | Description  |
  +--------------+--------------+--------------+
  | 1            | U8 [4]       | message-type |
  | 1            | U8           | down-flag    |
  | 2            |              | padding      |
  | 4            | U32          | key          |
  +--------------+--------------+--------------+
```

- message-type: 固定为 `0x4`
- down-flag: `1` 表示键位按下，`0` 表示弹起
- pending: 对齐字节，方便解析
- key: 表示具体的键位

其中 key 的值在 X 系统中有[明确定义](https://www.x.org/releases/X11R7.6/doc/xproto/x11protocol.html#keysym_encoding)

```
 +-----------------+--------------------+
 | Key name        | Keysym value (hex) |
 +-----------------+--------------------+
 | BackSpace       | 0xff08             |
 | Tab             | 0xff09             |
 | Return or Enter | 0xff0d             |
 | Escape          | 0xff1b             |
 | Insert          | 0xff63             |
 | Delete          | 0xffff             |
 | Home            | 0xff50             |
 | End             | 0xff57             |
 | Page Up         | 0xff55             |
 | Page Down       | 0xff56             |
 | Left            | 0xff51             |
 | Up              | 0xff52             |
 | Right           | 0xff53             |
 | Down            | 0xff54             |
 | F1              | 0xffbe             |
 | F2              | 0xffbf             |
 | F3              | 0xffc0             |
 | F4              | 0xffc1             |
 | ...             | ...                |
 | F12             | 0xffc9             |
 | Shift (left)    | 0xffe1             |
 | Shift (right)   | 0xffe2             |
 | Control (left)  | 0xffe3             |
 | Control (right) | 0xffe4             |
 | Meta (left)     | 0xffe7             |
 | Meta (right)    | 0xffe8             |
 | Alt (left)      | 0xffe9             |
 | Alt (right)     | 0xffea             |
 +-----------------+--------------------+
```

## 组合键

组合键指 `Ctrl + Alt + Del` 或 `Shift + 3` 等组合按键。受不同的操作系统、键盘布局影响，组合键是按键事件中容易发生歧义的一环。

RFB 基本遵循以下规则：

- 如果客户端 key 在 `keysym` 中存在，服务端应该遵循 `keysym` 的指示，尽可能的忽略客户端传递 `Shift`、`CpasLock` 等键位，在需要时，应该主动补充/忽略 `Shift` 等键位。例如，在 US 键盘布局中，`#` 需要按下 `Shift + 3`，但是在 UK 布局中不需要。这就意味着用户在输入 `#` 的时候不会输入 `Shift`。这种情况下，服务端应该主动模拟一个 `Shift` 状态，防止输入的键位是 `3`。同理，如果 key 输入的键位是 `A`，服务端统一要模拟一个 `Shift`，保证输入的是 `A` 而不是 `a`。
- 如果客户端 key 在 `keysym` 中不存在（例如 `Ctrl + A`），服务端应该遵循客户端指示，客户端应该主动在 `A` 前发送 `Ctrl` 的按键。
- 如果客户端通过 `Ctrl + Alt + Q` 来输入 `@`，客户端应该在发送 `Ctrl`/`Alt`/`@`后，主动发送`Ctrl`/`Alt`的弹起事件。
- 对于 `BackTab`，常见的有三种实现，`ISO_Left_Tab` `BackTab` 和 `Shift + Tab`。RFB 协议优先使用 `Shift + Tab`，但对于其他的键位，服务端和客户端应当尽量提供兼容。
- 优先使用 `ASCII` 而不是 `unicode`
- 对于 `Ctrl + Alt + Del` 等无法被客户端操作系统拦截的按键（系统拦截有更高优先级），客户端应该提供操作按钮
