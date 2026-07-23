# 新机器裸机部署流程

本文用于在新的 Linux 机器上直接部署 Drill Platform，不使用 Docker。部署形态为：

```text
浏览器
  │
  ▼
宿主机 nginx :80/:443
  ├── /          → /var/www/drill-platform
  ├── /api/      → 127.0.0.1:8080
  ├── /ws/       → 127.0.0.1:8080
  └── /ready     → 127.0.0.1:8080
                         │
                         ├── 外部 MySQL
                         └── 外部 Redis Cluster
```

假设：

- MySQL 数据库已经创建，但库内还没有任何表；
- Redis Cluster 已经部署完成；
- 后端由 systemd 托管；
- 前端静态文件由宿主机 nginx 提供；
- 当前先部署一个 `APP_ROLE=all` 实例。

> 单机部署不能抵御整台机器故障。以后扩展为多机时，每台机器使用相同的 MySQL、Redis、JWT 配置，但 `INSTANCE_ID` 必须唯一。

## 一、准备部署参数

上线前确认：

| 参数 | 示例 | 说明 |
|---|---|---|
| 应用域名 | `drill.example.com` | 没有域名时可先用机器 IP |
| MySQL 地址 | `mysql-vip.example.com:3306` | 新机器必须可达 |
| 数据库名 | `drill_platform` | 初始化脚本默认使用此名称 |
| 数据库迁移账号 | `drill_migrate` | 首次建表需要 DDL、INDEX、INSERT 权限 |
| 数据库运行账号 | `drill_app` | 建表后建议只保留业务 DML 权限 |
| Redis Cluster | `redis-1:6379,redis-2:6379,redis-3:6379` | 英文逗号分隔 |
| Redis 用户名 | `default` | 未启用 ACL 时留空 |
| Redis 密码 | `<secret>` | 不写入 Git |
| Redis TLS | `true` 或 `false` | 与 Redis 实际配置一致 |
| 登录方式 | 本地账号或 CAS/LDAP | 上线前必须确定 |

## 二、准备新机器

### 1. 安装依赖

运行环境需要：

- Linux x86_64；
- systemd；
- nginx；
- MySQL 8 客户端；
- ca-certificates、tzdata；
- curl、jq、openssl；
- 用于构建时的 Go 1.23、Node.js 20 和 npm。

如果二进制和前端由 CI 或其他编译机生成，生产机器不必安装 Go 和 Node.js。

检查：

```bash
systemctl --version
nginx -v
mysql --version
curl --version
```

在编译机检查：

```bash
go version
node --version
npm --version
```

### 2. 检查网络

生产机器必须能够访问：

- MySQL 服务地址和端口；
- Redis Cluster 公布的每个节点地址，而不只是种子节点；
- CAS/LDAP 服务地址（如果启用）；
- DNS 和 NTP 服务。

外部只需开放：

- `80/443`：提供 Web 服务；
- SSH 运维端口：仅允许受信网段。

后端 `8080` 建议只监听或只允许本机访问，不直接暴露公网。

### 3. 创建运行用户和目录

```bash
sudo useradd -r -s /sbin/nologin -d /opt/drill drill

sudo install -d -o drill -g drill -m 0750 /opt/drill
sudo install -d -o drill -g drill -m 0750 /opt/drill/bin
sudo install -d -o drill -g drill -m 0750 /opt/drill/configs
sudo install -d -o root -g root -m 0755 /var/www/drill-platform
```

## 三、验证 MySQL 和 Redis

### 1. 验证空数据库

```bash
mysql --protocol=TCP \
  -h <mysql-host> -P 3306 \
  -u <migration-user> -p \
  -D drill_platform \
  -e "SELECT VERSION(), DATABASE(); SHOW TABLES;"
```

首次初始化前，`SHOW TABLES` 应为空。如果已经有表，必须停止并确认数据来源，不能继续执行初始化脚本。

### 2. 验证 Redis Cluster

如果机器上安装了 `redis-cli`：

```bash
REDISCLI_AUTH='<redis-password>' redis-cli -c \
  -h <redis-node-1> -p 6379 \
  PING

REDISCLI_AUTH='<redis-password>' redis-cli -c \
  -h <redis-node-1> -p 6379 \
  CLUSTER INFO

REDISCLI_AUTH='<redis-password>' redis-cli -c \
  -h <redis-node-1> -p 6379 \
  CLUSTER NODES
```

启用 ACL 时增加 `--user <redis-username>`，启用 TLS 时增加 `--tls`。

