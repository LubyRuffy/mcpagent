// Package models provides error definitions for database models.
package models

import "errors"

// LLM配置相关错误
var (
	ErrLLMConfigNameEmpty    = errors.New("LLM配置名称不能为空")
	ErrLLMConfigTypeEmpty    = errors.New("LLM类型不能为空")
	ErrLLMConfigTypeInvalid  = errors.New("LLM类型无效，仅支持 openai 和 ollama")
	ErrLLMConfigBaseURLEmpty = errors.New("LLM Base URL不能为空")
	ErrLLMConfigModelEmpty   = errors.New("LLM模型名称不能为空")
	ErrLLMConfigAPIKeyEmpty  = errors.New("LLM API Key不能为空")
	ErrLLMConfigNotFound     = errors.New("LLM配置不存在")
	ErrLLMConfigNameExists   = errors.New("LLM配置名称已存在")
)

// MCP服务器配置相关错误
var (
	ErrMCPServerConfigNameEmpty    = errors.New("MCP服务器配置名称不能为空")
	ErrMCPServerConfigCommandEmpty = errors.New("MCP服务器启动命令不能为空")
	ErrMCPServerConfigNotFound     = errors.New("MCP服务器配置不存在")
	ErrMCPServerConfigNameExists   = errors.New("MCP服务器配置名称已存在")
)

// MCP工具相关错误
var (
	ErrMCPToolNameEmpty     = errors.New("MCP工具名称不能为空")
	ErrMCPToolServerIDEmpty = errors.New("MCP工具服务器ID不能为空")
	ErrMCPToolKeyEmpty      = errors.New("MCP工具唯一标识不能为空")
	ErrMCPToolNotFound      = errors.New("MCP工具不存在")
	ErrMCPToolKeyExists     = errors.New("MCP工具唯一标识已存在")
)

// 系统提示词相关错误
var (
	ErrSystemPromptNameEmpty    = errors.New("系统提示词名称不能为空")
	ErrSystemPromptContentEmpty = errors.New("系统提示词内容不能为空")
	ErrSystemPromptNotFound     = errors.New("系统提示词不存在")
	ErrSystemPromptNameExists   = errors.New("系统提示词名称已存在")
)

// 全局配置相关错误
var (
	ErrAppConfigNameEmpty      = errors.New("全局配置名称不能为空")
	ErrAppConfigMaxStepInvalid = errors.New("最大步数必须大于0")
	ErrAppConfigNotFound       = errors.New("全局配置不存在")
	ErrAppConfigNameExists     = errors.New("全局配置名称已存在")
)
