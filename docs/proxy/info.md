## Proxy

代码路径在`./cmd/proxy`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译

### 获取帮助信息
```shell
# 查看帮助信息
$ ./proxy --help

# 查看版本信息
$ ./proxy version

```
### 启动tcp服务

```shell

# 启动tcp server接受vnc viewer的连接
# vncHost  vnc服务器host
# vncPort  vnc服务器port
# vncPassword  vnc服务器密码
# tcpHost  本地监听的地址
# tcpPort  本地监听的端口
# proxyPassword  vnc连接的密码
# debug  使用debug模式启动服务

$ ./proxy start tcpServer --vncHost=192.168.1.2 \     
                          --vncPort=5901 \           
                          --vncPassword=vprix \       
                          --tcpHost=0.0.0.0 \        
                          --tcpPort=8989 \           
                          --proxyPassword=12345612 \  
                          --debug                    
```
### 启动WebSocket服务

```shell

# 启动ws server接受novnc的连接
# vncHost  vnc服务器host
# vncPort  vnc服务器port
# vncPassword  vnc服务器密码
# wsHost  本地监听的地址
# wsPort  本地监听的端口
# wsPath  websocket连接的地址
# proxyPassword  vnc连接的密码
# debug  使用debug模式启动服务

$ ./proxy start wsServer  --vncHost=192.168.1.2 \      
                          --vncPort=5901         \     
                          --vncPassword=vprix \       
                          --wsHost=0.0.0.0 \          
                          --wsPort=8988    \           
                          --wsPath=/websockify \         
                          --proxyPassword=12345612 \    
                          --debug              
```