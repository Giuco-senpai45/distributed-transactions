<script setup>
import { ref, onMounted } from 'vue'
import { accountService } from '@/services/api'

const accounts = ref([])
const selectedUserId = ref('')
const error = ref(null)

const fetchAccounts = async () => {
  try {
    if (!selectedUserId.value) return
    accounts.value = await accountService.listAccounts(selectedUserId.value)
  } catch (error) {
    console.error('Error fetching accounts:', error)
    error.value = 'Failed to fetch accounts'
  }
}

const handleUserSelect = () => {
  fetchAccounts()
}

onMounted(fetchAccounts)
</script>

<template>
  <div class="max-w-4xl p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Accounts</h2>

    <div class="mb-4">
      <label class="block mb-2 text-white">Select User ID</label>
      <input
        type="number"
        v-model="selectedUserId"
        @change="handleUserSelect"
        class="w-full px-3 py-2 text-gray-700 border rounded"
        placeholder="Enter user ID"
      />
    </div>

    <div v-if="error" class="p-4 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <div v-if="accounts.length" class="space-y-4">
      <div v-for="account in accounts" :key="account.id" class="p-4 bg-white border rounded">
        <h3 class="font-bold">Account #{{ account.id }}</h3>
        <p class="text-gray-600">Balance: ${{ account.balance }}</p>
        <p class="text-gray-600">User ID: {{ account.user_id }}</p>
      </div>
    </div>
    <p v-else class="text-gray-300">No accounts found for this user.</p>
  </div>
</template>
