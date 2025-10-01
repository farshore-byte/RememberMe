import React from 'react';
import { PRESET_ROLES } from '../services/api';

const RoleSelector = ({ onRoleSelect, currentRole }) => {
  const roles = Object.entries(PRESET_ROLES);

  return (
    <div className="role-selector">
      <h3 className="role-selector-title">选择对话角色</h3>
      <div className="role-grid">
        {roles.map(([key, role]) => (
          <button
            key={key}
            className={`role-card ${currentRole === key ? 'active' : ''}`}
            onClick={() => onRoleSelect(key)}
          >
            <div className="role-avatar">{role.avatar}</div>
            <div className="role-info">
              <div className="role-name">{role.name}</div>
              <div className="role-desc">{role.prompt}</div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
};

export default RoleSelector;
