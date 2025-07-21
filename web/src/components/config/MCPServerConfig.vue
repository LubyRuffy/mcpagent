<template>
  <el-card class="config-card">
    <template #header>
      <div class="card-header">
        <span>{{ $t('config.mcp.title') }}</span>
        <el-badge
            :value="`${activeServerCount}/${totalServerCount}`"
            type="primary"
            class="status-badge"
            :title="$t('config.mcp.serverCountHint', { active: activeServerCount, total: totalServerCount })"
        />
      </div>
    </template>

    <div class="mcp-config">


      <!-- 已选择的工具列表 -->
      <div class="tools-section">
        <div class="section-header">
          <span class="section-title">{{ $t('config.mcp.selectedTools') }}
            <el-tag v-if="selectedTools.length > 0" size="small" type="info" class="count-tag">
              {{ selectedTools.length }}
            </el-tag>
          </span>
          <el-button
              type="primary"
              size="small"
              @click="openAddToolDialog"
          >
            {{ $t('config.mcp.addTool') }}
          </el-button>
        </div>

        <div class="selected-tools-list">
          <div
              v-for="combinedName in selectedTools"
              :key="combinedName"
              class="selected-tool-item"
          >
            <div class="tool-info">
              <div class="tool-name" @click="toggleToolDescription(combinedName)">
                {{ getToolDisplayName(combinedName) }}
                <span class="server-tag" :title="getToolServer(combinedName)">
                  {{ getToolServer(combinedName) }}
                </span>
                <el-icon class="expand-icon" :class="{ 'expanded': expandedTools.has(combinedName) }">
                  <ArrowDown />
                </el-icon>
              </div>

              <div
                  v-if="expandedTools.has(combinedName)"
                  class="tool-description-content"
              >
                {{ getToolDescription(combinedName) }}
              </div>
            </div>

            <el-button
                size="small"
                type="danger"
                circle
                @click="removeTool(combinedName)"
                class="delete-btn"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>

          <el-empty
              v-if="selectedTools.length === 0"
              :description="$t('config.mcp.noToolsSelected')"
              :image-size="80"
          />
        </div>
      </div>
    </div>

    <!-- 添加/编辑服务器对话框 -->
    <el-dialog
        v-model="showAddServerDialog"
        :title="editingServer ? $t('config.mcp.editServer') : $t('config.mcp.addServer')"
        :width="serverDialogWidth"
        :close-on-click-modal="false"
        :close-on-press-escape="true"
        class="server-dialog"
        append-to-body
    >
      <el-form
          ref="serverFormRef"
          :model="serverForm"
          :rules="serverFormRules"
          label-width="180px"
      >
        <el-form-item :label="$t('config.mcp.serverName')" prop="name">
          <el-input
              v-model="serverForm.name"
              :disabled="!!editingServer"
              :placeholder="$t('config.mcp.serverNamePlaceholder')"
          />
        </el-form-item>

        <el-form-item :label="$t('config.mcp.transportType')" prop="transport_type">
          <el-select
              v-model="serverForm.transport_type"
              :teleported="false"
              :placeholder="$t('config.mcp.transportTypePlaceholder')"
              style="width: 100%"
          >
            <el-option
                value="stdio"
                :label="$t('config.mcp.transportStdio')"
            />
            <el-option
                value="sse"
                :label="$t('config.mcp.transportSSE')"
            />
            <el-option
                value="http"
                :label="$t('config.mcp.transportHTTP')"
            />
          </el-select>
        </el-form-item>

        <!-- STDIO specific fields -->
        <template v-if="serverForm.transport_type === 'stdio'">
          <el-form-item :label="$t('config.mcp.command')" prop="command">
            <el-input
                v-model="serverForm.command"
                :placeholder="$t('config.mcp.commandPlaceholder')"
            />
          </el-form-item>

          <el-form-item :label="$t('config.mcp.args')">
            <el-tag
                v-for="(arg, index) in serverForm.args"
                :key="index"
                closable
                @close="removeArg(index)"
                class="arg-tag"
            >
              {{ arg }}
            </el-tag>

            <el-input
                v-if="showArgInput"
                ref="argInputRef"
                v-model="newArg"
                size="small"
                @keyup.enter="addArg"
                @blur="addArg"
                class="arg-input"
            />

            <el-button
                v-else
                size="small"
                @click="showNewArgInput"
                class="add-arg-btn"
            >
              + {{ $t('config.mcp.addArg') }}
            </el-button>
          </el-form-item>

          <el-form-item :label="$t('config.mcp.env')">
            <div class="env-vars">
              <div
                  v-for="(value, key) in serverForm.env"
                  :key="key"
                  class="env-var-item"
              >
                <el-input
                    :model-value="key"
                    :placeholder="$t('config.mcp.varName')"
                    size="small"
                    @input="updateEnvKey(key, $event)"
                />
                <el-input
                    v-model="serverForm.env[key]"
                    :placeholder="$t('config.mcp.varValue')"
                    size="small"
                />
                <el-button
                    size="small"
                    type="danger"
                    @click="removeEnvVar(key)"
                >
                  {{ $t('common.delete') }}
                </el-button>
              </div>

              <el-button
                  size="small"
                  @click="addEnvVar"
              >
                + {{ $t('config.mcp.addEnvVar') }}
              </el-button>
            </div>
          </el-form-item>
        </template>

        <!-- SSE/HTTP specific fields -->
        <template v-if="serverForm.transport_type === 'sse' || serverForm.transport_type === 'http'">
          <el-form-item :label="$t('config.mcp.url')" prop="url">
            <el-input
                v-model="serverForm.url"
                :placeholder="$t('config.mcp.urlPlaceholder')"
            />
          </el-form-item>

          <el-form-item :label="$t('config.mcp.headers')">
            <el-tag
                v-for="(header, index) in serverForm.headers"
                :key="index"
                closable
                @close="removeHeader(index)"
                class="header-tag"
            >
              {{ header }}
            </el-tag>

            <el-input
                v-if="showHeaderInput"
                ref="headerInputRef"
                v-model="newHeader"
                size="small"
                @keyup.enter="addHeader"
                @blur="addHeader"
                class="header-input"
                :placeholder="$t('config.mcp.headerPlaceholder')"
            />

            <el-button
                v-else
                size="small"
                @click="showNewHeaderInput"
                class="add-header-btn"
            >
              + {{ $t('config.mcp.addHeader') }}
            </el-button>
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="cancelServerEdit">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveServer">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- 添加工具对话框 -->
    <el-dialog
        v-model="showAddToolDialog"
        :title="$t('config.mcp.selectTools')"
        :width="dialogWidth"
        :close-on-click-modal="false"
        :close-on-press-escape="true"
        destroy-on-close
    >
      <div class="tool-selection">
        <!-- 添加服务器按钮 -->
        <div class="tool-selection-header">
          <el-button
              type="primary"
              size="small"
              @click="showAddServerInToolDialog"
          >
            {{ $t('config.mcp.addServer') }}
          </el-button>
          <span class="selection-tip">{{ $t('config.mcp.selectionTip') }}</span>
        </div>

        <!-- 工具树 -->
        <el-tree
            ref="toolTreeRef"
            :data="toolTreeData"
            :props="treeProps"
            show-checkbox
            node-key="id"
            :default-checked-keys="getDefaultCheckedKeys()"
            @check="handleToolTreeCheck"
            class="tool-tree"
        >
          <template #default="{ node, data }">
            <div class="tree-node" :class="{ 'inner-server-node': data.isServer && data.isInnerServer }">
              <div class="node-content">
                <div class="node-label">{{ data.label }}</div>
                <div v-if="data.description" class="node-description">
                  {{ data.description }}
                </div>
              </div>
              <!-- 服务器节点的操作按钮 -->
              <div v-if="data.isServer" class="server-actions">
                <el-button
                    v-if="!data.isInnerServer"
                    size="small"
                    @click.stop="editServerInToolDialog(data.serverName)"
                >
                  {{ $t('common.edit') }}
                </el-button>
                <el-button
                    v-if="!data.isInnerServer"
                    size="small"
                    type="danger"
                    @click.stop="removeServerInToolDialog(data.serverName)"
                >
                  {{ $t('common.delete') }}
                </el-button>
                <el-tag v-if="data.isInnerServer" type="info" size="small">{{ $t('config.mcp.internalServer') }}</el-tag>
              </div>
            </div>
          </template>
        </el-tree>

        <!-- 空状态 -->
        <el-empty
            v-if="toolTreeData.length === 0"
                          :description="$t('config.mcp.noServers')"
            :image-size="80"
        >
          <template #default>
            <div class="empty-message">
              <p>{{ $t('config.mcp.emptyToolList') }}</p>
              <ol>
                <li>{{ $t('config.mcp.emptyReason1') }}</li>
                <li>{{ $t('config.mcp.emptyReason2') }}</li>
                <li>{{ $t('config.mcp.emptyReason3') }}</li>
              </ol>
              
              <div class="debug-info">
                <p>{{ $t('config.mcp.debugInfo') }}:</p>
                <ul>
                  <li>{{ $t('config.mcp.debugTools') }}: {{ configStore.availableTools?.length || 0 }}</li>
                  <li>{{ $t('config.mcp.debugServers') }}: {{ configStore.mcpServerConfigs?.length || 0 }}</li>
                  <li>{{ $t('config.mcp.debugApiResponse') }}: {{ (configStore.availableTools && configStore.availableTools.length > 0) ? $t('config.mcp.hasData') : $t('config.mcp.noData') }}</li>
                  <li>{{ $t('config.mcp.debugRefreshTime') }}: {{ new Date().toISOString() }}</li>
                </ul>
                
                <p v-if="configStore.availableTools && configStore.availableTools.length > 0">
                  <el-button size="small" @click="refreshTree">{{ $t('config.mcp.forceRefresh') }}</el-button>
                </p>
              </div>
            </div>
            
            <el-button
                type="primary"
                @click="showAddServerInToolDialog"
            >
              {{ $t('config.mcp.addServer') }}
            </el-button>
          </template>
        </el-empty>
      </div>

      <template #footer>
        <el-button @click="cancelToolSelection">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmToolSelection">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Delete } from '@element-plus/icons-vue'
