# API 接口文档

## 基础信息

- **Base URL**: `/api/v1`
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON
- **字符编码**: UTF-8

## 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码 (0=成功, 其他=错误) |
| message | string | 提示信息 |
| data | any | 响应数据 |

## 常见状态码

| 状态码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/Token 无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 503 | 在线状态服务暂不可用 |

---

## 1. 认证模块

### 1.1 用户登录

- **POST** `/api/v1/auth/login`

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGci...",
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### 1.2 刷新 Token

- **POST** `/api/v1/auth/refresh`

**请求头**: `Authorization: Bearer <token>`

**响应**:
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGci..."
  }
}
```

### 1.3 修改密码

- **POST** `/api/v1/auth/password`

**请求体**:
```json
{
  "old_password": "admin123",
  "new_password": "new_password_123"
}
```

### 1.4 在线心跳

- **POST** `/api/v1/auth/heartbeat`
- **权限**：所有已认证角色，无请求体。
- **响应 data**：`{"recorded": true}`。
- Redis 不可用时返回 HTTP 503，不将失败伪装为已记录。登录后的前端每 60 秒发送心跳，退出或卸载后停止；其他认证 API 请求也会尽力更新在线状态，但记录失败不影响业务响应。

---

## 2. 用户管理

### 2.1 获取用户列表

- **GET** `/api/v1/users?page=1&size=20&role=&keyword=`

**响应**:
```json
{
  "code": 0,
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "username": "admin",
        "nickname": "系统管理员",
        "role": "admin",
        "status": 1,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 2.2 创建用户

- **POST** `/api/v1/users`

**请求体**:
```json
{
  "username": "executor1",
  "nickname": "执行人一",
  "role": "executor",
  "password": "password123"
}
```

### 2.3 获取用户详情

- **GET** `/api/v1/users/{id}`

### 2.4 更新用户

- **PUT** `/api/v1/users/{id}`

### 2.5 删除用户

- **DELETE** `/api/v1/users/{id}`

---

## 3. 演练模板

### 3.1 模板列表

- **GET** `/api/v1/templates?page=1&size=20&category=`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |
| category | string | 否 | 分类 (灾备/降级/发布/安全) |
| keyword | string | 否 | 名称搜索 |

**分类类型**: `灾备` | `降级` | `发布` | `安全`

### 3.2 创建模板

- **POST** `/api/v1/templates`

**请求体**:
```json
{
  "name": "Redis 故障切换演练",
  "category": "灾备",
  "description": "模拟 Redis 主节点宕机，验证自动切换流程",
  "steps": [
    {
      "name": "停止 Redis 主节点",
      "type": "serial",
      "timeout": 30,
      "assignee_ids": [1, 2]
    },
    {
      "name": "验证从节点切换",
      "type": "serial",
      "timeout": 60,
      "depends_on": [1]
    }
  ]
}
```

### 3.3 模板详情

- **GET** `/api/v1/templates/{id}`

### 3.4 更新模板

- **PUT** `/api/v1/templates/{id}`

### 3.5 删除模板

- **DELETE** `/api/v1/templates/{id}`

---

## 4. 演练实例

### 4.1 实例列表

- **GET** `/api/v1/instances?page=1&size=20&status=&template_id=`

**状态枚举**: `pending` | `running` | `paused` | `completed` | `terminated`

### 4.2 创建演练实例

- **POST** `/api/v1/instances`

**请求体**:
```json
{
  "template_id": 1,
  "name": "2024-Q1 Redis 切换演练",
  "description": "首次验证 Redis 故障切换流程"
}
```

### 4.3 启动演练

- **POST** `/api/v1/instances/{id}/start`

### 4.4 暂停演练

- **POST** `/api/v1/instances/{id}/pause`

### 4.5 恢复演练

- **POST** `/api/v1/instances/{id}/resume`

### 4.6 终止演练

- **POST** `/api/v1/instances/{id}/terminate`

### 4.7 实例详情

- **GET** `/api/v1/instances/{id}`

**响应**:
```json
{
  "code": 0,
  "data": {
    "id": 100,
    "name": "2024-Q1 Redis 切换演练",
    "template_id": 1,
    "status": "running",
    "current_step": 2,
    "total_steps": 5,
    "started_at": "2024-01-01T10:00:00Z",
    "steps": [
      {
        "id": 101,
        "name": "停止 Redis 主节点",
        "type": "serial",
        "status": "completed",
        "timeout": 30,
        "started_at": "2024-01-01T10:00:00Z",
        "completed_at": "2024-01-01T10:00:15Z"
      },
      {
        "id": 102,
        "name": "验证从节点切换",
        "type": "serial",
        "status": "running",
        "timeout": 60,
        "started_at": "2024-01-01T10:00:20Z"
      }
    ]
  }
}
```

---

## 5. 步骤执行

### 5.1 完成步骤

- **POST** `/api/v1/steps/{id}/complete`

**请求体**:
```json
{
  "remark": "Redis 主节点已停止，从节点切换成功"
}
```

### 5.2 上报问题

- **POST** `/api/v1/steps/{id}/issue`

**请求体**:
```json
{
  "issue_description": "切换超时，从节点未成功接管",
  "severity": "high"
}
```

### 5.3 跳过步骤

- **POST** `/api/v1/steps/{id}/skip`

**需要权限**: `director` 或 `admin`

### 5.4 强制完成

- **POST** `/api/v1/steps/{id}/force_complete`

**需要权限**: `director` 或 `admin`

---

## 6. WebSocket 实时通信

### 6.1 大屏展示端

- **WS** `/ws/display/{drill_id}`

**事件类型**:
```json
{
  "event_type": "step_update",
  "drill_id": 100,
  "payload": {
    "step_id": 102,
    "step_name": "验证从节点切换",
    "status": "running"
  },
  "timestamp": 1704067200
}
```

### 6.2 工作台任务推送

- **WS** `/ws/tasks`

### 6.3 指挥台事件

- **WS** `/ws/control/{drill_id}`

**支持的控制命令**:
- `start` - 启动演练
- `pause` - 暂停演练
- `resume` - 恢复演练
- `terminate` - 终止演练
- `skip_step` - 跳过步骤

---

## 7. 健康检查

- **GET** `/api/v1/health`

**响应**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": 1704067200
}
```

---

## 8. 看板统计

所有接口需要 Bearer Token，使用统一响应格式。统计在后端对全量记录聚合，不受列表分页影响；数据库查询失败返回 HTTP 500。

| 接口 | 允许角色 | 响应 data 示例 |
|------|----------|---------------|
| GET `/api/v1/dashboard/team` | admin | `{"team_online_count": 3, "team_total_count": 8}` |
| GET `/api/v1/dashboard/my-templates` | admin、director | `{"my_template_count": 12}` |
| GET `/api/v1/dashboard/step-duration` | admin、director、viewer | `{"avg_step_duration_seconds": 65}` |

- 团队人数：总人数为当前启用的用户账号数；在线人数为其中最近 5 分钟有认证 API 活动或心跳的去重账号数，使用共享 Redis 跨节点统计。停用和删除的账号不计入。Redis 不可用时仅 `team_online_count` 为 `null`，总人数仍返回真实值。
- 我的模板：仅统计当前登录用户创建、未软删除的模板，包含已停用模板；不接受指定其他用户的查询参数。
- 平均步骤耗时：统计所有现存演练中正常完成的叶子步骤，以有效起止时间差计算平均值并四舍五入到整数秒。排除父节点、非空阶段与阶段步骤同名的占位节点、跳过/超时/异常等未正常完成步骤、缺失或倒置时间及孤立步骤。无样本时 `avg_step_duration_seconds` 为 `null`，有效零耗时返回 `0`。
- 未认证返回 HTTP 401；角色不在上表中返回 HTTP 403（执行员不能访问步骤耗时接口）。前端每 60 秒独立刷新；加载、失败或空样本显示 `—`，不将未知值显示为 `0`。

---

## 权限模型

| 角色 | 权限范围 |
|------|----------|
| admin | 全权限 |
| director | 指挥: 启动/暂停/恢复/终止演练，管理步骤 |
| executor | 执行: 完成/上报自身任务的步骤 |
| viewer | 只读: 查看演练状态和日志 |
