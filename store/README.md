# 存储结构

> qbase引用了cosmos-sdk的存储实现，在此对cosmos团队表示诚挚的感谢
> 此文档意在对cosmos-sdk存储模块进行分析，并说明qbase对其的修改，以供开发者参考

## 继承关系
![](https://raw.githubusercontent.com/QOSGroup/static/master/store_implements.jpg)
为了理解存储设计思路，突出两大类存储接口：

### commiter
为配合abci接口调用的进程，实现对version的控制和提交到底层存储

### cachewraper
为配合abci接口调用的数据生命周期，实现双层缓存/提交的控制逻辑

## 查询过程解释及接口定义
通过abci的Query(req RequestQuery)查询时：
##### req.Data：查询的key
##### req.Path：查询的路由信息，组成："store/$storeName/$subpath"
###### "store" 为固定字段，表示查询的对象时存储
###### $storeName 为某一iavlStore（或其他）指定的名字
###### $subpath：
* 为"store"/"key"时，表示当前查询目标为key为req.Data的value
* 为"subspace"时，比哦是当前查询目标为前缀为req.Data的子树（可展开）
需要注意存取的对象是否需要解码

## 存储过程解释及接口定义
![](https://raw.githubusercontent.com/QOSGroup/static/master/store_layers.png)
#### 直接存储数值
调用可调用实例提供的mapper
mapper直接访问BaseApp.cms实现存储
#### abci提交
两层缓存结构，在BaseApp.cms上，分别包装成两个缓存：checkState、deliverState
