# 北辰 ERP

面向小型批发业务的 ERP 首版，技术栈为 Go、Gin、GORM、Fx、SQLite，以及 Vue 3、JavaScript、Element Plus。当前提供经营概览、客户、商品、销售订单、库存和采购入库流程。

## 本地运行

需要 Go 1.27.2 或更新版本、Node.js 20+、npm 和 CGO C 编译器（例如 GCC）。

```bash
# 先构建 Vue 静态文件；产物写入 backend/internal/web/dist，随后由 Go 编译进可执行文件
cd frontend
npm ci
npm run build

# 启动单体应用，Vue 页面和 API 共用 :8080
cd ../backend
go run ./cmd/server
```

打开 `http://localhost:8080`。Vue Router 的前端路由也由 Go 服务回退到嵌入的 `index.html`。

开发 Vue 界面时可另开终端使用 Vite 热更新：

```bash
# 终端一：Go API
cd backend
go run ./cmd/server

# 终端二：前端
cd frontend
npm run dev
```

打开 Vite 输出的本地地址（默认 `http://localhost:5173`）。前端开发服务器会把 `/api` 请求代理到 `http://127.0.0.1:8080`。Go 服务直接提供嵌入的前端文件时，前端改动后需重新运行 `npm run build` 并重启 Go 进程。

首次启动会在 `backend/data/erp.db` 自动建表，并插入一组明确用于演示的客户、商品和订单数据。后续启动只做迁移，不会覆盖已有数据库。配置可通过环境变量调整：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ADDR` | `:8080` | API 监听地址 |
| `DB_PATH` | `data/erp.db` | SQLite 数据文件路径，相对 `backend/` 目录 |
| `APP_TIMEZONE` | `Asia/Shanghai` | 经营日期和报表时区 |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://127.0.0.1:5173` | 允许访问 API 的前端源，逗号分隔 |
| `VITE_API_BASE` | 空 | 前端 API 前缀；本地开发通常保留为空并使用 Vite 代理 |

## 主要流程

- 新建销售订单时，服务端重新读取商品售价并在 SQLite 事务中校验和预占库存；库存不足时整笔订单回滚。
- 订单可以从待发货变为已发货或取消；取消待发货订单会返还库存。已发货订单可标记完成。
- 采购收货和库存盘点调整会写库存流水；低于安全库存的商品会显示预警。
- 金额以分为单位保存，界面按元显示。销售订单保留创建时的商品价格快照。
- SQLite 开启外键、忙等待和 WAL 模式，适合单机或轻量内网部署。

## 检查

```bash
cd frontend
npm ci
npm run build
cd ../backend
go test ./...
go vet ./...
go build ./cmd/server
```

服务健康检查：`GET /api/health`。核心 API：`GET /api/dashboard`、`/api/customers`、`/api/products`、`/api/orders`、`/api/inventory`、`/api/purchases`；对应创建或库存操作使用各模块的 `POST` 路由。订单状态更新使用 `PATCH /api/orders/:id/status`。

## 当前边界

这是可运行的业务基础版，演示数据只在空数据库首次启动时写入。尚未包含用户登录、角色权限、审批、多仓库、税务/会计凭证、数据导入导出和自动备份。API 当前没有身份认证，应先在受控本机或内网试用；面向真实业务或公网部署前需要完成认证授权、备份恢复和财务对账设计。
