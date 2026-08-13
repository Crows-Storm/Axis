# Auth Service TDD 测试文档

## 📋 概述

本文档描述了 auth 服务的测试驱动开发 (TDD) 测试套件，包括：
1. **Register REST API 测试** (`http_test.go`)
2. **UserGRPC 客户端测试** (`adapters/grpc/user_grpc_test.go`)

---

## 🏗️ 测试架构

### 依赖关系图
```
HTTPServer (http.go)
    ↓
RegisterUserCommandHandler
    ↓
UserGRPC (adapters/grpc/user_grpc.go)
    ↓
UserServiceClient (gRPC stub)
```

### Mock 策略
- **HTTP 层测试**: Mock `RegisterUserCommandHandler`
- **gRPC 层测试**: Mock `userpb.UserServiceClient`

---

## 📦 测试文件

### 1. `http_test.go` - REST API 测试

#### 测试覆盖场景

| 测试函数 | 场景 | 期望结果 |
|---------|------|----------|
| `TestRegister_Success` | 正常注册流程 | 200 OK, 返回成功消息 |
| `TestRegister_InvalidJSON` | 无效 JSON 格式 | 返回错误消息 |
| `TestRegister_MissingRequiredFields` | 缺少必填字段 | 返回验证错误 |
| `TestRegister_EmptyBody` | 空请求体 | 返回错误 |
| `TestRegister_HandlerError` | Handler 返回错误（如用户已存在） | 返回业务错误 |
| `TestRegister_WrongContentType` | 错误的 Content-Type | 返回错误 |
| `TestRegister_LargePayload` | 超大 payload | 测试大小限制 |
| `TestRegister_ConcurrentRequests` | 并发请求 | 验证线程安全 |
| `BenchmarkRegister_Success` | 性能基准测试 | 测量吞吐量 |

#### 运行测试
```bash
# 运行所有 HTTP 测试
cd /Users/crow/project/goProject/Axis/auth
go test -v -run "TestRegister" ./...

# 运行单个测试
go test -v -run "TestRegister_Success" ./...

# 运行基准测试
go test -bench=BenchmarkRegister -benchmem ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

### 2. `adapters/grpc/user_grpc_test.go` - gRPC 客户端测试

#### 测试覆盖场景

##### CreateUser 测试
| 测试函数 | 场景 | 期望结果 |
|---------|------|----------|
| `TestCreateUser_Success` | 成功创建用户 | 无错误，返回空响应 |
| `TestCreateUser_DuplicateUser` | 用户已存在 | AlreadyExists 错误 |
| `TestCreateUser_InvalidRequest` | 无效请求（空字段、格式错误） | InvalidArgument 错误 |
| `TestCreateUser_ContextTimeout` | 上下文超时 | DeadlineExceeded 错误 |
| `TestCreateUser_NetworkError` | 网络故障 | Unavailable 错误 |

##### GetUserById 测试
| 测试函数 | 场景 | 期望结果 |
|---------|------|----------|
| `TestGetUserById_Success` | 成功获取用户 | 返回用户信息 |
| `TestGetUserById_NotFound` | 用户不存在 | NotFound 错误 |
| `TestGetUserById_EmptyId` | 空 ID | InvalidArgument 错误 |

##### GetUserByLoginId 测试
| 测试函数 | 场景 | 期望结果 |
|---------|------|----------|
| `TestGetUserByLoginId_Success` | 成功获取用户 | 返回用户信息 |
| `TestGetUserByLoginId_NotFound` | 用户不存在 | NotFound 错误 |
| `TestGetUserByLoginId_EmptyLoginId` | 空 LoginId | InvalidArgument 错误 |

##### 特殊场景测试
| 测试函数 | 场景 |
|---------|------|
| `TestUserGRPC_NilClient` | Nil 客户端处理 |
| `TestUserGRPC_ContextCancellation` | 上下文取消 |
| `TestUserGRPC_ConcurrentCalls` | 并发调用 |
| `TestUserGRPC_TransientError` | 瞬时错误 |
| `TestUserGRPC_InternalServerError` | 内部服务器错误 |
| `TestUserGRPC_PermissionDenied` | 权限拒绝 |
| `TestUserGRPC_UnknownError` | 未知错误 |

#### 运行测试
```bash
# 运行所有 gRPC 测试
cd /Users/crow/project/goProject/Axis/auth
go test -v ./adapters/grpc/...

# 运行特定测试
go test -v -run "TestCreateUser" ./adapters/grpc/...
go test -v -run "TestGetUserById" ./adapters/grpc/...

# 运行基准测试
go test -bench=. -benchmem ./adapters/grpc/...

# 生成覆盖率
go test -coverprofile=grpc_coverage.out ./adapters/grpc/...
go tool cover -html=grpc_coverage.out -o grpc_coverage.html
```

---

## 🔧 依赖安装

测试使用以下第三方库：

```bash
# 安装测试依赖
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/require
```

更新 `go.mod`:
```bash
cd /Users/crow/project/goProject/Axis/auth
go mod tidy
```

---

## 📊 测试覆盖率目标

| 模块 | 目标覆盖率 | 当前状态 |
|------|-----------|---------|
| HTTP Handler | 80%+ | ✅ 待运行 |
| gRPC Client | 90%+ | ✅ 待运行 |
| Command Handler | 75%+ | 🔄 待实现 |

---

## 🎯 TDD 工作流程

### 开发新功能时遵循的 TDD 流程：

```
1️⃣ RED (编写失败的测试)
   ├── 明确需求和预期行为
   ├── 编写测试用例
   └── 运行测试（应该失败）

