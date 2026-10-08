---
name: axis-makefile
description: >-
  Run Axis repository tasks through the root Makefile instead of ad-hoc
  scripts. Use when generating protobuf, OpenAPI, or mocks, running tests,
  tidying Go modules, or cleaning generated code in this repo.
---

# Axis Makefile

在仓库根目录执行 `make`。不要手写等价的 `protoc`、`oapi-codegen`、`mockery` 或跨模块循环。不要手改生成物。

生成物：`*.pb.go`、`*.gen.go`、`mock_*.go`。

## 选哪个目标

| 改了什么 | 命令 |
| --- | --- |
| `api/*pb/*.proto` | `make genproto` |
| `api/openapi/*-swagger.yaml` | `make genopenapi` |
| 接口，且该模块有 `.mockery.yaml` | `make gen-mock-svc SVC=<service>` |
| 上面几项一起，或不确定 | `make gen` |
| 全部模块测试 | `make test` |
| 只要 auth | `make test-auth` |
| `go.mod` / `go.sum` 漂移 | `make gotidy` |

`make gen` 依次执行 `genproto`、`genopenapi`、`gen-mock`、`gotidy`。单项目标已经够用时不要跑全量 `gen`。

`make clean`、`make clean-mock`、`make clean-mock-svc` 会删生成文件。只有用户明确要求清理时才跑。

## 目标

代码生成：

- `make genproto` — `api/*pb/*.proto` → `common/genproto/<name>pb/`。依赖 `protoc`、`protoc-gen-go`、`protoc-gen-go-grpc`。
- `make genopenapi` — `api/openapi/<svc>-swagger.yaml`（或 `.yml`）。客户端写入 `common/client/<svc>/`，服务端写入 `<svc>/protos/`。依赖 `oapi-codegen`。
- `make gen-mock` — 先 `common`，再每个同时有 `go.mod` 和 `.mockery.yaml` 的服务。mock 与接口同目录，文件名 `mock_*.go`。依赖 `mockery`（`github.com/vektra/mockery/v2`）。
- `make gen-mock-svc SVC=<service>` — 只生成一个服务。省略 `SVC` 会失败。

测试与维护：

- `make test` — 每个 `go.mod`（排除 `*/dagger/*`）执行 `go test -v -race ./...`。任一模块失败即停。
- `make test-auth` — `auth` 模块，`-race`。
- `make test-auth-coverage` — 覆盖率报告在 `auth/coverage.html`。
- `make test-auth-grpc` — `auth/adapters/grpc`。
- `make test-bench` — `auth` 的 benchmark。
- `make gotidy` — 每个 `go.mod`（排除 `*/dagger/*`）执行 `go mod tidy`。任一模块失败即停。
- `make help` — 列出目标。`make` 默认就是 help。

现有模块：`common`、`auth`、`user`、`account`、`ledger`、`audit`、`notification`、`gateway`、`market`、`exchange`、`chain`。mock 只覆盖带 `.mockery.yaml` 的模块。

单包测试没有对应目标。用户只要一个包时，进入该模块目录执行 `go test`，不要为此改 Makefile。
