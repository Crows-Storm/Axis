// A generated module for Dagger functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/dagger/internal/dagger"
	"fmt"
	"time"
)

type Service struct {
	Source     *dagger.Directory
	SdkVersion string
	AppName    string
}

func NewService(
	source *dagger.Directory,
	sdkVersion string,
	appName string,
) *Service {
	return &Service{
		Source:     source,
		SdkVersion: sdkVersion,
		AppName:    appName,
	}
}

type Dagger struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Dagger) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Dagger) GrepDir(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directoryArg).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."}).
		Stdout(ctx)
}

// Build process

func (m *Service) Test(ctx context.Context) (string, error) {
	return dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", m.SdkVersion)).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedDirectory("/src", m.Source).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-v", "-race", "-coverprofile=coverage.out", "./..."}).
		Stdout(ctx)
}

// Build packs
func (m *Service) Build(
	ctx context.Context,
	// +default="my-go-service"
	imageName string,
	// +default="latest"
	tag string,
) *dagger.Container {

	// 使用 Pack CLI (CNCF Buildpacks) 构建
	// Builder: Google Buildpacks (原生支持 Go)
	packContainer := dag.Container().
		From("buildpacksio/pack:latest").
		WithMountedDirectory("/workspace", m.Source).
		WithWorkdir("/workspace").
		// 设置环境变量优化 Go 构建
		WithEnvVariable("BP_GO_TARGETS", ".").
		WithEnvVariable("CGO_ENABLED", "0").
		// 执行 pack build (零 Dockerfile)
		WithExec([]string{
			"pack", "build", fmt.Sprintf("%s:%s", imageName, tag),
			"--builder", "gcr.io/buildpacks/builder:google-22",
			"--path", ".",
			"--env", "GO_BUILD_FLAGS=-ldflags='-s -w'",
		})

	return packContainer.
		WithExec([]string{"docker", "save", "-o", "/output/image.tar", fmt.Sprintf("%s:%s", imageName, tag)}).
		WithFile("/output/image.tar", packContainer.File("/output/image.tar"))
}

// ========== 阶段 2 (替代方案): 纯 Go 多阶段构建 ==========
// 当 Buildpacks 不满足需求时使用
func (m *Service) BuildNative(
	ctx context.Context,
	// +default="my-go-service"
	imageName string,
	// +default="latest"
	tag string,
) *dagger.Container {

	// Stage 1: 编译
	builder := dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", m.SdkVersion)).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedDirectory("/src", m.Source).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOOS", "linux").
		WithExec([]string{
			"go", "build",
			"-ldflags", fmt.Sprintf("-s -w -X main.Version=%s", tag),
			"-o", "/output/app",
			".",
		})

	// Stage 2: 运行镜像 (distroless)
	return dag.Container().
		From("gcr.io/distroless/static-debian12:nonroot").
		WithLabel("org.opencontainers.image.title", m.AppName).
		WithLabel("org.opencontainers.image.version", tag).
		WithLabel("org.opencontainers.image.created", time.Now().Format(time.RFC3339)).
		WithFile("/app", builder.File("/output/app")).
		WithEntrypoint([]string{"/app"}).
		WithExposedPort(8080)
}

// ========== 阶段 3: Cosign 签名 & 推送 ==========
func (m *Service) Publish(
	ctx context.Context,
	registry string,
	// +default="latest"
	tag string,
	// Cosign 实验性功能标志
	// +default="true"
	cosignExperimental string,
) (string, error) {

	imageName := fmt.Sprintf("%s/%s", registry, m.AppName)

	// 构建镜像
	builtImage := m.BuildNative(ctx, imageName, tag)

	// 推送到 Registry
	addr, err := builtImage.
		WithRegistryAuth(registry, "admin", dag.SetSecret("registry-password", "")).
		Publish(ctx, fmt.Sprintf("%s:%s", imageName, tag))

	if err != nil {
		return "", fmt.Errorf("publish failed: %w", err)
	}

	// 使用 Cosign 进行 Keyless 签名
	// 在 CI 环境中 (GitHub Actions)，Cosign 会自动使用 OIDC Token
	_, err = dag.Container().
		From("gcr.io/projectsigstore/cosign:v2.4.1").
		WithEnvVariable("COSIGN_EXPERIMENTAL", cosignExperimental).
		WithExec([]string{
			"cosign", "sign",
			"--yes",       // 非交互模式
			"--recursive", // 签名所有 tag
			addr,
		}).
		Sync(ctx)

	if err != nil {
		return "", fmt.Errorf("cosign failed: %w", err)
	}

	// 生成 SLSA Provenance
	_, err = dag.Container().
		From("gcr.io/projectsigstore/cosign:v2.4.1").
		WithEnvVariable("COSIGN_EXPERIMENTAL", cosignExperimental).
		WithExec([]string{
			"cosign", "attest",
			"--yes",
			"--predicate", "/src/slsa-provenance.json",
			"--type", "slsaprovenance",
			addr,
		}).
		Sync(ctx)

	return addr, nil
}

// The CIPipeline Complete pipeline (concurrent execution)
func (m *Service) CIPipeline(
	ctx context.Context,
	registry string,
	tag string,
) (string, error) {

	// 并发执行: 测试 + Lint
	result, testErr := m.Test(ctx)
	if testErr != nil {
		return "", fmt.Errorf("tests failed:\n%s", testErr)
	}
	fmt.Printf("CIPPING RESULT: \n%s\n", result)

	addr, err := m.Publish(ctx, registry, tag, "")
	if err != nil {
		return "", err
	}

	fmt.Printf("✅ Tests passed\n📦 Published: %s\n🔏 Signed with Cosign\n", addr)
	return addr, nil
}
