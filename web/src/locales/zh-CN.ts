export default {
  // 通用
  common: {
    confirm: '确认',
    cancel: '取消',
    save: '保存',
    delete: '删除',
    edit: '编辑',
    add: '添加',
    close: '关闭',
    loading: '加载中...',
    success: '成功',
    error: '错误',
    warning: '警告',
    info: '信息',
    copy: '复制',
    export: '导出',
    import: '导入',
    reset: '重置',
    clear: '清空',
    search: '搜索',
    refresh: '刷新',
    connect: '连接',
    disconnect: '断开连接',
    connected: '已连接',
    disconnected: '已断开',
    connecting: '连接中...',
    default: '默认',
    description: '描述',
    noDescription: '无描述',
    confirmAction: '确认操作',
    raw: '原始文本',
    markdown: 'Markdown'
  },

  // 应用标题
  app: {
    title: 'MCP Agent Web UI',
    subtitle: 'Model Context Protocol Agent 交互管理界面'
  },

  // 欢迎屏幕
  welcome: {
    title: '欢迎使用 MCP Agent',
    description: '请在下方输入您的任务描述，我将帮助您完成各种信息收集和分析任务。',
    exampleTitle: '示例任务',
    example1: '分析网络安全领域的最新研究趋势',
    example2: '搜索人工智能在医疗领域的应用',
    example3: '获取最新的科技新闻摘要',
    example4: '分析某个网站的技术架构'
  },

  // 主题和语言
  theme: {
    light: '浅色主题',
    dark: '深色主题',
    auto: '跟随系统',
    toggle: '切换主题'
  },

  language: {
    'zh-CN': '简体中文',
    'en-US': 'English',
    toggle: '切换语言'
  },

  // 配置面板
  config: {
    title: '配置管理',
    llm: {
      title: 'LLM配置',
      type: '模型类型',
      baseUrl: 'Base URL',
      model: '模型名称',
      apiKey: 'API Key',
      temperature: '温度',
      maxTokens: '最大Token数',
      status: '连接状态'
    },
    mcp: {
      title: 'MCP服务器',
      servers: '服务器列表',
      tools: '工具选择',
      selectedTools: '已选择的工具',
      addTool: '添加工具',
      configFile: '配置文件',
      addServer: '添加服务器',
      editServer: '编辑服务器',
      serverName: '名称',
      serverNamePlaceholder: '服务器名称',
      command: '命令',
      commandPlaceholder: '启动命令',
      args: '参数',
      addArg: '添加参数',
      env: '环境变量',
      varName: '变量名',
      varValue: '变量值',
      addEnvVar: '添加环境变量',
      status: '状态',
      selectTools: '选择工具',
      selectionTip: '选择需要的工具，按服务器分组显示',
      noToolsSelected: '暂无选择的工具',
      noServers: '暂无MCP服务器，请先添加服务器',
      emptyToolList: '当前工具列表为空。可能的原因:',
      emptyReason1: '没有配置MCP服务器',
      emptyReason2: 'MCP服务器没有提供工具',
      emptyReason3: '工具加载失败',
      debugInfo: '调试信息',
      debugTools: '可用工具数量',
      debugServers: 'MCP服务器配置数量',
      debugApiResponse: 'API响应工具数据',
      debugRefreshTime: '组件刷新时间',
      hasData: '有数据',
      noData: '无数据',
      forceRefresh: '强制刷新工具树',
      serverCountHint: '已选工具涉及 {active} 个服务器 / 总共 {total} 个活跃服务器',
      serverNameRequired: '请输入服务器名称',
      commandRequired: '请输入启动命令',
      deleteServerConfirm: '确定要删除服务器 "{name}" 吗？',
      deleteConfirmTitle: '确认删除',
      serverDeleteSuccess: '服务器删除成功',
      serverUpdateSuccess: '服务器配置更新成功',
      serverCreateSuccess: '服务器配置创建成功',
      saveServerConfigFailed: '保存服务器配置失败: {message}',
      toolsSelected: '已选择 {count} 个工具',
      noToolsSelectedInfo: '未选择任何工具',
      toolSelectionFailed: '工具选择失败: {message}',
      toolRemoveSuccess: '工具删除成功',
      internalServer: '内置服务器',
      cannotEditInnerServer: '内置服务器不能编辑',
      cannotDeleteInnerServer: '内置服务器不能删除'
    },
    prompt: {
      title: '系统提示词',
      template: '模板选择',
      custom: '自定义',
      placeholders: '占位符配置',
      preview: '预览',
      createNew: '新建配置',
      systemPrompt: '系统提示词',
      selectConfig: '选择系统提示词配置',
      savedConfigs: '保存的配置',
      presetTemplates: '预设模板',
      collapseDetails: '收起详情',
      expandDetails: '展开详情',
      configName: '配置名称',
      noPlaceholders: '暂无占位符',
      content: '提示词内容',
      contentPlaceholder: '请选择或创建系统提示词',
      hidePreview: '隐藏预览',
      showPreview: '显示预览',
      selectOrCreate: '请选择或创建系统提示词配置',
      createConfig: '创建配置',
      deleteConfirmText: '确定要删除系统提示词配置 "{name}" 吗？此操作不可恢复。',
      deleteConfirmTitle: '删除确认',
      deleteSuccess: '系统提示词配置删除成功',
      deleteFailed: '删除失败，请重试',
      contentRequired: '提示词内容不能为空',
      newConfig: '新建配置',
      templateCreatedDesc: '从模板创建的配置',
      fillNameAndSave: '请为此配置填写名称并保存',
      defaultField: '网络安全研究领域'
    },
    other: {
      title: '其他配置',
      proxy: '网络代理',
      logLevel: '日志级别',
      maxStep: '最大步数',
      advanced: {
        title: '高级设置',
        requestTimeout: '请求超时',
        retryCount: '重试次数',
        concurrencyLimit: '并发限制',
        debugMode: '调试模式',
        chatHistory: '聊天记录',
        clearChatHistory: '清空聊天记录',
        enable: '开启',
        disable: '关闭',
        seconds: '秒'
      }
    },
    actions: {
      save: '保存配置',
      export: '导出配置',
      import: '导入配置'
    }
  },

  // 聊天区域
  chat: {
    title: '聊天交互',
    input: {
      placeholder: '请输入您的任务描述...',
      send: '发送',
      stop: '停止',
      clear: '清空对话',
      history: '历史记录',
      filePrompt: '请分析以下文件内容：\n\n文件名：{filename}\n内容：\n{content}'
    },
    message: {
      user: '用户',
      assistant: '助手',
      system: '系统',
      timestamp: '时间戳',
      copy: '复制消息',
      thinking: 'AI 正在思考...',
      toolCall: '工具调用',
      result: '结果',
      error: '错误'
    },
    status: {
      typing: '正在输入...',
      processing: '处理中...'
    }
  },

  // 工具调用显示
  tool: {
    title: '工具调用',
    name: '工具名称',
    parameters: '参数',
    result: '结果',
    status: {
      calling: '调用中',
      success: '成功',
      error: '失败'
    },
    expand: '展开详情',
    collapse: '收起详情'
  },

  // 错误消息
  error: {
    networkError: '网络连接错误',
    configInvalid: '配置无效',
    saveConfigFailed: '保存配置失败',
    loadConfigFailed: '加载配置失败',
    sseError: 'SSE连接错误',
    sendMessageFailed: '发送消息失败',
    fileUploadFailed: '文件上传失败',
    importConfigFailed: '导入配置失败',
    exportConfigFailed: '导出配置失败',
    fileTooLarge: '文件大小不能超过{size}',
    stopTaskFailed: '停止任务失败: {message}',
    unknown: '未知错误',
    clearChatHistoryFailed: '清空聊天记录失败'
  },

  // 成功消息
  success: {
    configSaved: '配置保存成功',
    configImported: '配置导入成功',
    configExported: '配置导出成功',
    messageSent: '消息发送成功',
    connected: '连接成功',
    chatCleared: '对话已清空',
    chatHistoryCleared: '聊天记录已清空'
  },

  // 确认对话框
  confirm: {
    deleteServer: '确定要删除这个服务器吗？',
    clearChat: '确定要清空所有对话吗？',
    disconnect: '确定要断开连接吗？'
  },

  // 任务执行
  task: {
    running: '任务执行中',
    processing: '处理中...',
    steps: '{total} 步骤',
    completed: '任务已完成',
    failed: '任务失败',
    stopRequest: '已发送停止任务请求'
  },

  // 验证消息
  validation: {
    required: '此字段为必填项',
    invalidUrl: '无效的URL格式',
    invalidNumber: '请输入有效的数字',
    minLength: '长度不能少于{min}个字符',
    maxLength: '长度不能超过{max}个字符'
  }
}
