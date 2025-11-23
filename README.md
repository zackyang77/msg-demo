# 站内信示例应用

前端使用 Vue 3 + Vite，后端采用 go-zero + MySQL，展示站内信（站内消息）功能的端到端范例，支持「个人联系」与「系统通知」两类信息。

## 目录结构

```
msg-demo/
├── backend/           # go-zero 服务
│   └── inbox/
│       ├── etc/       # 配置（含 MySQL 连接）
│       ├── internal/  # handler、logic、svc、types
│       └── inbox.api  # API 定义
├── frontend/          # Vue 3 + Vite 项目
└── README.md
```

## 后端（go-zero）

1. 先启动 MySQL（Docker）：
   ```bash
   docker compose up -d mysql
   ```
   - 默认 root 密码 `password`，数据存储在 `./mysql-data`。
   - `backend/inbox/db/schema.sql` 会自动建立三张表：`direct_messages`（个人信息）、`system_notifications`（通知模板）、`system_notification_receipts`（通知发送记录与读取状态）。
2. 进入 `backend/inbox`，启动服务：
   ```bash
   go run inbox.go -f etc/inbox-api.yaml
   ```
3. 调整 `etc/inbox-api.yaml` 内的 `Mysql.DataSource` 以符合本地环境；REST 服务监听在 `:8888`。
4. API 接口（`inbox.api`）已提供：
   - `/api/v1/messages`：发送个人或系统信息（`channel` 传 `personal|system`）。
   - `/api/v1/messages` GET：依照 `channel + userId` 分页查询。
   - `/api/v1/messages/:id/read`：依 `channel` 标记已读。
   - `/api/v1/messages/unread/count`：取得个人 / 系统 / 总未读数。

## 前端（Vue）

1. 进入 `frontend` 安装依赖：
   ```bash
   npm install
   ```
2. 开发模式：
   ```bash
   npm run dev
   ```
3. Vite dev server 会通过 `vite.config.ts` 的 proxy 将 `/api` 调用转发到 `http://127.0.0.1:8888`。
4. 页面提供注册/登录面板，登录后会自动保存 token，后续刷新页面仍会保持会话，并自动按当前登录用户加载站内信。

## 后续工作建议

- 接入用户身份／登录机制，免去手动输入 `userId`。
- 将系统通知扩展为多用户或群组推送（目前为单用户）。
- 编写单元测试与集成测试，并补上 seed data / fixtures。
- 将后端与前端一起 Docker 化，或加入 CI/CD、自動部署流程。

  test git
  222
