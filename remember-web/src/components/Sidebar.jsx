import React, { useState } from 'react';

const Sidebar = ({ tasks, onTaskSelect, selectedTask, onNewTask }) => {
  const [activeFilter, setActiveFilter] = useState('all');
  const [taskSearch, setTaskSearch] = useState('');

  const filteredTasks = tasks.filter(task => {
    const matchesSearch = task.title.toLowerCase().includes(taskSearch.toLowerCase()) ||
                         task.description.toLowerCase().includes(taskSearch.toLowerCase());
    
    if (activeFilter === 'all') return matchesSearch;
    if (activeFilter === 'favourite') return matchesSearch && task.isFavourite;
    if (activeFilter === 'timed') return matchesSearch && task.hasTimer;
    
    return matchesSearch;
  });

  const formatTime = (date) => {
    return new Date(date).toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h3>任务管理</h3>
        <button className="new-task-btn" onClick={onNewTask}>
          <span className="plus-icon">+</span>
          新建任务
        </button>
      </div>

      <div className="task-search">
        <input
          type="text"
          placeholder="搜索任务..."
          value={taskSearch}
          onChange={(e) => setTaskSearch(e.target.value)}
          className="task-search-input"
        />
      </div>

      <div className="task-filters">
        <button 
          className={`filter-btn ${activeFilter === 'all' ? 'active' : ''}`}
          onClick={() => setActiveFilter('all')}
        >
          全部
        </button>
        <button 
          className={`filter-btn ${activeFilter === 'favourite' ? 'active' : ''}`}
          onClick={() => setActiveFilter('favourite')}
        >
          收藏
        </button>
        <button 
          className={`filter-btn ${activeFilter === 'timed' ? 'active' : ''}`}
          onClick={() => setActiveFilter('timed')}
        >
          已定时
        </button>
      </div>

      <div className="task-list">
        {filteredTasks.length === 0 ? (
          <div className="empty-tasks">
            <span className="empty-icon">📝</span>
            <p>暂无任务</p>
          </div>
        ) : (
          filteredTasks.map(task => (
            <div
              key={task.id}
              className={`task-item ${selectedTask?.id === task.id ? 'active' : ''}`}
              onClick={() => onTaskSelect(task)}
            >
              <div className="task-header">
                <div className="task-title">{task.title}</div>
                <div className="task-meta">
                  {task.isFavourite && <span className="favourite-icon">⭐</span>}
                  {task.hasTimer && <span className="timer-icon">⏰</span>}
                </div>
              </div>
              <div className="task-description">{task.description}</div>
              <div className="task-footer">
                <span className="task-time">{formatTime(task.createdAt)}</span>
                <span className={`task-status ${task.status}`}>
                  {task.status === 'pending' ? '进行中' : 
                   task.status === 'completed' ? '已完成' : '已超时'}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Sidebar;
