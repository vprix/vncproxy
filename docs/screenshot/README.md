## Screenshot

代码路径在`./cmd/screenshot`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译

### 获取帮助信息
```shell
# 查看帮助信息
$ ./screenshot --help

# 查看版本信息
$ ./screenshot version
```

### 启动Screenshot 获取vnc服务器的屏幕截图

```shell

# imageFile  要生成的截图地址,暂时只支持jpeg格式(必填)
# vncHost   要连接的vnc服务端地址(必填)
# vncPort   要连接的vnc服务端端口(必填)
# vncPassword  要连接的vnc服务端密码，不传则使用auth none

$ ./screenshot --imageFile=./screen.jpeg --vncHost=127.0.0.1 --vncPort=5900 --vncPassword=12345612       
```