<script setup>
import { ref, onMounted } from 'vue'
import { userService, accountService } from '@/services/api'

const users = ref([])
const selectedUser = ref(null)
const accounts = ref([])
const depositAmount = ref('')
const error = ref('')
const loading = ref(false)

const fetchUsers = async () => {
  loading.value = true
  error.value = ''

  try {
    const currentUser = userService.getCurrentUser()
    if (!currentUser) {
      throw new Error('Not authenticated')
    }

    users.value = await userService.listUsers()
    users.value = users.value.filter(user => user.id !== currentUser.id)
  } catch (err) {
    console.error('Error fetching users:', err)
    error.value = err.message || 'Failed to fetch users'
  } finally {
    loading.value = false
  }
}

const selectUser = async (user) => {
  selectedUser.value = user
  try {
    accounts.value = await accountService.listAccounts(user.id)
  } catch (error) {
    console.error('Error fetching accounts:', error)
  }
}

const handleDeposit = async (accountId) => {
  if (!depositAmount.value) {
    error.value = 'Please enter an amount'
    return
  }

  try {
    await accountService.deposit({
      account_id: accountId,
      amount: parseInt(depositAmount.value)
    })
    // Refresh accounts after deposit
    accounts.value = await accountService.listAccounts(selectedUser.value.id)
    depositAmount.value = ''
    error.value = ''
  } catch (err) {
    console.log('Error making deposit:', err);

    error.value = 'Failed to make deposit'
  }
}

onMounted(fetchUsers)
</script>

<template>
  <div class="max-w-4xl p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Users</h2>

    <!-- Error display -->
    <div v-if="error" class="p-4 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <div v-if="loading" class="text-gray-600">
      Loading users...
    </div>

    <div v-else class="grid gap-4 md:grid-cols-2">
      <div v-if="users.length" class="space-y-4">
        <div
          v-for="user in users"
          :key="user.id"
          class="p-4 bg-white border rounded cursor-pointer hover:bg-gray-50"
          @click="selectUser(user)"
        >
          <h3 class="font-bold">User #{{ user.id }}</h3>
          <p class="text-gray-600">Username: {{ user.username }}</p>
        </div>
      </div>
      <p v-else class="text-gray-300">No other users found.</p>

      <!-- Selected user's accounts -->
      <div v-if="selectedUser" class="mt-8">
        <h3 class="mb-4 text-xl font-bold text-white">
          {{ selectedUser.username }}'s Accounts
        </h3>

        <div v-if="accounts.length" class="space-y-4">
          <div v-for="account in accounts" :key="account.id" class="p-4 bg-white border rounded">
            <h4 class="font-bold">Account #{{ account.id }}</h4>
            <p class="text-gray-600">Balance: ${{ account.balance }}</p>

            <!-- Deposit form -->
            <div class="mt-4">
              <input
                type="number"
                v-model="depositAmount"
                class="px-3 py-2 mr-2 border rounded"
                placeholder="Amount to deposit"
              />
              <button
                @click="handleDeposit(account.id)"
                class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600"
              >
                Make Deposit
              </button>
            </div>
          </div>
        </div>
        <p v-else class="text-gray-300">No accounts found for this user.</p>
      </div>
    </div>
  </div>
</template>
