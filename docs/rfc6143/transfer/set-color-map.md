# 设置颜色表

当 PIXEL_FORMAT 的 true-color-flag 字段被设置为 0 时，服务端使用颜色表表示像素的颜色。
`SetColorMapEntries` 用于设置颜色表的内容。

```
+--------------+--------------+------------------+
| No. of bytes | Type [Value] | Description      |
+--------------+--------------+------------------+
| 1            | U8 [1]       | message-type     |
| 1            |              | padding          |
| 2            | U16          | first-color      |
| 2            | U16          | number-of-colors |
+--------------+--------------+------------------+
```

- message-type: 消息类型，固定是 `1`
- first-color: [未知](https://github.com/rfbproto/rfbproto/issues/42)
- number-of-colors: 颜色的数量

## 色值

颜色表的值总是 3 个 16 bits，代表红、绿、蓝三种颜色，每个颜色的范围是 0-65535。
例如，白色的色值是 65535,65535,65535。

```
+--------------+--------------+-------------+
| No. of bytes | Type [Value] | Description |
+--------------+--------------+-------------+
| 2            | U16          | red         |
| 2            | U16          | green       |
| 2            | U16          | blue        |
+--------------+--------------+-------------+
```
