## Recorder

代码路径在`./cmd/recorder`,如果单独编译该组件，也可以到该目录下自行执行`go build`命令编译

### 获取帮助信息
```shell
# 查看帮助信息
$ ./recorder --help

# 查看版本信息
$ ./recorder version

```
### 启动Recorder服务

```shell

# rbsFile  要保存的rbs文件路径(必填)
# vncHost  vnc服务器host
# vncPort  vnc服务器port
# vncPassword  vnc服务器密码
# debug  使用debug模式启动服务

$ ./recorder start --rbsFile=/path/to/foo.rbs
							--vncHost=192.168.1.2 
							--vncPort=5901
							--vncPassword=vprix
							--debug             
```