# 裸机多活高可用部署指南

适用场景：直接在物理机/虚拟机上部署后端二进制，前端打包后放在 nginx 机器上做静态托管，nginx 同时承担负载均衡。**不使用 Docker**。

---

## 一、部署拓扑

```
                    ┌─────────────┐
        用户 ──────► │   nginx     │  (静态前端 + 反向代理)
                    │  机器 (80)   │
                    └──────┬──────┘
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │backend-a │  │backend-b │  │backend-c │   每台跑一个 drill-server
        │ :8080    │  │ :8080    │  │ :8080    │   不同 INSTANCE_ID
        └────┬─────┘  └────┬─────┘  └────┬─────┘
             │             │             │
             └─────────────┼─────────────┘
                           ▼
                   ┌───────────────┐
                   │ MySQL + Redis │  (外部 HA，多 backend 共享)
                   └───────────────┘
```

**两种典型拓扑**：

| 拓扑 | 适用场景 | nginx 机器 | backend 机器 |
|---|---|---|---|
| 多机多实例 | 中等规模生产 | 1 台 | ≥ 2 台（每台 1 个 drill-server，端口 8080） |
| 单机多实例 | 小规模/测试 | 1 台 | 1 台，跑 3 个 drill-server 监听 8081/8082/8083 |

---

## 二、前置条件

### 1. 外部依赖（必须先准备）

- **MySQL HA**：MGR / RDS / ProxySQL / VIP，backend 通过单一入口连接
- **Redis Cluster**：≥ 3 主节点，backend 用 `REDIS_CLUSTER_ADDRS` 连接

### 2. 软件版本

| 组件 | 版本 |
|---|---|
| Go | 1.23+（仅编译机需要） |
| Node.js | 18+（仅编译机需要，用于构建前端） |
| nginx | 1.18+（带 `http_realip` 和 `stream` 模块） |
| systemd | 任何现代 Linux 自带 |
| MySQL client | 8.0+（执行迁移用，可选） |

---

## 三、编译产物（在编译机上执行）

```bash
# 假设项目根目录为 ~/drill-platform
cd ~/drill-platform

# 1. 编译后端二进制（Linux amd64）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o deploy/drill-server ./cmd/server

# 2. 构建前端
cd web
npm install
npm run build
# 产物在 web/dist/

# 3.（可选）编译 deploy-ha 部署助手
cd ..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o deploy/deploy-ha ./cmd/deploy-ha
```

产物清单：

```
deploy/
├── drill-server      # 后端二进制（每台 backend 机器一份）
├── deploy-ha         # 部署助手（可选，仅迁移/健康检查时需要）
└── web/dist/         # 前端静态文件（仅 nginx 机器需要）
```

---

## 四、数据库迁移（在任意能连 MySQL 的机器上执行）

### 方式 A：用 deploy-ha 工具（推荐）

```bash
cd ~/drill-platform

# 准备 .env
cp .env.example .env
vim .env   # 填写 DATABASE_HOST / DATABASE_PORT / DATABASE_USER / DATABASE_PASSWORD / DATABASE_NAME

# 执行迁移（幂等，可重复执行）
go run ./cmd/deploy-ha migrate
# 或用预编译二进制
./deploy/deploy-ha migrate
```

输出末尾应看到：

```
[4/4] 校验多活 schema 关键对象
  ✓ drill_flow_command 表
  ✓ drill_worker_epoch 表
  ...
  → 多活表结构校验通过

✓ 数据库迁移完成
```

### 方式 B：直接执行 SQL（无 Go 环境）

```bash
mysql -h <MYSQL_HOST> -P 3306 -u <USER> -p<PASS> drill_platform \
  < scripts/migration/2026-07-05-migrate-to-multi-active.sql
```

---

## 五、部署后端实例

### 1. 拷贝二进制到每台 backend 机器

```bash
# 在编译机上
scp deploy/drill-server user@backend-a:/opt/drill/
scp deploy/drill-server user@backend-b:/opt/drill/
scp deploy/drill-server user@backend-c:/opt/drill/

# 同时拷贝一份 .env 模板到每台机器
scp .env.example user@backend-a:/opt/drill/.env
```

### 2. 在每台 backend 机器上配置 .env

每台机器创建 `/opt/drill/.env`，**INSTANCE_ID 各不相同**：

**backend-a 的 .env**：
```ini
# 应用角色：多活模式下每台都是 all（同时跑 API + Worker）
APP_ROLE=all

# ★ 每台机器必须不同！
INSTANCE_ID=backend-a

SERVER_MODE=release

# 数据库（指向外部 HA 入口）
DATABASE_HOST=mysql-vip.example.com
DATABASE_PORT=3306
DATABASE_USER=drill
DATABASE_PASSWORD=<your-db-password>
DATABASE_NAME=drill_platform

# Redis Cluster（多节点逗号分隔）
REDIS_CLUSTER_ADDRS=redis-1:6379,redis-2:6379,redis-3:6379
REDIS_PASSWORD=<your-redis-password>
REDIS_TLS=false

# JWT 密钥（所有 backend 必须相同！）
JWT_SECRET=<your-strong-jwt-secret>

# 公共访问 URL（用于 CAS 回调等）
PUBLIC_BASE_URL=https://drill.example.com

# 登录日志走 stdout，由 systemd/journald 收集
LOGIN_LOG_FILE=
```

