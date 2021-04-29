
### Credible

#### Brief

- 基于TCP
- 传递指令加数据交互，具体实现由注册指令端实现
- 消息记录，每一条消息都有一个36B的唯一id标识
- 自定义json格式日志（level，time，line，func，data）

#### Protocol



固定八个字节：NTE [数据约定] [数据长度]
1~3B ：'N','T','E'作为协议约定字段
4B：高4位判断该数据是作为回复消息，发送消息，debug消息还是其他；低4位说明数据以什么格式发送过来（raw字节，json数据，其他）

```c
//高4位
ocSendFlag (0000): 此为对端发过来的消息
ocRespFlag (0001): 此为对端回复当前端的消息，主要作出队操作
ocDebugFlag (0010):其他
...

//低4位
ocRaw (0000):数据为raw数据
ocJson(0001):数据为Json格式
...
```

5-8B：数据长度

JSON数据格式：

```json
{
    "uuid":"c5fe250d-5cc0-3c42-bffd-75ff46912735", //标识信息的唯一值
    "order":"hello", //指令，
    "data":interface{} // 实际发给对端的数据
}
```

#### Example

```go
//server端
func main() {
    //IP and port
    ip := "0.0.0.0"
    port := "17178"

    //新建server数据，传入IP，和port
    credibleSrv := credible.NewServer(ip, port)

    //注册指令以及操作函数
    credibleSrv.
    RegisterHandler("hello", hello).
    RegisterHandler("test", test)

    //非阻塞
    credibleSrv.Run()

    for {
        time.Sleep(3*time.Second)
        //获取当前连接到
        fmt.Println("current connect to server:",credibleSrv.CurrentConnected())

        //获取当前发送出去的消息有多少条没有被回复
        fmt.Println("current msg in queue still not being responed:",credible.LenOfMsgQueue())
    }
}

func hello(data interface{}) {
    fmt.Println("hello world")
}

func test(data interface{}) {
    fmt.Println("this is test handler, data:",data)
}
```

```go
//client端
func main(){
    //实例一个客户端
    credibleCli := credible.NewClient("127.0.0.1","17178")
    //注册指令及操作函数
    credibleCli.RegisterHandler("test", test)

    err := credibleCli.Connect()
    if nil != err {
        fmt.Println("cannot connect to order server")
        return
    }
    cnt := 0
    for {
        cnt++
        time.Sleep(3 * time.Second)
        if cnt < 5 {
        go func(c int) {
             if c%2 == 0 {
                for i := 0; i < 50; i++ {
                // time.Sleep(500 * time.Millisecond)
                    if i%2 != 0 {
                        go cli.Send("hello", nil)
                    } else {
                        go cli.Send("test", map[string]interface{}{
                         "hello":  "jack",
                         "hi":     "Mary",
                         "Nothin": 1,
                         "Peter":  1.23,
                     })

                    }
                }
             } else {
                for i := 0; i < 50; i++ {
                 // time.Sleep(500 * time.Millisecond)
                if i%2 == 0 {
                    go cli.Send("test", "nothing")
                } else {
                    go cli.Send("hello", "nothing")
                }
            }
            }
        }(cnt)
  }
  fmt.Println(credible.LenOfMsgQueue())
}

func test(interface{}) {
    fmt.Println("hello, this is client test")
}
```

```go
//LOG
func main() {
    //将错误LEVEL重定向到文件error.log中(不设置则输出到标准出错流）
    credible.RedirectLevel(credible.LEVEL_ERROR,credible.NewFile("error.log"))
    credible.Error("hello")
}
```
