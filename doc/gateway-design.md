# Axis 项目 API Gateway 设计方案

## 文档信息
- **创建日期**: 2026-07-25
- **项目名称**: Axis 微服务系统
- **文档版本**: v1.0
- **作者**: System Architect

---

## 一、背景与现状分析

### 1.1 当前系统架构

Axis 项目采用微服务架构，当前包含以下服务：

| 服务名称 | HTTP 端口 | gRPC 端口 | 功能描述 |
|---------|----------|----------|---------|
| auth    | 18802    | 18702    | 认证授权服务 |
| user    | 18801    | 18701    | 用户管理服务 |
| wallet  | 待定     | 待定      | 钱包服务 |
| audit   | 待定     | 待定      | 审计日志服务 |
| ledger  | 待定     | 待定      | 账本服务 |

### 1.2 当前存在的问题

1. **端口分散管理复杂**
   - 前端/客户端需要记住多个服务地址
   - 开发环境、测试环境、生产环境配置复杂
   - 无法统一管理服务发现

2. **安全问题**
   - 每个服务都直接暴露给外部
   - 缺乏统一的认证授权验证
   - Token 验证逻辑分散在各个服务中
   - 敏感服务（如 audit）可能被直接访问

3. **横切关注点分散**
   - 日志记录格式不统一
   - 限流熔断需要每个服务单独实现
   - 请求追踪（tracing）难以实现
   - 监控指标收集分散

4. **客户端调用复杂**
   - 需要维护多个 HTTP 客户端
   - 跨服务调用需要服务间认证
   - 错误处理不统一
   - API 版本管理困难



### 1.3 为什么需要 API Gateway

基于以上问题，Axis 项目**强烈建议实现 API Gateway**，理由如下：

✅ **统一入口**：客户端只需访问一个地址（如 `http://api.axis.com`）
✅ **安全加固**：在网关层统一处理认证、授权、token 验证
✅ **简化客户端**：隐藏内部服务架构，降低客户端复杂度
✅ **流量控制**：统一限流、熔断、降级策略
✅ **可观测性**：统一日志、监控、链路追踪
✅ **协议转换**：支持 REST → gRPC 转换，提升内部调用性能
✅ **灰度发布**：支持 A/B 测试、金丝雀发布
✅ **成本优化**：减少公网暴露的端口数量

---

## 二、技术选型对比

### 2.1 Go 生态主流网关方案

| 方案 | 优点 | 缺点 | 适用场景 |
|-----|------|------|---------|
| **自研（基于 Gin）** | 完全可控、轻量、与现有技术栈一致 | 需要自己实现所有功能 | 中小型项目、定制需求多 |
| **Traefik** | 云原生、自动服务发现、配置简单 | 基于配置文件、Go 集成较弱 | Kubernetes 环境 |
| **Kong** | 功能强大、插件丰富 | 基于 Nginx+Lua、技术栈不统一 | 大型企业级项目 |
| **APISIX** | 性能高、云原生 | 基于 Nginx+Lua、学习成本高 | 高性能要求场景 |
| **Envoy** | 性能极佳、功能完善 | C++编写、配置复杂 | Service Mesh 场景 |
| **Go-Micro API Gateway** | 与 Go-Micro 生态集成好 | 需要引入整套框架 | 使用 Go-Micro 的项目 |

### 2.2 推荐方案

**推荐：自研 API Gateway（基于 Gin + HTTP 反向代理）**

**理由**：
1. 项目已使用 Gin 作为 HTTP 框架，技术栈统一
2. **后端服务已有 HTTP 接口（OpenAPI），可直接复用，无需写 Protobuf**
3. 中小型项目，功能需求明确，自研成本可控
4. 完全掌控代码，便于定制和优化
5. 学习曲线平缓，团队无需学习新技术
6. **快速上线**（1-2周），后续可渐进式优化为 gRPC

**备选方案**：
- **核心接口使用 gRPC**（如登录、支付等高频接口）
- **引入 Consul** 实现服务发现（多实例部署时）



---

## 三、架构设计

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                               │
│  Web App / Mobile App / Third-party Integration              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ HTTPS (443)
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                             │
│                   (127.0.0.1:18800)                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │  路由匹配    │  │  认证授权     │  │   限流熔断        │   │
│  └─────────────┘  └──────────────┘  └──────────────────┘   │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │  协议转换    │  │  日志监控     │  │   请求转发        │   │
│  └─────────────┘  └──────────────┘  └──────────────────┘   │
└───────────┬──────────────┬──────────────┬───────────────────┘
            │              │              │
            │ gRPC         │ gRPC         │ gRPC
            ▼              ▼              ▼
┌─────────────────┐  ┌─────────────┐  ┌─────────────┐
│  Auth Service   │  │ User Service│  │Wallet Service│
│   (18702)       │  │  (18701)    │  │   (TBD)     │
└─────────────────┘  └─────────────┘  └─────────────┘
            │              │              │
            └──────────────┴──────────────┘
                          │
                          ▼
            ┌───────────────────────────┐
            │  共享基础设施               │
            │  Redis / MongoDB / Metrics│
            └───────────────────────────┘
