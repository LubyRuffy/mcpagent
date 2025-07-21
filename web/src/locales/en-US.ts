export default {
  // Common
  common: {
    confirm: 'Confirm',
    cancel: 'Cancel',
    save: 'Save',
    delete: 'Delete',
    edit: 'Edit',
    add: 'Add',
    close: 'Close',
    loading: 'Loading...',
    success: 'Success',
    error: 'Error',
    warning: 'Warning',
    info: 'Info',
    copy: 'Copy',
    export: 'Export',
    import: 'Import',
    reset: 'Reset',
    clear: 'Clear',
    search: 'Search',
    refresh: 'Refresh',
    connect: 'Connect',
    disconnect: 'Disconnect',
    connected: 'Connected',
    disconnected: 'Disconnected',
    connecting: 'Connecting...',
    default: 'Default',
    description: 'Description',
    noDescription: 'No description',
    confirmAction: 'Confirm Action',
    raw: 'Raw Text',
    markdown: 'Markdown'
  },

  // App title
  app: {
    title: 'MCP Agent Web UI',
    subtitle: 'Model Context Protocol Agent Management Interface'
  },

  // Welcome screen
  welcome: {
    title: 'Welcome to MCP Agent',
    description: 'Please enter your task description below, I will help you complete various information collection and analysis tasks.',
    exampleTitle: 'Example Tasks',
    example1: 'Analyze the latest research trends in network security',
    example2: 'Search for applications of AI in the medical field',
    example3: 'Get a summary of the latest tech news',
    example4: 'Analyze the technical architecture of a website'
  },

  // Theme and language
  theme: {
    light: 'Light Theme',
    dark: 'Dark Theme',
    auto: 'Follow System',
    toggle: 'Toggle Theme'
  },

  language: {
    'zh-CN': '简体中文',
    'en-US': 'English',
    toggle: 'Toggle Language'
  },

  // Configuration panel
  config: {
    title: 'Configuration',
    llm: {
      title: 'LLM Configuration',
      type: 'Model Type',
      baseUrl: 'Base URL',
      model: 'Model Name',
      apiKey: 'API Key',
      temperature: 'Temperature',
      maxTokens: 'Max Tokens',
      status: 'Connection Status'
    },
    mcp: {
      title: 'MCP Server',
      servers: 'Server List',
      tools: 'Tool Selection',
      selectedTools: 'Selected Tools',
      addTool: 'Add Tool',
      configFile: 'Config File',
      addServer: 'Add Server',
      editServer: 'Edit Server',
      serverName: 'Server Name',
      serverNamePlaceholder: 'Server Name',
      transportType: 'Transport Type',
      transportTypePlaceholder: 'Select Transport Type',
      transportStdio: 'STDIO (Command Line)',
      transportSSE: 'SSE (Server-Sent Events)',
      transportHTTP: 'HTTP (Network Request)',
      command: 'Command',
      commandPlaceholder: 'Start Command',
      args: 'Arguments',
      addArg: 'Add Argument',
      env: 'Environment Variables',
      varName: 'Variable Name',
      varValue: 'Variable Value',
      addEnvVar: 'Add Environment Variable',
      url: 'Server URL',
      urlPlaceholder: 'e.g. http://localhost:8000/sse',
      headers: 'HTTP Headers',
      addHeader: 'Add Header',
      headerPlaceholder: 'e.g. Authorization: Bearer token',
      status: 'Status',
      selectTools: 'Select Tools',
      selectionTip: 'Select needed tools, grouped by server',
      noToolsSelected: 'No tools selected',
      noServers: 'No MCP servers, please add a server first',
      emptyToolList: 'The tool list is currently empty. Possible reasons:',
      emptyReason1: 'No MCP server is configured',
      emptyReason2: 'MCP server does not provide tools',
      emptyReason3: 'Failed to load tools',
      debugInfo: 'Debug Information',
      debugTools: 'Available Tools',
      debugServers: 'MCP Server Configurations',
      debugApiResponse: 'API Response Tool Data',
      debugRefreshTime: 'Component Refresh Time',
      hasData: 'Has Data',
      noData: 'No Data',
      forceRefresh: 'Force Refresh Tool Tree',
      serverCountHint: 'Selected tools involve {active} servers / Total {total} active servers',
      serverNameRequired: 'Please input server name',
      commandRequired: 'Please input start command',
      transportTypeRequired: 'Please select transport type',
      urlRequired: 'Please input server URL',
      urlInvalid: 'Please input a valid HTTP(S) URL',
      deleteServerConfirm: 'Are you sure you want to delete server "{name}"?',
      deleteConfirmTitle: 'Confirm Deletion',
      serverDeleteSuccess: 'Server deleted successfully',
      serverUpdateSuccess: 'Server configuration updated successfully',
      serverCreateSuccess: 'Server configuration created successfully',
      saveServerConfigFailed: 'Failed to save server configuration: {message}',
      toolsSelected: '{count} tools selected',
      noToolsSelectedInfo: 'No tools selected',
      toolSelectionFailed: 'Tool selection failed: {message}',
      toolRemoveSuccess: 'Tool removed successfully',
      internalServer: 'Internal Server',
      cannotEditInnerServer: 'Internal server cannot be edited',
      cannotDeleteInnerServer: 'Internal server cannot be deleted'
    },
    prompt: {
      title: 'System Prompt',
      template: 'Template Selection',
      custom: 'Custom',
      placeholders: 'Placeholders',
      preview: 'Preview',
      createNew: 'Create New',
      systemPrompt: 'System Prompt',
      selectConfig: 'Select System Prompt Configuration',
      savedConfigs: 'Saved Configurations',
      presetTemplates: 'Preset Templates',
      collapseDetails: 'Collapse Details',
      expandDetails: 'Expand Details',
      configName: 'Configuration Name',
      noPlaceholders: 'No Placeholders',
      content: 'Prompt Content',
      contentPlaceholder: 'Please select or create a system prompt',
      hidePreview: 'Hide Preview',
      showPreview: 'Show Preview',
      selectOrCreate: 'Please select or create a system prompt configuration',
      createConfig: 'Create Configuration',
      deleteConfirmText: 'Are you sure you want to delete the system prompt configuration "{name}"? This action cannot be undone.',
      deleteConfirmTitle: 'Delete Confirmation',
      deleteSuccess: 'System prompt configuration deleted successfully',
      deleteFailed: 'Delete failed, please try again',
      contentRequired: 'Prompt content cannot be empty',
      newConfig: 'New Configuration',
      templateCreatedDesc: 'Configuration created from template',
      fillNameAndSave: 'Please fill in a name for this configuration and save',
      defaultField: 'Network Security Research Field'
    },
    other: {
      title: 'Other Configuration',
      proxy: 'Network Proxy',
      logLevel: 'Log Level',
      maxStep: 'Max Steps',
      advanced: {
        title: 'Advanced Settings',
        requestTimeout: 'Request Timeout',
        retryCount: 'Retry Count',
        concurrencyLimit: 'Concurrency Limit',
        debugMode: 'Debug Mode',
        chatHistory: 'Chat History',
        clearChatHistory: 'Clear Chat History',
        enable: 'Enable',
        disable: 'Disable',
        seconds: 'seconds'
      }
    },
    actions: {
      save: 'Save Configuration',
      export: 'Export Configuration',
      import: 'Import Configuration'
    }
  },

  // Chat area
  chat: {
    title: 'Chat Interaction',
    input: {
      placeholder: 'Please enter your task description...',
      send: 'Send',
      stop: 'Stop',
      clear: 'Clear Chat',
      history: 'History',
      filePrompt: 'Please analyze the following file content:\n\nFilename: {filename}\nContent:\n{content}'
    },
    message: {
      user: 'User',
      assistant: 'Assistant',
      system: 'System',
      timestamp: 'Timestamp',
      copy: 'Copy Message',
      thinking: 'AI is thinking...',
      toolCall: 'Tool Call',
      result: 'Result',
      error: 'Error'
    },
    status: {
      typing: 'Typing...',
      processing: 'Processing...'
    }
  },

  // Tool call display
  tool: {
    title: 'Tool Call',
    name: 'Tool Name',
    parameters: 'Parameters',
    result: 'Result',
    status: {
      calling: 'Calling',
      success: 'Success',
      error: 'Failed'
    },
    expand: 'Expand Details',
    collapse: 'Collapse Details'
  },

  // Error messages
  error: {
    networkError: 'Network connection error',
    configInvalid: 'Invalid configuration',
    saveConfigFailed: 'Failed to save configuration',
    loadConfigFailed: 'Failed to load configuration',
    sseError: 'SSE connection error',
    sendMessageFailed: 'Failed to send message',
    fileUploadFailed: 'File upload failed',
    importConfigFailed: 'Failed to import configuration',
    exportConfigFailed: 'Failed to export configuration',
    fileTooLarge: 'File size cannot exceed {size}',
    stopTaskFailed: 'Failed to stop task: {message}',
    unknown: 'Unknown error',
    clearChatHistoryFailed: 'Failed to clear chat history'
  },

  // Success messages
  success: {
    configSaved: 'Configuration saved successfully',
    configImported: 'Configuration imported successfully',
    configExported: 'Configuration exported successfully',
    messageSent: 'Message sent successfully',
    connected: 'Connected successfully',
    chatCleared: 'Conversation cleared',
    chatHistoryCleared: 'Chat history cleared successfully'
  },

  // Confirmation dialogs
  confirm: {
    deleteServer: 'Are you sure you want to delete this server?',
    clearChat: 'Are you sure you want to clear all conversations?',
    disconnect: 'Are you sure you want to disconnect?'
  },

  // Task execution
  task: {
    running: 'Task in progress',
    processing: 'Processing...',
    steps: '{total} steps',
    completed: 'Task completed',
    failed: 'Task failed',
    stopRequest: 'Task stop request sent'
  },

  // Validation messages
  validation: {
    required: 'This field is required',
    invalidUrl: 'Invalid URL format',
    invalidNumber: 'Please enter a valid number',
    minLength: 'Length cannot be less than {min} characters',
    maxLength: 'Length cannot exceed {max} characters'
  }
}
