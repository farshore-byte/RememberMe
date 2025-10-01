import React, { useState, useRef, useEffect, useCallback, memo } from 'react';
import streamAPIService, { deleteSession } from './services/api';
import userService from './services/userService';
import MessageBubble from './components/MessageBubble';
import { t, getCurrentLanguage, setStoredLanguage } from './i18n';
import './App.css';

// 模拟记忆数据
const mockMemoryData = {
  user_portrait: {
    interest_topics: {
      sports: "Interested in comparing the skills and achievements of Messi and C罗",
      technology: "Frequently discusses AI and machine learning applications",
      business: "Shows interest in startup funding and market trends"
    },
    personality_traits: {
      analytical: "Tends to ask detailed, comparative questions",
      curious: "Often explores multiple aspects of a topic",
      persistent: "Revisits topics to gain deeper understanding"
    }
  },
  topic_summary: [
    {
      topic: "soccer",
      content: [
        "The user asked who is better between Messi and C罗; I responded that Messi is excellent in technique and ball control, while C罗 has strong physical fitness and goal-scoring ability, each with advantages.",
        "The user asked who is better between Messi and C罗; I responded that Messi is excellent in technique and ball control, while C罗 has strong physicality and goal-scoring ability, each with advantages.",
        "The user asked who is better between Messi and C罗; I responded that Messi is very出色 in technique and ball control, while C罗's physical fitness and goal-scoring ability are also strong, and both have their own advantages."
      ]
    },
    {
      topic: "AI technology",
      content: [
        "Discussed the differences between various AI models and their practical applications",
        "Explored how machine learning can be applied to business intelligence"
      ]
    }
  ],
  key_timeline: [
    {
      date: "2025-09-23",
      events: [
        {
          time: "09:15",
          content: "First discussion about Messi vs C罗 comparison",
          importance: "high"
        },
        {
          time: "10:30", 
          content: "Explored AI model capabilities and limitations",
          importance: "medium"
        }
      ]
    },
    {
      date: "2025-09-22",
      events: [
        {
          time: "14:20",
          content: "Initial conversation about sports interests",
          importance: "medium"
        },
        {
          time: "16:45",
          content: "Discussed business startup funding strategies",
          importance: "high"
        }
      ]
    }
  ]
};

