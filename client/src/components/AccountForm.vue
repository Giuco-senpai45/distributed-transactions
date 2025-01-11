<script setup>
import { ref } from 'vue'
import { accountService } from '@/services/api'

const accountData = ref({
  user_id: ''
})

const createAccount = async () => {
  try {
    const response = await accountService.createAccount(accountData.value)
    console.log('Account created:', response)
    accountData.value = { user_id: '' }
  } catch (error) {
    console.error('Error creating account:', error)
  }
}
</script>

<template>
  <div class="max-w-md p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Create Account</h2>
    <form @submit.prevent="createAccount">
      <div class="mb-4">
        <label class="block mb-2 text-white" for="user_id">User ID</label>
        <input
          type="number"
          id="user_id"
          v-model="accountData.user_id"
          class="w-full px-3 py-2 text-gray-700 border rounded"
          required
        />
      </div>
      <button type="submit" class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600">
        Create Account
      </button>
    </form>
  </div>
</template>
