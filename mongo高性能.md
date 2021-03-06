### 0. mongo的使用背景
MongoDB是NoSQL的一种，它是面向文档存储。为什么使用MongoDB？这要先从大数据说起，其中一个经典问题就是从互联网上抓数据。从互联网上我们能抓取大量的数据，那么就面临着存储，更新，查找，错误处理等问题。概括而言就是：

1. 保存，更新，查找
2. 处理错误
3. 处理大数据

### 1. 保存，更新，查找
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 0C49F3730359A14518585931BC711F9BA15703C6

**save:**

首先是如何保存从互联网上抓取的大量数据。这涉及到数据的存储格式：JSON。它是一种轻量级的数据交换格式，建构于“名称/值”对的集合，简单易读，数据体积小，速度快。
 
MongoDB的数据存储在document中，它是类似JSON的数据结构，由“名称/值”对组成。也就是说，数据块用JSON形式保存下来叫document。而很多document合在一起叫做collection，它类似于MySQL里的table。用JSON的名称/值对，到document，再到collection，大量的数据就保存了下来。

**update:**

从互联网上抓取的数据，会包含很多属性，比如url, time, author, title, content。而当我们把数据保存完后，如何添加新的属性？比如我们抓取了url, content，然后存储，后来发现还要抓取title并保存。如果之前一块块的数据连续存储完了，这时候就会发现没有空间来添加这个新的属性。一个直接的想法是把要update的原有数据块从原来位置删除，加上新的属性后再添加到后面。

但是这种做法在mongoDB中也有问题：在添加数据的过程中，因为数据从原来的位置上移走，留有了空位，就会产生碎片。为了解决这个问题，可以预留空间。在存完一个数据块之后，留一块空间（padding），这样再加新的属性的时候，直接加在padding里，不用把这个数据块移动。Padding的大小是一个tradeoff，如果太少，起不到效果，而如果太大则浪费空间。
 
那么如何设计padding的大小呢？一种做法是预留document的10%空间，这样document越大padding越大。此外，数据块移动一次，这个百分比就上涨一次，比如从10%变为15%，再移动一次，则变为20%。这种方法类似于TCP/IP中的连接重试算法。


**find:**
除了保存，更新，还有一个重要操作是find。首先讨论一下基本的find：scan。Scan是我们熟知的遍历，也就是一个个扫数据。比如要寻找某个url，我们一个个数据块，一个个数据查找。但是如果直接遍历，效率低，因为我们扫了很多无用的信息，比如content。为了跳过不需要扫描的东西，我们可以存储数据的长度length，那么我们通过length就能算出下一个url的位置，这样就能只扫描url，略过content等无用的信息。这就是BSON的第二个好处，也是对JSON的一大改进：它将JSON的每一个元素的长度存在元素的头部，这样只需读取到元素长度就能直接找到指定的点上进行读取。

### 2. 处理错误

#### 2.1 处理硬盘失败

比如一个数据A=3同时存在disk和memory里，我们想把A改为5。我们需要同时修改disk和memory里的数据。但是这样很慢，因为我们涉及到对disk的读写。

解决方法：把memory里A改写成5就认为可以了。

新问题：如果此时机器崩溃，A写成的5就没有了。

解决方法：写log/journal来处理，把log存到disk里。
 
虽然log也要写到disk里，但是把log写入disk要比把数据存入disk随机的位置快，这是因为log是sequence写的，而如果是在disk里写数据，指针要不断移动到新的位置，时间要多很多。还有一个tricky的方法：使用两块disk，一个写数据，一个写log。

这时候我们遇到另一个问题：如何写log？log有两种：behavior log和binary log。举例说，比如要把A=3改成A=5。behavior log写法就是记录所有信息：time, update, A, 3, 5。而binary log写法相对简单，记录位置和更新后的数据。而在MongoDB里使用第一种写法，具体原因，接下来会解释。//todo

前面我们提过，机器随时可能崩溃，为了保证数据的读取，我们需要备份。这样如果一个机器坏了，还可以使用备份。但是新的问题产生了，如何解决数据的同步？

#### 2.2 如何同步主从数据库

