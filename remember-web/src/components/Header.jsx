import React, { useState } from 'react';

const Header = ({ unreadNotifications, userPoints, onSearch }) => {
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e) => {
    e.preventDefault();
    onSearch(searchQuery);
  };

  return (
    <header className="header">
      <div className="header-left">
        <div className="app-logo">
          <span className="logo-icon">📋</span>
          <span className="app-name">Manus</span>
        </div>
      </div>

      <div className="header-center">
        <form onSubmit={handleSearch} className="search-form">
          <div className="search-container">
            <input
              type="text"
              placeholder="搜索任务、消息或内容..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="search-input"
            />
            <button type="submit" className="search-button">
              🔍
            </button>
          </div>
        </form>
      </div>

      <div className="header-right">
        <div className="header-actions">
          <button className="app-button">
            <span className="app-icon">📱</span>
            App
          </button>
          
          <div className="notification-badge">
            <button className="notification-button">
              <span className="notification-icon">🔔</span>
              {unreadNotifications > 0 && (
                <span className="notification-count">{unreadNotifications}</span>
              )}
            </button>
          </div>

          <div className="points-display">
            <span className="points-icon">⭐</span>
            <span className="points-value">{userPoints}</span>
          </div>

          <div className="user-avatar">
            <span className="avatar-icon">👤</span>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
