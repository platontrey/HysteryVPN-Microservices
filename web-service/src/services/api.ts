import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Xray API functions
export const xrayAPI = {
  // Generate Xray configuration for user
  generateConfig: (userId: string, protocol: string = 'vless', deviceId?: string) => {
    const params = new URLSearchParams();
    params.append('protocol', protocol);
    if (deviceId) params.append('deviceId', deviceId);

    return api.post(`/users/${userId}/xray/config?${params}`);
  },

  // Update Xray configuration
  updateConfig: (userId: string, configId: string, config: any, deviceId?: string) => {
    const params = new URLSearchParams();
    if (deviceId) params.append('deviceId', deviceId);

    return api.put(`/users/${userId}/xray/config/${configId}?${params}`, config);
  },

  // Get Xray configurations for user
  getConfigs: (userId: string) => {
    return api.get(`/users/${userId}/xray/config`);
  },

  // Get supported protocols
  getSupportedProtocols: () => {
    return api.get('/xray/protocols');
  },

  // Get Xray service status
  getStatus: () => {
    return api.get('/xray/status');
  },

  // Get active connections
  getConnections: (page: number = 1, limit: number = 50) => {
    return api.get(`/xray/connections?page=${page}&limit=${limit}`);
  },

  // Disconnect user
  disconnectUser: (userId: string, deviceId?: string) => {
    const params = new URLSearchParams();
    if (deviceId) params.append('deviceId', deviceId);

    return api.post(`/xray/users/${userId}/disconnect?${params}`);
  },

  // Reload configuration
  reloadConfig: () => {
    return api.post('/xray/reload');
  },
};

export default api;