# daemon类程序封装

- [daemon.go](daemon.go) 封装了daemon程序生命周期的管理
- [do.go](do.go) 提供了`Always`方法 封装一个无限循环的执行单元调度入口
- [ticker.go](ticker.go) 提供了`Ticker`方法 封装一个定时器执行单元调度入口
- [opts.go](opts.go) 提供了若干可选项 可以使`Always`和`Ticker`具有更多自定义特性