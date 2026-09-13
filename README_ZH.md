<div align="center">

# Axis 🚀

**现代化的个人财务与生产力工作台**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Code Style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://golang.org/doc/effective_go.html)

[English](README.md) · [简体中文](README_ZH.md)

</div>

---

## 📖 关于项目

**Axis** 是一个基于 Go 微服务架构的综合性个人工作台平台，旨在帮助您在一个统一的系统中管理财务生活、生产力工具和数字资产。

### 🎯 愿景

创建一个私有、安全且可扩展的平台，整合财务管理、生产力工具和 Web3 功能，赋能个人全面掌控数字金融生活。

---

## ✨ 功能特性

### 💰 财务管理

- [ ] **记账与簿记**
  - [ ] 多币种支持
  - [ ] 分类管理
  - [ ] 交易追踪
  - [ ] 票据上传与 OCR 识别
  
- [ ] **投资管理**
  - [ ] 投资组合追踪
  - [ ] 股票市场集成
  - [ ] 加密货币追踪
  - [ ] 实时价格更新
  - [ ] 盈亏分析
  
- [ ] **财富分析**
  - [ ] 净资产追踪
  - [ ] 资产配置可视化
  - [ ] 业绩报告
  - [ ] 税务优化建议

### 📈 市场数据

- [ ] **股票市场**
  - [ ] 实时行情
  - [ ] 历史数据
  - [ ] 技术指标
  - [ ] 自选股管理
  
- [ ] **加密货币**
  - [ ] 多交易所支持
  - [ ] DeFi 协议集成
  - [ ] 价格提醒
  - [ ] 市场情绪分析

### 🗓️ 生产力工具

- [ ] **日历与日程安排**
  - [ ] 事件管理
  - [ ] 提醒与通知
  - [ ] Google 日历同步
  - [ ] 时区支持
  
- [ ] **任务管理**
  - [ ] 待办事项
  - [ ] 项目追踪
  - [ ] 优先级管理
  - [ ] 截止日期提醒

### 🕷️ 自动化

- [ ] **网页爬虫工作流**
  - [ ] 可视化工作流设计器
  - [ ] 定时任务
  - [ ] 数据提取规则
  - [ ] 输出格式化
  
- [ ] **自定义集成**
  - [ ] Webhook 支持
  - [ ] API 连接器
  - [ ] 数据管道
  - [ ] 事件驱动自动化

### 🌐 Web3 集成

- [ ] **钱包管理**
  - [ ] HD 钱包创建
  - [ ] 多链支持（ETH、BTC、Solana 等）
  - [ ] 私钥加密
  - [ ] 硬件钱包集成
  
- [ ] **区块链交互**
  - [ ] 交易签名与广播
  - [ ] 智能合约交互
  - [ ] 代币转账
  - [ ] Gas 优化
  
- [ ] **链上监控**
  - [ ] 地址追踪
  - [ ] 交易提醒
  - [ ] 余额监控
  - [ ] DeFi 持仓追踪

---

## 🏗️ 架构设计

Axis 采用 **微服务架构**，结合 **领域驱动设计（DDD）** 和 **CQRS** 模式。

### 核心服务

```
┌─────────────────────────────────────────────────────────────┐
│                        API 网关                              │
│                  (路由、鉴权、限流)                           │
└────────────┬────────────────────────────────┬───────────────┘
             │                                │
    ┌────────▼────────┐              ┌───────▼────────┐
    │  认证服务       │              │  用户服务      │
    │  (JWT, RBAC)    │              │  (用户画像)    │
    └────────┬────────┘              └───────┬────────┘
             │                                │
    ┌────────▼────────────────────────────────▼────────┐
    │              事件总线 (Kafka)                     │
    └────────┬────────────────────────────┬─────────────┘
             │                            │
    ┌────────▼────────┐          ┌───────▼──────────┐
    │   钱包服务      │          │   账本服务       │
    │ (加密货币/股票) │          │   (记账簿记)     │
    └─────────────────┘          └──────────────────┘
             │                            │
    ┌────────▼────────┐          ┌───────▼──────────┐
    │   通知服务      │          │   审计服务       │
    │                 │          │   (日志记录)     │
    └─────────────────┘          └──────────────────┘
```

