import { useState, useRef, useCallback } from 'react';
import { memoryApply, sendMessage, getSessionHistory } from '../services/api';

export const useChat = () => {
  const [messages, setMessages] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [currentRole, setCurrentRole] = useState(null);
  const [sessionId, setSessionId] = useState('');
  const messagesEndRef = useRef(null);

  // 滚动到底部
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  // 初始化会话
  const initializeSession = useCallback(async (roleKey, customSessionId = '') => {
    const role = roleKey;
    const sessionId = customSessionId || `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    setIsLoading(true);
    try {
      const response = await memoryApply(sessionId, role);
      if (response.code === 0) {
        setSessionId(sessionId);
        setCurrentRole(role);
        setMessages(response.data.messages || []);
        return { success: true, sessionId };
      } else {
        throw new Error(response.msg || '初始化失败');
      }
    } catch (error) {
      console.error('初始化会话失败:', error);
      return { success: false, error: error.message };
    } finally {
      setIsLoading(false);
    }
  }, []);

  // 发送消息
  const sendChatMessage = useCallback(async (message) => {
    if (!sessionId || !message.trim()) return;

    setIsLoading(true);

    try {
      // 在用户发起请求时刷新记忆
      await memoryApply(sessionId, currentRole);

      const userMessage = {
        role: 'user',
        content: message.trim(),
        timestamp: new Date().toISOString()
      };

      // 添加用户消息
      setMessages(prev => [...prev, userMessage]);

      const response = await sendMessage(sessionId, message);
      if (response.code === 0) {
        const assistantMessage = {
          role: 'assistant',
          content: response.data.messages?.[response.data.messages.length - 1]?.content || '抱歉，我无法理解您的消息。',
          timestamp: new Date().toISOString()
        };
        
        setMessages(prev => [...prev, assistantMessage]);
      } else {
        throw new Error(response.msg || '发送消息失败');
      }
    } catch (error) {
      console.error('发送消息失败:', error);
      const errorMessage = {
        role: 'assistant',
        content: '抱歉，发送消息时出现错误，请稍后重试。',
        timestamp: new Date().toISOString(),
        isError: true
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
      setTimeout(scrollToBottom, 100);
    }
  }, [sessionId, currentRole, scrollToBottom]);

  // 清除对话
  const clearChat = useCallback(() => {
    setMessages([]);
    setSessionId('');
    setCurrentRole(null);
  }, []);

  // 加载会话历史
  const loadSessionHistory = useCallback(async (targetSessionId) => {
    setIsLoading(true);
    try {
      const response = await getSessionHistory(targetSessionId);
      if (response.code === 0) {
        setMessages(response.data.messages || []);
        setSessionId(targetSessionId);
        return { success: true };
      }
    } catch (error) {
      console.error('加载会话历史失败:', error);
      return { success: false, error: error.message };
    } finally {
      setIsLoading(false);
    }
  }, []);

  return {
    messages,
    isLoading,
    currentRole,
    sessionId,
    messagesEndRef,
    initializeSession,
    sendChatMessage,
    clearChat,
    loadSessionHistory,
    scrollToBottom
  };
};