2️⃣ GREEN (最小化实现让测试通过)
   ├── 编写最简单的代码让测试通过
   ├── 不考虑优化和完美
   └── 运行测试（应该通过）

3️⃣ REFACTOR (重构代码)
   ├── 优化代码结构
   ├── 消除重复
   ├── 改进设计
   └── 运行测试（确保仍然通过）

4️⃣ 重复上述流程
```

---

## 🐛 已发现的问题（通过测试）

### HTTP Handler 问题

1. **拼写错误**
   ```go
   // http.go:41
   server.Error(c, server.CodeServerError, fmt.Sprintf("Faield: %v", err))
   //                                                     ^^^^^^ 应该是 "Failed"
   ```

2. **HTTP 状态码不正确**
   - 当前所有错误都返回 `200 OK`
   - 建议：
     - 验证错误 → `400 Bad Request`
     - 业务错误 → `422 Unprocessable Entity`
     - 服务器错误 → `500 Internal Server Error`

3. **缺少输入验证**
   - 没有验证必填字段（LoginId, Password, Email）
   - 没有验证格式（Email 格式、密码强度）
   - 没有验证长度限制

4. **缺少 Payload 大小限制**
   - 没有限制请求体大小
   - 可能导致内存溢出攻击

### UserGRPC 问题

1. **缺少 Nil 检查**
   ```go
   func NewUserGRPC(client userpb.UserServiceClient) *UserGRPC {
       // 应该添加：
       if client == nil {
           panic("client cannot be nil")
       }
       return &UserGRPC{client: client}
   }
   ```

2. **没有重试逻辑**
   - 对于瞬时错误（Unavailable），应该自动重试
   - 建议使用 exponential backoff

3. **没有超时保护**
   - 应该为每个 gRPC 调用设置默认超时
   - 防止无限期等待

---

## ✅ 建议的改进

### 1. HTTP Handler 改进

```go
// 改进后的 Register 函数
func (H HTTPServer) Register(c *gin.Context) {
    var req command.RegisterUserCommand
    if err := c.ShouldBindJSON(&req); err != nil {
        server.Error(c, server.CodeInvalidParams, fmt.Sprintf("Invalid request body: %v", err))
        c.Status(http.StatusBadRequest)
        return
    }

    // 验证必填字段
    if req.LoginId == "" || req.Password == "" || req.Email == "" {
        server.Error(c, server.CodeInvalidParams, "LoginId, Password, and Email are required")
        c.Status(http.StatusBadRequest)
        return
    }

    // 验证 Email 格式
    if !isValidEmail(req.Email) {
        server.Error(c, server.CodeInvalidParams, "Invalid email format")
        c.Status(http.StatusBadRequest)
        return
    }

    _, err := H.app.Commands.RegisterUser.Handle(c, req)
    if err != nil {
        // 根据错误类型返回不同的状态码
        if isUserExistsError(err) {
            server.Error(c, server.CodeUserExists, "User already exists")
            c.Status(http.StatusConflict)
        } else {
            server.Error(c, server.CodeServerError, fmt.Sprintf("Failed: %v", err))
            c.Status(http.StatusInternalServerError)
        }
        return
    }

    server.Success(c, "successfully")
    c.Status(http.StatusCreated)
}
```

### 2. UserGRPC 改进

```go
// 添加重试和超时逻辑
func NewUserGRPC(client userpb.UserServiceClient) *UserGRPC {
    if client == nil {
        panic("client cannot be nil")
    }
    return &UserGRPC{client: client}
}

func (u *UserGRPC) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*emptypb.Empty, error) {
    // 添加默认超时
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // 调用客户端（可以添加重试逻辑）
    return u.client.CreateUser(ctx, req)
}
```

---

## 📈 持续集成建议

### GitHub Actions / GitLab CI 配置示例

```yaml
name: Auth Service Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: |
          cd auth
          go mod download
      
      - name: Run tests
        run: |
          cd auth
          go test -v -race -coverprofile=coverage.out ./...
      
      - name: Coverage report
        run: |
          cd auth
          go tool cover -func=coverage.out
      
      - name: Benchmark
        run: |
          cd auth
          go test -bench=. -benchmem ./...
```

---

## 📚 参考资料

- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Testify Documentation](https://github.com/stretchr/testify)
- [TDD in Go](https://quii.gitbook.io/learn-go-with-tests/)
- [gRPC Testing Guide](https://grpc.io/docs/languages/go/basics/#testing)

---

## 🤝 贡献指南

添加新测试时，请遵循以下规范：

1. **命名约定**:
   - 测试函数: `Test<FunctionName>_<Scenario>`
   - 基准测试: `Benchmark<FunctionName>`

2. **测试结构**:
   ```go
   func TestXxx_Scenario(t *testing.T) {
       // Arrange - 准备测试数据和 mock
       
       // Act - 执行被测试的函数
       
       // Assert - 验证结果
   }
   ```

3. **Mock 使用**:
   - 使用 `testify/mock` 创建 mock
   - 明确定义期望的调用和返回值
   - 使用 `AssertExpectations` 验证 mock 调用

4. **测试隔离**:
   - 每个测试应该独立运行
   - 不依赖测试执行顺序
   - 清理测试资源

---

## 📞 联系方式

如有问题或建议，请：
- 创建 Issue
- 提交 Pull Request
- 联系团队成员

---

**最后更新**: 2026-08-11  
**版本**: 1.0.0
