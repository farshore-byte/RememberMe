import React, { useState } from 'react';
import MessageBubble from './MessageBubble';

const ChatArea = ({ selectedTask, messages, onSendMessage, isLoading }) => {
  const [messageInput, setMessageInput] = useState('');

  const handleSendMessage = (e) => {
    e.preventDefault();
    if (messageInput.trim() && !isLoading) {
      onSendMessage(messageInput);
      setMessageInput('');
    }
  };

  const insertMedia = (type) => {
    const mediaPlaceholders = {
      image: '[图片]',
      webpage: '[网页链接]',
      spreadsheet: '[电子表格]',
      document: '[文档]'
    };
    
    setMessageInput(prev => prev + mediaPlaceholders[type]);
  };

  return (
    <div className="chat-area">
      {selectedTask ? (
        <>
          <div className="chat-header">
            <div className="task-chat-info">
              <h4>{selectedTask.title}</h4>
              <span className="task-status-badge">{selectedTask.status}</span>
            </div>
          </div>

          <div className="messages-container">
            <div className="messages-list">
              {messages.length === 0 ? (
                <div className="empty-chat">
                  <span className="chat-icon">💬</span>
                  <p>开始对话</p>
                  <small>发送第一条消息开始讨论这个任务</small>
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
                <div className="typing-indicator">
                  <span></span>
                  <span></span>
                  <span></span>
                  <span>正在输入...</span>
                </div>
              )}
            </div>
          </div>

          <div className="chat-input-section">
            <div className="media-toolbar">
              <button onClick={() => insertMedia('image')} className="media-btn">
                🖼️ 图片
              </button>
              <button onClick={() => insertMedia('webpage')} className="media-btn">
                🌐 网页
              </button>
              <button onClick={() => insertMedia('spreadsheet')} className="media-btn">
                📊 电子表格
              </button>
              <button onClick={() => insertMedia('document')} className="media-btn">
                📄 文档
              </button>
            </div>

            <form onSubmit={handleSendMessage} className="message-form">
              <div className="input-container">
                <textarea
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  placeholder="输入消息..."
                  className="message-input"
                  rows={3}
                  disabled={isLoading}
                />
                <div className="input-actions">
                  <button type="button" className="voice-btn" title="语音输入">
                    🎤
                  </button>
                  <button 
                    type="submit" 
                    className="send-btn"
                    disabled={!messageInput.trim() || isLoading}
                  >
                    {isLoading ? '发送中...' : '发送'}
                  </button>
                </div>
              </div>
            </form>
          </div>
        </>
      ) : (
        <div className="no-task-selected">
          <div className="no-task-content">
            <span className="no-task-icon">📋</span>
            <h3>选择一个任务开始对话</h3>
            <p>从左侧任务列表中选择一个任务来查看和参与讨论</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default ChatArea;