import { useConfigStore } from '@/stores/config'
import { mcpApi } from '@/utils/api'
import type { MCPServer, MCPConfig, MCPToolConfig } from '@/types/config'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const configStore = useConfigStore()

// 响应式状态
const showAddServerDialog = ref(false)
const showAddToolDialog = ref(false)
const editingServer = ref<string | null>(null)
const showArgInput = ref(false)
const newArg = ref('')
const argInputRef = ref()
const showHeaderInput = ref(false)
const newHeader = ref('')
const headerInputRef = ref()
const toolTreeRef = ref()
// 工具详情展开状态
const expandedTools = ref<Set<string>>(new Set())

// 表单数据
const serverForm = ref({
  name: '',
  transport_type: 'stdio',
  // STDIO fields
  command: '',
  args: [] as string[],
  env: {} as Record<string, string>,
  // SSE/HTTP fields
  url: '',
  headers: [] as string[],
  // Common fields
  disabled: false
})

  // 表单验证规则 - 使用computed实现条件验证
const serverFormRules = computed(() => {
  const rules: any = {
    name: [
      { required: true, message: t('config.mcp.serverNameRequired'), trigger: 'blur' }
    ],
    transport_type: [
      { required: true, message: t('config.mcp.transportTypeRequired'), trigger: 'change' }
    ]
  }

  // 根据传输类型添加条件验证
  if (serverForm.value.transport_type === 'stdio') {
    rules.command = [
      { required: true, message: t('config.mcp.commandRequired'), trigger: 'blur' }
    ]
  } else if (serverForm.value.transport_type === 'sse' || serverForm.value.transport_type === 'http') {
    rules.url = [
      { required: true, message: t('config.mcp.urlRequired'), trigger: 'blur' },
      { 
        pattern: /^https?:\/\/.+/, 
        message: t('config.mcp.urlInvalid'), 
        trigger: 'blur' 
      }
    ]
  }

  return rules
})

