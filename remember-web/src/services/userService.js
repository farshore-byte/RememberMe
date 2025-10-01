// 用户管理服务 - 使用JSON文件存储用户数据
import usersData from '../data/users.json';

class UserService {
  constructor() {
    this.users = usersData.users || [];
  }

  // 获取所有用户
  getAllUsers() {
    return this.users;
  }

  // 添加新用户
  addUser(userId) {
    if (!userId.trim()) return false;
    
    // 检查是否已存在
    if (this.users.includes(userId)) {
      return false;
    }
    
    // 添加新用户
    this.users.push(userId);
    
    // 保存到JSON文件（这里需要后端支持，暂时使用localStorage作为前端存储）
    this.saveToLocalStorage();
    
    return true;
  }

  // 删除用户
  removeUser(userId) {
    const index = this.users.indexOf(userId);
    if (index > -1) {
      this.users.splice(index, 1);
      this.saveToLocalStorage();
      return true;
    }
    return false;
  }

  // 保存到localStorage（前端存储）
  saveToLocalStorage() {
    localStorage.setItem('registeredUsers', JSON.stringify(this.users));
  }

  // 从localStorage加载用户数据
  loadFromLocalStorage() {
    const savedUsers = localStorage.getItem('registeredUsers');
    if (savedUsers) {
      this.users = JSON.parse(savedUsers);
    }
    return this.users;
  }

  // 获取当前选中的用户（从localStorage）
  getCurrentUser() {
    return localStorage.getItem('currentUserId') || '';
  }

  // 设置当前用户
  setCurrentUser(userId) {
    localStorage.setItem('currentUserId', userId);
  }

  // 清除当前用户
  clearCurrentUser() {
    localStorage.removeItem('currentUserId');
  }
}

// 创建单例实例
const userService = new UserService();

export default userService;
