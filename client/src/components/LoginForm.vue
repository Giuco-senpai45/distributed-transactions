<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { userService } from '@/services/api'

const router = useRouter()
const username = ref('')
const error = ref('')

const handleLogin = async () => {
  try {
    await userService.login(username.value)
    router.push('/users')
  } catch (err) {
    console.log(err);
    error.value = 'Login failed. Please try again.'
  }
}
</script>

<template>
  <div class="max-w-md p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Login</h2>
    <form @submit.prevent="handleLogin">
      <div class="mb-4">
        <label class="block mb-2 text-white" for="username">Username</label>
        <input
          type="text"
          id="username"
          v-model="username"
          class="w-full px-3 py-2 text-gray-700 border rounded"
          required
        />
      </div>
      <div v-if="error" class="mb-4 text-red-500">
        {{ error }}
      </div>
      <div class="flex items-center justify-between">
        <button type="submit" class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600">
          Login
        </button>
        <RouterLink
          to="/users/create"
          class="px-4 py-2 text-white bg-green-500 rounded hover:bg-green-600"
        >
          Create Account
        </RouterLink>
      </div>
    </form>
  </div>
</template>
