<script setup>
import { ref } from 'vue'
import { accountService } from '@/services/api'

const depositData = ref({
  account_id: '',
  amount: ''
})

const makeDeposit = async () => {
  try {
    const response = await accountService.deposit(depositData.value)
    console.log('Deposit made:', response)
    depositData.value = { account_id: '', amount: '' }
  } catch (error) {
    console.error('Error making deposit:', error)
  }
}
</script>

<template>
  <div class="max-w-md p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Make a Deposit</h2>
    <form @submit.prevent="makeDeposit">
      <div class="mb-4">
        <label class="block mb-2 text-white" for="account_id">Account ID</label>
        <input
          type="number"
          id="account_id"
          v-model="depositData.account_id"
          class="w-full px-3 py-2 text-gray-700 border rounded"
          required
        />
      </div>
      <div class="mb-4">
        <label class="block mb-2 text-white" for="amount">Amount</label>
        <input
          type="number"
          id="amount"
          v-model="depositData.amount"
          class="w-full px-3 py-2 text-gray-700 border rounded"
          required
        />
      </div>
      <button type="submit" class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600">
        Make Deposit
      </button>
    </form>
  </div>
</template>
