import fs from 'fs';
import yaml from 'js-yaml';

// 读取配置文件
function readConfig() {
  try {
    const configPath = '../remember/config.yaml';
    const configContent = fs.readFileSync(configPath, 'utf8');
    const config = yaml.load(configContent);
    return config;
  } catch (error) {
    console.error('读取配置文件失败:', error);
    return null;
  }
}

// 获取端口配置
function getPorts() {
  const config = readConfig();
  if (!config || !config.server) {
    return {
      web: 6006,
      main: 6006,
      openai: 8344
    };
  }

  return {
    web: config.server.web || 6006,
    main: config.server.main || 6006,
    openai: config.server.openai || 8344
  };
}

export { readConfig, getPorts };
