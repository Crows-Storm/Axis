# 流程集合

同一套请求有两份，按客户端分开：

| 目录 | 客户端 |
|------|--------|
| `bruno/` | Bruno |
| `postman/` | Postman |

## 端口（与 `.env` 一致）

| 服务 | REST |
|------|------|
| user | `http://localhost:18801/api` |
| auth | `http://localhost:18802/api` |
| account | `http://localhost:18803/api` |
| ledger | `http://localhost:18804/api` |
| notification | `http://localhost:18805/api` |
| audit | `http://localhost:18806/api` |
| market | `http://localhost:18807/api` |
| exchange | `http://localhost:18808/api` |
| chain | `http://localhost:18809/api` |

## Bruno

1. Open Collection → `doc/run_collection/bruno/auth`
2. 再打开 `doc/run_collection/bruno/user`
3. 环境选 **local**（各自的 `environments/local.bru`）
4. Collection Runner 按文件夹顺序跑

`local` 里只有 `auth_base` 和 `user_base`。`login_id`、`password`、`Token`、各类 id 由 00-setup 和登录脚本用 `bru.setVar` 写到当次运行的变量里。

## Postman

1. Import `postman/Axis.local.postman_environment.json`
2. Import `postman/auth/Axis_Auth.postman_collection.json`
3. Import `postman/user/Axis_User.postman_collection.json`
4. 环境选 **Axis Local**
5. Collection Runner 按文件夹顺序跑

`auth_base` / `user_base` 在环境里。账号、token、id 是 collection 变量，由脚本写入。

## 请求头与脚本

两份脚本做同一件事，API 不同：Bruno 用 `bru` / `req` / `res`，Postman 用 `pm`。

- Collection 请求前生成 `X-Request-ID`（`xxxx-xxxx-xxxx`）并写入请求头，请求后清空。
- Bruno 关闭响应 JSON 自动解析。需要留下的 id 从原始响应文本用正则取出，避免 flake `uint64` 被当成 JS number。
- 注册 / 创建用户：明文密码在发送前变为 `SHA256(password)`（H1）。
- 登录：`clientHash = SHA256(H1 + ":" + timestampSec + ":" + requestId)`，与后端 `VerifyPassword` 一致。
- 登录成功后把 `Token` 设为 `Bearer <access_token>`。标了 noauth 的请求不带 `Authorization`，其余请求由 collection 脚本自动带上。

响应里 token 字段是 `data.token`（JSON 小写）；脚本同时兼容 `data.Token`。

业务成功判定是 HTTP 200 **且** `code === 0`（错误也常是 HTTP 200 + 非 0 `code`）。

`POST /register` 请求体包含 `login_id`、`username`、`password`、`email`。`username` 与本次 `login_id` 相同。

## 依赖

- **auth**：auth HTTP + user gRPC（`POST /register` 会建用户）
- **user**：user HTTP + auth HTTP（先注册/登录拿 JWT；user 除 ping 外都要 Bearer）

## 流程摘要

### auth

注册 → 失败登录 → 登录拿 token → 菜单 CRUD → 角色 CRUD → 位掩码 grant/replace → 绑角色 → `/rbac/check` → 解绑/撤销 → 启停 → 登出后再登

`mask=15` 为 CRUD（READ|CREATE|UPDATE|DELETE），`grant=true` 再 OR `16`（ENABLE）。

### user

经 auth 建操作员并登录 → 当前用户 GET/PUT → 创建目标用户 → exists 单字段/非法组合 → 用目标账号登录拿 `target_user_id` → batch → disable → soft delete

创建用户接口不回 id，所以用目标账号再登录 auth 取 `user_id`。