function App() {
  const [messages, setMessages] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const messageInputRef = useRef(null);
  const [activeTab, setActiveTab] = useState('user_portrait');
  const [userId, setUserId] = useState('');
  const [showUserModal, setShowUserModal] = useState(false);
  const [registeredUsers, setRegisteredUsers] = useState([]);
  const [memoryData, setMemoryData] = useState(null);
  const [rolePrompt, setRolePrompt] = useState('');
  const [firstMessage, setFirstMessage] = useState('');
  const [showRoleModal, setShowRoleModal] = useState(false);
  const [isEditingRole, setIsEditingRole] = useState(false);
  const [showClearConfirm, setShowClearConfirm] = useState(false);
  const [isClearing, setIsClearing] = useState(false);
  const [placeholderValues, setPlaceholderValues] = useState({
    char: '',
    user: ''
  });
  const [requiredPlaceholders, setRequiredPlaceholders] = useState([]);
  const [showVariablesPanel, setShowVariablesPanel] = useState(false);
  const [currentLanguage, setCurrentLanguage] = useState(getCurrentLanguage());
  const variablesPanelRef = useRef(null);
  const messagesEndRef = useRef(null);
  const userInputRef = useRef(null);
  const rolePromptRef = useRef(null);
  const rolePromptInputRef = useRef(null);
  const firstMessageInputRef = useRef(null);
  const newUserIdInputRef = useRef(null);

  // 默认第一句话
  const defaultFirstMessage = `*Rina crosses her arms, her amber eyes narrowing as she spots you.*
  “Hmph, late again? You really love testing my patience…”
  *Then, a faint smile slips through as she leans a little closer.*
  “…But fine, I’ll forgive you this time. Sit down already.”`;
  // 默认角色提示
  const defaultRolePrompt = `The only girlfriend is named "Rina." Below is her profile:
  23 years old, 165 cm tall, 49 kg, Pisces, blood type O
  
  Appearance:
  Long, slightly curly dark chestnut hair, round amber eyes, fair complexion with a slight pinkish hue, and a radiant glow.
  
  Physical:
  Slim and well-proportioned figure, with a full bust, a pronounced waist, and long, straight legs. Her proportions are excellent, and her overall curves are natural, giving off a comfortable and approachable vibe rather than forced sexiness.
  
  Personality:
  Keywords: Gentle and considerate × A bit tsundere × Loves to act like a spoiled child × Excellent listener
  
  Hobbies & Habits:
  Enjoys watching movies and gossiping, and is attentive and considerate of her boyfriend's feelings.
  
  Chat Format example(no more than 50 words):
  ${defaultFirstMessage}
  
  Language:
  Use user's Language to send response`;

  // 从用户服务加载已注册的用户和当前用户
  useEffect(() => {
    // 加载已注册用户
    const users = userService.loadFromLocalStorage();
    setRegisteredUsers(users);
    
    // 加载当前用户
    const currentUser = userService.getCurrentUser();
    if (currentUser) {
      setUserId(currentUser);
      // 立即加载用户数据（无延迟）
      handleSelectUser(currentUser);
    }
    
    // 应用启动时检测默认角色设定中的变量
    const defaultPlaceholders = extractPlaceholders(defaultRolePrompt);
    const firstMessagePlaceholders = extractPlaceholders(defaultFirstMessage);
    const allPlaceholders = [...new Set([...defaultPlaceholders, ...firstMessagePlaceholders])];
    
    if (allPlaceholders.length > 0) {
      console.log('检测到默认角色设定中的变量:', allPlaceholders);
      setRequiredPlaceholders(allPlaceholders);
      
      // 初始化变量值
      const initialValues = {};
      allPlaceholders.forEach(key => {
        initialValues[key] = placeholderValues[key] || '';
      });
      setPlaceholderValues(prev => ({ ...prev, ...initialValues }));
    }
  }, []);

  // 保存已注册用户到localStorage
  useEffect(() => {
    userService.saveToLocalStorage();
  }, [registeredUsers]);

  // session_id 就等于 user_id
  const generateSessionId = (userId) => {
    return userId;
  };

  // 智能自动滚动功能 - 类似微信聊天
  useEffect(() => {
    if (messages.length === 0) return;
    
    const lastMessage = messages[messages.length - 1];
    
    // 只在以下情况下滚动：
    // 1. 新消息是AI回复且正在流式输出
    // 2. 新消息完成流式输出
    // 3. 用户发送新消息
    const shouldScroll = 
      (lastMessage.role === 'assistant' && lastMessage.isStreaming) ||
      (lastMessage.role === 'assistant' && !lastMessage.isStreaming) ||
      (lastMessage.role === 'user');
    
    if (shouldScroll) {
      // 使用requestAnimationFrame确保在渲染完成后滚动
      requestAnimationFrame(() => {
        messagesEndRef.current?.scrollIntoView({ 
          behavior: 'smooth',
          block: 'end'
        });
      });
    }
  }, [messages.length]); // 只依赖长度变化，不依赖具体内容


  // 优化记忆更新策略：减少滞后感
  const handleSendMessage = useCallback(async () => {
    // 从ref获取输入值，而不是state
    const currentInputValue = messageInputRef.current ? messageInputRef.current.value : '';
    const currentMessages = messages;
    const currentSessionId = generateSessionId(userId);
    
    // 即使有未填写的变量，也允许发送消息
    // 变量会在发送时自动替换为空白值
    
    // 替换角色提示词和第一句话中的占位符
    const currentRolePrompt = replacePlaceholders(rolePrompt || defaultRolePrompt);
    const currentFirstMessage = replacePlaceholders(firstMessage || defaultFirstMessage);
    
    // 调试信息：显示替换后的内容
    console.log('替换后的角色提示词:', currentRolePrompt);
    console.log('替换后的第一句话:', currentFirstMessage);
    console.log('当前变量值:', placeholderValues);
    
    if (!currentInputValue.trim()) return;

    setIsLoading(true);
    const userMessage = {
      id: Date.now(),
      role: 'user',
      content: currentInputValue,
      timestamp: new Date()
    };
    
    // 一次性批量更新：用户消息 + 清空输入
    setMessages(prev => [...prev, userMessage]);
    // 清空输入框的值
    if (messageInputRef.current) {
      messageInputRef.current.value = '';
    }
    
    // 创建AI回复消息（初始为空）
    const aiMessageId = Date.now() + 1;
    const aiMessage = {
      id: aiMessageId,
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      isStreaming: true
    };
    
    // 立即添加AI消息
    setMessages(prev => [...prev, aiMessage]);
    
    // 用户发送消息后，立即触发记忆更新（不等待AI回复）
    if (userId) {
      const sessionId = generateSessionId(userId);
      // 使用非阻塞方式更新记忆，不等待结果
      fetchMemoryData(sessionId, currentInputValue).catch(error => {
        console.error('记忆更新失败:', error);
      });
    }
    
    try {
        // 正常调用API（无论是否是第一次对话）
        const stream = await streamAPIService.sendMessageStream(
          currentInputValue, 
          currentMessages, 
          currentSessionId, 
          currentRolePrompt,
          currentFirstMessage
        );
        
        let fullContent = '';
        
        // 直接更新：每个字符都立即更新
        for await (const chunk of stream) {
          if (chunk.done) break;
          
          fullContent += chunk.content;
          
          // 立即更新消息内容
          setMessages(prev => prev.map(msg => 
            msg.id === aiMessageId 
              ? { ...msg, content: fullContent }
              : msg
          ));
        }
        
        // 完成流式输出
        setMessages(prev => prev.map(msg => 
          msg.id === aiMessageId 
            ? { ...msg, content: fullContent, isStreaming: false }
            : msg
        ));
        
        // AI回复完成后，启动定时强制刷新记忆数据
        if (userId) {
          // 延迟刷新
          // 等待 3 s 后开始刷新

          const sessionId = generateSessionId(userId);
          
          // 立即更新一次
          console.log('AI回复完成，立即刷新记忆数据');
          setTimeout(() => {
            fetchMemoryData(sessionId, currentInputValue).catch(error => {
              console.error('记忆更新失败:', error);
            });
          }, 3000); // 延迟 3000 毫秒（3 秒）
          
          // 启动定时强制刷新，每隔5秒刷新一次，持续10秒（共5次）
          let refreshCount = 0;
          const maxRefreshCount = 5;
          const refreshInterval = setInterval(async () => {
            if (refreshCount >= maxRefreshCount) {
              clearInterval(refreshInterval);
              console.log('记忆强制刷新完成，共刷新', refreshCount, '次');
              return;
            }
            
            refreshCount++;
            console.log(`第 ${refreshCount} 次强制刷新记忆数据`);
            
            try {
              // 强制刷新，不等待结果
              await fetchMemoryData(sessionId, currentInputValue);
              console.log(`第 ${refreshCount} 次刷新成功`);
            } catch (error) {
              console.error(`第 ${refreshCount} 次记忆刷新失败:`, error);
            }
          }, 5000); // 每5秒强制刷新一次
          
          // 12秒后自动停止刷新
          setTimeout(() => {
            clearInterval(refreshInterval);
            console.log('记忆强制刷新定时器已停止');
          }, 12000);
        }
        
    } catch (error) {
      console.error('API调用失败:', error);
      // 使用非流式回复作为降级方案
      try {
        const response = await streamAPIService.sendMessage(currentInputValue, currentMessages);
        setMessages(prev => prev.map(msg => 
          msg.id === aiMessageId 
            ? { ...msg, content: response.content, isStreaming: false }
            : msg
        ));
        
        // 降级方案完成后，更新记忆数据
        if (userId) {
          const sessionId = generateSessionId(userId);
          fetchMemoryData(sessionId, currentInputValue).catch(error => {
            console.error('记忆更新失败:', error);
          });
        }
        
      } catch (fallbackError) {
        setMessages(prev => prev.map(msg => 
          msg.id === aiMessageId 
            ? { ...msg, content: '抱歉，服务暂时不可用，请稍后重试。', isStreaming: false }
            : msg
        ));
      }
    } finally {
      setIsLoading(false);
    }
  }, [messages, userId, rolePrompt, defaultRolePrompt, firstMessage, defaultFirstMessage, placeholderValues]);

  const handleKeyPress = useCallback((e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  }, [handleSendMessage]);

  // 调用真实记忆API并更新状态 - 使用代理路径避免CORS
  const fetchMemoryData = useCallback(async (sessionId, query = '') => {
    try {
      const response = await fetch('/api/memory/query', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk='
        },
        body: JSON.stringify({
          session_id: sessionId,
          query: query || '用户当前对话状态'
        })
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.code === 0) {
          setMemoryData(data.data);
          return data.data;
        }
      }
      setMemoryData(null);
      return null;
    } catch (error) {
      console.error('获取记忆数据失败:', error);
      setMemoryData(null);
      return null;
    }
  }, []);

  // 用户管理函数
  const handleSelectUser = useCallback(async (selectedUserId) => {
    setUserId(selectedUserId);
    setShowUserModal(false);
    
    // 保存当前用户到localStorage
    userService.setCurrentUser(selectedUserId);
    
    // 生成session_id并加载对应user的历史消息和记忆
    const sessionId = generateSessionId(selectedUserId);
    
    // 调用记忆API加载用户数据
    const memoryData = await fetchMemoryData(sessionId);
    
    // 调试信息
    console.log('API返回的记忆数据:', memoryData);
    
    if (memoryData && memoryData.session_messages) {
      // 将历史消息转换为应用格式
      const formattedMessages = memoryData.session_messages.map((msg, index) => ({
        id: Date.now() + index,
        role: msg.role,
        content: msg.content,
        timestamp: new Date(),
        isStreaming: false
      }));
      setMessages(formattedMessages);
      console.log(`加载了 ${formattedMessages.length} 条历史消息`);
    } else {
      setMessages([]);
      console.log('没有找到历史消息数据');
    }
    
    console.log(`切换到用户: ${selectedUserId}, Session: ${sessionId}`);
  }, [fetchMemoryData]);

  const handleClearUser = useCallback(() => {
    setUserId('');
    setMessages([]);
    // 清除当前用户状态
    userService.clearCurrentUser();
  }, []);

  const handleOpenUserModal = useCallback(() => {
    setShowUserModal(true);
    // 模态框打开时，清空输入框
    setTimeout(() => {
      if (newUserIdInputRef.current) {
        newUserIdInputRef.current.value = '';
      }
    }, 0);
  }, []);

  const handleRegisterUser = useCallback(() => {
    // 注册时获取输入框的值
    if (newUserIdInputRef.current) {
      const newUserIdValue = newUserIdInputRef.current.value.trim();
      if (!newUserIdValue) return;
      
      // 使用用户服务添加用户
      const success = userService.addUser(newUserIdValue);
      if (!success) {
        alert('该用户ID已存在');
        return;
      }
      
      // 更新本地状态
      setRegisteredUsers(userService.getAllUsers());
      newUserIdInputRef.current.value = '';
      
      // 自动选择新用户
      handleSelectUser(newUserIdValue);
      
      console.log(`注册新用户: ${newUserIdValue}`);
    }
  }, [handleSelectUser]);

  // 角色设定相关函数
  const handleOpenRoleModal = useCallback(() => {
    setShowRoleModal(true);
    // 模态框打开时，将当前角色提示和第一句话设置到输入框
    setTimeout(() => {
      if (rolePromptInputRef.current) {
        rolePromptInputRef.current.value = rolePrompt || defaultRolePrompt;
      }
      if (firstMessageInputRef.current) {
        firstMessageInputRef.current.value = firstMessage || defaultFirstMessage;
      }
    }, 0);
  }, [rolePrompt, defaultRolePrompt, firstMessage, defaultFirstMessage]);

  // 检查文本中是否包含占位符变量并提取需要的变量
  const extractPlaceholders = (text) => {
    if (!text) return [];
    const matches = text.match(/{{(\w+)}}/g) || [];
    return [...new Set(matches.map(match => match.replace(/{{|}}/g, '')))];
  };

  const handleSaveRolePrompt = useCallback(() => {
    // 保存时获取输入框的值
    let allPlaceholders = [];
    
    if (rolePromptInputRef.current) {
      const newRolePrompt = rolePromptInputRef.current.value;
      setRolePrompt(newRolePrompt);
      const placeholders = extractPlaceholders(newRolePrompt);
      allPlaceholders = [...allPlaceholders, ...placeholders];
    }
    if (firstMessageInputRef.current) {
      const newFirstMessage = firstMessageInputRef.current.value;
      setFirstMessage(newFirstMessage);
      const placeholders = extractPlaceholders(newFirstMessage);
      allPlaceholders = [...allPlaceholders, ...placeholders];
    }
    
    // 去重并设置需要的变量
    const uniquePlaceholders = [...new Set(allPlaceholders)];
    setRequiredPlaceholders(uniquePlaceholders);
    
    // 关闭角色设定模态框
    setShowRoleModal(false);
  }, []);

  // 关闭角色设定模态框
  const handleCloseRoleModal = useCallback(() => {
    // 检查是否有未处理的占位符
    let allPlaceholders = [];
    
    if (rolePromptInputRef.current) {
      const currentPrompt = rolePromptInputRef.current.value;
      const placeholders = extractPlaceholders(currentPrompt);
      allPlaceholders = [...allPlaceholders, ...placeholders];
    }
    if (firstMessageInputRef.current) {
      const currentFirstMessage = firstMessageInputRef.current.value;
      const placeholders = extractPlaceholders(currentFirstMessage);
      allPlaceholders = [...allPlaceholders, ...placeholders];
    }
    
    // 去重并设置需要的变量
    const uniquePlaceholders = [...new Set(allPlaceholders)];
    setRequiredPlaceholders(uniquePlaceholders);
    
    // 关闭模态框
    setShowRoleModal(false);
  }, []);

  const handleResetRolePrompt = useCallback(() => {
    // 重置时直接设置输入框的值
    if (rolePromptInputRef.current) {
      rolePromptInputRef.current.value = defaultRolePrompt;
    }
    if (firstMessageInputRef.current) {
      firstMessageInputRef.current.value = defaultFirstMessage;
    }
  }, [defaultRolePrompt, defaultFirstMessage]);

  const handleRoleModalClose = useCallback(() => {
    setShowRoleModal(false);
  }, []);

  // 清空会话相关函数
  const handleOpenClearConfirm = useCallback(() => {
    if (userId && messages.length > 0) {
      setShowClearConfirm(true);
    } else {
      alert('当前没有可清空的会话');
    }
  }, [userId, messages.length]);

  const handleCloseClearConfirm = useCallback(() => {
    setShowClearConfirm(false);
    setIsClearing(false);
  }, []);

  const [toast, setToast] = useState({ show: false, message: '', type: 'success' });

  // 显示Toast通知
  const showToast = useCallback((message, type = 'success') => {
    setToast({ show: true, message, type });
    // 3秒后自动隐藏
    setTimeout(() => {
      setToast({ show: false, message: '', type: 'success' });
    }, 3000);
  }, []);

  const handleClearSession = useCallback(async () => {
    if (!userId) return;
    
    setIsClearing(true);
    try {
      const sessionId = generateSessionId(userId);
      const result = await deleteSession(sessionId);
      
      if (result.code === 0) {
        // 清空成功，重置本地状态
        setMessages([]);
        setMemoryData(null);
        showToast('✅ 会话已成功清空', 'success');
      } else {
        showToast(`❌ 清空会话失败: ${result.msg}`, 'error');
      }
    } catch (error) {
      console.error('清空会话失败:', error);
      showToast('❌ 清空会话失败，请检查网络连接', 'error');
    } finally {
      setIsClearing(false);
      setShowClearConfirm(false);
    }
  }, [userId, showToast]);

  // 处理占位符变量编辑
  const handlePlaceholderChange = (key, value) => {
    setPlaceholderValues(prev => ({
      ...prev,
      [key]: value
    }));
  };

  // 检查是否所有必需的变量都已填写
  const allPlaceholdersFilled = () => {
    return requiredPlaceholders.every(key => placeholderValues[key] && placeholderValues[key].trim());
  };

  // 替换模板中的占位符
  const replacePlaceholders = (template) => {
    if (!template || typeof template !== 'string') return template || '';
    
    let result = template;
    // 安全地处理变量替换，避免任何可能的JavaScript变量引用
    try {
      // 确保我们只处理字符串类型的模板
      if (typeof result === 'string') {
        Object.entries(placeholderValues).forEach(([key, value]) => {
          const placeholder = `{{${key}}}`;
          // 使用简单的字符串替换，避免正则表达式问题
          // 确保value是字符串类型
          const safeValue = String(value || '');
          result = result.split(placeholder).join(safeValue);
        });
        
        console.log(`变量替换结果:`, {
          原始模板: template,
          替换后结果: result,
          使用的变量: placeholderValues
        });
      }
    } catch (error) {
      console.error('变量替换出错:', error);
      // 如果替换出错，返回原始模板
      result = template;
    }
    
    return result;
  };

  // 点击空白处隐藏变量面板
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (showVariablesPanel && variablesPanelRef.current && 
          !variablesPanelRef.current.contains(event.target)) {
        setShowVariablesPanel(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showVariablesPanel]);

  // 切换变量面板显示状态
  const toggleVariablesPanel = () => {
    setShowVariablesPanel(!showVariablesPanel);
  };

  // 切换语言
  const toggleLanguage = () => {
    const newLanguage = currentLanguage === 'zh' ? 'en' : 'zh';
    setCurrentLanguage(newLanguage);
    setStoredLanguage(newLanguage);
  };

  // 当所有变量填写完成时，显示完成状态，但不自动隐藏面板
  // 让用户手动关闭面板，以便可以随时修改变量
  const [allVariablesFilled, setAllVariablesFilled] = useState(false);
  
  useEffect(() => {
    const filled = allPlaceholdersFilled();
    setAllVariablesFilled(filled);
  }, [placeholderValues, requiredPlaceholders]);

  // 角色设定模态框 - 简化版本，删除变量编辑功能
  const RoleModal = () => (
    <div className="role-modal-overlay">
      <div className="role-modal">
        <div className="modal-header">
          <h3>{t('roleSettings')}</h3>
          <button 
            className="close-btn"
            onClick={handleCloseRoleModal}
          >
            ×
          </button>
        </div>
        
        <div className="modal-content">
          <div className="first-message-section">
            <h4>{t('firstMessage')}</h4>
            <textarea
              ref={firstMessageInputRef}
              className="first-message-input"
              placeholder={t('enterMessage')}
              rows={3}
            />
          </div>
          
          <div className="role-prompt-section">
            <h4>{t('rolePrompt')}</h4>
            <textarea
              ref={rolePromptInputRef}
              className="role-prompt-input"
              placeholder={t('enterMessage')}
              rows={10}
            />
            <div className="role-actions">
              <button 
                className="reset-btn"
                onClick={handleResetRolePrompt}
              >
                {t('resetToDefault')}
              </button>
              <button 
                className="save-btn"
                onClick={handleSaveRolePrompt}
              >
                {t('saveSettings')}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  // 用户选择模态框
  const UserModal = () => (
    <div className="user-modal-overlay">
      <div className="user-modal">
        <div className="modal-header">
          <h3>{t('selectUser')}</h3>
          <button 
            className="close-btn"
            onClick={() => setShowUserModal(false)}
          >
            ×
          </button>
        </div>
        
        <div className="modal-content">
          {/* 注册新用户 */}
          <div className="register-user-section">
            <h4>{t('registerNewUser')}</h4>
            <div className="register-input">
              <input
                ref={newUserIdInputRef}
                type="text"
                placeholder={t('userID')}
                onKeyPress={(e) => e.key === 'Enter' && handleRegisterUser()}
              />
              <button onClick={handleRegisterUser}>{t('registerNewUser')}</button>
            </div>
          </div>

          {/* 选择已有用户 */}
          <div className="existing-users-section">
            <h4>{t('existingUsers')}</h4>
            <div className="users-list">
              {registeredUsers.map((user) => (
                <div 
                  key={user}
                  className={`user-item ${userId === user ? 'active' : ''}`}
                  onClick={() => handleSelectUser(user)}
                >
                  <div className="user-info">
                    <div className="user-id">{user}</div>
                    <div className="user-session">{t('session')}: {generateSessionId(user)}</div>
                  </div>
                  <div className="user-status">
                    {userId === user ? t('current') : t('select')}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  // 清空会话确认模态框
  const ClearConfirmModal = () => (
    <div className="clear-confirm-overlay">
      <div className="clear-confirm-modal">
        <div className="modal-header">
          <h3>{t('clearSessionConfirm')}</h3>
          <button 
            className="close-btn"
            onClick={handleCloseClearConfirm}
            disabled={isClearing}
          >
            ×
          </button>
        </div>
        
        <div className="modal-content">
          <div className="warning-section">
            <div className="warning-icon">⚠️</div>
            <div className="warning-text">
              <h4>{t('deleteOperation')}</h4>
              <p>{t('clearSessionWarning')}</p>
              <ul>
                {t('clearSessionItems').map((item, index) => (
                  <li key={index}>{item}</li>
                ))}
              </ul>
              <p>{t('confirmClear')}</p>
            </div>
          </div>
          
          <div className="confirm-actions">
            <button 
              className="cancel-btn"
              onClick={handleCloseClearConfirm}
              disabled={isClearing}
            >
              {t('cancel')}
            </button>
            <button 
              className="confirm-btn"
              onClick={handleClearSession}
              disabled={isClearing}
            >
              {isClearing ? t('clearing') : t('confirm')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  // 记忆面板组件 - 使用真实API数据
  const MemoryPanel = () => (
    <div className="memory-panel">
      <div className="memory-header">
        <h3>{t('userMemory')}</h3>
        <div className="user-info">
          {userId ? (
            <div className="current-user">
              <span className="user-label">{t('currentUser')}</span>
              <span className="user-id">{userId}</span>
              <span className="session-id">{t('session')}: {generateSessionId(userId)}</span>
              <button 
                className="change-user-btn"
                onClick={handleOpenUserModal}
              >
                {t('switchUser')}
              </button>
            </div>
          ) : (
            <button 
              className="select-user-btn"
              onClick={handleOpenUserModal}
            >
              {t('selectUser')}
            </button>
          )}
        </div>
        <div className="memory-tabs">
          <button 
            className={`tab-btn ${activeTab === 'user_portrait' ? 'active' : ''}`}
            onClick={() => setActiveTab('user_portrait')}
          >
            {t('userPortrait')}
          </button>
          <button 
            className={`tab-btn ${activeTab === 'topic_summary' ? 'active' : ''}`}
            onClick={() => setActiveTab('topic_summary')}
          >
            {t('topicSummary')}
          </button>
          <button 
            className={`tab-btn ${activeTab === 'key_timeline' ? 'active' : ''}`}
            onClick={() => setActiveTab('key_timeline')}
          >
            {t('eventTimeline')}
          </button>
        </div>
      </div>

      <div className="memory-content">
        {activeTab === 'user_portrait' && (
          <div className="user-portrait-section">
            {/* 基本信息栏 */}
            <div className="basic-info-section">
              <h4>{t('basicInformation')}</h4>
              <div className="info-content">
                {memoryData && memoryData.user_portrait && memoryData.user_portrait.basic_information ? (
                  Object.entries(memoryData.user_portrait.basic_information).map(([key, value], index) => (
                    <div key={index} className="info-item">
                      <span className="info-label">{key}:</span>
                      <span className="info-value">{value}</span>
                    </div>
                  ))
                ) : (
                  <div className="no-data">{t('noData')}</div>
                )}
              </div>
            </div>

            {/* 兴趣爱好栏 */}
            <div className="interest-topics-section">
              <h4>{t('interests')}</h4>
              <div className="interest-content">
                {memoryData && memoryData.user_portrait && memoryData.user_portrait.interest_topics ? (
                  Object.entries(memoryData.user_portrait.interest_topics).map(([category, description], index) => (
                    <div key={index} className="interest-item">
                      <span className="interest-label">{category}:</span>
                      <span className="interest-value">{description}</span>
                    </div>
                  ))
                ) : (
                  <div className="no-data">{t('noData')}</div>
                )}
              </div>
            </div>

            {/* 性取向栏 */}
            <div className="sexual-orientation-section">
              <h4>{t('sexualOrientation')}</h4>
              <div className="orientation-content">
                {memoryData && memoryData.user_portrait && memoryData.user_portrait.sexual_orientation ? (
                  Object.entries(memoryData.user_portrait.sexual_orientation).map(([category, values], index) => (
                    <div key={index} className="orientation-item">
                      <span className="orientation-label">{category}:</span>
                      <span className="orientation-value">{Array.isArray(values) ? values.join('; ') : values}</span>
                    </div>
                  ))
                ) : (
                  <div className="no-data">{t('noData')}</div>
                )}
              </div>
            </div>

            {/* 需求栏 */}
            <div className="needs-section">
              <h4>{t('needs')}</h4>
              <div className="needs-content">
                {memoryData && memoryData.user_portrait && memoryData.user_portrait.fulfilled_needs ? (
                  Object.entries(memoryData.user_portrait.fulfilled_needs).map(([category, description], index) => (
                    <div key={index} className="needs-item">
                      <span className="needs-label">{category}:</span>
                      <span className="needs-value">{description}</span>
                    </div>
                  ))
                ) : (
                  <div className="no-data">{t('noData')}</div>
                )}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'topic_summary' && (
          <div className="topic-summary-section">
            <h4>{t('topicSummary')}</h4>
            <div className="topic-summary-content">
              {memoryData && memoryData.topic_summary ? (
                memoryData.topic_summary.map((topic, index) => {
                  // 获取奖杯图标
                  const getTrophyIcon = (idx) => {
                    if (idx === 0) return '🏆'; // 金牌
                    if (idx === 1) return '🥈'; // 银牌
                    if (idx === 2) return '🥉'; // 铜牌
                    return '🎖️'; // 其他奖牌
                  };

                  return (
                    <div key={index} className="topic-summary-item">
                      <div className="topic-title">
                        <span className="topic-icon">{getTrophyIcon(index)}</span>
                        <span className="topic-name">{topic.topic}</span>
                        {topic.last_active && (
                          <span className="topic-time">
                            {new Date(topic.last_active).toLocaleString('zh-CN', {
                              year: 'numeric',
                              month: '2-digit',
                              day: '2-digit',
                              hour: '2-digit',
                              minute: '2-digit',
                              second: '2-digit'
                            })}
                          </span>
                        )}
                      </div>
                      <div className="summary-content">
                        {topic.content && topic.content.map((summary, summaryIndex) => (
                          <div key={summaryIndex} className="summary-item">
                            <div className="summary-text">{summary}</div>
                          </div>
                        ))}
                      </div>
                    </div>
                  );
                })
              ) : (
                <div className="no-data">{t('noData')}</div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'key_timeline' && (
          <div className="timeline-section">
            <h4>{t('eventTimeline')}</h4>
            <div className="timeline-content">
              {memoryData && memoryData.chat_events ? (
                <div className="timeline-events">
                  {/* 已完成事件 */}
                  {memoryData.chat_events.completed && memoryData.chat_events.completed.length > 0 && (
                    <div className="completed-events">
                      <h5>{t('completed')}</h5>
                      {memoryData.chat_events.completed.map((event, index) => (
                        <div key={`completed-${index}`} className="timeline-event completed">
                          <div className="event-content">
                            <div className="event-text">{event.content || event}</div>
                            {event.timestamp && (
                              <div className="event-time">
                                {new Date(event.timestamp).toLocaleString('zh-CN', {
                                  year: 'numeric',
                                  month: '2-digit',
                                  day: '2-digit',
                                  hour: '2-digit',
                                  minute: '2-digit',
                                  second: '2-digit'
                                })}
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                  
                  {/* 待办事件 */}
                  {memoryData.chat_events.todo && memoryData.chat_events.todo.length > 0 && (
                    <div className="todo-events">
                      <h5>{t('todo')}</h5>
                      {memoryData.chat_events.todo.map((event, index) => (
                        <div key={`todo-${index}`} className="timeline-event todo">
                          <div className="event-content">
                            <div className="event-text">{event.content || event}</div>
                            {event.created_at && (
                              <div className="event-time">
                                {t('createdAt')}: {new Date(event.created_at).toLocaleString('zh-CN', {
                                  year: 'numeric',
                                  month: '2-digit',
                                  day: '2-digit',
                                  hour: '2-digit',
                                  minute: '2-digit'
                                })}
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                  
                  {(!memoryData.chat_events.completed || memoryData.chat_events.completed.length === 0) && 
                   (!memoryData.chat_events.todo || memoryData.chat_events.todo.length === 0) && (
                    <div className="no-data">{t('noData')}</div>
                  )}
                </div>
              ) : (
                <div className="no-data">{t('noData')}</div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );


  return (
    <div className="app">
      {/* 用户选择模态框 */}
      {showUserModal && <UserModal />}
      
      {/* 角色设定模态框 */}
      {showRoleModal && <RoleModal />}
      
      {/* 主要内容区域 */}
      <div className="main-content">
        <MemoryPanel />
        
        {/* 聊天区域 */}
        <div className="chat-container">
          {/* 简化聊天头部 */}
          <div className="chat-header">
            <h2>{t('appTitle')}</h2>
          <div className="header-actions">
            <button 
              className="role-btn"
              onClick={handleOpenRoleModal}
            >
              {t('roleSettings')}
            </button>
            {userId && (
              <button 
                className="clear-session-btn"
                onClick={handleOpenClearConfirm}
              >
                {t('clearSession')}
            </button>
            )}
            {userId && (
              <button 
                className="clear-btn"
                onClick={handleClearUser}
              >
                {t('clearUser')}
              </button>
            )}
            <button 
              className="user-btn"
              onClick={handleOpenUserModal}
            >
              {userId ? t('switchUser') : t('selectUser')}
            </button>
          </div>
          </div>

          <div className="messages-container">
            <div className="messages-list">
              {messages.length === 0 ? (
                <div className="empty-state">
                  <div className="empty-icon">💬</div>
                  <h3>{t('emptyStateTitle')}</h3>
                  <p>{userId ? t('emptyStateMessage') : t('emptyStateNoUser')}</p>
                </div>
              ) : (
                messages.map((message) => (
                  <MessageBubble
                    key={message.id}
                    message={message}
                    isUser={message.role === 'user'}
                  />
                ))
              )}
              {isLoading && (
                <div className="loading-message">
                  <div className="typing-indicator">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>
          </div>

          {/* 悬浮变量编辑按钮 - 始终显示，让用户可以随时修改变量 */}
          <div className="floating-variables-container">
            {/* 悬浮按钮 - 显示当前变量状态 */}
            <button 
              className="floating-variables-btn"
              onClick={toggleVariablesPanel}
              title={t('roleVariables')}
            >
              <span className="btn-icon">⚙️</span>
              <span className="btn-text">{t('roleVariables')}</span>
              {requiredPlaceholders.length > 0 && (
                <span className={`badge ${allVariablesFilled ? 'complete' : 'incomplete'}`}>
                  {allVariablesFilled ? '✓' : '!'}
                </span>
              )}
            </button>

            {/* 悬浮变量面板 */}
            {showVariablesPanel && (
              <div 
                ref={variablesPanelRef}
                className="floating-variables-panel"
              >
                <div className="panel-header">
                  <h4>{t('roleVariables')}</h4>
                  <div className="panel-actions">
                    <span className="panel-status">
                      {requiredPlaceholders.length > 0 ? (
                        allVariablesFilled ? (
                          <span className="status-complete">{t('completedStatus')}</span>
                        ) : (
                          <span className="status-incomplete">{t('incompleteStatus')}</span>
                        )
                      ) : (
                        <span className="status-none">{t('noVariables')}</span>
                      )}
                    </span>
                    <button 
                      className="close-panel-btn"
                      onClick={() => setShowVariablesPanel(false)}
                      title="关闭面板"
                    >
                      ×
                    </button>
                  </div>
                </div>
                <div className="panel-content">
                  {requiredPlaceholders.length > 0 ? (
                    <>
                      <div className="variables-inputs">
                        {requiredPlaceholders.map((key) => (
                          <div key={key} className="variable-input-group">
                            <label className="variable-label">
                              {"{{" + key + "}}"}
                              {placeholderValues[key] && placeholderValues[key].trim() && (
                                <span className="filled-indicator">✓</span>
                              )}
                            </label>
                            <input
                              type="text"
                              value={placeholderValues[key] || ''}
                              onChange={(e) => handlePlaceholderChange(key, e.target.value)}
                              placeholder={`${t('enterMessage')} ${key}`}
                              className={`variable-input ${placeholderValues[key] && placeholderValues[key].trim() ? 'filled' : ''}`}
                            />
                          </div>
                        ))}
                      </div>
                      <div className="variables-help">
                        <p>{t('variableHelp1')}</p>
                        <p>{t('variableHelp2')}</p>
                      </div>
                    </>
                  ) : (
                    <div className="no-variables-message">
                      <div className="no-variables-icon">📝</div>
                      <p>{t('noVariablesMessage')}</p>
                      <p className="hint">{t('variableHint')}</p>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* 输入区域 */}
          <div className="input-section">
            <div className="input-container">
            <button 
              className="language-btn input-language-btn"
              onClick={toggleLanguage}
              title={currentLanguage === 'zh' ? 'Switch to English' : '切换到中文'}
            >
              {currentLanguage === 'zh' ? '中文' : 'English'}
            </button>
              <textarea
                ref={messageInputRef}
                className="message-input"
                placeholder={userId ? t('enterMessage') : t('pleaseSelectUser')}
                onKeyPress={handleKeyPress}
                rows={1}
                disabled={!userId}
              />
              <button 
                className="send-btn"
                onClick={handleSendMessage}
                disabled={isLoading || !userId}
              >
                {isLoading ? t('sending') : t('send')}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* 清空会话确认模态框 */}
      {showClearConfirm && <ClearConfirmModal />}

      {/* Toast通知 */}
      {toast.show && (
        <div className={`toast ${toast.type}`}>
          <div className="toast-content">
            <span className="toast-message">{toast.message}</span>
          </div>
        </div>
      )}

      {/* 简化底部，只保留必要信息 */}
      <footer className="app-footer">
        <div className="footer-content">
          <span>© 2025 Farshore AI</span>
          <div className="footer-links">
            <a href="https://github.com/farshore-byte" target="_blank" rel="noopener noreferrer">{t('contactUs')}</a>
            <a href="#share">{t('share')}</a>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
