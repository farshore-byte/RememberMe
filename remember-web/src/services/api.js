// 真实流式API服务
class StreamAPIService {
  constructor() {
    this.baseURL = 'http://localhost:8444/v1'; // 真实后端地址
    this.authToken = 'GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk=';
  }

  // 真实流式回复
  async sendMessageStream(message, conversationHistory = [], sessionId = 'test_session_123', rolePrompt = '', firstMessage = '') {
    try {
      const requestBody = {
        query: message,
        session_id: sessionId,
        role_prompt: rolePrompt || this.getDefaultRolePrompt(),
        stream: true
      };
      
      // 如果提供了first_message，添加到请求体中
      if (firstMessage) {
        requestBody.first_message = firstMessage;
      }

      const response = await fetch('/api/response', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json',
          'Accept': 'text/event-stream'
        },
        body: JSON.stringify(requestBody)
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      return {
        [Symbol.asyncIterator]: async function* () {
          try {
            while (true) {
              const { done, value } = await reader.read();
              if (done) break;

              const chunk = decoder.decode(value);
              const lines = chunk.split('\n');

              for (const line of lines) {
                if (line.startsWith('data: ')) {
                  try {
                    const data = JSON.parse(line.slice(6));
                    if (data.code === 0 && data.data && data.data.content) {
                      yield {
                        content: data.data.content,
                        done: false
                      };
                    }
                  } catch (e) {
                    // 忽略JSON解析错误，继续处理其他行
                    continue;
                  }
                }
              }
            }
          } finally {
            reader.releaseLock();
          }
          yield { content: '', done: true };
        }
      };
    } catch (error) {
      console.error('流式API调用失败:', error);
      throw error;
    }
  }

  // 非流式回复（备用）
  async sendMessage(message, conversationHistory = [], sessionId = 'test_session_123', rolePrompt = '', firstMessage = '') {
    try {
      const requestBody = {
        query: message,
        session_id: sessionId,
        role_prompt: rolePrompt || this.getDefaultRolePrompt(),
        stream: false
      };
      
      // 如果提供了first_message，添加到请求体中
      if (firstMessage) {
        requestBody.first_message = firstMessage;
      }

      const response = await fetch('/api/response', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.authToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      if (data.code === 0 && data.data && data.data.content) {
        return {
          content: data.data.content,
          success: true
        };
      } else {
        throw new Error(data.msg || 'API返回错误');
      }
    } catch (error) {
      console.error('非流式API调用失败:', error);
      throw error;
    }
  }

  // 默认角色提示
  getDefaultRolePrompt() {
    return `[Basic character information]
1.{{char}} is named AI Assistant.
2.{{char}} is a helpful and knowledgeable AI assistant.
3.{{char}} provides accurate and useful information.
4.{{char}} responds in a friendly and professional manner.`;
  }

  // 获取对话历史
  async getConversationHistory(sessionId) {
    // 这个功能可能需要后端支持，暂时返回空
    return {
      messages: [],
      sessionId: sessionId,
      success: true
    };
  }

  // 清除对话历史
  async clearConversation(sessionId) {
    // 这个功能可能需要后端支持，暂时返回成功
    return {
      success: true,
      message: '对话历史已清除'
    };
  }
}

// 创建单例实例
const streamAPIService = new StreamAPIService();

// 记忆应用API
export const memoryApply = async (sessionId, role) => {
  try {
    const response = await fetch('/api/memory/apply', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        session_id: sessionId,
        role: role
      })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('记忆应用失败:', error);
    throw error;
  }
};

// 发送消息API
export const sendMessage = async (sessionId, message) => {
  try {
    const response = await fetch('/api/chat/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        session_id: sessionId,
        message: message
      })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('发送消息失败:', error);
    throw error;
  }
};

// 获取会话历史API
export const getSessionHistory = async (sessionId) => {
  try {
    const response = await fetch(`/api/chat/history/${sessionId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('获取会话历史失败:', error);
    throw error;
  }
};

// 删除会话API - 清空所有微服务中的相关数据
export const deleteSession = async (sessionId) => {
  try {
    const response = await fetch('/api/memory/delete', {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer GcXDgjUSGGEpy83Y9jeqbpFVf4O4GiP1jJJB36hoGJk='
      },
      body: JSON.stringify({
        session_id: sessionId
      })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('删除会话失败:', error);
    throw error;
  }
};

export default streamAPIService;
