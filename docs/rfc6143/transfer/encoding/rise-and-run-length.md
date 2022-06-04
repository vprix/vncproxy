# 上升和游程编码 Rise-and-run-length

RRE 是游程编码的二维变种。基本思想是将大的矩形拆分为子矩形，每个子矩形有单值的像素组成，所有小矩形的并集构成原始矩形区域。

编码由背景像素值 Vb 、计数 N ，以及 N 个子矩形列表组成。子矩形由元组 <v,x,y,w,h> 表示，其中 v 是像素值（v != Vb），x/y/w/h 表示子矩形相对主矩形的坐标，和大小。

绘制时，客户端先以背景像素值填充矩形，再绘制每个子矩形，叠加出原始图像。

```
+---------------+--------------+-------------------------+
| No. of bytes  | Type [Value] | Description             |
+---------------+--------------+-------------------------+
| 4             | U32          | number-of-subrectangles |
| bytesPerPixel | PIXEL        | background-pixel-value  |
+---------------+--------------+-------------------------+
```

- number-of-subrectangles: 子矩形数量
- background-pixel-value: 矩形背景色

对于子矩形

```
+---------------+--------------+---------------------+
| No. of bytes  | Type [Value] | Description         |
+---------------+--------------+---------------------+
| bytesPerPixel | PIXEL        | subrect-pixel-value |
| 2             | U16          | x-position          |
| 2             | U16          | y-position          |
| 2             | U16          | width               |
| 2             | U16          | height              |
+---------------+--------------+---------------------+
```

- subrect-pixel-value: 子矩形色值
- x-position/y-position: 与背景矩行的**相对位置**
- width/height: 子矩形宽度和高度
