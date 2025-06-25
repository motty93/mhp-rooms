// API通信の共通関数
export const apiClient = {
  async get(url, options = {}) {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Accept': 'application/json',
        ...options.headers
      }
    });
    return response;
  },

  async post(url, data, options = {}) {
    const response = await fetch(url, {
      ...options,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...options.headers
      },
      body: JSON.stringify(data)
    });
    return response;
  },

  async put(url, data, options = {}) {
    const response = await fetch(url, {
      ...options,
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...options.headers
      },
      body: JSON.stringify(data)
    });
    return response;
  },

  async delete(url, options = {}) {
    const response = await fetch(url, {
      ...options,
      method: 'DELETE',
      headers: {
        'Accept': 'application/json',
        ...options.headers
      }
    });
    return response;
  }
};