## 1. redis快照

通过RDB和AOF进行快照同步

**RDB**

1.时间间隔轮训

2.手动操作，SAVE & BGSAVE，其中SAVE是阻塞客户端请求，不可以高可用，阻塞县城，BGSAVE另开线程处理

//todo https://mp.weixin.qq.com/s/PJ1-D9XK3pd7fWUUpm4FyQ