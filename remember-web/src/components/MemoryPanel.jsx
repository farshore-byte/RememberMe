import React, { useState } from 'react';

const MemoryPanel = ({ memoryData }) => {
  const [activeTab, setActiveTab] = useState('user_portrait');

  // 格式化时间函数
  const formatTime = (timestamp) => {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  // 获取奖杯图标
  const getTrophyIcon = (index) => {
    if (index === 0) return '🏆'; // 金牌
    if (index === 1) return '🥈'; // 银牌
    if (index === 2) return '🥉'; // 铜牌
    return '🎖️'; // 其他奖牌
  };

  return (
    <div className="memory-panel">
      <div className="memory-header">
        <h3>记忆摘录</h3>
        <div className="memory-tabs">
          <button 
            className={`tab-btn ${activeTab === 'user_portrait' ? 'active' : ''}`}
            onClick={() => setActiveTab('user_portrait')}
          >
            用户画像
          </button>
          <button 
            className={`tab-btn ${activeTab === 'topic_summary' ? 'active' : ''}`}
            onClick={() => setActiveTab('topic_summary')}
          >
            话题摘要
          </button>
        </div>
      </div>

      <div className="memory-content">
        {activeTab === 'user_portrait' && (
          <div className="user-portrait-section">
            <h4>兴趣主题</h4>
            {memoryData.user_portrait && memoryData.user_portrait.interest_topics && (
              <div className="interest-topics">
                {Object.entries(memoryData.user_portrait.interest_topics).map(([topic, description], index) => (
                  <div key={topic} className="topic-item">
                    <div className="topic-header">
                      <span className="topic-icon">{getTrophyIcon(index)}</span>
                      <span className="topic-name">{topic}</span>
                    </div>
                    <p className="topic-description">{description}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'topic_summary' && (
          <div className="topic-summary-section">
            <h4>对话话题摘要</h4>
            {memoryData.topic_summary && memoryData.topic_summary.map((topicData, index) => (
              <div key={index} className="topic-summary-item">
                <div className="topic-title">
                  <span className="topic-icon">{getTrophyIcon(index)}</span>
                  <span className="topic-name">{topicData.topic}</span>
                  {topicData.last_active && (
                    <span className="topic-time">
                      {formatTime(topicData.last_active)}
                    </span>
                  )}
                </div>
                <div className="summary-content">
                  {topicData.content && topicData.content.map((summary, summaryIndex) => (
                    <div key={summaryIndex} className="summary-item">
                      <div className="summary-text">{summary}</div>
                      <div className="summary-meta">
                        <span className="summary-time">最近更新</span>
                        <span className="summary-count">{summaryIndex + 1}/{topicData.content.length}</span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default MemoryPanel;
