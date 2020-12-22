## websocket-demo

这是一个用 go 编写的 websocket 服务器 demo，实现了 websocket 协议中最基本的收发数据帧服务。

[websocket 文档](https://tools.ietf.org/html/rfc6455)

**使用**

```
go build
./main
```

然后就可以用任意 websocket 工具与 `ws://127.0.0.1:8888` 进行通信了。