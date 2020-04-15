# 简介
处理TS数据的插件

# 插件名称

TS

# 功能

1. 通过Publish发布一个TS流，然后通过Feed方法填入TS数据即可
2. 通过PublishDir可以读取服务器上文件夹内的所有ts文件进行发布
3. 通过UI界面操作，点击TS文件夹，可以将文件夹中的TS文件逐个进行发布，文件夹路径就是房间名

# 配置

```toml
[TS]
BufferLength = 2048
Path         = "ts"
AutoPublish  = true
```
- BufferLength指的是解析TS流的时候的缓存大小，单位是PES包的个数
- Path 指存放ts的目录
- AutoPublish 是指是否当订阅者订阅某个房间时自动触发发布TS

> 注意，如果开启AutoPublish，有可能会和其他开启自动发布的插件冲突，目前一个房间只可以有一个发布者