**backend-b / backend-c** 只需把 `INSTANCE_ID` 改为 `backend-b` / `backend-c`，其余完全相同。

> ⚠️ **JWT_SECRET 必须所有实例相同**，否则不同节点签发的 token 互不可验证。
> ⚠️ **INSTANCE_ID 必须所有实例不同**，否则 leader 选举会冲突。

### 3. 创建 systemd unit 文件

在每台 backend 机器上创建 `/etc/systemd/system/drill-server.service`：

```ini
[Unit]
Description=Drill Platform Backend (HA)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=drill
Group=drill

# 工作目录与二进制
WorkingDirectory=/opt/drill
ExecStart=/opt/drill/drill-server

# 从 .env 加载环境变量
EnvironmentFile=/opt/drill/.env

# 优雅关闭：发 SIGTERM，等待 30s
KillSignal=SIGTERM
TimeoutStopSec=30s

# 自动重启
Restart=always
RestartSec=5s

# 资源限制
LimitNOFILE=65536

# 日志走 journald
StandardOutput=journal
StandardError=journal
SyslogIdentifier=drill-server

[Install]
WantedBy=multi-user.target
```

### 4. 启动服务

在每台 backend 机器上：

```bash
# 创建专用用户（首次）
sudo useradd -r -s /sbin/nologin -d /opt/drill drill
sudo chown -R drill:drill /opt/drill

# 加载 + 启动
sudo systemctl daemon-reload
sudo systemctl enable drill-server
sudo systemctl start drill-server

# 查看状态
sudo systemctl status drill-server
sudo journalctl -u drill-server -f --no-pager
```

启动日志中应看到：

```
应用角色: all, 实例ID: backend-a
Worker 角色已启动，参与领导选举...
HTTP server listening on :8080
```

### 5. 验证 /ready

```bash
# 在每台 backend 机器本地
curl -s http://localhost:8080/ready | jq .

# 期望响应（leader 节点）：
# { "ready": true, "worker_status": "leader-ready", ... }
# 期望响应（standby 节点）：
# { "ready": true, "worker_status": "standby-ready", ... }
```

> 3 个节点中应有恰好 1 个 `leader-ready`，其余为 `standby-ready`。
> 如果全是 `standby-ready`，等 15-30s 等选举完成后再检查。

---

## 六、配置 nginx

### 1. 部署前端静态文件到 nginx 机器

```bash
# 在编译机
scp -r web/dist user@nginx-machine:/var/www/drill-platform/

# 在 nginx 机器
sudo mkdir -p /var/www/drill-platform
sudo chown -R nginx:nginx /var/www/drill-platform
```

### 2. 配置 nginx

创建 `/etc/nginx/conf.d/drill-platform.conf`：

```nginx
# WebSocket 升级映射
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}

upstream drill_backend {
    least_conn;

    # ★ 改成你的 backend 机器 IP（多机多实例模式）
    # 若是单机多实例模式，用 127.0.0.1:8081/8082/8083
    server 10.0.0.11:8080 max_fails=3 fail_timeout=10s;  # backend-a
    server 10.0.0.12:8080 max_fails=3 fail_timeout=10s;  # backend-b
    server 10.0.0.13:8080 max_fails=3 fail_timeout=10s;  # backend-c

    keepalive 32;
}

server {
    listen 80;
    server_name drill.example.com;  # ★ 改成你的域名

    # 前端 SPA 静态托管
    root /var/www/drill-platform/dist;
    index index.html;

    # SPA 路由：所有未匹配文件的路由都回退到 index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存（带 hash 的文件名可长期缓存）
    location ~* \.(?:js|css|woff2?|ttf|otf|eot|svg|png|jpg|jpeg|gif|ico)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # API 反向代理（保留 /api/v1 前缀，不要加尾随 /）
    location /api/ {
        proxy_pass http://drill_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # WebSocket 反向代理（长连接，超时调长）
    location /ws/ {
        proxy_pass http://drill_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }

    # 健康检查入口（被 LB / 监控用）
    location /health {
        proxy_pass http://drill_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_connect_timeout 5s;
        proxy_read_timeout 5s;
    }

    # 错误页面
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

### 3. 测试 + 重载 nginx

```bash
sudo nginx -t              # 语法检查
sudo systemctl reload nginx
```

---

## 七、端到端验证

### 1. 入口连通性

```bash
# 通过 nginx 访问
curl -s http://drill.example.com/health | jq .
curl -s http://drill.example.com/api/v1/auth/captcha | head -c 200
```

### 2. leader 选举正确性

3 个 backend 中应**恰好 1 个**是 leader：

```bash
for host in 10.0.0.11 10.0.0.12 10.0.0.13; do
    echo -n "$host: "
    curl -s http://$host:8080/ready | jq -r '.worker_status'
