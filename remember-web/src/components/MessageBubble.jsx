import React from 'react';

const MessageBubble = ({ message, isUser }) => {
  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const isToday = date.toDateString() === now.toDateString();
    
    if (isToday) {
      // 今天显示详细时间
      return date.toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit',
        second: '2-digit'
      });
    } else {
      // 非今天显示完整日期时间
      return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    }
  };

  // 格式化消息内容，将*号之间的内容加粗（引号内的*号不处理，不成对的*号也不处理）
  const formatMessageContent = (content) => {
    if (!content) return { __html: '' };
    
    let result = '';
    let inQuotes = false;
    let inAsterisk = false;
    let quoteChar = '';
    let asteriskStartIndex = -1;
    
    for (let i = 0; i < content.length; i++) {
      const char = content[i];
      
      // 处理引号（支持英文单引号、双引号和中文引号）
      if ((char === '"' || char === "'" || char === '“' || char === '”' || char === '‘' || char === '’') && !inAsterisk) {
        if (!inQuotes) {
          // 进入引号区域
          inQuotes = true;
          quoteChar = char;
        } else if (char === quoteChar || 
                  (quoteChar === '“' && char === '”') || 
                  (quoteChar === '”' && char === '“') ||
                  (quoteChar === '‘' && char === '’') ||
                  (quoteChar === '’' && char === '‘')) {
          // 退出引号区域
          inQuotes = false;
          quoteChar = '';
        }
        result += char;
        continue;
      }
      
      // 处理星号
      if (char === '*' && !inQuotes) {
        if (!inAsterisk) {
          // 进入星号区域，记录开始位置
          inAsterisk = true;
          asteriskStartIndex = i;
        } else {
          // 退出星号区域，添加加粗标签
          inAsterisk = false;
          // 检查星号之间是否有内容
          if (i > asteriskStartIndex + 1) {
            // 有内容，添加加粗标签
            result += '<strong>' + content.substring(asteriskStartIndex + 1, i) + '</strong>';
          } else {
            // 没有内容，保持原样
            result += content.substring(asteriskStartIndex, i + 1);
          }
          asteriskStartIndex = -1;
        }
        continue;
      }
      
      // 普通字符处理
      if (!inAsterisk) {
        result += char;
      }
    }
    
    // 如果星号没有正确闭合，保持原样输出
    if (inAsterisk && asteriskStartIndex !== -1) {
      result += content.substring(asteriskStartIndex);
    }
    
    // 使用dangerouslySetInnerHTML来渲染HTML内容
    return { __html: result };
  };

  return (
    <div className={`message-bubble ${isUser ? 'user' : 'assistant'}`}>
      <div className="message-avatar">
        {isUser ? '👤' : '🤖'}
      </div>
      <div className="message-content">
        <div 
          className="message-text"
          dangerouslySetInnerHTML={formatMessageContent(message.content)}
        />
        {message.isStreaming && (
          <span className="typing-cursor">|</span>
        )}
        <div className="message-time">
          {formatTime(message.timestamp)}
          {message.isStreaming && (
            <span className="streaming-indicator">正在输入...</span>
          )}
        </div>
      </div>
    </div>
  );
};

export default MessageBubble;
