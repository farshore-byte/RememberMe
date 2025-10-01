import React, { useEffect } from 'react';
import MessageBubble from './MessageBubble';
import ChatInput from './ChatInput';
import RoleSelector from './RoleSelector';

const ChatContainer = ({ 
  messages, 
  isLoading, 
  currentRole, 
  sessionId,
  messagesEndRef,
  onSendMessage,
  onRoleSelect,
  onClearChat 
}) => {
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, messagesEndRef]);

  return (
    <div className="chat-container">
      {/* 头部 */}
      <div className="chat-header">
        <div className="header-left">
          <h1>Kimi对话助手</h1>
          {currentRole && (
            <span className="current-role">
              当前角色: {currentRole}
            </span>
          )}
        </div>
        <div className="header-actions">
          {sessionId && (
            <button onClick={onClearChat} className="clear-button">
              清除对话
            </button>
          )}
        </div>
      </div>

      {/* 主要内容区域 */}
      <div className="chat-main">
        {/* 角色选择器 */}
        {!sessionId && (
          <div className="role-selector-section">
            <RoleSelector 
              onRoleSelect={onRoleSelect}
              currentRole={currentRole}
            />
          </div>
        )}

        {/* 消息区域 */}
        {sessionId && (
          <div className="messages-container">
            <div className="messages-list">
              {messages.length === 0 ? (
                <div className="empty-state">
                  <div className="empty-icon">💬</div>
                  <h3>开始对话</h3>
                  <p>选择一个角色开始智能对话吧！</p>
                </div>
              ) : (
                messages.map((message, index) => (
                  <MessageBubble
                    key={index}
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
        )}

        {/* 输入区域 */}
        {sessionId && (
          <div className="input-section">
            <ChatInput
              onSendMessage={onSendMessage}
              isLoading={isLoading}
              placeholder={currentRole ? `向${currentRole}发送消息...` : "输入消息..."}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default ChatContainer;
