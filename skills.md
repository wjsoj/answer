# Apache Answer 项目构建和测试命令

## 后端 (Go)

### 构建
```bash
# 构建二进制文件
go build -o answer ./cmd/answer

# 使用 Makefile 构建
make build

# 清理并构建
make all

# 生成代码（swagger, wire, mockgen）
make generate

# 检查代码格式和 lint
make check
make lint
```

### 启动
```bash
# 初始化数据库
INSTALL_PORT=8080 ./answer init -C ./data/

# 运行服务器
./answer run -C ./data/
```

### 测试
```bash
# 运行测试
go test ./...

# 运行特定包的测试
go test ./internal/repo/repo_test

# 使用 Makefile 运行测试
make test
```

## 前端 (React UI)

### 构建
```bash
cd ui

# 安装依赖（使用 pnpm）
pnpm install

# 开发模式启动
pnpm start

# 生产构建
pnpm build

# TypeScript 类型检查
pnpm tsc --noEmit

# ESLint 检查
pnpm lint

# 代码格式化
pnpm prettier
```

### 前端依赖安装问题解决
如果遇到 `pnpm install` 失败，可以尝试：
```bash
# 启用 corepack
corepack enable

# 准备 pnpm
corepack prepare pnpm@9.7.0 --activate

# 清理后重新安装
rm -rf node_modules
pnpm install
```

## Docker 构建

```bash
# 构建 Docker 镜像（当前架构）
make docker

# 构建多架构镜像（amd64 + arm64）并推送
make docker-multiarch
```

## 快速验证

### 验证后端编译
```bash
go build -o /dev/null ./cmd/answer && echo "Backend builds OK"
```

### 验证前端 TypeScript
```bash
cd ui && pnpm tsc --noEmit && echo "Frontend TypeScript OK"
```

## 常见问题

### Go 依赖问题
```bash
go mod tidy
go mod download
```

### 前端依赖问题
```bash
cd ui
rm -rf node_modules
pnpm store prune
pnpm install
```

### 数据库初始化问题
```bash
# 删除旧数据重新初始化
rm -rf ./data/
./answer init -C ./data/
```