必须满足：

- `PING` 返回 `PONG`；
- `CLUSTER INFO` 包含 `cluster_state:ok`；
- 新机器能够访问 `CLUSTER NODES` 返回的每个节点地址；
- ACL 允许连接探测、GET/SET/DEL、Lua、Leader 租约和 Pub/Sub；
- Redis 不与另一个 Drill Platform 环境共用，避免固定的 `drill:*` Key 和频道冲突。

## 四、编译发布产物

推荐在编译机或 CI 上从固定 tag/commit 构建，不要直接部署不断变化的开发分支。

### 1. 获取固定版本

```bash
git clone <repository-url> drill-platform
cd drill-platform
git checkout <release-tag-or-commit>
git rev-parse HEAD
```

记录 commit ID，作为上线和回退依据。

### 2. 编译后端

目标机器为 Linux x86_64：

```bash
mkdir -p build

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" \
  -o build/drill-server ./cmd/server
```

如果目标机器是 ARM64，将 `GOARCH=amd64` 改为 `GOARCH=arm64`。

检查：

```bash
file build/drill-server
```

### 3. 编译前端

```bash
cd web
npm ci
npm run build
cd ..

test -f web/dist/index.html
```

### 4. 上传文件

将以下文件复制到新机器：

```text
build/drill-server       → /opt/drill/bin/drill-server
web/dist/                → /var/www/drill-platform/
configs/config.yaml      → /opt/drill/configs/config.yaml
scripts/init-db.sql      → /opt/drill/init-db.sql
```

安装后端二进制：

```bash
sudo install -o drill -g drill -m 0755 \
  build/drill-server /opt/drill/bin/drill-server
```

复制前端后保证 nginx 可读：

```bash
sudo cp -a web/dist/. /var/www/drill-platform/
sudo chown -R root:root /var/www/drill-platform
sudo chmod -R a+rX /var/www/drill-platform
```

## 五、首次初始化数据库

### 1. 重要限制

`scripts/init-db.sql` 会：

- 执行 `DROP TABLE IF EXISTS`；
- 创建完整基础表；
- 写入初始用户和模板分类；
- 固定使用数据库名 `drill_platform`。

因此它只能在已经确认没有任何表的新数据库上执行一次，不能用于升级已有环境，也不能在服务运行后重复执行。

如果实际数据库名不是 `drill_platform`，必须复制并审核脚本中的 `CREATE DATABASE` 和 `USE` 语句后再执行。

### 2. 创建表

在代码目录执行：

```bash
mysql --protocol=TCP \
  -h <mysql-host> -P 3306 \
  -u <migration-user> -p \
  < scripts/init-db.sql
```

也可以在新机器使用已上传的脚本：

```bash
mysql --protocol=TCP \
  -h <mysql-host> -P 3306 \
  -u <migration-user> -p \
  < /opt/drill/init-db.sql
```

### 3. 验证表

```bash
mysql --protocol=TCP \
  -h <mysql-host> -P 3306 \
  -u <migration-user> -p \
  -D drill_platform \
  -e "SHOW TABLES;"
```

至少应看到：

```text
user
drill_template_category
drill_template
drill_template_step
drill_instance
drill_instance_step
drill_flow_command
drill_worker_epoch
drill_instance_step_log
drill_assignee
notification
```

检查多活关键字段：

```bash
mysql --protocol=TCP \
  -h <mysql-host> -P 3306 \
  -u <migration-user> -p \
  -D drill_platform \
  -e "
    SHOW COLUMNS FROM drill_flow_command LIKE 'worker_epoch';
    SHOW COLUMNS FROM drill_flow_command LIKE 'lease_token';
    SHOW COLUMNS FROM drill_instance_step_log LIKE 'command_id';
    SHOW COLUMNS FROM notification LIKE 'command_id';
  "
```

> `deploy-ha migrate` 是已有基础表后的增量迁移工具，不能代替空数据库的首次初始化。

初始化完成后，建议撤销迁移账号的日常访问权限，服务只使用运行账号。

## 六、配置应用

### 1. 配置 `/opt/drill/configs/config.yaml`

后端以 `/opt/drill` 为工作目录时，会读取 `/opt/drill/configs/config.yaml`。

数据库、Redis、JWT 等敏感配置使用环境变量覆盖。YAML 主要配置登录方式、CAS/LDAP 和业务参数。

仓库默认 YAML 指向本地模拟 CAS/LDAP，不能原样用于生产。生产必须二选一：

