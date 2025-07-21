import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AppConfig, ConfigState, SystemPromptTemplate, MCPServer, MCPToolConfig, LLMConfigModel, CreateLLMConfigForm, MCPServerConfigModel, CreateMCPServerConfigForm, SystemPromptModel, CreateSystemPromptForm } from '@/types/config'
import { validateConfig } from '@/utils/storage'
import { llmApi, mcpApi, configApi, systemPromptApi } from '@/utils/api'
import type { ApiResponse } from '@/utils/api'

export const useConfigStore = defineStore('config', () => {
  // 状态
  const config = ref<AppConfig>({
    proxy: '',
    mcp: {
      mcp_servers: {},
      tools: []
    },
    llm: {
      type: 'ollama',
      base_url: 'http://127.0.0.1:11434',
      model: 'qwen3:14b',
      api_key: 'ollama'
    },
    system_prompt: '你是精通互联网的信息收集专家，需要帮助用户进行信息收集，当前时间是：{date}。',
    max_step: 20,
    placeholders: {}
  })

  const availableModels = ref<string[]>([])
  const availableTools = ref<Array<{ name: string; description: string; server: string }>>([])
  const llmConfigs = ref<LLMConfigModel[]>([])
  const currentLLMConfigId = ref<number | null>(null)
  const mcpServerConfigs = ref<MCPServerConfigModel[]>([])
  const systemPrompts = ref<SystemPromptModel[]>([])
  const currentSystemPromptId = ref<number | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  // 配置是否已从后端加载完成
  const configLoaded = ref(false)

  // 计算属性
  const mcpServers = computed(() => {
    return config.value.mcp.mcp_servers || {}
  })

  const connectedServers = computed(() => {
    return Object.entries(mcpServers.value).filter(
      ([_, server]) => server.status === 'connected'
    ).length
  })

  const selectedTools = computed(() => {
    return config.value.mcp.tools || []
  })

  const currentLLMConfig = computed(() => {
    if (!currentLLMConfigId.value) return null
    return llmConfigs.value.find(config => config.id === currentLLMConfigId.value) || null
  })

  const defaultLLMConfig = computed(() => {
    return llmConfigs.value.find(config => config.is_default) || null
  })

  const currentSystemPrompt = computed(() => {
    if (!currentSystemPromptId.value) return null
    return systemPrompts.value.find(prompt => prompt.id === currentSystemPromptId.value) || null
  })

  const defaultSystemPrompt = computed(() => {
    return systemPrompts.value.find(prompt => prompt.is_default) || null
  })

  // 方法
  const loadConfiguration = async (force = false) => {
    try {
      console.log('【配置】loadConfiguration开始执行，强制更新:', force, new Date().toISOString())
      isLoading.value = true
      error.value = null

      // 从后端加载配置
      try {
        const backendConfig = await configApi.getConfig()
        if (backendConfig) {
          config.value = { ...config.value, ...backendConfig }
          console.log('【配置】从后端加载配置完成', new Date().toISOString())
        }
      } catch (err) {
        console.error('【配置】从后端加载配置失败，使用默认配置:', err)
      }

      // 加载SystemPrompt配置
      console.log('【配置】开始加载SystemPrompt配置', new Date().toISOString())
      try {
        await loadSystemPrompts(force)
        console.log('【配置】SystemPrompt配置加载完成', new Date().toISOString())
      } catch (err) {
        console.error('【配置】SystemPrompt配置加载失败:', err)
      }

      // 加载LLM配置
      console.log('【配置】开始加载LLM配置', new Date().toISOString())
      try {
        await loadLLMConfigs(force)
        console.log('【配置】LLM配置加载完成', new Date().toISOString())
      } catch (err) {
        console.error('【配置】LLM配置加载失败:', err)
      }

      // 加载MCP服务器配置
      console.log('【配置】开始加载MCP服务器配置', new Date().toISOString())
      try {
        await loadMCPServerConfigs(force)
        console.log('【配置】MCP服务器配置加载完成', new Date().toISOString())
      } catch (err) {
        console.error('【配置】MCP服务器配置加载失败:', err)
      }

      // 加载工具列表 - 只更新可用工具，不影响用户选择
      console.log('【配置】开始加载工具列表', new Date().toISOString())
      try {
        await loadToolsFromDatabase()
        console.log('【配置】工具列表加载完成', new Date().toISOString())
      } catch (err) {
        console.error('【配置】工具列表加载失败:', err)
      }

      // 构建MCP服务器配置，但保留用户已选择的工具
      console.log('【配置】开始构建MCP配置', new Date().toISOString())
      try {
        buildMCPConfigFromDatabase()
        console.log('【配置】MCP配置构建完成', new Date().toISOString())
      } catch (err) {
        console.error('【配置】MCP配置构建失败:', err)
        // 应急处理：创建默认服务器
        config.value.mcp.mcp_servers = {
          'ddg': {
            command: 'ddg',
            args: [],
            env: {},
            status: 'disconnected'
          }
        };
      }

      // 检查工具配置格式是否正确
      // 如果是旧版本的字符串数组格式，需要转换为新的MCPToolConfig格式
      if (config.value.mcp.tools && config.value.mcp.tools.length > 0) {
        console.log('【配置】检查工具配置格式', config.value.mcp.tools)
        
        // 检查是否是旧的字符串数组格式
        if (typeof config.value.mcp.tools[0] === 'string') {
          console.log('【配置】发现旧的字符串数组格式，转换为MCPToolConfig格式')
          
          // 将字符串数组转换为MCPToolConfig数组
          const toolConfigs = (config.value.mcp.tools as unknown as string[]).map(toolName => {
            // 尝试从availableTools中查找对应的服务器信息
            const tool = availableTools.value.find(t => t.name === toolName);
            return {
              server: tool?.server || 'inner',
              name: toolName
            };
          });
          
          // 更新配置
          config.value.mcp.tools = toolConfigs;
          console.log('【配置】转换后的工具配置:', config.value.mcp.tools)
        }
      }

      console.log('【配置】loadConfiguration执行完成，当前选择的工具:', config.value.mcp.tools, new Date().toISOString())
      
      // 标记配置加载完成
      configLoaded.value = true

    } catch (err) {
      console.error('【配置】加载配置失败:', err)
      error.value = err instanceof Error ? err.message : '加载配置失败'
    } finally {
      isLoading.value = false
    }
  }

  const saveConfiguration = async () => {
    try {
      isLoading.value = true
      error.value = null

      // 验证配置
      const validation = validateConfig(config.value)
      if (!validation.valid) {
        throw new Error(validation.errors.join(', '))
      }

      // 创建配置副本用于保存
      const configToSave = JSON.parse(JSON.stringify(config.value))
      
      // 无需进行格式转换，现在直接使用MCPToolConfig格式

      // 将配置保存到后端
      const response = await configApi.updateConfig(configToSave)
      
      // 显示成功消息
      if (response.success) {
        // 可以通过UI组件或其他方式通知用户
        console.log('配置已成功保存到数据库，后端重启时将自动加载')
        return response.message || '配置保存成功，后端重启时将自动加载'
      }

    } catch (err) {
      error.value = err instanceof Error ? err.message : '保存配置失败'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const updateLLMConfig = (llmConfig: Partial<AppConfig['llm']>) => {
    config.value.llm = { ...config.value.llm, ...llmConfig }
  }

  const updateMCPConfig = (mcpConfig: Partial<AppConfig['mcp']>) => {
    config.value.mcp = { ...config.value.mcp, ...mcpConfig }
  }

  const addMCPServer = (name: string, server: MCPServer) => {
    if (!config.value.mcp.mcp_servers) {
      config.value.mcp.mcp_servers = {}
    }
    config.value.mcp.mcp_servers[name] = server
  }

  const removeMCPServer = (name: string) => {
    if (config.value.mcp.mcp_servers) {
      delete config.value.mcp.mcp_servers[name]
    }
  }

  const updateMCPServerStatus = (name: string, status: MCPServer['status']) => {
    if (config.value.mcp.mcp_servers && config.value.mcp.mcp_servers[name]) {
      config.value.mcp.mcp_servers[name].status = status
    }
  }

  const updateSystemPrompt = (prompt: string) => {
    config.value.system_prompt = prompt
  }

  const updatePlaceholders = (placeholders: Record<string, any>) => {
    config.value.placeholders = { ...config.value.placeholders, ...placeholders }
  }

  const selectTools = (toolNames: string[]) => {
    // 将工具名称列表转换为 MCPToolConfig 对象数组
    config.value.mcp.tools = getToolConfigsFromNames(toolNames);
  }

  const updateAvailableTools = (tools: Array<{ name: string; description: string; server: string }>) => {
    availableTools.value = tools
  }

  // 将工具名称数组转换为MCPToolConfig对象数组
  const getToolConfigsFromNames = (toolNames: string[]): MCPToolConfig[] => {
    if (!toolNames || toolNames.length === 0) {
      return [];
    }
    
    return toolNames.map(toolName => {
      const tool = availableTools.value.find(t => t.name === toolName);
      return {
        server: tool?.server || 'inner',
        name: toolName
      };
    });
  }

  // 从MCPToolConfig对象数组中提取工具名称
  const getToolNamesFromConfigs = (toolConfigs: MCPToolConfig[]): string[] => {
    if (!toolConfigs || toolConfigs.length === 0) {
      return [];
    }
    
    return toolConfigs.map(tool => tool.name);
  }

  // 获取用于后端任务执行的配置（工具名称转换为server_tool格式）
  const getConfigForBackend = (): AppConfig => {
    // 使用深层复制，确保不会意外修改原始配置
    const backendConfig = JSON.parse(JSON.stringify(config.value));
    
    // 构建MCP服务器配置
    if (mcpServerConfigs.value && mcpServerConfigs.value.length > 0) {
      // 创建一个新的服务器映射
      const newMcpServers: Record<string, MCPServer> = {};
      
      // 从mcpServerConfigs创建服务器映射
      mcpServerConfigs.value.forEach(serverConfig => {
        if (!serverConfig.disabled) {
          try {
            // 解析JSON格式的参数和环境变量
            let args: any = [];
            let env: any = {};
            
            try {
              args = typeof serverConfig.args === 'string' 
                ? (serverConfig.args.trim() ? JSON.parse(serverConfig.args) : []) 
                : serverConfig.args || [];
            } catch (e) {
              console.error(`【配置】解析服务器${serverConfig.name}的args失败:`, e, '原始值:', serverConfig.args);
              args = [];
            }
            
            try {
              env = typeof serverConfig.env === 'string' 
                ? (serverConfig.env.trim() ? JSON.parse(serverConfig.env) : {}) 
                : serverConfig.env || {};
            } catch (e) {
              console.error(`【配置】解析服务器${serverConfig.name}的env失败:`, e, '原始值:', serverConfig.env);
              env = {};
            }
            
            // 创建服务器配置，确保数据类型正确
            newMcpServers[serverConfig.name] = {
              command: serverConfig.command || "",
              args: Array.isArray(args) ? args : [],
              env: typeof env === 'object' && env !== null ? env : {},
              status: 'disconnected'
            };
          } catch (e) {
            console.error(`创建服务器配置失败: ${serverConfig.name}`, e);
          }
        }
      });
      
      // 如果没有可用的服务器，添加一个默认服务器
      if (Object.keys(newMcpServers).length === 0) {
        newMcpServers['ddg'] = {
          command: 'ddg',
          args: [],
          env: {},
          status: 'disconnected'
        };
      }
      
      // 更新backendConfig中的mcp_servers
      backendConfig.mcp.mcp_servers = newMcpServers;
    } else {
      // 应急措施：添加一个默认服务器
      backendConfig.mcp.mcp_servers = {
        'ddg': {
          command: 'ddg',
          args: [],
          env: {},
          status: 'disconnected'
        }
      };
    }
    
    // 不需要转换，直接使用当前的工具配置格式
    
    return backendConfig;
  }

  // LLM配置管理方法
  // 缓存状态
  const llmConfigsLastLoaded = ref<number | null>(null)
  const mcpConfigsLastLoaded = ref<number | null>(null)
  const CACHE_TIMEOUT = 5 * 60 * 1000 // 5分钟缓存

  const loadLLMConfigs = async (force = false) => {
    try {
      console.log('【配置】开始加载LLM配置，强制刷新:', force, new Date().toISOString())
      if (llmConfigs.value.length === 0 || force) {
        const response = await llmApi.getConfigs()
        console.log('【配置】LLM配置加载结果:', response)
        
        if (response.data) {
          // 完全替换配置列表，而不是修改现有列表
          // 这确保了Vue的响应式系统可以检测到变化
          llmConfigs.value = [...response.data]
          console.log('【配置】LLM配置列表已更新，当前长度:', llmConfigs.value.length)
          
          // 如果有默认配置，使用默认配置
          const defaultConfig = llmConfigs.value.find(config => config.is_default)
          if (defaultConfig) {
            currentLLMConfigId.value = defaultConfig.id
            // 同步默认配置到应用
            syncLLMConfigToApp(defaultConfig)
            console.log('【配置】设置默认LLM配置:', defaultConfig.name)
          }
        }
      }
    } catch (error) {
      console.error('加载LLM配置失败:', error)
    }
  }

  const createLLMConfig = async (configForm: CreateLLMConfigForm) => {
    try {
      const response = await llmApi.createConfig(configForm)
      if (response.data) {
        llmConfigs.value.push(response.data)
      }
      return response.data
    } catch (error) {
      console.error('创建LLM配置失败:', error)
      throw error
    }
  }

  const updateLLMConfigById = async (id: number, updates: Partial<CreateLLMConfigForm>) => {
    try {
      await llmApi.updateConfig(id, updates)
      // 更新本地数据
      const index = llmConfigs.value.findIndex(config => config.id === id)
      if (index !== -1) {
        // 完全替换对象，而不是仅更新部分属性
        // 之前的实现可能会导致视图未更新
        const updatedConfig = { ...llmConfigs.value[index] };
        
        // 显式更新每个属性，确保响应式系统能检测到变化
        if (updates.name !== undefined) updatedConfig.name = updates.name;
        if (updates.description !== undefined) updatedConfig.description = updates.description;
        if (updates.type !== undefined) updatedConfig.type = updates.type;
        if (updates.base_url !== undefined) updatedConfig.base_url = updates.base_url;
        if (updates.model !== undefined) updatedConfig.model = updates.model;
        if (updates.api_key !== undefined) updatedConfig.api_key = updates.api_key;
        if (updates.temperature !== undefined) updatedConfig.temperature = updates.temperature;
        if (updates.max_tokens !== undefined) updatedConfig.max_tokens = updates.max_tokens;
        if (updates.is_default !== undefined) updatedConfig.is_default = updates.is_default;
        
        // 更新时间戳
        updatedConfig.updated_at = new Date().toISOString();
        
        // 替换整个对象以确保视图更新
        llmConfigs.value[index] = updatedConfig;
        
        // 如果当前正在使用该配置，同步更新到应用
        if (currentLLMConfigId.value === id) {
          syncLLMConfigToApp(updatedConfig);
        }
      }
    } catch (error) {
      console.error('更新LLM配置失败:', error)
      throw error
    }
  }

  const deleteLLMConfig = async (id: number) => {
    try {
      await llmApi.deleteConfig(id)
      // 更新本地数据
      const index = llmConfigs.value.findIndex(config => config.id === id)
      if (index !== -1) {
        llmConfigs.value.splice(index, 1)
      }
      
      // 如果删除的是当前选中的配置，重置选中状态
      if (currentLLMConfigId.value === id) {
        currentLLMConfigId.value = null
      }
    } catch (error) {
      console.error('删除LLM配置失败:', error)
      throw error
    }
  }

  const setDefaultLLMConfig = async (id: number) => {
    try {
      await llmApi.setDefaultConfig(id)
      // 更新本地数据
      llmConfigs.value.forEach(config => {
        config.is_default = config.id === id
      })
      
      // 更新当前选中的配置
      currentLLMConfigId.value = id
      
      // 同步到应用配置
      const selectedConfig = llmConfigs.value.find(config => config.id === id)
      if (selectedConfig) {
        syncLLMConfigToApp(selectedConfig)
      }
    } catch (error) {
      console.error('设置默认LLM配置失败:', error)
      throw error
    }
  }

  const selectLLMConfig = (id: number) => {
    currentLLMConfigId.value = id
    
    // 同步到应用配置
    const selectedConfig = llmConfigs.value.find(config => config.id === id)
    if (selectedConfig) {
      syncLLMConfigToApp(selectedConfig)
    }
  }

  const syncLLMConfigToApp = (llmConfigModel: LLMConfigModel) => {
    config.value.llm = {
      type: llmConfigModel.type,
      base_url: llmConfigModel.base_url,
      model: llmConfigModel.model,
      api_key: llmConfigModel.api_key,
      temperature: llmConfigModel.temperature,
      max_tokens: llmConfigModel.max_tokens
    }
  }

  // MCP服务器配置管理方法
  const loadMCPServerConfigs = async (force = false) => {
    try {
      if (mcpServerConfigs.value.length === 0 || force) {
        const response = await mcpApi.getServerConfigs()
        if (response.data) {
          // 确保mcpServerConfigs的类型与API返回的数据类型匹配
          mcpServerConfigs.value = response.data.map(config => {
            return {
              id: config.id,
              name: config.name,
              description: config.description,
              transport_type: config.transport_type || 'stdio', // 默认为stdio保持向后兼容
              command: config.command || '',
              args: typeof config.args === 'string' ? config.args : JSON.stringify(config.args),
              env: typeof config.env === 'string' ? config.env : JSON.stringify(config.env),
              url: config.url || '',
              headers: config.headers || '',
              disabled: config.disabled,
              is_active: config.is_active,
              created_at: config.created_at,
              updated_at: config.updated_at
            }
          })
        }
      }
    } catch (error) {
      console.error('加载MCP服务器配置失败:', error)
    }
  }

  const createMCPServerConfig = async (configForm: CreateMCPServerConfigForm) => {
    try {
      const response = await mcpApi.createServerConfig(configForm)
      if (response.data) {
        // 确保添加到mcpServerConfigs的数据类型正确
        const newConfig = {
          id: response.data.id,
          name: response.data.name,
          description: response.data.description,
          transport_type: response.data.transport_type || 'stdio',
          command: response.data.command || '',
          args: typeof response.data.args === 'string' ? response.data.args : JSON.stringify(response.data.args),
          env: typeof response.data.env === 'string' ? response.data.env : JSON.stringify(response.data.env),
          url: response.data.url || '',
          headers: response.data.headers || '',
          disabled: response.data.disabled,
          is_active: response.data.is_active,
          created_at: response.data.created_at,
          updated_at: response.data.updated_at
        }
        mcpServerConfigs.value.push(newConfig)
        return response.data
      }
      return response.data
    } catch (error) {
      console.error('创建MCP服务器配置失败:', error)
      throw error
    }
  }

  const updateMCPServerConfigById = async (id: number, updates: Partial<CreateMCPServerConfigForm>) => {
    try {
      await mcpApi.updateServerConfig(id, updates)
      // 更新本地数据
      const index = mcpServerConfigs.value.findIndex(config => config.id === id)
      if (index !== -1) {
        // 确保更新的数据类型正确
        const updatedConfig = { ...mcpServerConfigs.value[index] }
        
        if (updates.name) updatedConfig.name = updates.name
        if (updates.description) updatedConfig.description = updates.description
        if (updates.command) updatedConfig.command = updates.command
        if (updates.args) updatedConfig.args = typeof updates.args === 'string' ? updates.args : JSON.stringify(updates.args)
        if (updates.env) updatedConfig.env = typeof updates.env === 'string' ? updates.env : JSON.stringify(updates.env)
        if (typeof updates.disabled !== 'undefined') updatedConfig.disabled = updates.disabled
        
        mcpServerConfigs.value[index] = updatedConfig
      }
    } catch (error) {
      console.error('更新MCP服务器配置失败:', error)
      throw error
    }
  }

  const deleteMCPServerConfig = async (id: number) => {
    try {
      await mcpApi.deleteServerConfig(id)
      // 更新本地数据
      const index = mcpServerConfigs.value.findIndex(config => config.id === id)
      if (index !== -1) {
        mcpServerConfigs.value.splice(index, 1)
      }
    } catch (error) {
      console.error('删除MCP服务器配置失败:', error)
      throw error
    }
  }

  const loadToolsFromDatabase = async () => {
    try {
      console.log('【配置】loadToolsFromDatabase开始执行', new Date().toISOString())
      const response = await mcpApi.getToolsFromDB()
      console.log('【配置】loadToolsFromDatabase执行结果', response)
      
      if (response.tools) {
        updateAvailableTools(response.tools)
      }
    } catch (error) {
      console.error('从数据库加载工具失败:', error)
    }
  }

  const buildMCPConfigFromDatabase = () => {
    try {
      console.log('【配置】buildMCPConfigFromDatabase开始执行', new Date().toISOString());
      
      if (!mcpServerConfigs.value || mcpServerConfigs.value.length === 0) {
        console.warn('【配置】没有可用的MCP服务器配置，无法构建MCP配置');
        
        // 应急措施：创建一个默认的DDG服务器
        config.value.mcp.mcp_servers = {
          'ddg': {
            command: 'ddg',
            args: [],
            env: {},
            status: 'disconnected'
          }
        };
        
        console.log('【配置】添加了应急服务器配置:', JSON.stringify(config.value.mcp.mcp_servers));
        return;
      }
      
      // 保存当前选中的工具
      const selectedToolNames = [...config.value.mcp.tools];
      console.log('【配置】当前选中的工具:', selectedToolNames);
      
      // 创建新的MCP服务器映射
      const newMcpServers: Record<string, MCPServer> = {};
      
      // 遍历数据库服务器配置，构建MCP服务器映射
      mcpServerConfigs.value.forEach(serverConfig => {
        // 跳过禁用的服务器
        if (serverConfig.disabled) {
          console.log(`【配置】服务器 ${serverConfig.name} 已禁用，跳过`);
          return;
        }
          
        try {
          // 解析JSON格式的参数和环境变量
          let args: any = [];
          let env: any = {};
          
          try {
            args = typeof serverConfig.args === 'string' 
              ? (serverConfig.args.trim() ? JSON.parse(serverConfig.args) : []) 
              : serverConfig.args || [];
          } catch (e) {
            console.error(`【配置】解析服务器${serverConfig.name}的args失败:`, e, '原始值:', serverConfig.args);
            args = [];
          }
          
          try {
            env = typeof serverConfig.env === 'string' 
              ? (serverConfig.env.trim() ? JSON.parse(serverConfig.env) : {}) 
              : serverConfig.env || {};
          } catch (e) {
            console.error(`【配置】解析服务器${serverConfig.name}的env失败:`, e, '原始值:', serverConfig.env);
            env = {};
          }
            
          // 创建服务器配置
          newMcpServers[serverConfig.name] = {
            command: serverConfig.command || '',
            args: Array.isArray(args) ? args : [],
            env: typeof env === 'object' && env !== null ? env : {},
            status: 'disconnected' // 初始状态为断开连接
          };
          
          console.log(`【配置】添加服务器 ${serverConfig.name}:`, JSON.stringify(newMcpServers[serverConfig.name]));
        } catch (e) {
          console.error(`【配置】解析服务器配置失败: ${serverConfig.name}`, e);
        }
      });
      
      // 如果没有可用的服务器，添加一个默认的
      if (Object.keys(newMcpServers).length === 0) {
        console.warn('【配置】无可用服务器，添加默认DDG服务器');
        newMcpServers['ddg'] = {
          command: 'ddg',
          args: [],
          env: {},
          status: 'disconnected'
        };
      }
        
      // 更新MCP服务器配置
      config.value.mcp.mcp_servers = newMcpServers;
      console.log('【配置】更新后的mcp_servers:', JSON.stringify(config.value.mcp.mcp_servers));
        
      // 恢复选中的工具
      config.value.mcp.tools = selectedToolNames;
      console.log('【配置】恢复选中的工具:', config.value.mcp.tools);
        
      console.log('【配置】构建MCP配置完成');
    } catch (error) {
      console.error('【配置】构建MCP配置失败:', error);
      
      // 应急处理：创建默认服务器
      config.value.mcp.mcp_servers = {
        'ddg': {
          command: 'ddg',
          args: [],
          env: {},
          status: 'disconnected'
        }
      };
    }
  }

  // SystemPrompt配置管理方法
  // 缓存状态
  const systemPromptsLastLoaded = ref<number | null>(null)

  const loadSystemPrompts = async (force = false) => {
    try {
      console.log('【配置】开始加载SystemPrompt配置，强制刷新:', force, new Date().toISOString())
      if (systemPrompts.value.length === 0 || force) {
        const response = await systemPromptApi.getPrompts()
        console.log('【配置】SystemPrompt配置加载结果:', response)
        
        // 检查response是否为数组或包含data字段
        const responseData = Array.isArray(response) ? response : response.data;
        
        if (responseData) {
          // 处理API返回的数据，解析placeholders字段
          const processedData = responseData.map(item => {
            try {
              // 如果placeholders是字符串，尝试解析为数组
              if (typeof item.placeholders === 'string') {
                item.placeholders = JSON.parse(item.placeholders);
              }
              return item;
            } catch (e) {
              console.error(`解析placeholders失败:`, e, '原始值:', item.placeholders);
              // 如果解析失败，设置为空数组
              item.placeholders = [];
              return item;
            }
          });
          
          // 完全替换配置列表，而不是修改现有列表
          systemPrompts.value = [...processedData];
          console.log('【配置】SystemPrompt配置列表已更新，当前长度:', systemPrompts.value.length)
          
          // 如果有默认配置，使用默认配置
          const defaultPrompt = systemPrompts.value.find(prompt => prompt.is_default)
          if (defaultPrompt) {
            currentSystemPromptId.value = defaultPrompt.id
            // 同步默认配置到应用
            syncSystemPromptToApp(defaultPrompt)
            console.log('【配置】设置默认SystemPrompt配置:', defaultPrompt.name)
          }
        }
      }
    } catch (error) {
      console.error('加载SystemPrompt配置失败:', error)
    }
  }

  const createSystemPrompt = async (promptForm: CreateSystemPromptForm) => {
    try {
      const response = await systemPromptApi.createPrompt(promptForm)
      if (response.data) {
        systemPrompts.value.push(response.data)
      }
      return response.data
    } catch (error) {
      console.error('创建SystemPrompt配置失败:', error)
      throw error
    }
  }

  const updateSystemPromptById = async (id: number, updates: Partial<CreateSystemPromptForm>) => {
    try {
      await systemPromptApi.updatePrompt(id, updates)
      // 更新本地数据
      const index = systemPrompts.value.findIndex(prompt => prompt.id === id)
      if (index !== -1) {
        // 完全替换对象，确保响应式系统能检测到变化
        const updatedPrompt = { ...systemPrompts.value[index] }
        
        // 显式更新每个属性
        if (updates.name !== undefined) updatedPrompt.name = updates.name
        if (updates.description !== undefined) updatedPrompt.description = updates.description
        if (updates.content !== undefined) updatedPrompt.content = updates.content
        if (updates.placeholders !== undefined) updatedPrompt.placeholders = updates.placeholders
        if (updates.is_default !== undefined) updatedPrompt.is_default = updates.is_default
        
        // 更新时间戳
        updatedPrompt.updated_at = new Date().toISOString()
        
        // 替换整个对象以确保视图更新
        systemPrompts.value[index] = updatedPrompt
        
        // 如果当前正在使用该配置，同步更新到应用
        if (currentSystemPromptId.value === id) {
          syncSystemPromptToApp(updatedPrompt)
        }
      }
    } catch (error) {
      console.error('更新SystemPrompt配置失败:', error)
      throw error
    }
  }

  const deleteSystemPrompt = async (id: number) => {
    try {
      await systemPromptApi.deletePrompt(id)
      // 更新本地数据
      const index = systemPrompts.value.findIndex(prompt => prompt.id === id)
      if (index !== -1) {
        systemPrompts.value.splice(index, 1)
      }
      
      // 如果删除的是当前选中的配置，重置选中状态
      if (currentSystemPromptId.value === id) {
        currentSystemPromptId.value = null
      }
    } catch (error) {
      console.error('删除SystemPrompt配置失败:', error)
      throw error
    }
  }

  const setDefaultSystemPrompt = async (id: number) => {
    try {
      await systemPromptApi.setDefaultPrompt(id)
      // 更新本地数据
      systemPrompts.value.forEach(prompt => {
        prompt.is_default = prompt.id === id
      })
      
      // 更新当前选中的配置
      currentSystemPromptId.value = id
      
      // 同步到应用配置
      const selectedPrompt = systemPrompts.value.find(prompt => prompt.id === id)
      if (selectedPrompt) {
        syncSystemPromptToApp(selectedPrompt)
      }
    } catch (error) {
      console.error('设置默认SystemPrompt配置失败:', error)
      throw error
    }
  }

  const selectSystemPrompt = (id: number) => {
    currentSystemPromptId.value = id
    
    // 同步到应用配置
    const selectedPrompt = systemPrompts.value.find(prompt => prompt.id === id)
    if (selectedPrompt) {
      syncSystemPromptToApp(selectedPrompt)
    }
  }

  const syncSystemPromptToApp = (promptModel: SystemPromptModel) => {
    config.value.system_prompt = promptModel.content
    
    // 初始化占位符
    const placeholders: Record<string, any> = {}
    promptModel.placeholders.forEach(key => {
      if (key === 'date') {
        placeholders[key] = new Date().toLocaleDateString('zh-CN')
      } else {
        placeholders[key] = config.value.placeholders[key] || ''
      }
    })
    
    config.value.placeholders = placeholders
  }

  return {
    config,
    availableModels,
    availableTools,
    llmConfigs,
    currentLLMConfigId,
    mcpServerConfigs,
    systemPrompts,
    currentSystemPromptId,
    isLoading,
    error,
    configLoaded,

    // 计算属性
    mcpServers,
    connectedServers,
    selectedTools,
    currentLLMConfig,
    defaultLLMConfig,
    currentSystemPrompt,
    defaultSystemPrompt,

    // 方法
    loadConfiguration,
    saveConfiguration,
    updateLLMConfig,
    updateMCPConfig,
    addMCPServer,
    removeMCPServer,
    updateMCPServerStatus,
    updateSystemPrompt,
    updatePlaceholders,
    selectTools,
    updateAvailableTools,
    getToolConfigsFromNames,
    getToolNamesFromConfigs,
    getConfigForBackend,
    loadLLMConfigs,
    createLLMConfig,
    updateLLMConfigById,
    deleteLLMConfig,
    setDefaultLLMConfig,
    selectLLMConfig,
    syncLLMConfigToApp,
    loadMCPServerConfigs,
    createMCPServerConfig,
    updateMCPServerConfigById,
    deleteMCPServerConfig,
    loadToolsFromDatabase,
    buildMCPConfigFromDatabase,
    loadSystemPrompts,
    createSystemPrompt,
    updateSystemPromptById,
    deleteSystemPrompt,
    setDefaultSystemPrompt,
    selectSystemPrompt,
    syncSystemPromptToApp
  }
})
