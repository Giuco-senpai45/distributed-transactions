<script setup>
import { ref } from 'vue'
import { userService } from '@/services/api'

const userData = ref({
  username: ''
})

const createUser = async () => {
  try {
    const response = await userService.createUser(userData.value)
    console.log('User created:', response)
    userData.value = { username: '' }
  } catch (error) {
    console.error('Error creating user:', error)
  }
}
</script>

<template>
  <div class="max-w-md p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold">Create User</h2>
    <form @submit.prevent="createUser">
      <div class="mb-4">
        <label class="block mb-2 text-white" for="username">Username</label>
        <input
          type="text"
          id="username"
          v-model="userData.username"
          class="w-full px-3 py-2 text-gray-700 border rounded"
          required
        />
      </div>
      <button type="submit" class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600">
        Create User
      </button>
    </form>
  </div>
</template>
