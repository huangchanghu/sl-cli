# Project Development Prompts

This document records the chronological prompts used to generate the `sl-cli` project using an LLM. It serves as a development log and a guide for reproducing or extending the project.

## Phase 1: Inception & Architecture

**Prompt 1: Initial Requirements**
> 用go语言开发一个命令行程序，命名为sl-cli。
> 调研go进行命令行程序的流行库、规范等。
> 这个命令行程序需要有非常好的扩展性，包括但不限于：
>
> 1. 支持子命令
> 2. 方便继续用go语言扩展开发新命令
> 3. 支持通过配置restfull api进行功能扩展
> 4. 支持配置外部shell、python等脚本进行功能扩展
> 5. 支持配置调用系统中已安装的命令行工具进行功能扩展
> 6. 需要和mac osx系统有良好的适配
>
> 基于以上描述，请生成一个todo计划，不要着急生成代码，每一个步骤执行前需等待我确认。

**Prompt 2: Architecture Refinement**
> "http"执行器类型需要支持配置header，用于支持Bearer Token认证。添加header支持后，可以进行Step 2

## Phase 2: Core Implementation

**Prompt 3: Skeleton & Dynamic Engine**
> (After Step 2 confirmation)
> Step 3运行时未输出预期的Mock信息，调试发现initConfig()方法并未执行

**Prompt 4: Executor Implementation (HTTP/Shell/System)**
> (Proceed to Step 4: Implement the actual execution logic for HTTP, Shell, and System commands)

**Prompt 5: Bug Fix - Flag Parsing**
> 执行`./sl-cli ls -la`时报错：
> `Error: unknown shorthand flag: 'l' in -la`
> 检查原因并修复，确保系统命令能透传参数。

## Phase 3: UX & Mac Integration

**Prompt 6: Native Go Extensions**
> (Proceed to Step 5: Implement Native Go extension mechanism with `version` and `calc` examples)

**Prompt 7: UX Polish (Spinner & Completion)**
> (Proceed to Step 6: MacOS adaptation, adding Spinner for loading states and verifying Zsh auto-completion)

**Prompt 8: Makefile & Auto-completion Install**
> 在make install中为命令添加当前shell环境的自动补全

**Prompt 9: Bug Fix - Shell Detection**
> 系统当前shell环境为zsh， 运行`echo $SHELL`时输出:/bin/zsh
> 但是make install中检测到的shell环境是sh:`Detecting shell: sh`
> 检查是什么原因，并修正

**Prompt 10: Man Pages Support**
> 添加生成`man pages`的支持

## Phase 4: Advanced Features

**Prompt 11: Pipe Support (jq integration)**
> restfull api类型命令增强：
>
> 1. 支持配置管道命令，将api输出结果传递给管道命令（系统命令调用），并将管道命令的输出作为最终输出；
> 2. 注意只有restfull api请求成功的情况下才继续执行管道命令
> 3. 管道命令以jq进行举例，将restfull api的结果用jq进行json格式化输出

**Prompt 12: Config Management**
> 增加一个命令用于修改配置后检测是否有语法错误。
> 并生成一个`sl-cli.yaml`示例文件用于安装分发

## Phase 5: Distribution

**Prompt 13: Uninstallation**
> Makefile中添加卸载指令

**Prompt 14: Documentation**
> 生成README.md文件