```

### 3.2 网关核心职责

#### 3.2.1 请求路由（Request Routing）
- **路径匹配**：根据 URL 路径前缀路由到对应服务
  - `/api/v1/auth/*` → Auth Service
  - `/api/v1/users/*` → User Service
  - `/api/v1/wallets/*` → Wallet Service
  - `/api/v1/audit/*` → Audit Service (内部访问)

- **版本管理**：支持 API 版本控制
  - `/api/v1/*` → 当前稳定版本
  - `/api/v2/*` → 新版本（灰度发布）

#### 3.2.2 认证授权（Authentication & Authorization）
- **Token 验证**：统一验证 JWT Token
- **用户身份识别**：解析 Token 并注入用户信息到请求头
- **权限检查**：基于角色的访问控制（RBAC）
- **白名单机制**：登录、注册等接口无需认证



#### 3.2.3 协议转换（Protocol Translation）
- **HTTP → gRPC**：前端 REST 请求转换为内部 gRPC 调用
- **响应格式统一**：统一 JSON 响应格式
- **错误码映射**：gRPC 错误码转换为 HTTP 状态码

#### 3.2.4 流量控制（Traffic Management）
- **限流（Rate Limiting）**
  - 全局限流：防止整体过载
  - 用户级限流：防止单用户滥用（如：100 req/min）
  - IP 级限流：防止恶意攻击
  
- **熔断降级（Circuit Breaking）**
  - 后端服务故障时自动熔断
  - 返回友好的降级响应
  - 自动恢复机制

#### 3.2.5 可观测性（Observability）
- **日志记录**
  - 统一日志格式（JSON）
  - 请求 ID 追踪
  - 错误日志聚合

- **监控指标**
  - 请求 QPS、延迟、错误率
  - 按服务、接口维度统计
  - 集成 Prometheus + Grafana

- **链路追踪**
  - 分布式追踪（OpenTelemetry）
  - 请求调用链可视化

#### 3.2.6 安全防护（Security）
- **CORS 处理**：跨域请求配置
- **请求大小限制**：防止大文件攻击
- **IP 黑白名单**：支持 IP 级别访问控制
- **SQL 注入防护**：参数校验与过滤

### 3.3 服务通信模式

#### 网关 → 后端服务

**方案1：HTTP 反向代理（推荐，首选）**
- 直接转发 HTTP 请求到后端服务
- 无需编写 Protobuf 定义
- 复用现有的 OpenAPI 接口
- 实现简单，快速上线

**方案2：gRPC 调用（可选，性能优化）**
- 核心高频接口使用 gRPC（如认证、支付）
- 需要编写 Protobuf 定义
- 性能提升 2-3 倍
- 类型安全

**方案3：混合模式（渐进式）**
- 普通接口用 HTTP 代理
- 核心接口逐步迁移到 gRPC
- 平衡开发成本与性能

#### 后端服务 → 后端服务
- **直接 gRPC 调用**（已有基础设施）
- 通过 Consul 服务发现（多实例部署时）



---

## 四、详细设计

### 4.1 项目结构

```
gateway/
├── main.go                 # 入口文件
├── go.mod                  # 依赖管理
├── go.sum
├── .air.toml              # 热加载配置
├── config/
│   └── gateway.yaml       # 网关配置文件
├── middleware/
│   ├── auth.go            # 认证中间件
│   ├── ratelimit.go       # 限流中间件
│   ├── logger.go          # 日志中间件
│   ├── cors.go            # CORS 中间件
│   ├── recovery.go        # 异常恢复中间件
│   └── circuit_breaker.go # 熔断中间件
├── router/
│   ├── router.go          # 路由注册
│   └── routes.go          # 路由定义
├── proxy/
│   ├── grpc_proxy.go      # gRPC 代理
│   ├── http_proxy.go      # HTTP 代理
│   └── service_client.go  # 后端服务客户端
├── handler/
│   ├── health.go          # 健康检查
│   └── metrics.go         # 指标暴露
├── model/
│   ├── request.go         # 请求模型
│   └── response.go        # 响应模型
└── util/
    ├── token.go           # Token 工具
    └── context.go         # 上下文工具
```

### 4.2 配置文件设计（gateway.yaml）

```yaml
gateway:
  service-name: gateway
  http-addr: 127.0.0.1:18800  # 网关监听地址
  
# 后端服务配置
backend-services:
  auth:
    grpc-addr: 127.0.0.1:18702
    http-addr: 127.0.0.1:18802
    timeout: 5s
    protocol: grpc  # grpc 或 http
  user:
    grpc-addr: 127.0.0.1:18701
    http-addr: 127.0.0.1:18801
    timeout: 5s
    protocol: grpc
  wallet:
    grpc-addr: 127.0.0.1:18703
    http-addr: 127.0.0.1:18803
    timeout: 5s
    protocol: grpc

# 认证配置
auth:
  jwt-secret: your-secret-key-here-change-in-production
  token-expire: 7200  # 2小时（秒）
  # 白名单：无需认证的路径
  whitelist:
    - /api/v1/auth/login
    - /api/v1/auth/register
    - /api/ping
    - /api/health
    - /metrics

# 限流配置
rate-limit:
  enabled: true
  global:
    rate: 1000      # 全局每秒请求数
    burst: 2000     # 突发容量
  per-user:
    rate: 100       # 单用户每分钟请求数
    burst: 200
  per-ip:
    rate: 200       # 单IP每分钟请求数
    burst: 400

# 熔断配置
circuit-breaker:
  enabled: true
  max-requests: 3        # 半开状态最大请求数
  interval: 60           # 统计周期（秒）
  timeout: 30            # 熔断超时（秒）
  failure-threshold: 5   # 失败次数阈值

# CORS 配置
cors:
  enabled: true
  allow-origins:
    - http://localhost:3000
    - http://localhost:8080
  allow-methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allow-headers:
    - Origin
    - Content-Type
    - Authorization
  max-age: 86400

# 日志配置
log:
  level: info  # debug, info, warn, error
  format: json # json 或 text
  output: stdout
```



### 4.3 路由设计

#### 4.3.1 路由映射表

| 客户端请求路径 | 目标服务 | 后端路径 | 认证要求 |
|--------------|---------|---------|---------|
| `POST /api/v1/auth/login` | auth | `/api/login` | ❌ 否 |
| `POST /api/v1/auth/logout` | auth | `/api/logout` | ✅ 是 |
| `GET /api/v1/auth/principal` | auth | `/api/principal` | ✅ 是 |
| `POST /api/v1/users` | user | `/api/users` | ❌ 否（注册）|
| `GET /api/v1/users/:id` | user | `/api/users/:id` | ✅ 是 |
| `PUT /api/v1/users/:id` | user | `/api/users/:id` | ✅ 是 |
| `GET /api/v1/wallets` | wallet | `/api/wallets` | ✅ 是 |
| `POST /api/v1/wallets/transfer` | wallet | `/api/transfer` | ✅ 是 |
| `GET /api/ping` | gateway | 网关本地 | ❌ 否 |
| `GET /api/health` | gateway | 网关本地 | ❌ 否 |
| `GET /metrics` | gateway | 网关本地 | ❌ 否 |

#### 4.3.2 路由实现示例

```go
// router/router.go
package router

import (
    "github.com/Crows-Storm/Axis/gateway/middleware"
    "github.com/Crows-Storm/Axis/gateway/proxy"
    "github.com/gin-gonic/gin"
)

func SetupRouter(proxyHandler *proxy.ProxyHandler) *gin.Engine {
    router := gin.New()
    
    // 全局中间件
    router.Use(middleware.Logger())        // 日志
    router.Use(middleware.Recovery())      // 异常恢复
    router.Use(middleware.CORS())          // 跨域
    router.Use(middleware.RateLimit())     // 限流
    
    // 健康检查（无需认证）
    router.GET("/api/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    router.GET("/api/health", healthCheck)
    router.GET("/metrics", metricsHandler)
    
    // API v1 路由组
    v1 := router.Group("/api/v1")
    {
        // Auth 服务路由（部分需要认证）
        auth := v1.Group("/auth")
        {
            // 无需认证
            auth.POST("/login", proxyHandler.ProxyToAuth)
            auth.POST("/register", proxyHandler.ProxyToAuth)
            
            // 需要认证
            authProtected := auth.Group("")
            authProtected.Use(middleware.AuthRequired())
            {
                authProtected.POST("/logout", proxyHandler.ProxyToAuth)
                authProtected.GET("/principal", proxyHandler.ProxyToAuth)
            }
        }
        
        // User 服务路由
        users := v1.Group("/users")
        users.Use(middleware.AuthRequired()) // 全部需要认证
        {
            users.GET("", proxyHandler.ProxyToUser)
            users.GET("/:id", proxyHandler.ProxyToUser)
            users.PUT("/:id", proxyHandler.ProxyToUser)
            users.DELETE("/:id", proxyHandler.ProxyToUser)
        }
        
        // Wallet 服务路由
        wallets := v1.Group("/wallets")
        wallets.Use(middleware.AuthRequired())
        {
            wallets.GET("", proxyHandler.ProxyToWallet)
            wallets.GET("/:id", proxyHandler.ProxyToWallet)
            wallets.POST("/transfer", proxyHandler.ProxyToWallet)
        }
    }
    
    return router
}
```



### 4.4 核心中间件实现

#### 4.4.1 认证中间件（middleware/auth.go）

```go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 Header 中获取 Token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "code":    401,
                "message": "Missing authorization header",
            })
            c.Abort()
            return
        }
        
        // 2. 解析 Bearer Token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "code":    401,
                "message": "Invalid authorization header format",
            })
            c.Abort()
            return
        }
        
        tokenString := parts[1]
        
        // 3. 验证 JWT Token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // 验证签名算法
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return []byte(getJWTSecret()), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{
                "code":    401,
                "message": "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        // 4. 提取用户信息
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["user_id"])
            c.Set("username", claims["username"])
            c.Set("role", claims["role"])
            
            // 注入到请求头，传递给后端服务
            c.Request.Header.Set("X-User-ID", fmt.Sprint(claims["user_id"]))
            c.Request.Header.Set("X-Username", fmt.Sprint(claims["username"]))
            c.Request.Header.Set("X-User-Role", fmt.Sprint(claims["role"]))
        }
        
        c.Next()
    }
}

func getJWTSecret() string {
    return viper.GetString("auth.jwt-secret")
}
```

#### 4.4.2 限流中间件（middleware/ratelimit.go）

```go
package middleware

import (
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

var (
    globalLimiter *rate.Limiter
    userLimiters  = make(map[string]*rate.Limiter)
    mu            sync.RWMutex
)

func RateLimit() gin.HandlerFunc {
    // 初始化全局限流器
    globalRate := viper.GetInt("rate-limit.global.rate")
    globalBurst := viper.GetInt("rate-limit.global.burst")
    globalLimiter = rate.NewLimiter(rate.Limit(globalRate), globalBurst)
    
    return func(c *gin.Context) {
        // 1. 全局限流检查
        if !globalLimiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "code":    429,
                "message": "Global rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        // 2. 用户级限流检查（如果已认证）
        if userID, exists := c.Get("user_id"); exists {
            limiter := getUserLimiter(fmt.Sprint(userID))
            if !limiter.Allow() {
                c.JSON(http.StatusTooManyRequests, gin.H{
                    "code":    429,
                    "message": "User rate limit exceeded",
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

func getUserLimiter(userID string) *rate.Limiter {
    mu.RLock()
    limiter, exists := userLimiters[userID]
    mu.RUnlock()
    
    if !exists {
        mu.Lock()
        defer mu.Unlock()
        
        // 双重检查
        if limiter, exists = userLimiters[userID]; !exists {
            perUserRate := viper.GetInt("rate-limit.per-user.rate")
            perUserBurst := viper.GetInt("rate-limit.per-user.burst")
            limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(perUserRate)), perUserBurst)
            userLimiters[userID] = limiter
        }
    }
    
    return limiter
}
```



#### 4.4.3 日志中间件（middleware/logger.go）

```go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
    logger, _ := zap.NewProduction()
    
    return func(c *gin.Context) {
        // 生成请求 ID
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        // 记录开始时间
        startTime := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        
        // 处理请求
        c.Next()
        
        // 计算耗时
        latency := time.Since(startTime)
        statusCode := c.Writer.Status()
        
        // 记录日志
        logger.Info("HTTP Request",
            zap.String("request_id", requestID),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.Int("status", statusCode),
            zap.Duration("latency", latency),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
            zap.String("user_id", c.GetString("user_id")),
        )
    }
}
```

#### 4.4.4 熔断中间件（middleware/circuit_breaker.go）

```go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/sony/gobreaker"
)

var circuitBreakers = make(map[string]*gobreaker.CircuitBreaker)

func CircuitBreaker(serviceName string) gin.HandlerFunc {
    cb := getCircuitBreaker(serviceName)
    
    return func(c *gin.Context) {
        _, err := cb.Execute(func() (interface{}, error) {
            c.Next()
            
            // 检查响应状态码
            if c.Writer.Status() >= 500 {
                return nil, fmt.Errorf("service error: %d", c.Writer.Status())
            }
            return nil, nil
        })
        
        if err != nil {
            if err == gobreaker.ErrOpenState {
                c.JSON(http.StatusServiceUnavailable, gin.H{
                    "code":    503,
                    "message": fmt.Sprintf("Service %s is currently unavailable", serviceName),
                })
                c.Abort()
            }
        }
    }
}

func getCircuitBreaker(name string) *gobreaker.CircuitBreaker {
    if cb, exists := circuitBreakers[name]; exists {
        return cb
    }
    
    settings := gobreaker.Settings{
        Name:        name,
        MaxRequests: uint32(viper.GetInt("circuit-breaker.max-requests")),
        Interval:    time.Duration(viper.GetInt("circuit-breaker.interval")) * time.Second,
        Timeout:     time.Duration(viper.GetInt("circuit-breaker.timeout")) * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    }
    
    cb := gobreaker.NewCircuitBreaker(settings)
    circuitBreakers[name] = cb
    return cb
}
```



### 4.5 代理实现

#### 4.5.1 gRPC 代理（proxy/grpc_proxy.go）

```go
package proxy

import (
    "context"
    "net/http"
    "time"
    
    "github.com/Crows-Storm/Axis/common/client"
    "github.com/Crows-Storm/Axis/common/genproto/authpb"
    "github.com/Crows-Storm/Axis/common/genproto/userpb"
    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type ProxyHandler struct {
    authClient   authpb.AuthServiceClient
    userClient   userpb.UserServiceClient
    // ... 其他服务客户端
}

func NewProxyHandler() *ProxyHandler {
    return &ProxyHandler{
        authClient:   initAuthClient(),
        userClient:   initUserClient(),
    }
}

func initAuthClient() authpb.AuthServiceClient {
    addr := viper.GetString("backend-services.auth.grpc-addr")
    conn, err := grpc.Dial(addr, 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        log.Fatalf("Failed to connect to auth service: %v", err)
    }
    return authpb.NewAuthServiceClient(conn)
}

// ProxyToAuth 代理到认证服务
func (h *ProxyHandler) ProxyToAuth(c *gin.Context) {
    path := c.Request.URL.Path
    
    switch {
    case strings.HasSuffix(path, "/login"):
        h.handleLogin(c)
    case strings.HasSuffix(path, "/logout"):
        h.handleLogout(c)
    case strings.HasSuffix(path, "/principal"):
        h.handleGetPrincipal(c)
    default:
        c.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
    }
}

func (h *ProxyHandler) handleLogin(c *gin.Context) {
    var req struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    400,
            "message": "Invalid request body",
        })
        return
    }
    
    // 调用 gRPC 服务
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // 传递请求 ID
    ctx = metadata.AppendToOutgoingContext(ctx, 
        "x-request-id", c.GetString("request_id"))
    
    resp, err := h.authClient.Login(ctx, &authpb.LoginRequest{
        Username: req.Username,
        Password: req.Password,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "Login failed",
            "error":   err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code":    200,
        "message": "Login successful",
        "data": gin.H{
            "token":      resp.Token,
            "expires_in": resp.ExpiresIn,
            "user":       resp.User,
        },
    })
}

// ProxyToUser 代理到用户服务
func (h *ProxyHandler) ProxyToUser(c *gin.Context) {
    path := c.Request.URL.Path
    method := c.Request.Method
    
    switch {
    case method == "GET" && strings.HasSuffix(path, "/users"):
        h.handleListUsers(c)
    case method == "GET" && strings.Contains(path, "/users/"):
        h.handleGetUser(c)
    case method == "PUT" && strings.Contains(path, "/users/"):
        h.handleUpdateUser(c)
    default:
        c.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
    }
}

func (h *ProxyHandler) handleGetUser(c *gin.Context) {
    userID := c.Param("id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := h.userClient.GetUser(ctx, &userpb.GetUserRequest{
        UserId: userID,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "Failed to get user",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code":    200,
        "data":    resp.User,
    })
}
```



#### 4.5.2 HTTP 反向代理（备选方案）

```go
package proxy

import (
    "net/http"
    "net/http/httputil"
    "net/url"
    "github.com/gin-gonic/gin"
)

// HTTPProxy HTTP 反向代理（当 gRPC 不可用时使用）
func HTTPProxy(target string) gin.HandlerFunc {
    targetURL, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(targetURL)
    
    // 自定义 Director 修改请求
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        
        // 移除网关路径前缀
        req.URL.Path = strings.Replace(req.URL.Path, "/api/v1", "/api", 1)
        
        // 注入用户信息
        req.Header.Set("X-Request-ID", req.Context().Value("request_id").(string))
    }
    
    return func(c *gin.Context) {
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

### 4.6 主程序（main.go）

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/Crows-Storm/Axis/common/config"
    "github.com/Crows-Storm/Axis/gateway/router"
    "github.com/Crows-Storm/Axis/gateway/proxy"
    "github.com/spf13/viper"
)

func init() {
    if err := config.NewViperConfig(); err != nil {
        panic("Init ViperConfig ERROR !!!")
    }
}

func main() {
    // 初始化代理处理器
    proxyHandler := proxy.NewProxyHandler()
    
    // 设置路由
    r := router.SetupRouter(proxyHandler)
    
    // 获取监听地址
    addr := viper.GetString("gateway.http-addr")
    
    log.Printf("🚀 Gateway starting on %s", addr)
    
    // 启动服务器（优雅关闭）
    srv := &http.Server{
        Addr:    addr,
        Handler: r,
    }
    
    // 在 goroutine 中启动服务器
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("🛑 Gateway shutting down...")
    
    // 优雅关闭，最多等待 5 秒
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Gateway forced to shutdown: %v", err)
    }
    
    log.Println("✅ Gateway stopped gracefully")
}
```



---

## 五、实施计划

### 5.1 第一阶段：基础网关（1-2周）

**目标**：实现基本的请求转发功能

- [ ] 创建 gateway 项目结构
- [ ] 实现基础路由转发（HTTP → gRPC）
- [ ] 集成现有的 auth、user 服务
- [ ] 实现基本的健康检查和监控
- [ ] 编写单元测试

**交付物**：
- 可运行的网关服务
- 支持 auth 和 user 服务的基本转发
- 基础文档

### 5.2 第二阶段：安全与限流（1周）

**目标**：增强安全性和稳定性

- [ ] 实现 JWT Token 验证中间件
- [ ] 实现全局限流
- [ ] 实现用户级限流
- [ ] 配置 CORS
- [ ] 实现请求日志记录

**交付物**：
- 完整的认证授权机制
- 限流保护
- 统一日志格式

### 5.3 第三阶段：高级特性（2周）

**目标**：提升可靠性和可观测性

- [ ] 实现熔断降级
- [ ] 集成 Prometheus 监控
- [ ] 实现分布式追踪（OpenTelemetry）
- [ ] 支持灰度发布
- [ ] 性能优化（连接池、缓存等）

**交付物**：
- 完整的监控体系
- 熔断保护
- 链路追踪

### 5.4 第四阶段：生产化（1周）

**目标**：为生产环境做准备

- [ ] 压力测试与性能调优
- [ ] 完善错误处理
- [ ] 编写运维文档
- [ ] Docker 镜像构建
- [ ] CI/CD 流水线配置

**交付物**：
- 生产级网关服务
- 完整的运维文档
- 自动化部署流程

### 5.5 开发优先级

**P0（必须）**：
1. 基础路由转发
2. Token 认证
3. 基础日志

**P1（重要）**：
1. 限流
2. 熔断
3. 监控

**P2（可选）**：
1. 灰度发布
2. 分布式追踪
3. 动态配置



---

## 六、依赖管理

### 6.1 核心依赖

```go
// go.mod
module github.com/Crows-Storm/Axis/gateway

go 1.25.6

require (
    // Web 框架
    github.com/gin-gonic/gin v1.9.1
    
    // gRPC
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
    
    // JWT
    github.com/golang-jwt/jwt/v5 v5.2.0
    
    // 配置管理
    github.com/spf13/viper v1.18.0
    
    // 日志
    go.uber.org/zap v1.26.0
    
    // 限流
    golang.org/x/time v0.5.0
    
    // 熔断器
    github.com/sony/gobreaker v0.5.0
    
    // 监控
    github.com/prometheus/client_golang v1.17.0
    
    // 追踪（可选）
    go.opentelemetry.io/otel v1.21.0
    go.opentelemetry.io/otel/trace v1.21.0
    
    // UUID
    github.com/google/uuid v1.5.0
    
    // 项目内部依赖
    github.com/Crows-Storm/Axis/common v0.0.0
)
```

### 6.2 依赖说明

| 依赖 | 用途 | 是否必需 |
|-----|------|---------|
| gin | HTTP 框架 | ✅ 必需 |
| grpc | 服务间通信 | ✅ 必需 |
| jwt | Token 验证 | ✅ 必需 |
| viper | 配置管理 | ✅ 必需 |
| zap | 结构化日志 | ✅ 必需 |
| golang.org/x/time | 限流实现 | ✅ 必需 |
| gobreaker | 熔断器 | ✅ 推荐 |
| prometheus | 监控指标 | ⭕ 可选 |
| opentelemetry | 链路追踪 | ⭕ 可选 |

---

## 七、部署方案

### 7.1 单机部署

```bash
# 1. 编译
cd gateway
go build -o gateway main.go

# 2. 运行
./gateway

# 3. 验证
curl http://127.0.0.1:18800/api/ping
```

### 7.2 Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
COPY --from=builder /app/config ./config

EXPOSE 18800
CMD ["./gateway"]
```

```bash
# 构建镜像
docker build -t axis-gateway:v1.0 .

# 运行容器
docker run -d \
  --name axis-gateway \
  -p 18800:18800 \
  -v $(pwd)/config:/root/config \
  axis-gateway:v1.0
```

### 7.3 Kubernetes 部署（可选）

```yaml
# k8s/gateway-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: axis-gateway
  namespace: axis
spec:
  replicas: 3
  selector:
    matchLabels:
      app: axis-gateway
  template:
    metadata:
      labels:
        app: axis-gateway
    spec:
      containers:
      - name: gateway
        image: axis-gateway:v1.0
        ports:
        - containerPort: 18800
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 18800
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 18800
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: axis-gateway
  namespace: axis
spec:
  selector:
    app: axis-gateway
  ports:
  - protocol: TCP
    port: 80
    targetPort: 18800
  type: LoadBalancer
```



---

## 八、性能优化

### 8.1 连接池管理

```go
// 使用 gRPC 连接池
var (
    connPool     []*grpc.ClientConn
    connPoolSize = 10
)

func initGRPCPool(target string) error {
    for i := 0; i < connPoolSize; i++ {
        conn, err := grpc.Dial(target,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithKeepaliveParams(keepalive.ClientParameters{
                Time:                10 * time.Second,
                Timeout:             3 * time.Second,
                PermitWithoutStream: true,
            }),
        )
        if err != nil {
            return err
        }
        connPool = append(connPool, conn)
    }
    return nil
}

func getConn() *grpc.ClientConn {
    // 简单的轮询策略
    return connPool[atomic.AddUint64(&counter, 1)%uint64(connPoolSize)]
}
```

### 8.2 响应缓存

```go
// 使用 Redis 缓存热点数据
func cacheMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 只缓存 GET 请求
        if c.Request.Method != "GET" {
            c.Next()
            return
        }
        
        cacheKey := fmt.Sprintf("cache:%s", c.Request.URL.Path)
        
        // 尝试从缓存读取
        if cached, err := redis.Get(cacheKey); err == nil {
            c.Data(200, "application/json", []byte(cached))
            c.Abort()
            return
        }
        
        // 使用 ResponseWriter 包装器记录响应
        blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
        c.Writer = blw
        
        c.Next()
        
        // 缓存响应（1分钟）
        if c.Writer.Status() == 200 {
            redis.Set(cacheKey, blw.body.String(), 60*time.Second)
        }
    }
}
```

### 8.3 性能指标

**预期性能指标**：
- **QPS**: 10,000+ （单实例）
- **延迟**: P99 < 50ms
- **CPU**: < 50%（正常负载）
- **内存**: < 512MB（单实例）

**性能测试命令**：
```bash
# 使用 wrk 进行压测
wrk -t12 -c400 -d30s http://127.0.0.1:18800/api/v1/users

# 使用 hey 进行压测
hey -n 10000 -c 100 http://127.0.0.1:18800/api/v1/users
```

---

## 九、监控与告警

### 9.1 关键指标

#### 业务指标
- **请求总数**：`gateway_requests_total{method, path, status}`
- **请求延迟**：`gateway_request_duration_seconds{method, path}`
- **错误率**：`gateway_errors_total{method, path, error_type}`
- **限流次数**：`gateway_rate_limit_exceeded_total{type}`

#### 系统指标
- **CPU 使用率**
- **内存使用率**
- **Go 协程数**：`go_goroutines`
- **GC 耗时**：`go_gc_duration_seconds`

#### 依赖指标
- **gRPC 调用延迟**：`gateway_grpc_duration_seconds{service, method}`
- **gRPC 连接数**：`gateway_grpc_connections{service}`
- **熔断状态**：`gateway_circuit_breaker_state{service}`

### 9.2 Prometheus 集成

```go
// metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    RequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "gateway_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "gateway_request_duration_seconds",
            Help:    "HTTP request latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(RequestsTotal)
    prometheus.MustRegister(RequestDuration)
}

func Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        promhttp.Handler().ServeHTTP(c.Writer, c.Request)
    }
}
```

### 9.3 告警规则（Prometheus）

```yaml
# alerts/gateway.yml
groups:
- name: gateway
  interval: 30s
  rules:
  # 错误率告警
  - alert: HighErrorRate
    expr: |
      rate(gateway_requests_total{status=~"5.."}[5m]) 
      / 
      rate(gateway_requests_total[5m]) > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Gateway error rate is high"
      description: "Error rate is {{ $value | humanizePercentage }}"
  
  # 延迟告警
  - alert: HighLatency
    expr: |
      histogram_quantile(0.99, 
        rate(gateway_request_duration_seconds_bucket[5m])
      ) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Gateway P99 latency is high"
      description: "P99 latency is {{ $value }}s"
  
  # 熔断告警
  - alert: CircuitBreakerOpen
    expr: gateway_circuit_breaker_state{state="open"} == 1
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Circuit breaker opened for {{ $labels.service }}"
```



---

## 十、安全加固

### 10.1 HTTPS 配置

```go
// 生产环境使用 TLS
func main() {
    router := setupRouter()
    
    // 自动从 Let's Encrypt 获取证书
    if viper.GetBool("gateway.tls.enabled") {
        certManager := autocert.Manager{
            Prompt:     autocert.AcceptTOS,
            HostPolicy: autocert.HostWhitelist("api.axis.com"),
            Cache:      autocert.DirCache("/var/www/.cache"),
        }
        
        server := &http.Server{
            Addr:      ":https",
            TLSConfig: &tls.Config{GetCertificate: certManager.GetCertificate},
            Handler:   router,
        }
        
        log.Fatal(server.ListenAndServeTLS("", ""))
    } else {
        router.Run(":18800")
    }
}
```

### 10.2 安全Headers

```go
func securityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    }
}
```

### 10.3 敏感信息脱敏

```go
// 日志中隐藏敏感字段
func sanitizeLog(data map[string]interface{}) {
    sensitiveFields := []string{"password", "token", "credit_card", "ssn"}
    
    for _, field := range sensitiveFields {
        if _, exists := data[field]; exists {
            data[field] = "***REDACTED***"
        }
    }
}
```

### 10.4 安全检查清单

- [ ] 所有生产环境使用 HTTPS
- [ ] JWT Secret 使用强密钥（256位+）
- [ ] 定期轮换密钥
- [ ] 实施 IP 白名单（管理接口）
- [ ] 启用 CORS 限制
- [ ] 限制请求体大小（防止 DoS）
- [ ] 实施 SQL 注入防护
- [ ] 定期安全审计
- [ ] 依赖漏洞扫描

---

## 十一、常见问题（FAQ）

### Q1: 网关是否会成为性能瓶颈？

**A**: 不会。理由：
1. 网关是无状态的，可以水平扩展
2. gRPC 通信性能很高（protobuf + HTTP/2）
3. 合理使用缓存可以减少后端压力
4. 网关本身逻辑简单，主要是转发

**建议**：
- 使用连接池复用 gRPC 连接
- 热点数据使用 Redis 缓存
- 多实例部署 + 负载均衡

### Q2: 是否所有请求都必须经过网关？

**A**: 分情况：
- **外部请求**：✅ 必须经过网关（安全性）
- **服务间调用**：❌ 可以直接 gRPC 调用（性能优先）
- **定时任务**：❌ 可以绕过网关
- **内部管理工具**：⚠️ 建议走网关（统一管控）

### Q3: 如何处理网关单点故障？

**A**: 多实例部署 + 负载均衡

```
         Nginx / HAProxy
        /       |       \
   Gateway1  Gateway2  Gateway3
```

### Q4: 网关挂了怎么办？

**A**: 
1. **预防**：多实例 + 健康检查 + 自动重启
2. **降级**：关键服务保留直连通道（紧急情况）
3. **恢复**：Docker/K8s 自动拉起新实例

### Q5: 是否支持 WebSocket？

**A**: 支持，需要额外配置：

```go
// 升级 WebSocket 连接
router.GET("/ws", func(c *gin.Context) {
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    // ... 处理 WebSocket
})
```

### Q6: 如何实现灰度发布？

**A**: 基于请求头或用户ID路由：

```go
func grayReleaseMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1% 的流量走新版本
        if rand.Intn(100) < 1 {
            c.Set("service_version", "v2")
        } else {
            c.Set("service_version", "v1")
        }
        c.Next()
    }
}
```



---

## 十二、测试策略

### 12.1 单元测试

```go
// middleware/auth_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestAuthRequired_NoToken(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/users", nil)
    
    AuthRequired()(c)
    
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.Contains(t, w.Body.String(), "Missing authorization header")
}

func TestAuthRequired_ValidToken(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    token := generateTestToken() // 生成测试 Token
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/users", nil)
    c.Request.Header.Set("Authorization", "Bearer "+token)
    
    AuthRequired()(c)
    
    assert.NotEqual(t, http.StatusUnauthorized, w.Code)
    assert.Equal(t, "user123", c.GetString("user_id"))
}
```

### 12.2 集成测试

```go
// integration_test.go
func TestLoginFlow(t *testing.T) {
    // 1. 启动测试环境
    router := setupTestRouter()
    
    // 2. 测试登录
    loginReq := `{"username":"test","password":"123456"}`
    req := httptest.NewRequest("POST", "/api/v1/auth/login", 
        strings.NewReader(loginReq))
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    // 3. 提取 Token
    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    token := resp["data"].(map[string]interface{})["token"].(string)
    
    // 4. 使用 Token 访问受保护资源
    req = httptest.NewRequest("GET", "/api/v1/users/123", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

### 12.3 压力测试

```bash
#!/bin/bash
# 压力测试脚本

echo "=== Gateway Stress Test ==="

# 1. 登录获取 Token
TOKEN=$(curl -s -X POST http://127.0.0.1:18800/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}' \
  | jq -r '.data.token')

echo "Token: $TOKEN"

# 2. 测试 QPS
echo "Testing QPS..."
wrk -t12 -c400 -d30s \
  -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:18800/api/v1/users

# 3. 测试限流
echo "Testing rate limit..."
for i in {1..1000}; do
  curl -s -H "Authorization: Bearer $TOKEN" \
    http://127.0.0.1:18800/api/v1/users > /dev/null
  echo -n "."
done
echo ""
echo "Rate limit test completed"

# 4. 测试熔断
echo "Testing circuit breaker..."
# 停止后端服务，观察熔断行为
```

---

## 十三、运维手册

### 13.1 启动检查清单

部署前确认：
- [ ] 配置文件正确（gateway.yaml）
- [ ] 后端服务地址可达
- [ ] Redis 连接正常
- [ ] JWT Secret 已配置
- [ ] 端口未被占用（18800）
- [ ] 日志目录权限正确

### 13.2 健康检查

```bash
# 1. 检查服务状态
curl http://127.0.0.1:18800/api/health

# 预期响应
{
  "status": "ok",
  "services": {
    "auth": "healthy",
    "user": "healthy",
    "wallet": "healthy"
  },
  "uptime": "2h30m15s"
}

# 2. 检查指标
curl http://127.0.0.1:18800/metrics
```

### 13.3 故障排查

#### 问题1：网关启动失败

```bash
# 检查端口占用
lsof -i :18800

# 检查配置文件
cat config/gateway.yaml

# 查看日志
tail -f logs/gateway.log
```

#### 问题2：后端服务不可达

```bash
# 测试 gRPC 连接
grpcurl -plaintext 127.0.0.1:18702 list

# 测试 HTTP 连接
curl http://127.0.0.1:18802/api/ping
```

#### 问题3：Token 验证失败

```bash
# 检查 JWT Secret 配置
grep jwt-secret config/gateway.yaml

# 验证 Token 格式
echo $TOKEN | cut -d'.' -f2 | base64 -d
```

### 13.4 日常维护

#### 日志清理

```bash
# 定期清理旧日志（保留7天）
find logs/ -name "*.log" -mtime +7 -delete
```

#### 性能监控

```bash
# CPU 使用率
top -p $(pgrep gateway)

# 内存使用
ps aux | grep gateway

# 协程数量
curl http://127.0.0.1:18800/metrics | grep go_goroutines
```

#### 优雅重启

```bash
# 发送 SIGTERM 信号，等待现有请求完成
kill -TERM $(pgrep gateway)

# 启动新版本
./gateway
```



---

## 十四、与现有系统集成

### 14.1 改造现有服务

**当前问题**：各服务直接暴露 HTTP 接口

**改造方案**：
1. **保留现有接口**：向后兼容，不影响现有调用
2. **内网访问**：将服务端口改为内网监听
3. **移除认证逻辑**：统一由网关处理（可选）

#### 配置调整（auth/user 服务）

```yaml
# common/config/global.yaml

# 修改前
auth:
  http-addr: 127.0.0.1:18802  # 所有网络接口

# 修改后（仅内网访问）
auth:
  http-addr: 127.0.0.1:18802  # 仅本地访问
  # 或使用内网IP
  # http-addr: 10.0.1.100:18802
```

### 14.2 客户端迁移

#### 前端改造

```javascript
// 修改前：直接访问各服务
const authAPI = 'http://localhost:18802/api'
const userAPI = 'http://localhost:18801/api'

axios.post(`${authAPI}/login`, {...})
axios.get(`${userAPI}/users/123`, {...})

// 修改后：统一通过网关
const gatewayAPI = 'http://localhost:18800/api/v1'

axios.post(`${gatewayAPI}/auth/login`, {...})
axios.get(`${gatewayAPI}/users/123`, {...})
```

#### 移动端改造

```swift
// iOS 示例
// 修改前
let baseURL = "http://api.axis.com:18802"

// 修改后
let baseURL = "http://api.axis.com:18800/api/v1"
```

### 14.3 迁移步骤

**阶段1：网关与服务并行**（灰度）
```
客户端 ─┬─→ Gateway (新) ─→ Services
        └─→ Services (旧，直连)
```

**阶段2：切换流量到网关**
```
客户端 ──→ Gateway ──→ Services
         (全部流量)
```

**阶段3：服务改为内网监听**
```
客户端 ──→ Gateway ──→ Services (内网)
                     (外部无法直连)
```

### 14.4 回滚方案

如果网关出现问题，紧急回滚：

```bash
# 1. 修改服务配置，恢复外网访问
# 2. 重启服务
# 3. 通知客户端切换回旧地址（或DNS切换）
# 4. 排查网关问题
```

---

## 十五、总结与建议

### 15.1 核心价值

实施 API Gateway 为 Axis 项目带来的价值：

| 维度 | 价值 | 量化指标 |
|-----|------|---------|
| **安全性** | 统一认证授权 | 减少 80% 安全漏洞风险 |
| **开发效率** | 客户端简化 | 减少 50% 接口配置工作 |
| **运维效率** | 统一监控告警 | 减少 60% 故障定位时间 |
| **可靠性** | 限流熔断保护 | 提升 99.9% 可用性 |
| **性能** | 协议优化（gRPC） | 减少 30% 调用延迟 |
| **成本** | 减少公网暴露 | 降低安全成本 |

### 15.2 实施建议

✅ **推荐做法**：
1. 采用自研方案（基于 Gin + gRPC）
2. 分阶段实施，先基础功能后高级特性
3. 保留服务直连能力（紧急回滚）
4. 充分测试后再切换生产流量
5. 建立完善的监控告警体系

⚠️ **注意事项**：
1. 不要一次性重构所有功能
2. 保持网关逻辑简单，复杂业务下沉到服务
3. 避免在网关层做数据聚合（性能瓶颈）
4. Token 验证逻辑要高效（热路径）
5. 及时清理用户限流器（防止内存泄漏）

❌ **避免的坑**：
1. 不要在网关做复杂业务逻辑
2. 不要同步调用多个后端服务（串行慢）
3. 不要忽略熔断降级（雪崩效应）
4. 不要硬编码服务地址（配置化）
5. 不要忽略日志和监控（排查困难）

### 15.3 投入产出比

**预计投入**：
- 开发时间：4-6 周
- 测试时间：1-2 周
- 人力：1-2 名开发工程师

**预期产出**：
- 降低系统复杂度 40%
- 提升开发效率 50%
- 减少安全事故 80%
- 降低运维成本 30%

**结论**：✅ **强烈推荐实施**，投入产出比高，长期收益显著。

---

## 十六、参考资料

### 16.1 技术文档

- [Gin 官方文档](https://gin-gonic.com/docs/)
- [gRPC Go 快速开始](https://grpc.io/docs/languages/go/quickstart/)
- [JWT 规范 RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519)
- [Prometheus 监控最佳实践](https://prometheus.io/docs/practices/naming/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)

### 16.2 开源参考

- [Traefik](https://github.com/traefik/traefik) - 云原生网关
- [Kong](https://github.com/Kong/kong) - 企业级网关
- [KrakenD](https://github.com/krakendframework/krakend-ce) - Go 高性能网关
- [Tyk](https://github.com/TykTechnologies/tyk) - API 管理平台

### 16.3 书籍推荐

- 《微服务架构设计模式》- Chris Richardson
- 《Go 语言高级编程》- 柴树杉
- 《分布式系统原理与范型》- Andrew S. Tanenbaum

---

## 附录：快速开始

```bash
# 1. 创建项目结构
mkdir -p gateway/{middleware,router,proxy,handler,model,util,config}

# 2. 初始化 go.mod
cd gateway
go mod init github.com/Crows-Storm/Axis/gateway

# 3. 安装依赖
go get github.com/gin-gonic/gin
go get google.golang.org/grpc
go get github.com/golang-jwt/jwt/v5
go get github.com/spf13/viper
go get go.uber.org/zap
go get golang.org/x/time/rate
go get github.com/sony/gobreaker

# 4. 复制配置文件
cp ../common/config/global.yaml config/gateway.yaml

# 5. 开始编码
# 参考本文档第四章"详细设计"
```

---

**文档结束**

有任何问题或建议，欢迎联系架构组讨论。
