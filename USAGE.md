# sl-cli 使用文档

**sl-cli (Super Link CLI)** 是一个高度可扩展的现代命令行工具，旨在成为你日常工作流的"超级粘合剂"。它采用 Go 原生代码 + YAML 动态配置的混合驱动模式。

## 🚀 安装与使用

### 安装要求
- Go 1.24.0+
- Make 工具
- (可选) `jq` - 用于 JSON 格式化管道功能

### 编译与安装
```bash
# 克隆项目
git clone https://github.com/your-repo/sl-cli.git
cd sl-cli

# 编译并安装 (需要 sudo 权限)
make install

# 验证安装
sl-cli version
```

> **注意**: 安装完成后，请重新打开终端或运行 `source ~/.zshrc` 以使自动补全生效。

## 📋 基本命令

### 版本信息
```bash
sl-cli version
```

### 配置管理
```bash
# 初始化配置文件
sl-cli config init

# 检查配置文件语法和逻辑
sl-cli config check
```

### 生成文档
```bash
# 生成 Man Pages 文档
sl-cli gen-man [output-dir]
```

## 🛠️ 配置文件

`sl-cli` 会按以下顺序查找配置文件：
1. `$HOME/.config/sl-cli/sl-cli.yaml`（默认位置）
2. 当前目录下的 `sl-cli.yaml`
3. `$HOME/.sl-cli.yaml`（旧版兼容）

### 配置文件格式

```yaml
commands:
  # HTTP API 调用示例
  - name: "weather"
    usage: "查询天气 (使用方法: sl-cli weather London)"
    type: "http"
    api:
      url: "https://goweather.herokuapp.com/weather/{{index .args 0}}"
      method: "GET"
      headers:
        Authorization: "Bearer ${MY_API_TOKEN}"
      pipes:
        - command: "jq"
          args: ["."]

  # Shell 脚本执行示例
  - name: "deploy"
    usage: "执行部署脚本"
    type: "shell"
    script: |
      echo "正在构建项目..."
      sleep 1
      echo "部署到环境: {{index .args 0}}"
      echo "完成!"

  # 系统命令别名示例
  - name: "gl"
    usage: "优雅的 Git Log"
    type: "system"
    command: "git"
    args: ["log", "--graph", "--oneline", "--decorate"]
```

### 支持的命令类型

1. **HTTP**: 用于调用 RESTful API
2. **Shell**: 执行多行 Shell 脚本
3. **System**: 系统命令别名

## 🎯 功能特性

### 1. HTTP 执行器
- 支持自定义 Headers（支持环境变量注入）
- 支持 Go Template 语法动态渲染 URL 和 Body
- 管道支持 (Pipe)：支持将 API 响应直接传递给 `jq` 等工具处理
- 内置优雅的加载动画 (Spinner)

### 2. Shell/Script 集成
- 支持在配置中编写多行 Shell 脚本
- 支持交互式输入

### 3. 系统命令透传
- 可以作为 `git`, `docker`, `kubectl` 等复杂命令的快捷别名管理器

### 4. 自动补全支持
- 自动生成并安装 Zsh/Bash 自动补全脚本
- 支持通过 `completion` 字段为动态命令的参数配置补全脚本（见下文「参数动态补全」）

### 5. Man Pages 文档生成
- 自动生成并安装 Man Pages 文档

## 📖 示例配置详解

### HTTP API 调用
```yaml
- name: "myip"
  usage: "获取公网 IP 地址"
  type: "http"
  api:
    url: "https://httpbin.org/ip"
    method: "GET"
```

### 带管道的 HTTP API
```yaml
- name: "weather"
  usage: "获取城市天气信息"
  type: "http"
  api:
    url: "https://goweather.herokuapp.com/weather/{{index .args 0}}"
    method: "GET"
    pipes:
      - command: "jq"
        args: ["."]
```

### Shell 脚本
```yaml
- name: "greet"
  usage: "运行 Shell 脚本"
  type: "shell"
  script: |
    echo "--------------------------------"
    echo "Hello, {{index .args 0}}!"
    echo "当前目录: $(pwd)"
    echo "--------------------------------"
```

### 系统命令别名
```yaml
- name: "ll"
  usage: "列出文件详细信息 (别名 ls -laG)"
  type: "system"
  command: "ls"
  args: ["-l", "-a", "-G"]
```

### 参数动态补全
给命令增加 `completion` 字段，按 TAB 时执行脚本，stdout 每行一个候选：

```yaml
# 简写：直接给一段脚本
- name: "sw"
  type: "system"
  command: "git"
  args: ["switch"]
  completion: |
    git branch --format='%(refname:short)'

# 完整写法
- name: "deploy"
  type: "http"
  api:
    url: "{{.vars.api_base}}/deploy/{{index .args 0}}/{{index .args 1}}"
    method: "POST"
  completion:
    script: |
      case $# in
        0) printf 'prod\t生产环境\nstaging\t预发环境\n' ;;
        1) curl -s "{{.vars.api_base}}/services?env=$1" | jq -r '.[].name' ;;
      esac
    timeout: 5s          # 默认 3s，超时返回空候选
    filter: prefix       # prefix(默认): 按当前输入前缀过滤; none: 原样返回
    file_fallback: false # 无候选时是否回退到文件补全，默认 false
    no_space: false      # 选中后不追加空格，默认 false
```

脚本输入：
- `$#`：已输入的参数个数，即当前正在补全第几个参数（从 0 开始）
- `$1..$n` / `"$@"`：已输入的参数（`shell`/`system` 类型下包含 `-n 5` 这类 flag 原文）
- `$SL_CLI_COMP_CUR`：当前正在输入的词
- `$SL_CLI_COMP_CMD`：完整命令路径，如 `sl-cli deploy`
- 脚本同样经过模板渲染，可用 `{{.vars.KEY}}`；**参数请用 `$1` 而非 `{{index .args 0}}`**，补全时的半成品输入可能含引号等字符，直接拼进脚本会破坏语法

脚本输出：每行一个候选，可写成 `值<TAB>描述`（zsh/fish 会显示描述）；候选**按脚本输出的顺序展示**，shell 不会再按字母重排，需要字母序请在脚本末尾加 `| sort`（bash 需 4.4+，macOS 自带的 bash 3.2 仍会排序）；脚本退出码非 0 或超时时不显示候选，也不报错。

调试：

```bash
sl-cli __complete deploy prod ""                              # 最后一行 :N 是 directive
BASH_COMP_DEBUG_FILE=/tmp/sl-comp.log sl-cli __complete deploy ""  # 查看脚本报错/超时日志
```

> 已安装的补全脚本是运行时调用 `sl-cli __complete` 的，新增或修改 `completion` 配置后无需重装补全。
> 注意：按 TAB 也会执行脚本，`imports` 第三方配置前请确认其中的补全脚本可信。

## 🔧 开发扩展

### 添加原生 Go 命令
1. 在 `pkg/cmd/` 下新建文件（例如 `my_cmd.go`）
2. 定义 Cobra 命令
3. 在 `init()` 中调用 `rootCmd.AddCommand(yourCmd)`
4. 重新编译：`make install`

### 配置文件结构
- `name`: 命令名称（必须）
- `usage`: 命令使用说明
- `type`: 命令类型 (`http`, `shell`, `system`)
- `api`: HTTP 相关配置
- `script`: Shell 脚本内容
- `command`/`args`: 系统命令配置
- `completion`: 参数补全脚本（字符串简写，或包含 `script`/`timeout`/`filter`/`file_fallback`/`no_space` 的对象）

## 🗑 卸载
```bash
make uninstall
```

## 🔐 许可证

MIT License