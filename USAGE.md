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

## 🗑 卸载
```bash
make uninstall
```

## 🔐 许可证

MIT License