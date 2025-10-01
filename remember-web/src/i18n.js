// 国际化配置文件
export const translations = {
  en: {
    // 通用
    appTitle: "AI Conversation Assistant",
    farshoreAI: "Farshore AI",
    contactUs: "Contact Us",
    share: "Share",
    
    // 用户管理
    selectUser: "Select User",
    switchUser: "Switch User",
    registerNewUser: "Register New User",
    existingUsers: "Existing Users",
    current: "Current",
    select: "Select",
    userID: "User ID",
    session: "Session",
    clearUser: "Clear User",
    
    // 角色设定
    roleSettings: "Role Settings",
    firstMessage: "First Message",
    rolePrompt: "Role Prompt",
    resetToDefault: "Reset to Default",
    saveSettings: "Save Settings",
    
    // 会话管理
    clearSession: "Clear Session",
    clearSessionConfirm: "Clear Session Confirmation",
    deleteOperation: "Delete Operation",
    clearSessionWarning: "Clearing the session will delete all of the following data:",
    clearSessionItems: [
      "All conversation message records",
      "User profile information",
      "Topic summary data",
      "Key event timeline"
    ],
    confirmClear: "Are you sure you want to clear the current session?",
    cancel: "Cancel",
    confirm: "Confirm",
    clearing: "Clearing...",
    
    // 记忆面板
    userMemory: "User Memory",
    currentUser: "Current User:",
    userPortrait: "User Portrait",
    topicSummary: "Topic Summary",
    eventTimeline: "Event Timeline",
    basicInformation: "Basic Information",
    interests: "Interests",
    sexualOrientation: "Sexual Orientation",
    needs: "Needs",
    noData: "No data available",
    completed: "Completed",
    todo: "To Do",
    createdAt: "Created at",
    
    // 聊天界面
    startConversation: "Start Conversation",
    enterMessage: "Enter message...",
    pleaseSelectUser: "Please select a user first",
    send: "Send",
    sending: "Sending...",
    
    // 变量面板
    roleVariables: "Role Variables",
    completedStatus: "✅ Completed",
    incompleteStatus: "⚠️ Incomplete",
    noVariables: "No variables",
    variableHelp1: "💡 After filling in variables, the panel won't close automatically, you can modify anytime",
    variableHelp2: "💡 Click outside the panel or close button to hide it",
    noVariablesMessage: "No variables need to be filled in current role settings",
    variableHint: "If you need to use variables, use {{variable_name}} format in role settings",
    
    // Toast消息
    sessionCleared: "✅ Session cleared successfully",
    clearFailed: "❌ Failed to clear session",
    networkError: "❌ Failed to clear session, please check network connection",
    
    // 空状态
    emptyStateTitle: "Start Conversation",
    emptyStateMessage: "Enter message to continue conversation",
    emptyStateNoUser: "Please select a user to start conversation"
  },
  
  zh: {
    // 通用
    appTitle: "AI对话助手",
    farshoreAI: "Farshore AI",
    contactUs: "联系我们",
    share: "分享",
    
    // 用户管理
    selectUser: "选择用户",
    switchUser: "切换用户",
    registerNewUser: "注册新用户",
    existingUsers: "已注册用户",
    current: "当前",
    select: "选择",
    userID: "用户ID",
    session: "会话",
    clearUser: "清除用户",
    
    // 角色设定
    roleSettings: "角色设定",
    firstMessage: "第一句话",
    rolePrompt: "角色提示词",
    resetToDefault: "重置为默认",
    saveSettings: "保存设定",
    
    // 会话管理
    clearSession: "清空会话",
    clearSessionConfirm: "清空会话确认",
    deleteOperation: "删除操作",
    clearSessionWarning: "清空会话将删除以下所有数据：",
    clearSessionItems: [
      "所有对话消息记录",
      "用户画像信息",
      "主题归纳数据",
      "关键事件时间线"
    ],
    confirmClear: "确定要清空当前会话吗？",
    cancel: "取消",
    confirm: "确认",
    clearing: "清空中...",
    
    // 记忆面板
    userMemory: "用户记忆",
    currentUser: "当前用户:",
    userPortrait: "用户画像",
    topicSummary: "主题归纳",
    eventTimeline: "事件时间线",
    basicInformation: "基本信息",
    interests: "兴趣爱好",
    sexualOrientation: "性取向",
    needs: "需求",
    noData: "暂无数据",
    completed: "已完成",
    todo: "待办事项",
    createdAt: "创建时间",
    
    // 聊天界面
    startConversation: "开始对话",
    enterMessage: "输入消息...",
    pleaseSelectUser: "请先选择用户",
    send: "发送",
    sending: "发送中...",
    
    // 变量面板
    roleVariables: "角色设定变量",
    completedStatus: "✅ 已完成",
    incompleteStatus: "⚠️ 待完成",
    noVariables: "无变量",
    variableHelp1: "💡 变量填写完成后，面板不会自动关闭，您可以随时修改",
    variableHelp2: "💡 点击面板外部或关闭按钮可以隐藏面板",
    noVariablesMessage: "当前角色设定中没有需要填写的变量",
    variableHint: "如果需要使用变量，请在角色设定中使用 {{变量名}} 格式",
    
    // Toast消息
    sessionCleared: "✅ 会话已成功清空",
    clearFailed: "❌ 清空会话失败",
    networkError: "❌ 清空会话失败，请检查网络连接",
    
    // 空状态
    emptyStateTitle: "开始对话",
    emptyStateMessage: "输入消息继续对话",
    emptyStateNoUser: "请先选择用户开始对话"
  }
};

// 语言检测和存储
export const getStoredLanguage = () => {
  return localStorage.getItem('language') || 'zh';
};

export const setStoredLanguage = (lang) => {
  localStorage.setItem('language', lang);
};

// 获取当前语言
export const getCurrentLanguage = () => {
  return getStoredLanguage();
};

// 获取翻译文本
export const t = (key, lang = null) => {
  const currentLang = lang || getCurrentLanguage();
  const keys = key.split('.');
  let value = translations[currentLang];
  
  for (const k of keys) {
    if (value && value[k] !== undefined) {
      value = value[k];
    } else {
      console.warn(`Translation key not found: ${key} for language ${currentLang}`);
      return key; // 返回键名作为降级
    }
  }
  
  return value;
};
