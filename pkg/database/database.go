// Package database provides database connection and management functionality.
// It handles database initialization, migrations, and connection management.
package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/LubyRuffy/mcpagent/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// InitDatabase initializes the database connection and performs migrations.
// It creates the database file if it doesn't exist and runs auto-migrations.
func InitDatabase(dbPath string) error {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 配置GORM日志级别
	logLevel := logger.Silent
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.Info
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置全局数据库实例
	DB = db

	// 执行自动迁移
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化默认数据
	if err := initDefaultData(); err != nil {
		return fmt.Errorf("初始化默认数据失败: %w", err)
	}

	// 执行数据迁移脚本
	if err := RunMigrations(); err != nil {
		return fmt.Errorf("执行数据迁移脚本失败: %w", err)
	}

	log.Printf("数据库初始化成功: %s", dbPath)
	return nil
}

// autoMigrate performs automatic database migrations
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.LLMConfigModel{},
		&models.MCPServerConfigModel{},
		&models.MCPToolModel{},
		&models.SystemPromptModel{},
		&models.AppConfigModel{},
	)
}

// initDefaultData initializes default data if the database is empty
func initDefaultData() error {
	// 检查是否已有LLM配置
	var llmCount int64
	if err := DB.Model(&models.LLMConfigModel{}).Count(&llmCount).Error; err != nil {
		return err
	}

	// 如果没有配置，创建默认配置
	if llmCount == 0 {
		defaultConfig := &models.LLMConfigModel{
			Name:        "默认Ollama配置",
			Description: "默认的Ollama本地配置",
			Type:        "ollama",
			BaseURL:     "http://127.0.0.1:11434",
			Model:       "qwen3:14b",
			APIKey:      "ollama",
			IsDefault:   true,
			IsActive:    true,
		}

		if err := DB.Create(defaultConfig).Error; err != nil {
			return fmt.Errorf("创建默认LLM配置失败: %w", err)
		}

		log.Println("已创建默认LLM配置")
	}

	// 检查是否已有MCP服务器配置
	var mcpCount int64
	if err := DB.Model(&models.MCPServerConfigModel{}).Count(&mcpCount).Error; err != nil {
		return err
	}

	// 如果没有MCP服务器配置，创建默认配置
	if mcpCount == 0 {
		defaultMCPServers := []models.MCPServerConfigModel{
			{
				Name:        "ddg-search",
				Description: "DuckDuckGo搜索服务器",
				Command:     "uvx",
				IsActive:    true,
			},
			{
				Name:        "sequential-thinking",
				Description: "顺序思考服务器",
				Command:     "npx",
				IsActive:    true,
			},
			{
				Name:        "fetch",
				Description: "网页抓取服务器",
				Command:     "uvx",
				IsActive:    true,
			},
		}

		// 设置参数
		defaultMCPServers[0].SetArgs([]string{"duckduckgo-mcp-server"})
		defaultMCPServers[1].SetArgs([]string{"-y", "@modelcontextprotocol/server-sequential-thinking"})
		defaultMCPServers[2].SetArgs([]string{"mcp-server-fetch"})

		for _, server := range defaultMCPServers {
			if err := DB.Create(&server).Error; err != nil {
				return fmt.Errorf("创建默认MCP服务器配置失败: %w", err)
			}
		}

		log.Println("已创建默认MCP服务器配置")
	}

	// 检查是否已有全局配置
	var appConfigCount int64
	if err := DB.Model(&models.AppConfigModel{}).Count(&appConfigCount).Error; err != nil {
		return err
	}

	// 如果没有全局配置，创建默认全局配置
	if appConfigCount == 0 {
		defaultAppConfig := &models.AppConfigModel{
			Name:         "默认全局配置",
			Description:  "默认的MCP Agent全局配置",
			Proxy:        "",
			SystemPrompt: "你是精通互联网的信息收集专家，需要帮助用户进行信息收集，当前时间是：{date}。",
			MaxStep:      20,
			IsDefault:    true,
			IsActive:     true,
		}

		// 设置默认占位符
		if err := defaultAppConfig.SetPlaceHoldersFromMap(map[string]interface{}{}); err != nil {
			return fmt.Errorf("设置默认全局配置占位符失败: %w", err)
		}

		// 设置默认MCP配置
		defaultMCPConfig := &models.MCPConfig{
			ConfigFile: "",
			Tools:      []string{},
		}
		if err := defaultAppConfig.SetMCPConfig(defaultMCPConfig); err != nil {
			return fmt.Errorf("设置默认MCP配置失败: %w", err)
		}

		if err := DB.Create(defaultAppConfig).Error; err != nil {
			return fmt.Errorf("创建默认全局配置失败: %w", err)
		}

		log.Println("已创建默认全局配置")
	}

	// 检查是否已有系统提示词配置
	var promptCount int64
	if err := DB.Model(&models.SystemPromptModel{}).Count(&promptCount).Error; err != nil {
		return err
	}

	// 如果没有系统提示词配置，创建默认配置
	if promptCount == 0 {
		defaultSystemPrompts := []models.SystemPromptModel{
			{
				Name:        "学术研究写作系统提示词",
				Description: "学术研究写作系统提示词，适用于需要多步思考和工具调用的场景。",
				Content: `# 学术研究写作系统提示词

## 角色与专业背景
你是一位{field}资深专家，专精于国际顶级期刊的论文撰写标准。具备深厚的学术研究方法论基础，能够撰写符合同行评议要求的高质量学术文献。当前日期：{date}

## 语言风格规范
### 必须遵循
- **正式性**：采用学术正式语体，避免口语化和俚语表达
- **客观性**：保持中立立场，基于证据进行论述
- **精确性**：使用准确的专业术语，避免模糊表达
- **严谨性**：逻辑严密，论证充分

### 具体要求
- 完整形式：使用完整词汇而非缩写（如"do not"而非"don't"）
- 避免绝对化：除非有充分证据支撑，避免"完全"、"绝对"等绝对化表述
- 情感中性：避免情感化词汇，保持分析性和批判性语调

## 结构与组织规范
### 整体架构
- **文献综述**：按主题分类组织，每个主题下系统讨论关键研究成果，明确指出现有研究空白
- **段落构成**：每段包含清晰主题句、充分论证和逻辑连贯的结论
- **引用规范**：关键理论或模型需标注来源，采用(Author, Year)格式或自然融入式引用

### 内容深度要求
- 系统分析现有研究的方法论局限性
- 基于文献gap提出未来研究方向建议
- 避免未经实证支持的主观判断
- 确保论述的理论基础和实证依据

## 目标读者定位
{field}学者、政策制定者及相关从业者，具备该领域基础知识背景

## 研究过程要求
### 信息收集与分析
- **阶段性思考**：必须通过'sequentialthinking'工具进行结构化思考，总计思考次数可达{total_thoughts}次
- **信息搜索**：充分利用'search'工具获取相关学术资源，避免重复相同查询
- **深度获取**：针对有价值的搜索结果，使用'fetch'工具获取完整内容进行深入分析
- **持续推进**：在nextThoughtNeeded参数为true时，必须继续思考过程，不得提前终止

### 工具使用规范
- 确保所有工具调用参数格式正确（JSON格式）
- 注意'sequentialthinking'工具参数命名规范：使用驼峰式命名（如nextThoughtNeeded）
- 思考过程采用中文输出，便于理解和跟踪

## 输出标准
- **字数要求**：不少于2000字
- **直接输出**：无需开场白或额外解释，直接提供所需内容
- **质量标准**：符合国际顶级期刊的学术写作标准

## 质量检核要点
1. 论点是否有充分的文献支撑
2. 逻辑结构是否清晰连贯
3. 语言表达是否符合学术规范
4. 是否准确识别和阐述了研究空白
5. 未来研究建议是否具有可操作性和创新性`,
				IsDefault: true,
				IsActive:  true,
			},
			{
				Name:        "网络安全专家",
				Description: "专注于网络安全领域的系统提示词",
				Content:     "你是一位资深的网络安全专家，拥有丰富的{field}经验。请以专业的方式回答用户关于网络安全的问题，使用精确的术语，并在必要时说明潜在风险。今天是{date}。",
				IsActive:    true,
			},
		}

		// 设置占位符
		if err := defaultSystemPrompts[0].SetPlaceholdersFromStringSlice([]string{}); err != nil {
			return fmt.Errorf("设置默认系统提示词占位符失败: %w", err)
		}

		if err := defaultSystemPrompts[1].SetPlaceholdersFromStringSlice([]string{"field", "date"}); err != nil {
			return fmt.Errorf("设置默认系统提示词占位符失败: %w", err)
		}

		for _, prompt := range defaultSystemPrompts {
			if err := DB.Create(&prompt).Error; err != nil {
				return fmt.Errorf("创建默认系统提示词配置失败: %w", err)
			}
		}

		log.Println("已创建默认系统提示词配置")
	}

	return nil
}

// GetDB returns the global database instance
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