1. 使用真实 CAS/LDAP：填写真实的 CAS、LDAP URL、Base DN、绑定账号和回调地址；
2. 使用本地登录：设置 `auth.mode: local`、`cas.enabled: false`、`ldap.enabled: false`。

使用 CAS 时，服务端 CAS 地址和 LDAP 参数放在 YAML；对外回调地址还可在 `.env` 中设置：

```dotenv
CAS_PUBLIC_URL=https://sso.example.com/cas
CAS_SERVICE_URL=https://drill.example.com/api/v1/auth/cas/callback
```

配置权限：

```bash
sudo chown drill:drill /opt/drill/configs/config.yaml
sudo chmod 600 /opt/drill/configs/config.yaml
```

### 2. 创建 `/opt/drill/.env`

```dotenv
APP_ROLE=all
INSTANCE_ID=backend-a
SERVER_MODE=release
SERVER_PORT=8080

DATABASE_HOST=<mysql-host>
DATABASE_PORT=3306
DATABASE_USER=<runtime-app-user>
DATABASE_PASSWORD=<runtime-app-password>
DATABASE_NAME=drill_platform

REDIS_CLUSTER_ADDRS=<redis-node-1>:6379,<redis-node-2>:6379,<redis-node-3>:6379
REDIS_USERNAME=
REDIS_PASSWORD=<redis-password>
REDIS_TLS=false

JWT_SECRET=<random-secret>
PUBLIC_BASE_URL=https://drill.example.com

LOGIN_LOG_FILE=
```

生成 JWT 密钥：

```bash
openssl rand -hex 32
```

然后：

```bash
sudo chown drill:drill /opt/drill/.env
sudo chmod 600 /opt/drill/.env
```

注意：

- `SERVER_MODE=release` 时 `JWT_SECRET` 不能为空；
- `INSTANCE_ID` 不能为空；
- 增加其他后端机器时，每台机器的 `INSTANCE_ID` 必须不同；
- 所有机器的 `JWT_SECRET`、MySQL 和 Redis 配置必须一致；
- `.env` 由 systemd 读取，不需要在终端中 `source`；
- 密码中不要包含未经验证的换行或 systemd EnvironmentFile 特殊格式。

## 七、配置 systemd

创建 `/etc/systemd/system/drill-server.service`：

```ini
[Unit]
Description=Drill Platform Backend
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=drill
Group=drill
WorkingDirectory=/opt/drill
EnvironmentFile=/opt/drill/.env
ExecStart=/opt/drill/bin/drill-server

KillSignal=SIGTERM
TimeoutStopSec=60s
Restart=always
RestartSec=5s
LimitNOFILE=65536

StandardOutput=journal
StandardError=journal
SyslogIdentifier=drill-server

[Install]
WantedBy=multi-user.target
```

加载配置：

```bash
sudo systemctl daemon-reload
sudo systemctl enable drill-server
```

先不要立即对外开放 nginx，先启动并验证后端：

```bash
sudo systemctl start drill-server
sudo systemctl status drill-server --no-pager
sudo journalctl -u drill-server -n 200 --no-pager
```

日志应包含：

```text
数据库连接成功
Redis连接成功 (mode=cluster, addr=...)
Worker 已启动
服务已就绪
```

如果出现 `Redis连接失败 (可忽略)`，不能继续上线。Redis 对 Leader 选举和事件分发是必需依赖。

## 八、验证后端

```bash
curl -fsS http://127.0.0.1:8080/live | jq .
curl -fsS http://127.0.0.1:8080/ready | jq .
curl -fsS http://127.0.0.1:8080/health | jq .
```

`/ready` 必须满足：

- `status` 为 `ready`；
- `components.db` 为 `ok`；
- `components.redis` 为 `ok`；
- `components.publisher` 为 `ok`；
- `components.subscriber` 为 `ok`；
- 单机 `APP_ROLE=all` 时，等待选举后 `worker_status` 为 `leader-ready`。

不要只看 HTTP 200。Redis 初始化失败时后端进程仍可能启动，必须同时检查组件字段和启动日志。

检查监听端口：

```bash
sudo ss -lntp
```

应看到后端监听 `8080`。

## 九、配置宿主机 nginx

将下面配置保存为 `/etc/nginx/conf.d/drill-platform.conf`。如果发行版使用 `sites-available/sites-enabled`，放到对应目录并建立启用链接。

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