// 树形组件配置
const treeProps = {
  children: 'children',
  label: 'label'
}

// 选中的工具列表，只保存工具名称，方便处理
const selectedTools = ref<string[]>([]);

// 获取用户真实选择的工具
const loadSelectedToolsFromBackend = async () => {
  try {
    console.log('【工具】开始从后端加载已选择的工具...', new Date().toISOString());
    
    // 加载可用工具列表
    await configStore.loadToolsFromDatabase();
    console.log('【工具】可用工具列表加载完成:', configStore.availableTools.length);
    
    // 从配置中获取工具列表
    await configStore.loadConfiguration(true);
    console.log('【工具】从后端加载的配置:', configStore.config);
    
    // 从配置中提取工具信息
    const toolList: string[] = [];
    
    // 处理工具数据
    if (configStore.config.mcp.tools && configStore.config.mcp.tools.length > 0) {
      // 使用any类型暂时解决类型问题
      const mcpTools: any[] = configStore.config.mcp.tools;
      
      mcpTools.forEach(tool => {
        try {
          let toolName: string;
          let serverName: string;
          
          if (typeof tool === 'object' && tool !== null) {
            toolName = tool.name;
            serverName = tool.server || 'inner';
          } else {
            toolName = String(tool);
            const matchedTool = configStore.availableTools.find(t => t.name === toolName);
            serverName = matchedTool?.server || 'inner';
          }
          
          // 使用"工具名@服务器名"格式
          toolList.push(`${toolName}@${serverName}`);
        } catch (e) {
          console.warn('【工具】处理工具数据失败:', e, tool);
        }
      });
    }
    
    console.log('【工具】从后端加载的已选择工具:', toolList);
    selectedTools.value = toolList;
  } catch (error) {
    console.error('【工具】加载已选择工具失败:', error);
    ElMessage.error(t('error.loadSelectedToolsFailed'));
  }
};

// 计算属性
const mcpConfig = computed({
  get: () => configStore.config.mcp,
  set: (value: MCPConfig) => configStore.updateMCPConfig(value)
})

// 计算活跃服务器数量和总服务器数量
const activeServerCount = computed(() => {
  // 获取已选择工具对应的服务器
  const selectedServers = new Set<string>()
  
  // 从选择的工具中提取对应的服务器
  selectedTools.value.forEach(toolName => {
    const tool = configStore.availableTools.find(t => t.name === toolName)
    if (tool && tool.server) {
      selectedServers.add(tool.server)
    }
  })
  
  // 只统计已选择工具相关的活跃服务器
  return selectedServers.size
})

const totalServerCount = computed(() => {
  // 统计所有活跃的服务器总数
  return configStore.mcpServerConfigs.filter(config => config.is_active && !config.disabled).length
})

// 响应式弹窗宽度
const dialogWidth = computed(() => {
  if (typeof window !== 'undefined') {
    const width = window.innerWidth
    if (width <= 768) return '98%'
    if (width <= 1200) return '90%'
    return '800px'
  }
  return '800px'
})

// 服务器配置弹窗宽度
const serverDialogWidth = computed(() => {
  if (typeof window !== 'undefined') {
    const width = window.innerWidth
    if (width <= 768) return '95%'
    if (width <= 1024) return '80%'
    return '500px'
  }
  return '500px'
})

