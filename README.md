# sl-cli (Super Link CLI)

**sl-cli** 是一个高度可扩展的现代命令行工具，旨在成为你日常工作流的“超级粘合剂”。

它采用 **Go 原生代码** + **YAML 动态配置** 的混合驱动模式。你既可以通过编写 Go 代码开发高性能的核心命令，也可以通过修改配置文件瞬间挂载 RESTful API、Shell 脚本或系统命令别名，而无需重新编译。

## ✨ 核心特性

* **⚡️ 混合驱动引擎**：支持原生 Go 命令扩展 + YAML 配置动态加载。
* **🌐 强大的 HTTP 执行器**：
  * 支持自定义 Headers（支持环境变量注入，如 `${TOKEN}`）。
  * 支持 Go Template 语法动态渲染 URL 和 Body。
  * **管道支持 (Pipe)**：支持将 API 响应直接传递给 `jq` 等工具处理。
  * 内置优雅的加载动画 (Spinner)。
* **🐚 Shell/Script 集成**：支持直接在配置中编写多行 Shell 脚本，支持交互式输入。
* **🖥 系统命令透传**：可以作为 `git`, `docker`, `kubectl` 等复杂命令的快捷别名管理器。
* **🍎 MacOS 深度适配**：
  * 自动生成并安装 **Zsh/Bash 自动补全**。
  * 支持彩色输出。
* **🛠 配置管理**：内置配置语法检查 (`config check`) 和初始化工具 (`config init`)。

## 🚀 安装指南

### 前置要求

* Go 1.24.0+
* Make
* (可选) `jq` - 用于测试 JSON 格式化管道功能

### 编译与安装

我们提供了完善的 `Makefile`，一键完成编译、安装二进制、补全脚本和文档。

```bash
# 1. 克隆项目
git clone https://github.com/huangchanghu/sl-cli.git
cd sl-cli

# 2. 安装 (需要 sudo 权限以写入 /usr/local/bin)
make install

# 3. 验证安装
sl-cli version
```

> **注意**：安装完成后，请重新打开终端或运行 `source ~/.zshrc` 以使自动补全生效。

---

## ⚙️ 配置指南

`sl-cli` 会按以下顺序查找配置文件：
1. `$HOME/.config/sl-cli/sl-cli.yaml`（默认位置）
2. 当前目录下的 `sl-cli.yaml`
3. `$HOME/.sl-cli.yaml`（旧版兼容）

### 快速上手

生成默认示例配置：

```bash
sl-cli config init
```

### 配置文件详解 (`sl-cli.yaml`)

```yaml
commands:
  # ------------------------------------------------------------------
  # 场景 1: 调用 RESTful API (带 Token 和 管道处理)
  # ------------------------------------------------------------------
  - name: "weather"
    usage: "查询天气并格式化 (用法: sl-cli weather London)"
    type: "http"
    api:
      # 使用 {{index .args 0}} 获取命令行第一个参数
      url: "https://goweather.herokuapp.com/weather/{{index .args 0}}"
      method: "GET"
      headers:
        Authorization: "Bearer ${MY_API_TOKEN}" # 读取环境变量
      # [高级特性] 将 API 结果作为 stdin 传给 jq 命令
      pipes:
        - command: "jq"
          args: ["."]

  # ------------------------------------------------------------------
  # 场景 2: 执行多行 Shell 脚本
  # ------------------------------------------------------------------
  - name: "deploy"
    usage: "执行模拟部署脚本"
    type: "shell"
    script: |
      echo "正在构建项目..."
      sleep 1
      echo "部署到环境: {{index .args 0}}"
      echo "完成!"

  # ------------------------------------------------------------------
  # 场景 3: 系统命令别名 (Alias)
  # ------------------------------------------------------------------
  - name: "gl"
    usage: "优雅的 Git Log"
    type: "system"
    command: "git"
    # 预设参数，用户输入的额外参数会追加在最后
    args: ["log", "--graph", "--oneline", "--decorate"]
```

### 检查配置

修改配置后，建议运行检查命令以确保语法正确：

```bash
sl-cli config check
```

---

## 💻 开发与扩展

### 目录结构

```text
sl-cli/
├── cmd/sl-cli/main.go      # 入口文件
├── sl-cli.yaml             # 配置文件
├── internal/
│   ├── executor/           # 执行器核心 (HTTP, Shell, System, Pipe)
│   └── config/             # 配置解析
├── pkg/cmd/                # 命令存放目录
│   ├── root.go             # 根命令 & 动态加载逻辑
│   ├── version.go          # [原生] 版本命令
│   └── config_ops.go       # [原生] 配置管理命令
└── Makefile                # 构建脚本
```

### 如何添加原生 Go 命令

如果你需要非常复杂的逻辑（如数据库操作、交互式 UI），建议添加原生命令：

1. 在 `pkg/cmd/` 下新建文件（例如 `my_cmd.go`）。
2. 定义 `cobra.Command`。
3. 在 `init()` 中调用 `rootCmd.AddCommand(yourCmd)`。

示例：

```go
package cmd
import "github.com/spf13/cobra"

var myCmd = &cobra.Command{
    Use: "hello",
    Run: func(cmd *cobra.Command, args []string) {
        println("Hello from Native Go!")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

4. 重新编译：`make install`

---

## 🗑 卸载

如果你想彻底移除 sl-cli（包括二进制、文档和补全脚本）：

```bash
make uninstall
```

## 📄 License

MIT License
