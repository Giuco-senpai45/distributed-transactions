const API_URL = 'http://localhost:8080'

export const userService = {
  async getUser(userId) {
    const response = await fetch(`${API_URL}/users?user_id=${userId}`)
    return response.json()
  },

  async createUser(userData) {
    const response = await fetch(`${API_URL}/users`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData),
    })
    return response.json()
  },

  async login(username) {
    const response = await fetch(`${API_URL}/users/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    })
    if (!response.ok) {
      throw new Error('Login failed')
    }
    const user = await response.json()
    localStorage.setItem('user', JSON.stringify(user))
    return user
  },

  async logout() {
    localStorage.removeItem('user')
  },

  async listUsers() {
    const response = await fetch(`${API_URL}/users/list`)
    if (!response.ok) {
      throw new Error('Failed to fetch users')
    }
    return response.json()
  },

  getCurrentUser() {
    const user = localStorage.getItem('user')
    return user ? JSON.parse(user) : null
  }
}

export const accountService = {
  async listAccounts(userId) {
    const response = await fetch(`${API_URL}/accounts?user_id=${userId}`)
    return response.json()
  },

  async createAccount(userData) {
    const response = await fetch(`${API_URL}/accounts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData),
    })
    return response.json()
  },

  async deposit(accountData) {
    const response = await fetch(`${API_URL}/accounts`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(accountData),
    })
    return response.json()
  }
}

export const auditService = {
  async getAudit(auditId) {
    const response = await fetch(`${API_URL}/audits?audit_id=${auditId}`)
    if (!response.ok) {
      throw new Error('Failed to fetch audit log')
    }
    return response.json()
  },

  async createAudit(auditData) {
    const response = await fetch(`${API_URL}/audits`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(auditData),
    })
    if (!response.ok) {
      throw new Error('Failed to create audit log')
    }
    return response.json()
  }
}
