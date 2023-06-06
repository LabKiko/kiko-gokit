# 服务端数据埋点 SDK
> 提供日志打点 + kafka + 神策 多通道能力

## 1. 准备工作
- 安装 `gokit`
> go get -u github.com/LabKiko/kiko-gokit

## 2. 使用示例

```go
// 初始化数据打点配置项
datalog, err = Dial("infra.bff.feeds",
    WithBasePath("../data"),
)
if err != nil {
    panic(err)
}

defer datalog.Close()

attributes := make([]attribute.KeyValue, 0)
attributes = append(attributes,
    attribute.String("biz_system", "sona"),
    attribute.String("platform_type", "server"),
)

err := datalog.Write(context.Background(), &Event{
    Name:         "addpersonal",
    DistinctId:   "10000",
    DistinctType: User,
    Time:         time.Now(),
}, attributes...)
```

- ctx 上下文，用于 `链路追踪` 获取 trace 信息
- `event` 信息，kafka topic 会自动加上初始化传递的 `app_id`，组合最终显示为 `${app_id}-${event}`, kafka topic 规范请[参考](https://wiki.mter.io/pages/viewpage.action?pageId=27821929) 
- attributes 键值

## 3. 其他注意事项

### 神策保留字段
为了保证查询时属性名不与系统变量名冲突，设置如下保留字段，请避免其作为事件名和属性名（properties 中的 key）使用：

```
date
datetime
distinct_id
event
events
event_id
first_id
id
original_id
device_id
properties
second_id
time
user_id
users
user_group 开头
user_tag 开头
```