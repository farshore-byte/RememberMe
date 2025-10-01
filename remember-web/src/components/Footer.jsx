import React from 'react';

const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer-content">
        <div className="footer-left">
          <button className="footer-btn">
            <span className="footer-icon">❓</span>
            帮助
          </button>
          <button className="footer-btn">
            <span className="footer-icon">💬</span>
            支持
          </button>
        </div>
        
        <div className="footer-center">
          <span className="app-version">Manus v1.0.0</span>
        </div>
        
        <div className="footer-right">
          <button className="footer-btn">
            <span className="footer-icon">📤</span>
            分享
          </button>
          <button className="footer-btn">
            <span className="footer-icon">⚙️</span>
            设置
          </button>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