done

# 期望输出（顺序无关）：
# 10.0.0.11: leader-ready
# 10.0.0.12: standby-ready
# 10.0.0.13: standby-ready
```

### 3. 故障切换验证

```bash
# 找到当前 leader，假设是 backend-a
LEADER_IP=10.0.0.11

# 停掉 leader
ssh user@$LEADER_IP sudo systemctl stop drill-server

# 立即访问应用，应仍可用（standby 接管）
curl -s http://drill.example.com/health | jq .

# 等待 15-30s，新 leader 应被选出
for host in 10.0.0.11 10.0.0.12 10.0.0.13; do
    echo -n "$host: "
    curl -s http://$host:8080/ready 2>/dev/null | jq -r '.worker_status // "down"'
done

# 恢复原 leader
ssh user@$LEADER_IP sudo systemctl start drill-server
```

### 4. 前端页面

浏览器打开 `http://drill.example.com/`，应能看到登录页。

---

## 八、运维命令速查

```bash
# 查看某节点状态
ssh user@10.0.0.11 sudo systemctl status drill-server
ssh user@10.0.0.11 sudo journalctl -u drill-server -f --no-pager

# 重启某节点（优雅关闭，30s 内完成）
ssh user@10.0.0.11 sudo systemctl restart drill-server

# 滚动重启（每次只重启一个，确认 ready 后再重启下一个）
for host in 10.0.0.11 10.0.0.12 10.0.0.13; do
    ssh user@$host sudo systemctl restart drill-server
    # 等待 ready
    until curl -sf http://$host:8080/ready >/dev/null; do sleep 1; done
done

# 更新二进制（滚动）
for host in 10.0.0.11 10.0.0.12 10.0.0.13; do
    scp deploy/drill-server user@$host:/opt/drill/drill-server.new
    ssh user@$host bash -c '
        sudo mv /opt/drill/drill-server.new /opt/drill/drill-server
        sudo chown drill:drill /opt/drill/drill-server
        sudo systemctl restart drill-server
    '
    until curl -sf http://$host:8080/ready >/dev/null; do sleep 1; done
done

# 更新前端
cd web && npm run build
scp -r dist/* user@nginx-machine:/var/www/drill-platform/dist/
ssh user@nginx-machine sudo nginx -s reload
```

---

## 九、常见问题

### Q1：所有节点都显示 `standby-ready`，没有 leader？

等待 15-30s（默认租约 TTL）。若仍无 leader：

- 检查 Redis Cluster 是否可达：`redis-cli -c -h <redis-host> -p 6379 -a <pass> ping`
- 检查 .env 中 `REDIS_CLUSTER_ADDRS` 是否逗号分隔多个节点
- 查看 backend 日志：`journalctl -u drill-server -f`

### Q2：前端访问 404 / 刷新后 404？

确认 nginx 配置了 SPA 回退：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### Q3：WebSocket 连接断开？

- 确认 nginx 的 `/ws/` location 配置了 `Upgrade` / `Connection` 头
- 确认 `proxy_read_timeout` ≥ 300s
- 检查 firewall 是否允许长连接

### Q4：登录后 token 在另一个节点验证失败？

`JWT_SECRET` 在所有 backend 实例必须完全相同。检查每台机器的 .env。

### Q5：可以单机跑 3 个实例吗？

可以。把 3 个实例监听不同端口（8081/8082/8083），nginx upstream 改成 `127.0.0.1:8081` 等。每个实例用不同的 systemd unit（如 `drill-server@a.service` 用 systemd template unit）。

template unit `/etc/systemd/system/drill-server@.service`：

```ini
[Service]
EnvironmentFile=/opt/drill/.env.%i
ExecStart=/opt/drill/drill-server --instance %i
...
```

启动：
```bash
sudo systemctl start drill-server@a
sudo systemctl start drill-server@b
sudo systemctl start drill-server@c
```

每个 `.env.a` / `.env.b` / `.env.c` 的 `INSTANCE_ID` 不同。

---

## 十、安全建议（生产环境）

1. **HTTPS**：在 nginx 启用 TLS（Let's Encrypt 或自签证书），后端仍走 HTTP
2. **网络隔离**：backend 机器仅允许 nginx 机器访问 8080 端口
3. **数据库账号**：backend 用专用账号，仅授予 drill_platform 库的 CRUD 权限
4. **JWT_SECRET**：≥ 32 字节随机串，`openssl rand -hex 32` 生成
5. **防火墙**：backend 机器关闭公网入站，仅开放 SSH + 内网 8080
6. **日志轮转**：配置 logrotate 或依赖 journald 自动轮转
