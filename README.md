1. 将补全脚本移动到 zsh 的补全目录 (通常是 /usr/local/share/zsh/site-functions 或 fpath 中的路径)

这里演示临时加载

source <(./sl-cli completion zsh)

现在尝试输入 ./sl-cli [TAB]

你应该能看到 myip, greet, ls 等命令的提示
