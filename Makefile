BINARY_NAME=sl-cli
INSTALL_PATH=/usr/local/bin

# 探测当前 Shell 类型 (zsh 或 bash)
# 如果探测失败，默认 fallback 到 zsh (Mac 默认)
SHELL_TYPE := $(shell basename $$SHELL)

.PHONY: all build clean install install-completion

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) cmd/sl-cli/main.go

clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)

# install 依赖 build，安装二进制文件后，尝试安装补全
install: build
	@echo "Installing binary to $(INSTALL_PATH)..."
	@# 使用 sudo 移动二进制文件，确保有权限写入 /usr/local/bin
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Binary installed."
	@$(MAKE) install-completion

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