### 技术栈

| 类别 | 技术 |
|------|------|
| **编程语言** | Go 1.26+ |
| **框架** | Gin, gRPC, Protocol Buffers |
| **数据库** | MySQL, Redis, MongoDB |
| **消息队列** | Kafka |
| **服务发现** | Consul |
| **监控** | Prometheus, Grafana, Jaeger |
| **容器化** | Docker, Kubernetes |
| **CI/CD** | GitHub Actions |

### 设计模式

- ✅ **领域驱动设计（DDD）**
- ✅ **命令查询职责分离（CQRS）**
- ✅ **事件溯源**（用于审计追踪）
- ✅ **Saga 模式**（用于分布式事务）
- ✅ **Outbox 模式**（保证事件可靠性）
- ✅ **仓储模式**
- ✅ **装饰器模式**（用于横切关注点）
- ✅ **依赖注入**（使用 Wire）

---

## 🚀 快速开始

### 环境要求

- **Go 1.26+**
- **Docker & Docker Compose**
- **MySQL 8.0+**
- **Redis 7.0+**
- **Kafka 3.0+**（可选，用于异步消息）
- **Consul**（可选，用于服务发现）

### 安装步骤

1. **克隆仓库**

```bash
git clone https://github.com/Crows-Storm/Axis.git
cd Axis
```

2. **启动基础设施服务**

```bash
docker-compose up -d mysql redis consul
```

3. **运行数据库迁移**

```bash
make migrate-up
```

4. **启动服务**

```bash
# 启动所有服务
make run-all

# 或单独启动服务
make run-auth
make run-user
make run-wallet
```

5. **访问 API**

```bash
# 健康检查
curl http://localhost:8000/health

# API 文档
open http://localhost:8000/swagger
```

### 开发环境设置

```bash
# 安装依赖
make deps

# 生成 protobuf 文件
make proto

# 生成 OpenAPI 客户端
make openapi

# 运行测试
make test

# 运行代码检查
make lint

# 构建所有服务
make build
```

---

## 📁 项目结构

```
Axis/
├── api/                        # API 定义
│   ├── proto/                  # gRPC proto 文件
│   └── openapi/                # OpenAPI/Swagger 规范
│
├── services/                   # 微服务
│   ├── auth/                   # 认证授权服务
│   ├── user/                   # 用户管理服务
│   ├── wallet/                 # 钱包服务
│   ├── ledger/                 # 账本服务
│   ├── audit/                  # 审计服务
│   ├── notification/           # 通知服务
│   └── gateway/                # API 网关
│
├── common/                     # 共享库
│   ├── domain/                 # 领域原语
│   ├── decorator/              # CQRS 装饰器
│   ├── server/                 # 服务器工具
│   └── client/                 # 服务客户端
│
├── pkg/                        # 公共库
│   ├── logger/                 # 日志
│   ├── metrics/                # 监控指标
│   ├── discovery/              # 服务发现
│   └── security/               # 安全工具
│
├── deployments/                # 部署配置
│   ├── docker/                 # Dockerfile
│   └── kubernetes/             # K8s 清单
│
├── scripts/                    # 构建和部署脚本
├── docs/                       # 文档
├── test/                       # 集成测试与 E2E 测试
│
├── docker-compose.yml          # 本地开发环境
├── Makefile                    # 构建自动化
├── go.work                     # Go 工作空间
└── README.md                   # 本文件
```

### 服务结构（DDD 分层架构）

