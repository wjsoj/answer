# Deployment Guide

## Architecture Overview

```
Makefile          →  构建镜像、推送到 Registry
docker-compose.yml  →  部署运行（只拉取镜像，不构建）
```

**原则：构建和部署完全分离。** `docker-compose.yml` 中没有 `build` 配置，它只负责从 Registry 拉取镜像并运行。

---

## 1. Version Management

版本号自动从 git tag 派生，无需手动维护：

```bash
git tag v2.1.0       # 打 tag = 定义版本号
make version         # 查看当前版本信息
```

| 场景 | VERSION | VERSION_FULL |
|------|---------|-------------|
| HEAD 在 `v2.1.0` tag 上 | `2.1.0` | `2.1.0` |
| tag 后有 3 个新 commit | `2.1.0` | `2.1.0-3-g1a2b3c4` |
| 无 tag | `0.0.0-dev` | `commit hash` |

---

## 2. Build (构建镜像)

### 2.1 本地开发构建（当前架构）

```bash
make docker
```

- 仅构建当前机器架构的镜像（如 amd64）
- 镜像保存在本地，不推送
- 自动打 tag：`<REPO>:<VERSION_FULL>` 和 `<REPO>:latest`

构建后本地测试：

```bash
docker compose up -d
```

### 2.2 多架构构建 + 推送（发布）

```bash
make docker-multiarch
```

- 同时构建 `linux/amd64` + `linux/arm64`
- 自动推送到 Registry
- 自动打 3 个 tag：`<VERSION>`、`<VERSION_FULL>`、`latest`

**前置条件**（仅首次需要）：

```bash
# 安装 buildx 插件（Arch Linux）
sudo pacman -S docker-buildx

# 创建多架构 builder
docker buildx create --name multiarch --use

# 安装 QEMU 模拟器（x86 上交叉编译 ARM）
docker run --privileged --rm tonistiigi/binfmt --install all
```

---

## 3. Release (发版流程)

```bash
# 1. 确保代码已提交
git add -A && git commit -m "feat: some feature"

# 2. 打版本 tag
git tag v2.1.0

# 3. 推送代码和 tag
git push && git push --tags

# 4. 构建多架构镜像并推送到 Registry
make docker-multiarch
```

完成后 Registry 中有：
- `answer:2.1.0` — 版本号
- `answer:latest` — 最新版

---

## 4. Deploy (部署)

### 4.1 部署配置

在目标服务器上创建 `docker-compose.yml`：

```yaml
services:
  answer:
    image: git.pku.edu.cn/2200011523/answer:${IMAGE_TAG:-latest}
    container_name: answer
    restart: unless-stopped
    ports:
      - "9080:80"
    volumes:
      - ./data:/data
    environment:
      - TZ=Asia/Shanghai
      - AUTO_INSTALL=true
      - DB_TYPE=sqlite3
      - DB_FILE=/data/answer.db
      - LANGUAGE=zh_CN
      - SITE_NAME=My Forum
      - SITE_URL=https://forum.example.com
      - CONTACT_EMAIL=admin@example.com
      - ADMIN_NAME=admin
      - ADMIN_PASSWORD=changeme123456
      - ADMIN_EMAIL=admin@example.com
```

### 4.2 首次部署

```bash
docker compose up -d
```

Docker 会自动根据机器架构（amd64/arm64）拉取对应镜像。

### 4.3 指定版本部署

```bash
# 部署特定版本
IMAGE_TAG=2.1.0 docker compose up -d

# 部署最新版
docker compose up -d
```

### 4.4 升级

```bash
# 拉取最新镜像
docker compose pull

# 重启容器（数据在 ./data 中持久化，不会丢失）
docker compose up -d
```

### 4.5 回滚

```bash
IMAGE_TAG=2.0.0 docker compose up -d
```

---

## 5. Environment Variables

| 变量 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `AUTO_INSTALL` | Yes | 启用自动安装 | `true` |
| `DB_TYPE` | Yes | 数据库类型 | `sqlite3` / `mysql` / `postgres` |
| `DB_FILE` | sqlite3 | SQLite 文件路径 | `/data/answer.db` |
| `DB_HOST` | mysql/pg | 数据库地址 | `db:3306` |
| `DB_USERNAME` | mysql/pg | 数据库用户名 | `root` |
| `DB_PASSWORD` | mysql/pg | 数据库密码 | `password` |
| `DB_NAME` | mysql/pg | 数据库名 | `answer` |
| `LANGUAGE` | Yes | 界面语言 | `zh_CN` / `en_US` |
| `SITE_NAME` | Yes | 站点名称 | `My Forum` |
| `SITE_URL` | Yes | 站点 URL | `https://forum.example.com` |
| `CONTACT_EMAIL` | Yes | 联系邮箱 | `admin@example.com` |
| `ADMIN_NAME` | Yes | 管理员用户名 | `admin` |
| `ADMIN_PASSWORD` | Yes | 管理员密码（>=8位） | `changeme123456` |
| `ADMIN_EMAIL` | Yes | 管理员邮箱 | `admin@example.com` |

---

## Quick Reference

```bash
make version            # 查看版本
make docker             # 本地构建
make docker-multiarch   # 多架构构建+推送
docker compose up -d    # 部署/启动
docker compose pull     # 拉取最新镜像
docker compose logs -f  # 查看日志
```