想要同步primary（简称P）机器里的数据和secondary（简称S）机器里的数据，P需要把log传给S，S依据log来更改数据。这也是MongoDB用behavior log的原因，因为在binary log中，address是local的，那么P中的log的地址是P里的，即使是传给S，S还是无法找到数据。


### 3.如何处理大数据

#### 3.1 how to save 100 TB of documents?
当今主流的计算机硬件比较便宜而且可以扩展，因此对于海量的数据，可以把数据（比如100 TB） 存在不同的机器上，形成一个cluster。
 
在MongoDB中，使用sharding（分片）机制来在不同机器上存储资料。每个shard（碎片）都是一个独立的资料库，很多个shards可以组成一个资料库。比如一个1 TB的collection可以分 成4个shard，每个shard存256 GB。如果分成40个shard，那么每个shard只需管理25 GB的资料。
 
#### 3.2 how to save document of 100TB?
如果一个document就有100 TB，那么要如何存储呢？我们可以把100 TB分成小的数据块。拆成255k每块。为什么不用256k呢？这是因为我们要存metadata，如果用256k，那么就没有空间存metadata。
 
从前面的这些介绍可以看出，每种数据结构或者技术都有它产生的原因。就像MongoDB的产生，就是因为现今的数据量越来越大，传统的SQL在处理海量数据时有它的局限性。为了应对各种新的问题，MongoDB才逐渐发展壮大。


### 4.mongo双活架构
当组织考虑在多个跨数据中心（或区域云）部署应用时，他们通常会希望使用“双活”的架构，即所有数据中心的应用服务器同时处理所有的请求。


//todo

### 4. mongo 性能测试

#### 4.1 单机mongo tps/qps 性能测试
鉴于了解较少，我们先简单的测试工具来测试mongo单机写入能力，于是找到了**mongo-mload**(三年前更新的代码，但是不影响简单的测试)

插入1000000条数据，mongo tps性能

![mongostat单机](./mongostat单机.png) 

对数据进行查询，mongo qps性能

![单机mongostat_qps数据](./单机mongostat_qps数据.png)

#### 4.2 sharding mongo tps/qps 性能测试







### 5.分区(分片)数据库

在Mongodb里面存在另一种集群，就是分片技术,可以满足MongoDB数据量大量增长的需求。

当MongoDB存储海量的数据时，一台机器可能不足以存储数据，也可能不足以提供可接受的读写吞吐量。这时，我们就可以通过在多台机器上分割数据，使得数据库系统能存储和处理更多的数据。

我么接下来操作一个分片实例：

#### 5.0 分片端口分布
```
Shard Server 1：27020
Shard Server 2：27021
Shard Server 3：27022
Shard Server 4：27023
Config Server ：27100
Route Process：40000
```
mongo分片集群：

![mongo分片集群](./mongo分片集群.png)


#### 5.1 启动Shard Server
```
mkdir -p /www/mongoDB/shard/s0
mkdir -p /www/mongoDB/shard/s1
mkdir -p /www/mongoDB/shard/s2
mkdir -p /www/mongoDB/shard/s3
mkdir -p /www/mongoDB/shard/log
mongod --port 27020 --dbpath=/www/mongoDB/shard/s0 --logpath=/www/mongoDB/shard/log/s0.log --logappend --fork
....
mongod --port 27023 --dbpath=/www/mongoDB/shard/s3 --logpath=/www/mongoDB/shard/log/s3.log --logappend --fork

```
#### 5.2 启动Config Server
```
mkdir -p /www/mongoDB/shard/config
mongod --port 27100 --dbpath=/www/mongoDB/shard/config --logpath=/www/mongoDB/shard/log/config.log --logappend --fork

```

#### 5.3 启动Route Process
```
/usr/bin/mongos --port 40000 --configdb conf/localhost:27100 --fork --logpath=/www/mongoDB/shard/log/route.log  &
```

#### 5.4 配置Sharding



### 99. 相关参考博客


// 深入浅出mongodb的设计与实现 https://yq.aliyun.com/articles/54424

// mongo（mongostate工具） 运作状态，性能监控分析  http://m.myexception.cn/go/1998284.html

// mongo-mload工具 https://github.com/eshujiushiwo/mongo-mload


//mongo分片概念和原理 https://blog.51cto.com/13941177/2309939 