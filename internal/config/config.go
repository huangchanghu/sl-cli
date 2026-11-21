package config

import "gopkg.in/yaml.v3"

type Config struct {
	Imports  []string          `mapstructure:"imports"`
	Vars     map[string]string `mapstructure:"vars"` // Global variables
	Commands []CommandConfig   `mapstructure:"commands"`
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

	// 参数动态补全配置
	// 注意：loader 使用 yaml.v3 解析，只识别 yaml tag，mapstructure tag 不生效
	Completion *CompletionConfig `yaml:"completion"`
}

// CompletionConfig 定义参数补全脚本
// 支持简写 `completion: "<script>"` 和完整写法 `completion: {script: ..., timeout: ...}`
type CompletionConfig struct {
	Script       string `yaml:"script"`
	Timeout      string `yaml:"timeout"`       // 脚本超时，time.ParseDuration 格式，默认 3s
	Filter       string `yaml:"filter"`        // prefix(默认): 按当前输入前缀过滤; none: 原样返回
	FileFallback bool   `yaml:"file_fallback"` // 无候选时是否回退到文件补全
	NoSpace      bool   `yaml:"no_space"`      // 选中候选后不追加空格
}

// UnmarshalYAML 让 completion 字段既可以是脚本字符串，也可以是完整的对象
func (c *CompletionConfig) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		c.Script = node.Value
		return nil
	}
	type raw CompletionConfig // 避免递归调用 UnmarshalYAML
	return node.Decode((*raw)(c))
}

// APIConfig 定义 HTTP 请求细节
type APIConfig struct {
	URL         string            `mapstructure:"url"`
	Method      string            `mapstructure:"method"`
	Headers     map[string]string `mapstructure:"headers"` // 支持 Header
	QueryParams map[string]string `mapstructure:"query_params"`
	Body        string            `mapstructure:"body"`
	Pipes       []PipeConfig      `mapstructure:"pipes"`
}

// PipeConfig 定义后续处理命令
type PipeConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}
