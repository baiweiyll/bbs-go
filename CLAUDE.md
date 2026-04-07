# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

bbs-go 是一个基于 Go 语言开发的开源社区论坛系统，采用模块化架构：

- **server**: Go 后端 API 服务 (Iris + GORM)
- **site**: Nuxt.js 3 前端站点 (Vue 3 + Element Plus)
- **admin**: Vue 3 管理后台 (Arco Design + Vite)

## Development Commands

### Root Level (Makefile)

```bash
# 构建所有组件
make build

# 构建特定模块
make build-site    # 构建前端站点
make build-admin   # 构建管理后台
make build-all-platforms  # 构建多平台服务端

# 清理构建产物
make clean
```

### Server (Go)

```bash
cd server

# 运行服务
go run main.go

# 编译
go build -v -o bbs-go main.go

# 依赖管理
go mod tidy
```

配置位于 `server/bbs-go.yaml`，参考 `bbs-go.example.yaml` 创建。

### Site (Nuxt.js)

```bash
cd site

# 安装依赖
pnpm install

# 开发模式
pnpm dev           # 使用 .env.local 配置

# 构建
pnpm build         # 使用 .env.production 配置
pnpm generate      # 静态生成 (NUXT_SSR=false)

# 预览
pnpm preview
```

### Admin (Vue 3 + Vite)

```bash
cd admin

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建
pnpm build

# 类型检查
pnpm type:check

# 代码检查
pnpm lint-staged
```

## Architecture

### 后端架构 (server/)

采用分层架构：

```
internal/
├── controllers/    # HTTP 处理器 (API / Admin / Render)
├── services/       # 业务逻辑层
├── repositories/   # 数据访问层
├── models/         # 数据模型 (GORM)
├── middleware/     # 中间件
├── cache/          # 缓存层
├── scheduler/      # 定时任务
└── pkg/            # 工具包
```

关键入口：
- `main.go` - 服务启动入口
- `bbs-go.yaml` - 配置文件
- `migrations/` - 数据库迁移脚本

API 路由前缀：
- `/bbsapi/` - 前端 API
- `/bbsadmin/` - 管理后台 API
- `/bbsoidc/` - OIDC 认证

### 前端站点 (site/)

Nuxt.js 3 项目，配置在 `nuxt.config.ts`：

```
src/
├── pages/          # 页面路由
├── components/     # Vue 组件
├── composables/    # 组合式函数
├── stores/         # Pinia 状态管理
├── layouts/        # 布局组件
├── plugins/        # Nuxt 插件
├── middleware/     # 路由中间件
└── locales/        # i18n 国际化
```

代理配置（开发时）：
- `/bbsapi/**` → Go 后端
- `/bbsoidc/**` → Go 后端（不跟随重定向）
- `/admin/**` → Go 后端

站点使用 `baseURL: '/forum/'`，部署时注意路径。

### 管理后台 (admin/)

Vue 3 + TypeScript + Arco Design：

```
src/
├── api/            # API 接口
├── views/          # 页面视图
├── components/     # 组件
├── router/         # 路由配置
├── store/          # Pinia 状态管理
├── locale/         # 国际化
└── utils/          # 工具函数
```

路由使用 hash 模式 (`createWebHashHistory`)，支持动态路由加载。

## 技术栈

- **后端**: Go 1.24, Iris, GORM, MySQL, JWT, Bleve (搜索)
- **站点前端**: Nuxt.js 3, Vue 3, Element Plus, Pinia, VueUse
- **管理后台**: Vue 3, Vite, Arco Design, Pinia
- **工具**: pnpm, Docker

## 重要配置

### 数据库
MySQL 连接配置在 `server/bbs-go.yaml`：
```yaml
DB:
  Url: username:password@tcp(localhost:3306)/bbsgo_db?charset=utf8mb4&parseTime=True&multiStatements=true&loc=Local
```

### 上传存储
支持阿里云 OSS、腾讯云 COS，配置在 `Uploader` 段。

### OIDC 登录
配置 OIDC 认证参数实现单点登录集成。

## 开发注意事项

1. **包管理器**: 使用 `pnpm`，不要混用 npm/yarn
2. **Node 版本**: >= 14.0.0
3. **Go 版本**: 1.24.0
4. **站点路径**: 生产环境部署在 `/forum/` 路径下
5. **API 代理**: 开发时通过 Nuxt Nitro 代理到后端
