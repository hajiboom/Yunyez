# Yunyez 项目云服务器部署指南

## 1. 服务器准备

### 1.1 系统要求
- **操作系统**: Ubuntu 22.04 LTS
- **CPU**: 至少2核 (推荐4核以上)
- **内存**: 至少4GB RAM (推荐8GB以上)
- **存储**: 至少20GB可用空间
- **网络**: 稳定的互联网连接

### 1.2 安全设置
1. 创建非root用户并配置sudo权限
2. 配置SSH密钥认证
3. 配置防火墙 (ufw)

```bash
# 创建新用户
adduser deploy
usermod -aG sudo deploy

# 配置SSH密钥认证
mkdir -p /home/deploy/.ssh
echo "你的公钥内容" >> /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys

# 配置防火墙
ufw allow OpenSSH
ufw allow 8080    # 主服务端口
ufw allow 18083   # EMQX Dashboard
ufw enable
```

## 2. 项目部署

### 2.1 上传项目文件
将项目文件上传到服务器上的 `/home/deploy/yunyez` 目录：

```bash
# 登录到服务器
ssh deploy@your_server_ip

# 创建项目目录
mkdir -p ~/yunyez
cd ~/yunyez

# 上传项目文件 (使用scp或git clone)
# 示例：使用git clone
git clone https://github.com/your-username/yunyez.git .
```

### 2.2 运行部署脚本
使用提供的部署脚本自动化部署：

```bash
# 给脚本执行权限
chmod +x deploy.sh

# 运行部署脚本
./deploy.sh
```

### 2.3 手动部署步骤 (备选)
如果需要手动部署，请按照以下步骤：

1. 安装 Docker 和 Docker Compose
2. 配置环境变量
3. 启动服务

```bash
# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker deploy

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 创建环境变量文件
cat > .env << EOF
POSTGRES_PASSWORD=your_secure_password
MONGO_USER=mongo
MONGO_PASSWORD=your_mongo_password
EOF

# 启动服务
docker-compose -f docker-compose.prod.yml up -d
```

## 3. 服务配置

### 3.1 数据库配置
- **PostgreSQL**: 用于存储结构化数据 (用户、景点、打卡记录等)
- **MongoDB**: 用于存储非结构化数据 (图片、音频等媒体文件)
- **Redis**: 用于缓存热点数据 (设备状态、会话信息等)

### 3.2 MQTT 配置
- **EMQX**: 作为 MQTT Broker 处理设备通信
- 默认端口: 1883 (MQTT), 8083 (MQTT/SSL), 8883 (MQTT/SSL), 18083 (Dashboard)

### 3.3 应用配置
修改 `configs/prod/` 目录下的配置文件以适应生产环境：

- `database.yaml`: 数据库连接配置
- `mqtt.yaml`: MQTT 服务配置
- `config.yaml`: 应用基本配置

## 4. 服务管理

### 4.1 查看服务状态
```bash
docker-compose -f docker-compose.prod.yml ps
```

### 4.2 查看服务日志
```bash
# 查看所有服务日志
docker-compose -f docker-compose.prod.yml logs

# 实时查看日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f yunyez_service
```

### 4.3 重启服务
```bash
docker-compose -f docker-compose.prod.yml restart
```

### 4.4 停止服务
```bash
docker-compose -f docker-compose.prod.yml down
```

### 4.5 更新服务
```bash
# 拉取最新代码
git pull origin main

# 重新构建并启动服务
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
```

## 5. 监控和维护

### 5.1 系统资源监控
```bash
# 查看容器资源使用情况
docker stats

# 查看系统资源使用情况
htop
df -h
```

### 5.2 日志管理
- 应用日志位于 `storage/logs/` 目录
- 定期清理旧日志以节省磁盘空间
- 考虑使用日志轮转工具如 logrotate

### 5.3 数据备份
定期备份数据库和重要文件：

```bash
# 备份 PostgreSQL
docker exec yunyez_postgres pg_dump -U postgres yunyez > backup_$(date +%Y%m%d_%H%M%S).sql

# 备份 MongoDB
docker exec yunyez_mongodb mongodump --out /tmp/mongo_backup_$(date +%Y%m%d_%H%M%S)
```

## 6. 故障排除

### 6.1 服务无法启动
1. 检查日志输出
```bash
docker-compose -f docker-compose.prod.yml logs
```

2. 检查端口占用
```bash
netstat -tlnp | grep :8080
netstat -tlnp | grep :1883
```

### 6.2 数据库连接问题
1. 检查数据库服务是否运行
2. 检查网络连接
3. 验证凭据配置

### 6.3 MQTT 连接问题
1. 检查 EMQX 服务状态
2. 验证设备认证信息
3. 检查防火墙设置

## 7. 安全建议

1. **使用强密码**: 为所有服务设置强密码
2. **定期更新**: 定期更新系统和容器镜像
3. **限制访问**: 通过防火墙限制不必要的端口访问
4. **启用认证**: 为数据库和 MQTT 服务启用认证
5. **监控日志**: 定期检查安全相关日志
6. **备份策略**: 实施定期备份和恢复测试

## 8. 性能优化

1. **调整数据库参数**: 根据实际负载调整数据库连接池大小
2. **缓存策略**: 合理使用 Redis 缓存减少数据库压力
3. **负载均衡**: 对于高并发场景，考虑使用负载均衡器
4. **CDN**: 对于静态资源，考虑使用 CDN 加速