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
      body: JSON.stringify({ username: userData.username }),
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    return await response.json()
  },

  async login(username) {
    const response = await fetch(`${API_URL}/users/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username })
    })

    if (!response.ok) {
      throw new Error('Login failed')
    }

    const user = await response.json()
    localStorage.setItem('user', JSON.stringify(user))
    return user
  },

  async listUsers() {
    const response = await fetch(`${API_URL}/users`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })
    if (!response.ok) {
      throw new Error('Failed to fetch users')
    }
    return response.json()
  },

  getCurrentUser() {
    const userStr = localStorage.getItem('user')
    return userStr ? JSON.parse(userStr) : null
  },

  async logout() {
    localStorage.removeItem('user')
  },
}

export const accountService = {
  async listAccounts(userId) {
    const response = await fetch(`${API_URL}/accounts/${userId}`)
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
    console.log(accountData);

    const response = await fetch(`${API_URL}/accounts`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(accountData),
    })
    return response.json()
  },

  async transfer(transferData) {
    try {
      console.log('Starting transfer:', transferData)
      const response = await fetch(`${API_URL}/accounts/transfer`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          from_account_id: transferData.from_account_id,
          to_account_id: transferData.to_account_id,
          amount: transferData.amount
        })
      })

      const responseText = await response.text()
      console.log('Transfer response:', responseText)

      if (!response.ok) {
        throw new Error(responseText || 'Transfer failed')
      }

      try {
        return JSON.parse(responseText)
      } catch {
        return { success: true }
      }
    } catch (err) {
      console.error('Transfer error details:', err)
      throw new Error(err.message || 'Transfer failed')
    }
  }
}

export const auditService = {
  async getAuditsByUser(userId) {
    const response = await fetch(`${API_URL}/audits/${userId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })
    if (!response.ok) {
      throw new Error('Failed to fetch audit logs')
    }
    return response.json()
  }
}
