package config

// Config 对应整个配置文件的根结构
type Config struct {
	Commands []CommandConfig `mapstructure:"commands"`
}

// CommandConfig 定义单个命令的配置
type CommandConfig struct {
	Name        string          `mapstructure:"name"`
	Usage       string          `mapstructure:"usage"`
	Type        string          `mapstructure:"type"` // http, shell, system
	SubCommands []CommandConfig `mapstructure:"subcommands"`

	// HTTP 相关配置
	API APIConfig `mapstructure:"api"`

	// Shell/Script 相关配置
	Script string `mapstructure:"script"`

	// System Command 相关配置
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

// APIConfig 定义 HTTP 请求细节
type APIConfig struct {
	URL         string            `mapstructure:"url"`
	Method      string            `mapstructure:"method"`
	Headers     map[string]string `mapstructure:"headers"` // 支持 Header
	QueryParams map[string]string `mapstructure:"query_params"`
	Body        string            `mapstructure:"body"`
	Pipe        PipeConfig        `mapstructure:"pipe"`
}

// PipeConfig 定义后续处理命令
type PipeConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}