```
service/
├── cmd/
│   └── api/                    # 应用入口
│       ├── main.go
│       └── wire.go             # 依赖注入
│
├── internal/                   # 私有代码
│   ├── domain/                 # 领域层（业务逻辑）
│   │   ├── user/
│   │   │   ├── user.go         # 聚合根
│   │   │   ├── repository.go   # 仓储接口
│   │   │   └── events.go       # 领域事件
│   │
│   ├── application/            # 应用层（用例）
│   │   ├── command/            # 命令（写操作）
│   │   ├── query/              # 查询（读操作）
│   │   └── service/            # 应用服务
│   │
│   ├── infrastructure/         # 基础设施层
│   │   ├── persistence/        # 数据库实现
│   │   ├── messaging/          # 消息队列
│   │   └── external/           # 外部服务客户端
│   │
│   └── ports/                  # 端口层（接口）
│       ├── http/               # HTTP 处理器
│       ├── grpc/               # gRPC 服务
│       └── events/             # 事件消费者
│
├── configs/                    # 配置文件
├── go.mod
└── Makefile
```

---

## 🧪 测试

```bash
# 运行单元测试
make test

# 运行集成测试
make test-integration

# 运行 E2E 测试
make test-e2e

# 生成覆盖率报告
make coverage

# 运行特定服务的测试
cd services/auth && go test ./...
```

---

## 📚 文档

- [架构概览](docs/architecture-review.md)
- [领域事件设计](docs/domain-event-design.md)
- [API 文档](docs/api/)
- [部署指南](docs/deployment/)
- [开发指南](docs/development/)
- [贡献指南](CONTRIBUTING.md)

---

## 🛣️ 路线图

### 第一阶段：基础设施（2026 Q1）🏗️

- [ ] 核心微服务搭建
- [ ] 认证与授权
- [ ] 用户管理
- [ ] 基础 API 网关
- [ ] 数据库架构设计
- [ ] Docker 容器化

### 第二阶段：财务核心（2026 Q2）💰

- [ ] 账本服务（记账）
- [ ] 交易管理
- [ ] 多币种支持
- [ ] 基础报表
- [ ] 分类管理
- [ ] 预算追踪

### 第三阶段：投资追踪（2026 Q3）📈

- [ ] 钱包服务
- [ ] 股票市场集成
- [ ] 加密货币追踪
- [ ] 投资组合分析
- [ ] 实时价格更新
- [ ] 盈亏计算

### 第四阶段：自动化与 Web3（2026 Q4）🤖

- [ ] 网页爬虫工作流
- [ ] 定时任务
- [ ] Web3 钱包创建
- [ ] 区块链集成
- [ ] DeFi 追踪
- [ ] 智能合约交互

### 第五阶段：高级功能（2027+）🚀

- [ ] 移动应用（React Native）
- [ ] 高级分析与机器学习
- [ ] 社交功能
- [ ] 插件系统
- [ ] 白标解决方案
- [ ] 企业版功能

---

## 🤝 贡献

我们欢迎贡献！请查看我们的[贡献指南](CONTRIBUTING.md)了解详情。

### 开发工作流

1. Fork 本仓库
2. 创建特性分支（`git checkout -b feature/amazing-feature`）
3. 提交你的更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 提交 Pull Request

### 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- 为新功能编写单元测试
- 更新文档
- 使用常规提交格式
- 提交前运行代码检查工具

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 🙏 致谢

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) 提供项目结构灵感
- Eric Evans 的《领域驱动设计》
- Vaughn Vernon 的《实现领域驱动设计》
- Chris Richardson 的《微服务模式》

---

## 📞 联系与支持

- **问题反馈**：[GitHub Issues](https://github.com/Crows-Storm/Axis/issues)
- **讨论**：[GitHub Discussions](https://github.com/Crows-Storm/Axis/discussions)
- **邮箱**：crow@axis-platform.com

---

## ⭐ Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=Crows-Storm/Axis&type=Date)](https://star-history.com/#Crows-Storm/Axis&Date)

---

<div align="center">

**用 ❤️ 制作 by [Crow](https://github.com/Crows-Storm)**

如果觉得这个项目有帮助，请考虑给它一个 ⭐！

</div>
