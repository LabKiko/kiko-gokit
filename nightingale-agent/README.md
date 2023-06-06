# Prometheus 远程写入代理

> 该代理用于将指标写入远程 Prometheus 服务器.

## 核心组件

- input 接口：实现该接口的类将被用于接收指标数据
- reader 接口：读取 input 接口的数据，并将其转换为指标数据，一个 input 对应一个 reader，写入 output 接口
- output 接口：转换 reader 接口的数据为 Prometheus 协议数据 ，并将其写入远程 Prometheus 服务器
- remote write 接口：实现 Prometheus Remote Write 协议，用于将指标写入远程 Prometheus 服务器

## 配置

## 使用

## 示例

## 后续计划

- [] 增加 `Prometheus` metrics 接口注册和注销功能