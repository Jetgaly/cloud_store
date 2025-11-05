# gin-cloud-store

## 简单架构图

![](https://raw.githubusercontent.com/Jetgaly/cloud_store/refs/heads/master/imgs/structure.png)

## desc

CloudStore 是一个支持 **大文件上传、断点续传、秒传、模糊搜索与异步存储** 的云端文件服务系统。系统采用 **MinIO 临时层 + OSS 冷存层** 的架构，通过 **Redis、RabbitMQ、Canal、Elasticsearch** 等组件构建高性能、高扩展性的分布式文件平台。

```bash
jjet@jet:~/projects/GoProjects/cloud_store$ tree -L 1
.
├── api
├── conf.yml 
├── config
├── core          
├── cron          #cron task
├── global        #global dependency
├── go.mod
├── go.sum
├── main.go       #entry
├── middleware    #jwt,limitBucket
├── model         #mysql model
├── readme.md
├── router        
├── scripts
├── temp
├── test
└── utils
```

该项目使用**gin**框架实现了一个**分布式云盘项目**，主要运用到：

1. golang
2. gin
3. rabbitmq
4. minio
5. mysql
6. redis
7. es
8. docker
9. nginx
10. canal
11. aliyunOSS
12. JWT

## detail

**1. 用户体系**

- 基于邮箱验证码注册与登录
- JWT + 中间件实现用户认证与权限校验
- 自动刷新 Token、权限路由控制

**2. 接口安全与限流**

- 基于 **Redis + 分布式令牌桶** 实现接口限流
- 防止恶意刷接口、暴力请求，提高系统稳定性

**3. 文件上传体系**

- 支持 **大文件分片上传**
- 断点续传、失败重传、超时自动清理
- 客户端 `sha256` + Redis 分布式锁实现：
  - **秒传**（命中已存在文件直接返回）
  - **去重**（避免重复计算和重复上传）
- MinIO 作为分片与临时对象层
- 最终通过 **异步任务合并并上传至 OSS**，显著降低前端等待时长

**4. 文件下载体系（支持在线预览）**

- 支持 **HTTP Range** 请求
- 返回 `206 Partial Content`，仅传输指定字节段
- 浏览器 video/audio 可直接预览大文件媒体内容
- 支持断点续传，失败后可从中断位置恢复下载
- 后端流式读取文件，避免大文件占用过多内存

**5. 存储架构设计**

- **MinIO 本地存储集群**：用于分片、临时文件、高频访问文件
- **阿里云 OSS 冷数据层**：通过异步迁移降低长期存储成本
- **RabbitMQ** 解耦上传流程，实现文件的 **异步合并与转存 OSS**
- **Canal + MySQL Binlog**：
  - 监听增量变化
  - 实时同步至 Elasticsearch
  - 保证搜索索引延迟低、数据一致性高

**6. 文件检索搜索系统**

- Elasticsearch 实现文件名称、关键词、链路字段的 **模糊查询**



