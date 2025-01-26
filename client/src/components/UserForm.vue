<script setup>
import { ref } from 'vue'
import { userService } from '@/services/api'
import { useRouter } from 'vue-router'

const router = useRouter()
const userData = ref({
  username: ''
})
const error = ref('')
const loading = ref(false)

const createUser = async () => {
  if (!userData.value.username.trim()) {
    error.value = 'Username is required'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await userService.createUser(userData.value)
    router.push('/login')
  } catch (err) {
    error.value = err.message || 'Failed to create user'
    console.error('Error:', err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Create User</h2>

    <div v-if="error" class="p-3 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <form @submit.prevent="createUser">
      <div class="mb-4">
        <label class="block mb-2 text-white">Username</label>
        <input
          v-model="userData.username"
          type="text"
          required
          class="w-full px-3 py-2 text-black border rounded"
          :disabled="loading"
        />
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600 disabled:opacity-50"
      >
        {{ loading ? 'Creating...' : 'Create User' }}
      </button>
    </form>
  </div>
</template>
