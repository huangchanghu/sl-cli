BINARY_NAME=sl-cli
INSTALL_PATH=/usr/local/bin
MAN_PATH=/usr/local/share/man/man1
CONFIG_DIR=$(HOME)/.config/sl-cli

# 探测当前 Shell 类型 (zsh 或 bash)
# 如果探测失败，默认 fallback 到 zsh (Mac 默认)
SHELL_TYPE := $(shell basename $$SHELL)

.PHONY: all build clean install install-completion install-config

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) cmd/sl-cli/main.go

clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)

# install 依赖 build，安装二进制文件后，尝试安装补全
install: build install-config
	@echo "Installing binary to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Binary installed."
	@$(MAKE) install-completion

# 安装默认配置文件到 $HOME/.config/sl-cli
# sl-cli.yaml 是用户主配置，已存在则跳过；sl-cli.example.yaml 是参考样例，每次覆盖以同步最新示例
install-config:
	@echo "Provisioning config files to $(CONFIG_DIR)..."
	@mkdir -p $(CONFIG_DIR)
	@if [ -f "$(CONFIG_DIR)/sl-cli.yaml" ]; then \
		echo "  ⚠️  $(CONFIG_DIR)/sl-cli.yaml already exists, skipping."; \
	else \
		cp sl-cli.yaml $(CONFIG_DIR)/sl-cli.yaml; \
		echo "  ✅ Installed sl-cli.yaml"; \
	fi
	@# 先删后拷：兼容早期用 sudo 安装、属主为 root 的旧文件
	@rm -f $(CONFIG_DIR)/sl-cli.example.yaml
	@cp sl-cli.example.yaml $(CONFIG_DIR)/sl-cli.example.yaml
	@echo "  ✅ Installed sl-cli.example.yaml (reference file, always overwritten)"
	@if grep -qsE "^[[:space:]]*-[[:space:]]*[\"']?sl-cli\.example\.yaml" "$(CONFIG_DIR)/sl-cli.yaml"; then \
		echo "  ℹ️  sl-cli.yaml imports sl-cli.example.yaml, example commands are updated as well."; \
	fi

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
	
	@# 4. 清理旧版本安装的 Man pages (已不再生成，保留清理以免残留)
	@sudo rm -f $(MAN_PATH)/$(BINARY_NAME)*.1
	@echo "  - Removed legacy man pages"
	
	@echo "✅ Uninstallation complete."
