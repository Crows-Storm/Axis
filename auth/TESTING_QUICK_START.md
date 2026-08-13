# 🚀 Auth Service TDD 快速入门

## 📦 安装测试依赖

```bash
cd /Users/crow/project/goProject/Axis/auth
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/require
go mod tidy
```

---

## 🧪 运行测试

### 方式 1: 使用 Makefile（推荐）

```bash
# 从项目根目录运行
cd /Users/crow/project/goProject/Axis

# 运行所有 auth 测试
make test-auth

# 生成覆盖率报告
make test-auth-coverage

# 只运行 gRPC 测试
make test-auth-grpc

# 运行性能测试
make test-bench
```

### 方式 2: 直接使用 go test

```bash
cd /Users/crow/project/goProject/Axis/auth

# 运行所有测试（详细输出）
go test -v ./...

# 运行所有测试（带竞态检测）
go test -v -race ./...

# 只运行 HTTP 测试
go test -v -run "TestRegister" ./...

# 只运行 gRPC 测试
go test -v ./adapters/grpc/...

# 运行特定测试
go test -v -run "TestRegister_Success" ./...
go test -v -run "TestCreateUser_Success" ./adapters/grpc/...

# 生成覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

---

## 📊 查看测试结果

### 成功的测试输出示例
```
=== RUN   TestRegister_Success
--- PASS: TestRegister_Success (0.00s)
=== RUN   TestCreateUser_Success
--- PASS: TestCreateUser_Success (0.00s)
PASS
ok      github.com/Crows-Storm/Axis/auth        0.123s
```

### 失败的测试输出示例
```
=== RUN   TestRegister_Success
    http_test.go:45: 
        Error:      Not equal: 
                    expected: 200
                    actual  : 500
--- FAIL: TestRegister_Success (0.00s)
FAIL
```

---

## 🎯 测试覆盖率

### 查看覆盖率摘要
```bash
cd auth
go test -cover ./...
```

输出示例:
```
ok      github.com/Crows-Storm/Axis/auth                    0.123s  coverage: 85.2% of statements
ok      github.com/Crows-Storm/Axis/auth/adapters/grpc      0.089s  coverage: 92.5% of statements
```

### 详细覆盖率报告
```bash
cd auth
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# 输出示例:
# github.com/Crows-Storm/Axis/auth/http.go:23:    Register        85.7%
# github.com/Crows-Storm/Axis/auth/adapters/grpc/user_grpc.go:15:  CreateUser  100.0%
# total:                                          (statements)    88.3%
```

---

## ⚡ 性能测试

### 运行基准测试
```bash
cd auth
go test -bench=. -benchmem ./...
```

输出示例:
```
BenchmarkRegister_Success-8        50000    25840 ns/op    4512 B/op    48 allocs/op
BenchmarkCreateUser-8            100000    12345 ns/op    2048 B/op    24 allocs/op
```

解读：
- `50000`: 测试执行次数
- `25840 ns/op`: 每次操作耗时（纳秒）
- `4512 B/op`: 每次操作分配的内存（字节）
- `48 allocs/op`: 每次操作的内存分配次数

---

## 🐛 调试失败的测试

### 1. 查看详细错误信息
```bash
go test -v -run "TestRegister_Success" ./...
```

### 2. 添加调试日志
```go
func TestRegister_Success(t *testing.T) {
    // ...
    t.Logf("Request body: %+v", requestBody)
    t.Logf("Response: %s", w.Body.String())
    // ...
}
```

### 3. 使用 Delve 调试器
```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试测试
dlv test -- -test.run TestRegister_Success
```

---

## 📝 编写新测试

### 模板 - HTTP 测试
```go
func TestRegister_YourScenario(t *testing.T) {
    // Arrange (准备)
    mockHandler := new(MockRegisterUserHandler)
    router := setupTestRouter(mockHandler)
    
    requestBody := command.RegisterUserCommand{
        LoginId:  "testuser",
        Password: "Test@1234",
        Email:    "test@example.com",
    }
    
    mockHandler.On("Handle", mock.Anything, requestBody).Return(struct{}{}, nil)
    
    body, _ := json.Marshal(requestBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    // Act (执行)
    router.ServeHTTP(w, req)
    
    // Assert (验证)
    assert.Equal(t, http.StatusOK, w.Code)
    mockHandler.AssertExpectations(t)
}
```

### 模板 - gRPC 测试
```go
func TestCreateUser_YourScenario(t *testing.T) {
    // Arrange (准备)
    mockClient := new(MockUserServiceClient)
    userGRPC := NewUserGRPC(mockClient)
    
    ctx := context.Background()
    req := &userpb.CreateUserRequest{
        LoginId:  "testuser",
        Password: "Test@1234",
        Email:    "test@example.com",
    }
    
    expectedResp := &emptypb.Empty{}
    mockClient.On("CreateUser", ctx, req).Return(expectedResp, nil)
    
    // Act (执行)
    resp, err := userGRPC.CreateUser(ctx, req)
    
    // Assert (验证)
    require.NoError(t, err)
    assert.Equal(t, expectedResp, resp)
    mockClient.AssertExpectations(t)
}
```

---

## 🔄 TDD 工作流程

```
1. 🔴 RED - 编写失败的测试
   └─> go test -v -run TestYourFeature

2. 🟢 GREEN - 编写最少代码让测试通过
   └─> go test -v -run TestYourFeature

3. 🔵 REFACTOR - 重构代码
   └─> go test -v ./...  # 确保所有测试仍然通过

4. 🔁 重复
```

---

## 📋 常用测试命令速查表

| 命令 | 说明 |
|------|------|
| `go test ./...` | 运行所有测试 |
| `go test -v ./...` | 详细输出 |
| `go test -race ./...` | 竞态检测 |
| `go test -run TestXxx` | 运行特定测试 |
| `go test -cover ./...` | 覆盖率 |
| `go test -bench=.` | 基准测试 |
| `go test -short ./...` | 跳过长时间测试 |
| `go test -timeout 30s ./...` | 设置超时 |
| `go test -count=1 ./...` | 禁用缓存 |
| `go test -parallel 4 ./...` | 并行测试 |

---

## 🎓 学习资源

### 官方文档
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Test Command](https://pkg.go.dev/cmd/go#hdr-Test_packages)

### 第三方库
- [Testify Documentation](https://github.com/stretchr/testify)
- [Testify Assert](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [Testify Mock](https://pkg.go.dev/github.com/stretchr/testify/mock)

### 教程
- [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests/)
- [Go by Example: Testing](https://gobyexample.com/testing)

---

## ❓ 常见问题

### Q: 测试运行缓慢怎么办？
A: 
```bash
# 使用并行测试
go test -parallel 8 ./...

# 只运行短测试
go test -short ./...

# 跳过基准测试
go test -run=^Test ./...
```

### Q: Mock 没有被调用？
A: 确保在测试结束时调用：
```go
mockHandler.AssertExpectations(t)
```

### Q: 覆盖率报告在哪里？
A: 
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Q: 如何测试并发代码？
A: 使用 `-race` 标志：
```bash
go test -race ./...
```

### Q: 如何跳过某些测试？
A: 
```go
func TestSomething(t *testing.T) {
    t.Skip("Skipping this test for now")
    // 或者
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }
}
```

---

## 📞 获取帮助

- 查看完整文档: `TEST_README.md`
- 报告问题: 创建 Issue
- 提出建议: 提交 Pull Request

---

**Happy Testing! 🎉**
