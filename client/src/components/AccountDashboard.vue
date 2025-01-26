<!-- AccountDashboard.vue -->
<script setup>
import { ref, onMounted } from 'vue'
import { userService, accountService } from '@/services/api'

const myAccounts = ref([])
const users = ref([])
const selectedAccount = ref(null)
const targetAccount = ref(null)
const depositAmount = ref('')
const transferAmount = ref('')
const error = ref('')
const loading = ref(false)

const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchMyAccounts(), fetchUsers()])
  } catch (err) {
    console.error('Error refreshing data:', err)
    error.value = 'Failed to refresh data'
  } finally {
    loading.value = false
  }
}

const createNewAccount = async () => {
  try {
    const currentUser = userService.getCurrentUser()
    if (!currentUser) {
      error.value = 'User not authenticated'
      return
    }

    await accountService.createAccount({ user_id: currentUser.id })
    // Refresh accounts list after creation
    await fetchMyAccounts()
  } catch (err) {
    error.value = 'Failed to create account'
    console.error(err)
  }
}

const fetchMyAccounts = async () => {
  try {
    const currentUser = userService.getCurrentUser()
    if (!currentUser) return
    myAccounts.value = await accountService.listAccounts(currentUser.id)
  } catch (err) {
    console.log(err)
    error.value = 'Failed to fetch your accounts'
  }
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await userService.listUsers()
    const currentUser = userService.getCurrentUser()

    const usersWithAccounts = await Promise.all(
      response
        .filter((u) => u.id !== currentUser.id)
        .map(async (user) => {
          console.log(`Fetching accounts for user ${user.id}`) // Debug log
          const accounts = await accountService.listAccounts(user.id)
          console.log(`Accounts for user ${user.id}:`, accounts) // Debug log
          return {
            ...user,
            accounts: Array.isArray(accounts) ? accounts : [],
          }
        }),
    )

    console.log('Users with accounts:', usersWithAccounts) // Debug log
    users.value = usersWithAccounts
  } catch (err) {
    console.error('Error fetching users and accounts:', err)
    error.value = 'Failed to fetch users and their accounts'
  } finally {
    loading.value = false
  }
}

const selectAccount = (account) => {
  selectedAccount.value = account
  targetAccount.value = null // Reset target selection
}

const handleDeposit = async () => {
  if (!selectedAccount.value || !depositAmount.value) {
    error.value = 'Please enter an amount'
    return
  }

  loading.value = true
  try {
    await accountService.deposit({
      account_id: selectedAccount.value.id,
      amount: parseInt(depositAmount.value),
    })
    await refreshData()
    depositAmount.value = ''
    error.value = ''
  } catch (err) {
    console.error('Deposit error:', err)
    error.value = 'Deposit failed'
  } finally {
    loading.value = false
  }
}

const handleTransfer = async () => {
  if (!selectedAccount.value || !targetAccount.value || !transferAmount.value) {
    error.value = 'Please select accounts and enter amount'
    return
  }

  const amount = parseInt(transferAmount.value)
  if (isNaN(amount) || amount <= 0) {
    error.value = 'Please enter a valid amount'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const result = await accountService.transfer({
      from_account_id: selectedAccount.value.id,
      to_account_id: targetAccount.value.id,
      amount: amount,
    })

    // Update source account locally
    if (result.from_account) {
      const sourceAccount = myAccounts.value.find((a) => a.id === result.from_account.id)
      if (sourceAccount) {
        sourceAccount.balance = result.from_account.balance
      }
    }

    // Update target account locally
    if (result.to_account) {
      const targetUser = users.value.find((u) =>
        u.accounts?.some((a) => a.id === result.to_account.id),
      )
      if (targetUser) {
        const targetAccount = targetUser.accounts.find((a) => a.id === result.to_account.id)
        if (targetAccount) {
          targetAccount.balance = result.to_account.balance
        }
      }
    }

    // Force refresh to ensure consistency
    await fetchMyAccounts()
    await fetchUsers()

    transferAmount.value = ''
    targetAccount.value = null
    error.value = `Successfully transferred $${amount}`
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchMyAccounts()
  fetchUsers()
})
</script>

<template>
  <div class="p-6 mx-auto mt-8 max-w-7xl">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-bold text-white">Account Dashboard</h2>
      <button
        @click="createNewAccount"
        class="px-4 py-2 font-semibold text-white bg-green-500 rounded hover:bg-green-800"
      >
        Create New Account
      </button>
    </div>

    <!-- Error display -->
    <div v-if="error" class="p-4 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <div class="grid grid-cols-2 gap-6">
      <!-- Left side - My Accounts -->
      <div class="p-6 rounded-lg bg-slate-700">
        <h3 class="mb-4 text-xl font-bold text-white">My Accounts</h3>
        <div class="space-y-4">
          <div
            v-for="account in myAccounts"
            :key="account.id"
            :class="[
              'p-4 bg-slate-600 rounded cursor-pointer transition-colors',
              selectedAccount?.id === account.id ? 'ring-2 ring-blue-500' : '',
            ]"
            @click="selectAccount(account)"
          >
            <h4 class="font-bold text-white">Account #{{ account.id }}</h4>
            <p class="text-gray-300">Balance: ${{ account.balance }}</p>
          </div>
        </div>

        <!-- Deposit Section -->
        <div v-if="selectedAccount" class="mt-6">
          <h4 class="mb-2 font-bold text-white">Make a Deposit</h4>
          <div class="flex gap-2">
            <input
              type="number"
              v-model="depositAmount"
              class="flex-1 input-field"
              placeholder="Amount"
            />
            <button
              @click="handleDeposit"
              class="px-4 py-2 font-semibold text-white bg-green-500 rounded hover:bg-green-800"
            >
              Deposit
            </button>
          </div>
        </div>
      </div>

      <!-- Right side - Transfer -->
      <div class="p-6 rounded-lg bg-slate-700">
        <h3 class="mb-4 text-xl font-bold text-white">Transfer Money</h3>
        <div v-if="selectedAccount" class="space-y-4">
          <div v-for="user in users" :key="user.id" class="mb-4">
            <h4 class="mb-2 font-bold text-white">{{ user.username }}'s Accounts</h4>
            <div v-if="user.accounts && user.accounts.length" class="space-y-2">
              <div
                v-for="account in user.accounts"
                :key="account.id"
                :class="[
                  'p-4 bg-slate-600 rounded cursor-pointer transition-colors',
                  targetAccount?.id === account.id ? 'ring-2 ring-blue-500' : '',
                ]"
                @click="targetAccount = account"
              >
                <p class="text-white">Account #{{ account.id }}</p>
                <p class="text-gray-300">Balance: ${{ account.balance }}</p>
              </div>
            </div>
            <p v-else class="text-gray-300">No accounts found</p>
          </div>

          <!-- Transfer Amount Input -->
          <div v-if="targetAccount" class="mt-4">
            <div class="flex gap-2">
              <input
                type="number"
                v-model="transferAmount"
                class="flex-1 input-field"
                placeholder="Amount to transfer"
              />
              <button
                @click="handleTransfer"
                class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600"
              >
                Transfer
              </button>
            </div>
          </div>
        </div>
        <p v-else class="text-gray-300">Select one of your accounts to make a transfer</p>
      </div>
    </div>
  </div>
</template>

<style>
.input-field {
  @apply text-black font-bold border rounded px-3 py-2;
}
</style>
