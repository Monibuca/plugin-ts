# tsplugin
ts publisher for monibuca

处理TS数据的插件

## 插件名称

TS

## 功能

1. 通过Publish发布一个TS流，然后通过Feed方法填入TS数据即可
2. 通过PublishDir可以读取服务器上文件夹内的所有ts文件进行发布

## 配置

```toml
[Plugins.TS]
BufferLength = 2048
Path         = "ts"
AutoPublish  = true
```
BufferLength指的是解析TS流的时候的缓存大小，单位是PES包的个数
Path 指存放ts的目录
AutoPublish 是指是否当订阅者订阅某个房间时自动触发发布TS