// 构建工具树数据
const toolTreeData = computed(() => {
  const treeData: any[] = []
  // 用于存储inner服务器节点
  let innerServerNode: any = null

  console.log('【工具树】开始构建工具树数据...', new Date().toISOString())
  console.log('【工具树】可用工具列表详情:', JSON.stringify(configStore.availableTools))
  console.log('【工具树】MCP服务器配置列表:', configStore.mcpServerConfigs)
  console.log('【工具树】MCP服务器配置数量:', configStore.mcpServerConfigs.length)

  if (!configStore.availableTools || configStore.availableTools.length === 0) {
    console.warn('【工具树】警告: 没有可用工具数据，树将为空')
    return treeData
  }

  // 工具服务器分布统计
  const toolServerCounts: Record<string, number> = {}
  configStore.availableTools.forEach(tool => {
    toolServerCounts[tool.server] = (toolServerCounts[tool.server] || 0) + 1
  })
  console.log('【工具树】工具按服务器分布:', toolServerCounts)

  // 记录所有遇到的服务器名称，以便找出不在配置中的服务器
  const processedServers = new Set<string>()

  // 遍历所有数据库中的MCP服务器配置
  configStore.mcpServerConfigs.forEach(serverConfig => {
    if (!serverConfig.is_active) {
      console.log(`【工具树】服务器 ${serverConfig.name} 未激活，跳过`)
      return
    }

    processedServers.add(serverConfig.name)

    let args: string[] = []
    try {
      args = serverConfig.args ? JSON.parse(serverConfig.args) : []
      console.log(`【工具树】服务器 ${serverConfig.name} 参数解析成功:`, args)
    } catch (e) {
      console.error(`【工具树】解析服务器 ${serverConfig.name} 参数失败:`, e)
    }

    const serverNode = {
      id: `server_${serverConfig.name}`,
      label: serverConfig.name,
      isServer: true,
      serverName: serverConfig.name,
      isInnerServer: serverConfig.name === 'inner', // 标记内置服务器
      description: `${serverConfig.command} ${args.join(' ')}`,
      children: [] as any[]
    }

    // 获取该服务器下的工具
    const serverTools = configStore.availableTools.filter(tool => {
      const match = tool.server === serverConfig.name
      if (match) {
        console.log(`【工具树】工具 ${tool.name} 匹配到服务器 ${serverConfig.name}`)
      }
      return match
    })
    
    console.log(`【工具树】服务器 ${serverConfig.name} 的工具数量: ${serverTools.length}`)
    console.log(`【工具树】服务器 ${serverConfig.name} 的工具:`, JSON.stringify(serverTools))

    if (serverTools.length === 0) {
      console.log(`【工具树】警告: 服务器 ${serverConfig.name} 没有关联的工具`)
    }

    serverTools.forEach(tool => {
      // 创建唯一ID，格式为 tool_[工具名]_[服务器名]，确保在不同服务器上的同名工具能够被区分
      const uniqueId = `tool_${tool.name}_${serverConfig.name}`
      serverNode.children.push({
        id: uniqueId,
        label: tool.name,
        description: tool.description,
        serverName: serverConfig.name,
        toolName: tool.name,
        isServer: false,
        // 保存组合名称，方便后续处理
        combinedName: `${tool.name}@${serverConfig.name}`
      })
    })

    // 如果是inner服务器，保存到innerServerNode变量，等待置顶
    if (serverConfig.name === 'inner') {
      innerServerNode = serverNode
      console.log(`【工具树】数据库中的inner服务器将被置顶显示`)
    } else {
      // 添加其他服务器节点
      treeData.push(serverNode)
      console.log(`【工具树】已添加服务器 ${serverConfig.name} 到工具树`)
    }
  })

  // 处理不在数据库配置中但在工具列表中出现的服务器（如inner）
  const serversInTools = new Set(configStore.availableTools.map(tool => tool.server))
  
  for (const serverName of serversInTools) {
    if (!processedServers.has(serverName)) {
      console.log(`【工具树】发现未在数据库配置中的服务器: ${serverName}`)
      
      // 为这个服务器创建一个节点
      const serverNode = {
        id: `server_${serverName}`,
        label: serverName,
        isServer: true,
        serverName: serverName,
        isInnerServer: serverName === 'inner', // 标记内置服务器
        description: serverName === 'inner' ? '内置工具服务器' : '外部工具服务器',
        children: [] as any[]
      }
      
      // 获取该服务器下的工具
      const serverTools = configStore.availableTools.filter(tool => tool.server === serverName)
      console.log(`【工具树】服务器 ${serverName} 的工具数量: ${serverTools.length}`)
      
      serverTools.forEach(tool => {
        // 创建唯一ID，格式为 tool_[工具名]_[服务器名]
        const uniqueId = `tool_${tool.name}_${serverName}`
        serverNode.children.push({
          id: uniqueId,
          label: tool.name,
          description: tool.description,
          serverName: serverName,
          toolName: tool.name,
          isServer: false,
          // 保存组合名称，方便后续处理
          combinedName: `${tool.name}@${serverName}`
        })
      })
      
      // 如果是inner服务器，保存到innerServerNode变量
      if (serverName === 'inner') {
        innerServerNode = serverNode
        console.log(`【工具树】未配置的inner服务器将被置顶显示`)
      } else {
        // 添加其他服务器节点
        treeData.push(serverNode)
        console.log(`【工具树】已添加服务器 ${serverName} 到工具树`)
      }
    }
  }

  // 如果找到了inner服务器，将其放在数组最前面
  if (innerServerNode) {
    treeData.unshift(innerServerNode)
    console.log(`【工具树】inner服务器已置顶`)
  }

  console.log('【工具树】最终树形数据:', treeData)
  return treeData
})

// 方法
const getStatusType = (status?: string) => {
  switch (status) {
    case 'connected': return 'success'
    case 'connecting': return 'warning'
    case 'error': return 'danger'
    default: return 'info'
  }
}

const getStatusText = (status?: string) => {
  switch (status) {
    case 'connected': return '已连接'
    case 'connecting': return '连接中'
    case 'error': return '错误'
    default: return '未连接'
  }
}

const editServer = (name: string) => {
  // 禁止编辑inner服务器
  if (name === 'inner') {
    ElMessage.warning(t('config.mcp.cannotEditInnerServer'))
    return
  }

  const serverConfig = configStore.mcpServerConfigs.find(config => config.name === name)
  if (!serverConfig) return

  editingServer.value = name

  let args: string[] = []
  let env: Record<string, string> = {}

  try {
    args = serverConfig.args ? JSON.parse(serverConfig.args) : []
  } catch (e) {
    console.error('解析服务器参数失败:', e)
  }

  try {
    env = serverConfig.env ? JSON.parse(serverConfig.env) : {}
  } catch (e) {
    console.error('解析服务器环境变量失败:', e)
  }

  serverForm.value = {
    name: serverConfig.name,
    transport_type: serverConfig.transport_type || 'stdio',
    command: serverConfig.command,
    args: [...args],
    env: { ...env },
    url: serverConfig.url || '',
    headers: [], // Will be parsed below if needed
    disabled: serverConfig.disabled || false
  }
  
  // Parse headers if it's an SSE/HTTP server
  if (serverConfig.transport_type === 'sse' || serverConfig.transport_type === 'http') {
    try {
      const headers = serverConfig.headers ? JSON.parse(serverConfig.headers) : []
      serverForm.value.headers = [...headers]
    } catch (e) {
      console.error('解析服务器头部失败:', e)
      serverForm.value.headers = []
    }
  }
  showAddServerDialog.value = true
}

