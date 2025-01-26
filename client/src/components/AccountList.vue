<script setup>
import { ref, onMounted } from 'vue'
import { accountService } from '@/services/api'

const accounts = ref([])
const selectedAccount = ref(null)
const transferData = ref({
  toAccountId: '',
  amount: ''
})
const error = ref(null)

const fetchAccounts = async () => {
  try {
    const currentUser = JSON.parse(localStorage.getItem('user'))
    accounts.value = await accountService.listAccounts(currentUser.id)
  } catch (error) {
    error.value = 'Failed to fetch accounts'
  }
}

const handleTransfer = async () => {
  if (!selectedAccount.value || !transferData.value.toAccountId || !transferData.value.amount) {
    error.value = 'Please fill all fields'
    return
  }

  try {
    await accountService.transfer({
      from_account_id: selectedAccount.value.id,
      to_account_id: parseInt(transferData.value.toAccountId),
      amount: parseInt(transferData.value.amount)
    })
    await fetchAccounts()
    transferData.value = { toAccountId: '', amount: '' }
  } catch (err) {
    console.log(err);

    error.value = 'Transfer failed'
  }
}

onMounted(fetchAccounts)
</script>

<template>
  <div class="max-w-4xl p-6 mx-auto mt-8 bg-white rounded shadow">
    <h2 class="mb-4 text-2xl font-bold">My Accounts</h2>

    <!-- Error display -->
    <div v-if="error" class="p-4 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <!-- Account List -->
    <div class="grid gap-4 md:grid-cols-2">
      <div v-for="account in accounts"
           :key="account.id"
           :class="['p-4 border rounded cursor-pointer',
                   selectedAccount?.id === account.id ? 'border-blue-500' : '']"
           @click="selectedAccount = account">
        <h3 class="font-bold">Account #{{ account.id }}</h3>
        <p class="text-gray-600">Balance: ${{ account.balance }}</p>
      </div>
    </div>

    <!-- Transfer Form -->
    <div v-if="selectedAccount" class="mt-8">
      <h3 class="mb-4 text-xl font-bold">Transfer Money</h3>
      <form @submit.prevent="handleTransfer" class="space-y-4">
        <div>
          <label class="block mb-2">To Account ID</label>
          <input type="number"
                 v-model="transferData.toAccountId"
                 class="w-full px-3 py-2 border rounded" />
        </div>
        <div>
          <label class="block mb-2">Amount</label>
          <input type="number"
                 v-model="transferData.amount"
                 class="w-full px-3 py-2 border rounded" />
        </div>
        <button type="submit"
                class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600">
          Transfer
        </button>
      </form>
    </div>
  </div>
</template>