upstream drill_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name drill.example.com;

    root /var/www/drill-platform;
    index index.html;
    client_max_body_size 10m;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://drill_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    location /ws/ {
        proxy_pass http://drill_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }

    location = /health {
        proxy_pass http://drill_backend;
        proxy_connect_timeout 5s;
        proxy_read_timeout 5s;
    }

    location = /live {
        proxy_pass http://drill_backend;
        proxy_connect_timeout 5s;
        proxy_read_timeout 5s;
    }

    location = /ready {
        proxy_pass http://drill_backend;
        proxy_connect_timeout 5s;
        proxy_read_timeout 5s;
    }
}
```

检查并启动：

```bash
sudo nginx -t
sudo systemctl enable --now nginx
sudo systemctl reload nginx
```

在启用 SELinux 的 RHEL/CentOS 系统上，如果 nginx 无法连接 `127.0.0.1:8080`，需要按组织安全规范允许 nginx 建立上游网络连接。

正式生产应在 nginx 或上游负载均衡配置 HTTPS，不建议直接通过公网提供明文 HTTP。

## 十、上线验收

### 1. 入口检查

本机：

```bash
curl -fsS http://127.0.0.1/health
curl -fsS http://127.0.0.1/ready | jq .
curl -I http://127.0.0.1/
```

配置域名和 HTTPS 后：

```bash
curl -fsS https://drill.example.com/health
curl -fsS https://drill.example.com/ready | jq .
curl -I https://drill.example.com/
```

`/health` 和 `/live` 只表示进程存活；负载均衡和依赖监控应使用 `/ready`。

### 2. 修改默认账号

初始化脚本会写入演示账号，初始密码为 `admin123`。允许普通用户访问前必须：

- 修改管理员密码；
- 删除或禁用不需要的演示账号；
- 使用 CAS/LDAP 时验证角色映射；
- 确认日志和配置文件没有泄漏密码。

### 3. 业务冒烟测试

至少完成一次：

1. 登录；
2. 创建或查看演练模板；
3. 创建测试演练；
4. 启动演练并执行一个步骤；
5. 确认操作状态已写入 MySQL；
6. 使用两个浏览器会话确认 WebSocket 状态实时同步；
7. 确认 journal 中没有数据库、Redis、Leader 或订阅错误。

只有进程存活、依赖健康和业务冒烟测试全部通过，才算部署完成。

## 十一、常用运维命令

查看状态：

```bash
sudo systemctl status drill-server --no-pager
```

持续查看日志：

```bash
sudo journalctl -u drill-server -f
```

修改 `.env` 或 YAML 后重启：

```bash
sudo systemctl restart drill-server
curl -fsS http://127.0.0.1:8080/ready | jq .
```

systemd 重启会重新读取 `EnvironmentFile`。只有修改 unit 文件时才需要再次执行 `systemctl daemon-reload`。

停止：

```bash
sudo systemctl stop drill-server
```

## 十二、版本更新与回退

更新前备份当前二进制：

```bash
sudo cp /opt/drill/bin/drill-server /opt/drill/bin/drill-server.previous
```

安装新二进制后：

```bash
sudo install -o drill -g drill -m 0755 \
  drill-server.new /opt/drill/bin/drill-server

sudo systemctl restart drill-server
curl -fsS http://127.0.0.1:8080/ready | jq .
```

回退：

```bash
sudo systemctl stop drill-server
sudo cp /opt/drill/bin/drill-server.previous /opt/drill/bin/drill-server
sudo chown drill:drill /opt/drill/bin/drill-server
sudo chmod 0755 /opt/drill/bin/drill-server
sudo systemctl start drill-server
curl -fsS http://127.0.0.1:8080/ready | jq .
```

数据库发生过结构升级时，回退前必须单独评估 schema 兼容性，不能默认执行反向 SQL。

## 十三、部署失败时的排查顺序

1. `systemctl status drill-server` 查看进程状态；
2. `journalctl -u drill-server -n 200` 查看启动错误；
3. 检查 `/opt/drill/.env` 权限和变量值；
4. 检查 `/opt/drill/configs/config.yaml` 是否仍指向模拟 CAS/LDAP；
5. 从新机器验证 MySQL 和全部 Redis 节点；
6. 检查 `/ready` 的具体失败组件；
7. 检查 nginx 配置、静态文件权限和上游连通性；
8. 最后执行真实业务冒烟测试。

不要通过重复执行 `init-db.sql` 解决启动问题，该脚本会删除并重建表。