const removeServer = async (name: string) => {
  // 禁止删除inner服务器
  if (name === 'inner') {
    ElMessage.warning(t('config.mcp.cannotDeleteInnerServer'))
    return
  }

  try {
    await ElMessageBox.confirm(
        t('config.mcp.deleteServerConfirm', { name }),
        t('config.mcp.deleteConfirmTitle'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
    )

          // 从数据库删除服务器配置
      const serverConfig = configStore.mcpServerConfigs.find(config => config.name === name)
      if (serverConfig) {
        await configStore.deleteMCPServerConfig(serverConfig.id)
        ElMessage.success(t('config.mcp.serverDeleteSuccess'))

      // 重新加载工具列表
      await configStore.loadToolsFromDatabase()
    }
  } catch {
    // 用户取消操作
  }
}

const showNewArgInput = () => {
  showArgInput.value = true
  nextTick(() => {
    argInputRef.value?.focus()
  })
}

const addArg = () => {
  if (newArg.value.trim()) {
    serverForm.value.args.push(newArg.value.trim())
    newArg.value = ''
  }
  showArgInput.value = false
}

const removeArg = (index: number) => {
  serverForm.value.args.splice(index, 1)
}

const showNewHeaderInput = () => {
  showHeaderInput.value = true
  nextTick(() => {
    headerInputRef.value?.focus()
  })
}

const addHeader = () => {
  if (newHeader.value.trim()) {
    serverForm.value.headers.push(newHeader.value.trim())
    newHeader.value = ''
  }
  showHeaderInput.value = false
}

const removeHeader = (index: number) => {
  serverForm.value.headers.splice(index, 1)
}

const addEnvVar = () => {
  const key = 'VAR_' + (Object.keys(serverForm.value.env).length + 1)
  serverForm.value.env[key] = ''
}

const removeEnvVar = (key: string) => {
  delete serverForm.value.env[key]
}

const updateEnvKey = (oldKey: string, newKey: string) => {
  if (oldKey !== newKey && newKey) {
    const value = serverForm.value.env[oldKey]
    delete serverForm.value.env[oldKey]
    serverForm.value.env[newKey] = value
  }
}

const saveServer = async () => {
  try {
    const configData = {
      name: serverForm.value.name,
      description: '', // 可以添加描述字段
      transport_type: serverForm.value.transport_type,
      command: serverForm.value.command,
      args: serverForm.value.args,
      env: serverForm.value.env,
      url: serverForm.value.url,
      headers: serverForm.value.headers,
      disabled: serverForm.value.disabled
    }

    if (editingServer.value) {
      // 更新现有服务器配置
      const existingConfig = configStore.mcpServerConfigs.find(config => config.name === editingServer.value)
      if (existingConfig) {
        await configStore.updateMCPServerConfigById(existingConfig.id, configData)
        ElMessage.success(t('config.mcp.serverUpdateSuccess'))
      }
    } else {
      // 创建新的服务器配置
      await configStore.createMCPServerConfig(configData)
      ElMessage.success(t('config.mcp.serverCreateSuccess'))
    }

    // 重新加载工具列表
    await configStore.loadToolsFromDatabase()

    cancelServerEdit()
  } catch (error) {
    console.error('保存服务器配置失败:', error)
    ElMessage.error(t('config.mcp.saveServerConfigFailed', { message: (error as Error).message }))
  }
}

const cancelServerEdit = () => {
  showAddServerDialog.value = false
  editingServer.value = null
  serverForm.value = {
    name: '',
    transport_type: 'stdio',
    command: '',
    args: [],
    env: {},
    url: '',
    headers: [],
    disabled: false
  }
  showArgInput.value = false
  newArg.value = ''
  showHeaderInput.value = false
  newHeader.value = ''
}

const handleToolsChange = (tools: string[]) => {
  configStore.selectTools(tools)
}

// 获取工具显示名称（去除服务器信息）
const getToolDisplayName = (combinedName: string) => {
  const parts = combinedName.split('@');
  return parts[0]; // 只返回工具名部分
}

// 获取工具服务器名称
const getToolServer = (combinedName: string) => {
  const parts = combinedName.split('@');
  if (parts.length > 1) {
    return parts[1]; // 返回服务器名部分
  }
  
  // 兼容旧数据格式
  const toolName = parts[0];
  const exactMatch = configStore.availableTools.find(t => t.name === toolName);
  if (exactMatch) {
    return exactMatch.server || '';
  }
  
  // 如果没有精确匹配，检查工具名是否包含已知工具名作为前缀
  for (const tool of configStore.availableTools) {
    if (toolName.startsWith(tool.name) && tool.name.length > 3) {
      return tool.server || '';
    }
  }
  
  const matchingTools = configStore.availableTools.filter(t => 
    toolName.includes(t.name) || t.name.includes(toolName)
  );
  
  if (matchingTools.length > 0) {
    matchingTools.sort((a, b) => b.name.length - a.name.length);
    return matchingTools[0].server || '';
  }
  
  return '';
}

const getToolDescription = (combinedName: string) => {
  const toolName = getToolDisplayName(combinedName);
  
  // 直接查找精确匹配的工具
  const exactMatch = configStore.availableTools.find(t => t.name === toolName);
  if (exactMatch) {
    return exactMatch.description || '';
  }
  
  // 如果没有精确匹配，检查工具名是否包含已知工具名作为前缀
  for (const tool of configStore.availableTools) {
    if (toolName.startsWith(tool.name) && tool.name.length > 3) {
      return tool.description || '';
    }
  }
  
  // 如果上述匹配都失败，查找部分匹配的工具
  const matchingTools = configStore.availableTools.filter(t => 
    toolName.includes(t.name) || t.name.includes(toolName)
  );
  
  if (matchingTools.length > 0) {
    // 按工具名长度排序，优先使用最长的匹配
    matchingTools.sort((a, b) => b.name.length - a.name.length);
    return matchingTools[0].description || '';
  }
  
  // 找不到匹配，返回空字符串
  return '';
}

const removeTool = async (combinedName: string) => {
  try {
    const newSelectedTools = selectedTools.value.filter(name => name !== combinedName);
    
    // 提取工具名称
    const toolNames = newSelectedTools.map(name => {
      const parts = name.split('@');
      return parts[0]; // 只取工具名部分
    });
    
    // 更新到configStore
    configStore.selectTools(toolNames);
    
    // 保存到配置
    await configStore.saveConfiguration();
    console.log('【工具】工具选择保存成功');
    
    // 更新本地UI
    selectedTools.value = newSelectedTools;
    ElMessage.success(t('config.mcp.toolRemoveSuccess'));
    
    // 重新加载最新的工具选择状态
    setTimeout(() => {
      loadSelectedToolsFromBackend();
    }, 1000);
  } catch (error) {
    console.error('【工具】删除工具失败:', error);
    ElMessage.error(t('config.mcp.toolRemoveFailed'));
  }
}

const toggleToolDescription = (combinedName: string) => {
  if (expandedTools.value.has(combinedName)) {
    expandedTools.value.delete(combinedName)
  } else {
    expandedTools.value.add(combinedName)
  }
}

const getDefaultCheckedKeys = () => {
  const checkedKeys: string[] = []
  
  // 更新工具-服务器映射关系
  updateToolServerMap()
  
  // 遍历已选择的工具
  selectedTools.value.forEach(combinedName => {
    try {
      let toolName, serverName
      
      // 处理不同格式的工具名
      if (typeof combinedName === 'string') {
        if (combinedName.includes('@')) {
          [toolName, serverName] = combinedName.split('@')
        } else {
          toolName = combinedName
          serverName = getActualToolServer(toolName)
        }
      } else if (combinedName && typeof combinedName === 'object' && 'name' in combinedName && 'server' in combinedName) {
        toolName = (combinedName as any).name
        serverName = (combinedName as any).server
      }
      
      if (toolName && serverName) {
        // 使用新的节点ID格式: tool_[工具名]_[服务器名]
        const nodeId = `tool_${toolName}_${serverName}`
        console.log(`【工具选择】预选中工具: ${toolName}, 服务器: ${serverName}, 节点ID: ${nodeId}`)
        checkedKeys.push(nodeId)
      }
    } catch (e) {
      console.error('【工具选择】处理工具选择状态失败:', e, combinedName)
    }
  })
  
  console.log('【工具选择】默认选中的节点:', checkedKeys)
  return checkedKeys
}

const handleToolTreeCheck = (data: any, checked: any) => {
  // 记录选中状态变更
  console.log('【工具选择】选中状态变更:', data, checked)
}

const confirmToolSelection = async () => {
  try {
    if (!toolTreeRef.value) {
      console.error('【工具选择】工具树引用不存在');
      ElMessage.error('工具选择失败: 工具树组件未加载');
      return;
    }
    
    const checkedNodes = toolTreeRef.value.getCheckedNodes();
    console.log('【工具选择】选中的节点详情:', checkedNodes);
    
    // 只处理工具节点，跳过服务器节点
    const toolNodes = checkedNodes.filter((node: any) => !node.isServer && node.toolName && node.serverName);
    console.log('【工具选择】过滤后的工具节点:', toolNodes);
    
    const selectedCombinedNames: string[] = [];
    const selectedToolObjects: Array<{name: string, server: string}> = [];

    toolNodes.forEach((node: any) => {
      const toolName = node.toolName;
      const serverName = node.serverName;
      
      if (toolName && serverName) {
        console.log(`【工具选择】添加工具: ${toolName}, 服务器: ${serverName}`);
        
        // 使用combinedName属性，如果存在的话
        const combinedName = node.combinedName || `${toolName}@${serverName}`;
        selectedCombinedNames.push(combinedName);
        
        // 保存工具名和服务器名
        selectedToolObjects.push({
          name: toolName,
          server: serverName
        });
        
        // 更新工具服务器映射
        toolServerMap.value.set(toolName, serverName);
      }
    });

    // 更新到configStore，使用MCPToolConfig格式
    configStore.config.mcp.tools = selectedToolObjects;
    
    // 保存到配置
    await configStore.saveConfiguration();
    console.log('【工具选择】工具选择保存成功');
    
    // 更新本地UI
    selectedTools.value = selectedCombinedNames;
    
    // 关闭对话框
    showAddToolDialog.value = false;
    
    if (selectedCombinedNames.length > 0) {
      ElMessage.success(t('config.mcp.toolsSelected', { count: selectedCombinedNames.length }));
    } else {
      ElMessage.info(t('config.mcp.noToolsSelectedInfo'));
    }
    
    // 重新加载最新的工具选择状态
    setTimeout(() => {
      loadSelectedToolsFromBackend();
    }, 1000);
  } catch (error) {
    console.error('【工具选择】确认选择时出错:', error);
    ElMessage.error(t('config.mcp.toolSelectionFailed', {
      message: error instanceof Error ? error.message : t('error.unknown')
    }));
  }
}

const cancelToolSelection = () => {
  showAddToolDialog.value = false
}

// 打开添加工具对话框
const openAddToolDialog = async () => {
  try {
    console.log('【工具】开始加载MCP配置数据...', new Date().toISOString())
    console.log('【工具】当前可用工具:', JSON.stringify(configStore.availableTools))
    
    // 确保数据已加载
    await configStore.loadMCPServerConfigs()
    console.log('【工具】MCP服务器配置加载完成，数量:', configStore.mcpServerConfigs.length, new Date().toISOString())

    // 加载可用工具列表，但不更新用户的选择
    console.log('【工具】准备加载工具列表...', new Date().toISOString())
    try {
      await configStore.loadToolsFromDatabase()
      console.log('【工具】工具列表加载完成后的状态:')
      console.log('【工具】工具数量:', configStore.availableTools.length, new Date().toISOString())
      console.log('【工具】工具列表内容:', JSON.stringify(configStore.availableTools))
      console.log('【工具】当前选择的工具:', configStore.selectedTools, new Date().toISOString())
      
      // 更新工具-服务器映射关系
      updateToolServerMap()
      
      // 无论是否有工具，都显示工具选择对话框
      setTimeout(() => {
        // 延迟打开对话框，确保数据已经正确更新
        showAddToolDialog.value = true
        
        // 只在日志中提示，不再弹出警告消息
        if (configStore.availableTools.length === 0) {
          console.warn('【工具】没有找到可用工具，可能需要先添加MCP服务器配置')
        }
      }, 100)
      
    } catch (toolError) {
      console.error('【工具】加载工具列表失败:', toolError)
      ElMessage.error(`加载工具列表失败: ${toolError instanceof Error ? toolError.message : '未知错误'}`)
    }
  } catch (error) {
    console.error('【工具】加载MCP服务器配置失败:', error)
    ElMessage.error(`加载MCP配置失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 在工具选择弹窗中显示添加服务器对话框
const showAddServerInToolDialog = () => {
  showAddServerDialog.value = true
}

// 在工具选择弹窗中编辑服务器
const editServerInToolDialog = (serverName: string) => {
  editServer(serverName)
}

// 在工具选择弹窗中删除服务器
const removeServerInToolDialog = async (serverName: string) => {
  await removeServer(serverName)
}

// 加载所有服务器的工具列表
const loadAllServerTools = async () => {
  console.log('开始加载所有服务器的工具列表...')
  console.log('当前MCP服务器数量:', Object.keys(configStore.mcpServers).length)

  if (Object.keys(configStore.mcpServers).length === 0) {
    console.log('没有MCP服务器，跳过工具加载')
    return
  }

  try {
    // 获取所有MCP服务器配置
    const allMCPServers: Record<string, any> = {}
    Object.entries(configStore.mcpServers).forEach(([name, serverConfig]) => {
      allMCPServers[name] = {
        command: serverConfig.command,
        args: serverConfig.args || [],
        env: serverConfig.env || {}
      }
    })

    console.log('准备发送的服务器配置:', allMCPServers)

    // 调用后端API获取所有服务器的工具列表
    const response = await mcpApi.getTools(allMCPServers)
    console.log('API响应:', response)

    if (response.success && response.tools) {
      // 更新store中的可用工具列表
      configStore.updateAvailableTools(response.tools)
      console.log(`已加载 ${response.tools.length} 个工具:`, response.tools)
    } else {
      console.error('API调用失败:', response.message || '未知错误')
    }
  } catch (error) {
    console.error('加载工具列表失败:', error)
  }
}

// 强制刷新工具树
const refreshTree = () => {
  console.log('【工具树】手动强制刷新工具树', new Date().toISOString())
  console.log('【工具树】当前可用工具:', JSON.stringify(configStore.availableTools))
  
  // 使用nextTick确保DOM更新
  nextTick(() => {
    // 如果有可用工具但工具树为空，尝试重新加载
    if (configStore.availableTools && configStore.availableTools.length > 0 && toolTreeData.value.length === 0) {
      // 显示通知
      ElMessage.info('正在刷新工具树...')
      
      // 如果工具树为空但availableTools不为空，可能是计算属性没有正确触发更新
      // 我们可以尝试通过简单地重新获取工具数据来触发更新
      configStore.loadToolsFromDatabase().then(() => {
        console.log('【工具树】重新加载工具数据完成')
        console.log('【工具树】刷新后的工具数量:', configStore.availableTools.length)
        
        // 手动检查是否有工具显示
        if (toolTreeData.value.length === 0) {
          console.log('【工具树】警告: 刷新后工具树仍然为空')
          ElMessage.warning('刷新后工具树仍然为空，请检查控制台日志')
        } else {
          ElMessage.success('工具树刷新成功')
        }
      }).catch(err => {
        console.error('【工具树】刷新工具数据失败:', err)
        ElMessage.error('刷新工具数据失败')
      })
    }
  })
}

// 声明工具映射关系缓存
const toolServerMap = ref<Map<string, string>>(new Map())

// 更新工具和服务器的映射关系
const updateToolServerMap = () => {
  // 先清空缓存
  toolServerMap.value.clear()
  
  // 从树中获取所有工具和对应的服务器
  configStore.availableTools.forEach(tool => {
    if (tool.name && tool.server) {
      toolServerMap.value.set(tool.name, tool.server)
    }
  })
  
  console.log('【工具映射】更新工具-服务器映射关系:', Object.fromEntries(toolServerMap.value))
}

// 获取工具的真实服务器名称
const getActualToolServer = (toolName: string): string => {
  // 如果工具名中已包含服务器信息，直接解析
  if (toolName.includes('@')) {
    const parts = toolName.split('@')
    return parts[1] || ''
  }
  
  // 从映射缓存中查询
  return toolServerMap.value.get(toolName) || ''
}

// 组件挂载时更新工具映射并加载已选择的工具
onMounted(() => {
  updateToolServerMap()
  loadSelectedToolsFromBackend()
})

// 组件挂载时不需要重复加载数据，因为App.vue的loadConfiguration()已经加载过了
</script>

<style scoped>
.config-card {
  border: 1px solid var(--border-color);
  background-color: var(--bg-color);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.mcp-config {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.tools-section {
  margin-top: 6px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.section-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.count-tag {
  margin-left: 4px;
  transform: scale(0.85);
  transform-origin: left center;
}

.selected-tools-list {
  margin-top: 4px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  max-height: 160px;
  overflow-y: auto;
}

.selected-tool-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 4px 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.selected-tool-item:last-child {
  border-bottom: none;
}

.tool-info {
  flex: 1;
  overflow: hidden;
  padding-right: 6px;
}

.tool-name {
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  color: var(--el-color-primary);
}

.expand-icon {
  margin-left: 4px;
  font-size: 10px;
  transition: transform 0.2s;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.tool-server {
  font-size: 11px;
  color: var(--text-color-secondary);
  margin-top: 2px;
}

.tool-description-content {
  font-size: 11px;
  color: var(--text-color-regular);
  margin-top: 3px;
  padding: 3px 0;
  border-top: 1px dashed var(--el-border-color-lighter);
  word-break: break-word;
}

.delete-btn {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  padding: 0;
  font-size: 10px;
}

.delete-btn :deep(.el-icon) {
  font-size: 10px;
}

/* 工具选择对话框 */
.tool-selection {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tool-selection-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.selection-tip {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.tool-tree {
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  padding: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.tree-node {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  width: 100%;
  min-height: 24px;
}

.node-content {
  flex: 1;
}

.node-label {
  font-size: 12px;
}

.node-description {
  font-size: 11px;
  color: var(--text-color-secondary);
  margin-top: 2px;
}

.server-actions {
  display: flex;
  gap: 4px;
}

.empty-message {
  font-size: 12px;
  color: var(--text-color-secondary);
  text-align: left;
}

/* 环境变量表单 */
.env-vars {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.env-var-item {
  display: flex;
  gap: 6px;
  align-items: center;
}

/* 参数输入 */
.arg-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

.arg-input {
  width: 100px;
  margin-bottom: 4px;
}

/* 头部输入 */
.header-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

.header-input {
  width: 150px;
  margin-bottom: 4px;
}

:deep(.el-empty__image) {
  width: 50px !important;
  height: 50px !important;
}

:deep(.el-empty__description) {
  font-size: 12px;
  margin-top: 4px;
}

:deep(.el-dialog__body) {
  padding: 10px;
}

:deep(.el-dialog__header) {
  padding: 10px;
  margin-right: 0;
}

:deep(.el-dialog__footer) {
  padding: 10px;
}

@media (max-width: 500px) {
  .selected-tool-item {
    padding: 3px 5px;
  }
  
  .tool-name, .section-title {
    font-size: 12px;
  }
  
  .tool-server, .tool-description-content {
    font-size: 10px;
  }
  
  .delete-btn {
    width: 18px;
    height: 18px;
  }
  
  .delete-btn :deep(.el-icon) {
    font-size: 9px;
  }
}

.status-badge :deep(.el-badge__content) {
  font-size: 10px;
  padding: 2px 6px;
  min-width: auto;
  height: auto;
  line-height: 1;
}

/* 解决对话框层级问题 */
:deep(.server-dialog) {
  z-index: 3000 !important;
}

/* 确保弹窗内容不被遮挡 */
.el-dialog__wrapper {
  overflow-y: auto;
}

.el-dialog {
  position: relative;
  margin-top: 5vh;
  margin-bottom: 5vh;
}

/* 树形组件样式优化 */
.tool-tree :deep(.el-tree-node__content) {
  height: auto;
  min-height: 40px;
  padding: 0;
}

.tool-tree :deep(.el-tree-node__expand-icon) {
  padding: 6px;
  margin-right: 8px;
}

.tool-tree :deep(.el-checkbox) {
  margin-right: 12px;
}

.tool-tree :deep(.el-tree-node__label) {
  flex: 1;
  padding: 0;
}

/* 防止按钮文字被截断 */
.server-actions .el-button {
  white-space: nowrap;
  flex-shrink: 0;
}

/* 环境变量输入框样式 */
.env-var-item .el-input {
  flex: 1;
  min-width: 120px;
}

.debug-info {
  margin-top: 12px;
  padding: 8px;
  background-color: #f8f8f8;
  border-radius: 4px;
  font-size: 12px;
  font-family: monospace;
}

.debug-info p {
  font-weight: bold;
  margin-bottom: 4px;
}

.debug-info ul {
  margin: 0;
  padding-left: 16px;
}

/* inner服务器样式 */
.inner-server-node {
  background-color: rgba(var(--el-color-primary-rgb), 0.1);
  border-radius: 4px;
  padding: 4px;
  border-left: 3px solid var(--el-color-primary);
}

.inner-server-node .node-label {
  font-weight: bold;
  color: var(--el-color-primary);
}

.server-tag {
  font-size: 10px;
  color: var(--el-color-primary);
  background-color: rgba(var(--el-color-primary-rgb), 0.1);
  border-radius: 4px;
  padding: 1px 4px;
  margin-left: 4px;
  display: inline-block;
  border: 1px solid rgba(var(--el-color-primary-rgb), 0.2);
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
</style>
