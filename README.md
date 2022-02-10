# controller

这是一个Goroutine的任务控制池, 可以控制池中Goroutine的启停, 查看goroutine的运行状况

一. 任务池`TaskPool`中的任务分为两种:
   1. **指令型**: 创建一个任务`goTask`, 该任务在Goroutine中运行时会调度用户指定的指令`goTask.TaskOrder`
   2. **函数型**: 创建一个任务`goTask`, 该任务在Goroutine中运行时会执行用户编写的函数`goTask.Operate`
   
二. 任务的返回值:
   1. **指令型**: 将指令后终端打印存入`TaskPool`的`OutDataChan`中, 无则不存入
   2. **函数型**: 将函数的返回值存入`TaskPool`的`OutDataChan`中, 无则不存入
   3. 单个任务的返回值在`OutDataChan`的数据存储格式统一为`[]interface{}`, 同时该切片的第0个值为任务名称`goTask.TaskName`, 用来区分是哪个任务的返回值
   
三. 其他:
   1. 若用户对任务返回值数据有处理需求, 请自行提取`TaskPool.OutDataChan`的数据, 并转换\断言类型
   2. 提供一个`LogChan`, 里面存放着该框架的日志输出
   3. 提供一个函数参数生成器` ArgsMaker(arg ...interface{})`, 用在`goTask.TaskArgs  []interface{}`中, 即为用户编写的函数的参数(注意传参顺序)
   4. example后续更新
   
   







