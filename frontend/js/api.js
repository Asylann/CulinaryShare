// API Configuration
// Use relative path when served by nginx (Docker), absolute when running locally
const API_BASE_URL = window.location.port === '3000' ? '/api' : 'http://localhost:8080/api';

// Token management
const TokenManager = {
  getToken() {
    return localStorage.getItem('token');
  },
  setToken(token) {
    localStorage.setItem('token', token);
  },
  removeToken() {
    localStorage.removeItem('token');
  },
  getUser() {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },
  setUser(user) {
    localStorage.setItem('user', JSON.stringify(user));
  },
  removeUser() {
    localStorage.removeItem('user');
  },
  isLoggedIn() {
    return !!this.getToken();
  },
  logout() {
    this.removeToken();
    this.removeUser();
  }
};

// API Client
const api = {
  // Generic request method
  async request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    const token = TokenManager.getToken();

    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers
      },
      ...options
    };

    if (config.body && typeof config.body === 'object') {
      config.body = JSON.stringify(config.body);
    }

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw {
          status: response.status,
          message: data.error || data.message || 'Something went wrong',
          data
        };
      }

      return data;
    } catch (error) {
      if (error.status === 401) {
        TokenManager.logout();
        window.location.href = '/login.html';
      }
      throw error;
    }
  },

  // GET request
  async get(endpoint) {
    return this.request(endpoint, { method: 'GET' });
  },

  // POST request
  async post(endpoint, body) {
    return this.request(endpoint, { method: 'POST', body });
  },

  // PUT request
  async put(endpoint, body) {
    return this.request(endpoint, { method: 'PUT', body });
  },

  // PATCH request
  async patch(endpoint, body) {
    return this.request(endpoint, { method: 'PATCH', body });
  },

  // DELETE request
  async delete(endpoint) {
    return this.request(endpoint, { method: 'DELETE' });
  }
};

// Auth API
const AuthAPI = {
  async register(userData) {
    const response = await api.post('/auth/register', userData);
    if (response.data?.token) {
      TokenManager.setToken(response.data.token);
      TokenManager.setUser(response.data.user);
    }
    return response;
  },

  async login(credentials) {
    const response = await api.post('/auth/login', credentials);
    if (response.data?.token) {
      TokenManager.setToken(response.data.token);
      TokenManager.setUser(response.data.user);
    }
    return response;
  },

  async getMe() {
    return api.get('/users/me');
  },

  logout() {
    TokenManager.logout();
    window.location.href = '/index.html';
  }
};

// Recipes API
const RecipesAPI = {
  async getAll(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/recipes?${queryString}` : '/recipes';
    return api.get(endpoint);
  },

  async getById(id) {
    return api.get(`/recipes/${id}`);
  },

  async create(recipeData) {
    return api.post('/recipes', recipeData);
  },

  async update(id, recipeData) {
    return api.put(`/recipes/${id}`, recipeData);
  },

  async delete(id) {
    return api.delete(`/recipes/${id}`);
  },

  async updateIngredients(id, action, ingredients) {
    return api.patch(`/recipes/${id}/ingredients`, { action, ingredients });
  },

  async addReview(recipeId, reviewData) {
    return api.post(`/recipes/${recipeId}/reviews`, reviewData);
  },

  async deleteReview(recipeId, reviewId) {
    return api.delete(`/recipes/${recipeId}/reviews/${reviewId}`);
  }
};

// Categories API
const CategoriesAPI = {
  async getAll() {
    return api.get('/categories');
  },

  async create(categoryData) {
    return api.post('/categories', categoryData);
  }
};

// Analytics API
const AnalyticsAPI = {
  async getTopRated(limit = 10) {
    return api.get(`/analytics/top-rated?limit=${limit}`);
  }
};

// Export for use
window.TokenManager = TokenManager;
window.api = api;
window.AuthAPI = AuthAPI;
window.RecipesAPI = RecipesAPI;
window.CategoriesAPI = CategoriesAPI;
window.AnalyticsAPI = AnalyticsAPI;
