BINARY_NAME=sl-cli
INSTALL_PATH=/usr/local/bin
MAN_PATH=/usr/local/share/man/man1

# 探测当前 Shell 类型 (zsh 或 bash)
# 如果探测失败，默认 fallback 到 zsh (Mac 默认)
SHELL_TYPE := $(shell basename $$SHELL)

.PHONY: all build clean install install-completion install-man gen-man

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) cmd/sl-cli/main.go

clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf ./man1

# [新增] 生成 Man pages 的目标
gen-man: build
	@echo "Generating man pages..."
	@mkdir -p man1
	@./$(BINARY_NAME) gen-man ./man1

# install 依赖 build，安装二进制文件后，尝试安装补全
install: build install-man
	@echo "Installing binary to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Binary installed."
	@$(MAKE) install-completion

# 安装 Man pages
install-man: gen-man
	@echo "Installing man pages to $(MAN_PATH)..."
	@sudo mkdir -p $(MAN_PATH)
	@# 安装所有生成的 .1 文件
	@sudo cp man1/*.1 $(MAN_PATH)/
	@echo "✅ Man pages installed."

install-completion:
	@echo "Detecting shell: $(SHELL_TYPE)"
ifeq ($(SHELL_TYPE),zsh)
	@echo "Installing Zsh completion..."
	@# 创建标准的 Zsh site-functions 目录 (如果不存在)
	@sudo mkdir -p /usr/local/share/zsh/site-functions
	@# 生成补全脚本并写入文件，文件名为 _sl-cli (Zsh 规范)
	@$(INSTALL_PATH)/$(BINARY_NAME) completion zsh | sudo tee /usr/local/share/zsh/site-functions/_$(BINARY_NAME) > /dev/null
	@echo "✅ Zsh completion installed to /usr/local/share/zsh/site-functions/_$(BINARY_NAME)"
	@echo "👉 You may need to run 'rm -f ~/.zcompdump; compinit' to reload."
else ifeq ($(SHELL_TYPE),bash)
	@echo "Installing Bash completion..."
	@# 创建 Bash 补全目录 (兼容 Homebrew 和 Linux)
	@sudo mkdir -p /usr/local/etc/bash_completion.d
	@$(INSTALL_PATH)/$(BINARY_NAME) completion bash | sudo tee /usr/local/etc/bash_completion.d/$(BINARY_NAME) > /dev/null
	@echo "✅ Bash completion installed to /usr/local/etc/bash_completion.d/$(BINARY_NAME)"
	@echo "👉 Ensure you have bash-completion installed and sourced."
else
	@echo "⚠️  Shell '$(SHELL_TYPE)' not fully supported for auto-install."
	@echo "Please run '$(BINARY_NAME) completion --help' to install manually."
endif

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@# 1. 删除二进制文件
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "  - Removed binary"
	
	@# 2. 删除 Zsh 补全脚本
	@sudo rm -f /usr/local/share/zsh/site-functions/_$(BINARY_NAME)
	@echo "  - Removed Zsh completion"
	
	@# 3. 删除 Bash 补全脚本
	@sudo rm -f /usr/local/etc/bash_completion.d/$(BINARY_NAME)
	@echo "  - Removed Bash completion"
	
	@# 4. 删除 Man pages (删除 sl-cli.1 以及 sl-cli-xxx.1 所有子命令文档)
	@sudo rm -f $(MAN_PATH)/$(BINARY_NAME)*.1
	@echo "  - Removed Man pages"
	
	@echo "✅ Uninstallation complete."
