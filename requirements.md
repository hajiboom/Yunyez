# AI使用文档

## 环境
- os: ubuntu 22.04
- go 1.24.8
- node 20.19.0
- redis
- docker:container
-- emqx 
-- postgresql

## 项目
名称：Yunyez  
描述：电子宠物解决方案，AI陪伴机器人，提供情绪价值。  
使用场景：
- 家庭宠物陪伴
- 远程宠物监控
- 旅行搭子陪伴 + 旅行记录打卡

参考产品：
- 小智AI

截至目前项目结构如下：
```
02:46:53 pp@bb Yunyez ±|feat/mqtt ✗|→ tree
.
├── api
│   └── proto # protobuf 定义
├── cmd # 服务启动入口
│   ├── device # 设备服务
│   │   └── device.go
│   └── web # web 服务
├── configs # 项目配置文件
│   ├── config.yaml # 全局配置
│   ├── dev # 开发环境配置
│   │   ├── database.yaml
│   │   └── mqtt.yaml
│   ├── device.yaml # 设备配置（公共）
│   ├── pre # 预发布环境配置
│   └── test # 测试环境配置
├── docs # 项目文档
│   ├── checkList # 代码审查要点
│   │   └── check.md
│   ├── error_records.md # 错误码对照表
│   ├── protocol.md # 自定义协议文档
│   └── sql # 数据库 SQL 脚本
│       ├── default.sql # 默认数据库脚本
│       └── device # 设备数据库脚本
│           └── device.sql
├── example # 示例代码
│   └── mock # mock代码
│       ├── virtual_capture # mock摄像
│       └── virtual_voice # mock语音
│           ├── device_voice 
│           ├── device_voice.c # 设备语音处理函数（设备端）
│           └── voice_proto.h # 设备语音处理头文件
├── fronted # 前端代码
│   ├── admin # 管理平台前端目录
│   └── v1 # 其余前端目录 v1
│       └── welcome.html
├── go.mod
├── go.sum
├── internal # 项目内部代码
│   ├── app # 各服务主函数
│   │   └── device
│   │       ├── app.go
│   │       └── http.go
│   ├── common # 项目通用代码
│   │   ├── config # 配置相关代码
│   │   │   └── config.go
│   │   ├── constant # 常量相关代码
│   │   │   ├── default.go
│   │   │   └── error.go
│   │   ├── frequency # 频率相关代码
│   │   └── tools # 工具相关代码
│   │       ├── context.go
│   │       ├── file.go
│   │       └── file_test.go
│   ├── controller # 控制器相关代码
│   │   ├── deviceManage # 设备管理控制器
│   │   │   ├── delete.go
│   │   │   ├── fetch.go
│   │   │   └── update.go
│   │   └── voiceManage # 语音管理控制器
│   ├── middleware # 中间件相关代码
│   ├── model # 数据库模型相关代码
│   │   ├── device # 设备相关数据库模型
│   │   │   └── device.go
│   │   └── image # 图像相关数据库模型
│   │       └── imageMessage.go
│   ├── pkg # 项目中间件通用代码
│   │   ├── http # http 相关代码
│   │   │   ├── common
│   │   │   └── middleware
│   │   │       ├── authMiddleware.go
│   │   │       └── frequencyMiddleware.go
│   │   ├── logger # 日志相关代码
│   │   │   ├── default.go 
│   │   │   └── gorm.go
│   │   ├── mqtt # mqtt 相关代码
│   │   │   ├── connect.go
│   │   │   ├── constant
│   │   │   │   └── constant.go
│   │   │   ├── core
│   │   │   │   ├── client.go
│   │   │   │   ├── mqtt.go
│   │   │   │   └── topic.go
│   │   │   ├── handler
│   │   │   │   └── forward.go
│   │   │   ├── middleware
│   │   │   │   ├── device_handler.go
│   │   │   │   └── middleware.go
│   │   │   └── protocol
│   │   │       └── voice
│   │   │           ├── message.go
│   │   │           ├── voice.go
│   │   │           └── voice_test.go
│   │   ├── postgre # postgresql 相关代码
│   │   │   └── db.go
│   │   ├── redis # redis 相关代码
│   │   └── websocket # websocket 相关代码
│   ├── service # 服务相关代码
│   │   ├── device # 设备相关服务
│   │   │   ├── device.go # 设备服务
│   │   │   └── device_test.go
│   │   └── voice # 语音相关服务
│   └── types # 数据类型相关代码
│       ├── common # 通用数据类型
│       │   └── default.go
│       └── device # 设备相关数据类型
│           ├── device.go
│           └── dto.go
├── pkg # 外部依赖代码
├── README.md 
├── requirements.md
├── storage # 项目存储目录
│   ├── logs # 日志存储目录
│   └── tmp # 临时文件存储目录
└── test # 集成测试目录
```

## 行为
1. 在用户给出一个开发需求的时候，列举出你的技术规格说明(spec)给用户评审。

## 备注
1. 项目采用微服务架构，每个服务都有自己的数据库。
2. 项目采用 RESTful API 设计，所有接口都返回 JSON 格式数据。
3. 项目采用 MQTT / UDP 等协议进行设备与服务端通信。
4. 特别备注：前端代码有其他小伙伴在写 ./fronted/admin 目录下，我们测试时的前端代码你只需要关注 ./fronted/v1 目录下的代码即可。