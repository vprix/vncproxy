## Player

代码路径在`./cmd/player`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译

### 获取帮助信息
```shell
# 查看帮助信息
$ ./player --help

# 查看版本信息
$ ./player version
```

### 启动Player Tcp服务

```shell

# rbsFile  要保存的rbs文件路径(必填)
# tcpHost   本地监听的tcp协议地址 默认0.0.0.0
# tcpPort  本地监听的tcp协议端口 默认8989
# proxyPassword  连接到proxy的密码   不传入密码则使用auth none
# debug  使用debug模式启动服务

$ ./player start tcpServer  --rbsFile=/path/to/foo.rbs
                            --tcpHost=0.0.0.0
                            --tcpPort=8989
                            --proxyPassword=12345612
                            --debug             
```

### 启动Player WS服务

```shell

# rbsFile  要保存的rbs文件路径(必填)
# wsHost   启动websocket服务的本地地址  默认 0.0.0.0
# wsPort   启动websocket服务的本地端口 默认8988
# wsPath   启动websocket服务的url path 默认'/'
# proxyPassword  连接到proxy的密码   不传入密码则使用auth none
# debug  使用debug模式启动服务

$ ./player start wsServer --rbsFile=/path/to/foo.rbs
                          --wsHost=0.0.0.0
                          --wsPort=8989
                          --wsPath=/
                          --proxyPassword=12345612
                          --debug             
```