# Apache Answer 项目常见问题修复

## 1. 多回答限制 (Multi-Answer Restriction)

### 问题描述
用户无法在同一帖子下创建多个回答。

### 原因
数据库中 `restrict_answer` 设置为 `true`，限制了每个用户对每个帖子只能有一条回复。

### 解决方案

**方法一：通过数据库直接修改**
```bash
sqlite3 data/answer.db "SELECT content FROM site_info WHERE type = 'questions';"
# 查看当前设置

sqlite3 data/answer.db "UPDATE site_info SET content = '{\"min_content\":6,\"min_tags\":1,\"restrict_answer\":false}' WHERE type = 'questions';"
# 将 restrict_answer 设为 false
```

**方法二：通过 Admin 后台修改**
- 进入 Admin 后台
- 导航到 General → Questions 设置
- 找到 "每个用户对于每个帖子只能有一条回复" 选项
- 关闭该选项

---

## 2. 插件翻译不完整 (Plugin Translations Not Fully Displayed)

### 问题描述
插件系统中的翻译显示不完整，插件名称和描述显示为 `plugin.xxx.backend.info.name` 这样的 key 而非翻译后的中文。

### 原因分析

插件翻译有两种来源：

1. **内置插件 (Builtin Plugins)**：通过 side-effect import 初始化
   - 位于 `ui/src/plugins/builtin/`
   - 在 `i18n/index.ts` 中调用 `initI18nResource()` 加载翻译
   - 正常工作

2. **外部 Go 插件 (External Plugins)**：通过 Docker 构建时合并
   - 位于 `plugin_list` 中的插件（如 `reviewer-glm`、`search-meilisearch` 等）
   - 在 Docker 构建时通过 `answer i18n` 命令将插件翻译合并到主 i18n 文件
   - **本地开发时不会自动合并**，因此翻译缺失

### 解决方案

**方案一：本地合并插件翻译**

⚠️ **常见错误**：只使用 `./answer i18n -t ./i18n` 不会合并插件翻译，因为缺少 `-s` 参数指定插件 i18n 源目录。

```bash
# 步骤 1：创建 vendor 目录（包含所有插件的 i18n 文件）
# 如果 vendor 目录已存在可跳过此步骤
go mod vendor

# 步骤 2：合并插件翻译到主 i18n 目录
# -s: source，指定包含插件 i18n 的目录（vendor）
# -t: target，指定要合并到的 i18n 目录
./answer i18n -s ./vendor -t ./i18n

# 步骤 3：重新构建前端（i18n yaml 会被打包进 JS bundle）
make ui

# 步骤 4：重新构建后端（i18n yaml 也会被 embed 进二进制）
make build
```

**参数说明**：
- `-s, --source`：包含插件 i18n 文件的目录，本地开发时为 `./vendor`
- `-t, --target`：主 i18n 目录，即 `./i18n`

**方案二：使用 Docker 构建（推荐用于生产）**
```bash
make docker
```
Dockerfile 中已配置插件翻译合并步骤。

### 调试发现

**问题根因**：运行 `./answer i18n -t ./i18n` 后插件翻译仍未合并。

**排查过程**：
1. 检查 `i18n/zh_CN.yaml` 发现 `plugin.` 开头的 key 不存在
2. 检查 `vendor/github.com/apache/answer-plugins/*/i18n/zh_CN.yaml` 插件 i18n 文件存在
3. 检查 `internal/cli/i18n.go` 发现 `MergeI18nFilesLocal(originalI18nDir, targetI18nDir)` 函数
4. 检查 `cmd/command.go` 发现调用：`MergeI18nFilesLocal(i18nTargetPath, i18nSourcePath)`
5. 关键发现：`-t` 传入的是 `i18nTargetPath`（目标），`-s` 传入的是 `i18nSourcePath`（源）
6. 只传 `-t ./i18n` 时，`i18nSourcePath` 为空字符串，导致 `findI18nFileInDir("")` 无法找到任何插件 i18n

**正确逻辑**：
- `MergeI18nFilesLocal(originalI18nDir, targetI18nDir)` 的参数顺序是：原始(插件)目录 → 目标目录
- 命令行参数：`-s` 是源（vendor），`-t` 是目标（i18n）
- 因此正确的合并命令必须是：`./answer i18n -s ./vendor -t ./i18n`

### 相关文件
- `script/build_plugin.sh` - Docker 构建脚本
- `internal/cli/i18n.go` - i18n 合并逻辑
- `cmd/command.go` - i18n 命令定义
- `ui/src/utils/pluginKit/utils.ts` - 前端 i18n 初始化
- `plugin_list` - 外部插件列表

---

## 3. 本地开发加载全部插件

### 问题描述
本地 `make build` 只加载 2 个插件，而 Docker 镜像加载 8 个插件。

### 原因
`cmd/answer/main.go` 中只导入了 2 个插件，其他插件需要通过 `plugin_list` 在 Docker 构建时嵌入。

### 解决方案
确保 `cmd/answer/main.go` 中导入了所有插件：

```go
import (
    answercmd "github.com/apache/answer/cmd"
    _ "github.com/apache/answer-plugins/captcha-basic"
    _ "github.com/apache/answer-plugins/connector-basic"
    _ "github.com/apache/answer-plugins/editor-formula"
    _ "github.com/apache/answer-plugins/quick-links"
    _ "github.com/apache/answer-plugins/render-markdown-codehighlight"
    _ "github.com/apache/answer-plugins/search-meilisearch"
    _ "github.com/apache/answer-plugins/storage-s3"
    _ "github.com/wjsoj/answer-plugins/reviewer-glm"
)
```

然后运行：
```bash
go get github.com/apache/answer-plugins/captcha-basic@latest
go get github.com/apache/answer-plugins/quick-links@latest
go get github.com/apache/answer-plugins/editor-formula@latest
go get github.com/apache/answer-plugins/render-markdown-codehighlight@latest
go get github.com/apache/answer-plugins/storage-s3@latest
go get github.com/apache/answer-plugins/search-meilisearch@latest

make build
```

### 已加载的插件验证
```bash
./answer plugin
```
应显示 8 个插件。

---

## 4. 翻译文件同步

### 问题描述
修改 i18n 文件后，前端 UI 没有更新。

### 原因
前端通过 webpack alias `@i18n` 直接读取 `../i18n/*.yaml`，但需要重新构建才能生效。

### 解决方案
```bash
# 重新构建前端和后端
make ui && make build
```

---

## 5. 数据库初始化问题

### 问题描述
数据库配置丢失或需要重置。

### 解决方案
```bash
# 删除旧数据重新初始化
rm -rf ./data/
./answer init -C ./data/
```

---

## 常用开发命令

```bash
# 构建
make build          # 后端
make ui             # 前端
make clean build    # 清理并重新构建

# 数据库
sqlite3 data/answer.db "SELECT * FROM site_info;"

# 插件
./answer plugin     # 列出已加载插件

# 前端开发
cd ui && pnpm start
```
