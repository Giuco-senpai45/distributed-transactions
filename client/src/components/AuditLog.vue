<script setup>
import { ref, onMounted } from 'vue'
import { auditService } from '@/services/api'

const audits = ref([])
const loading = ref(false)
const error = ref(null)

// Add form data model
const auditForm = ref({
  user_id: '',
  operation: '',
  path_id: '',
  timestamp: new Date().toISOString()
})

// Add form submission handler
const handleSubmit = async () => {
  if (!auditForm.value.user_id || !auditForm.value.operation) {
    error.value = 'Please fill all required fields'
    return
  }

  await createAudit(auditForm.value)
  // Reset form
  auditForm.value = {
    user_id: '',
    operation: '',
    path_id: '',
    timestamp: new Date().toISOString()
  }
}

const fetchAudits = async () => {
  loading.value = true
  error.value = null
  try {
    const auditId = 1 // You may want to make this dynamic
    const response = await auditService.getAudit(auditId)
    audits.value = [response] // Adjust if you implement fetching multiple audits
  } catch (err) {
    error.value = 'Error fetching audit logs'
    console.error('Error:', err)
  } finally {
    loading.value = false
  }
}

const createAudit = async (auditData) => {
  loading.value = true
  error.value = null
  try {
    const response = await auditService.createAudit(auditData)
    audits.value.push(response)
  } catch (err) {
    error.value = 'Error creating audit log'
    console.error('Error:', err)
  } finally {
    loading.value = false
  }
}

onMounted(fetchAudits)
</script>

<template>
  <div class="max-w-4xl p-6 mx-auto mt-8 bg-white rounded shadow">
    <h2 class="mb-4 text-2xl font-bold">Create Audit Log</h2>

    <form @submit.prevent="handleSubmit" class="mb-8">
      <div class="grid gap-4">
        <div>
          <label class="block mb-2 text-sm font-bold">User ID</label>
          <input
            v-model="auditForm.user_id"
            type="number"
            required
            class="w-full px-3 py-2 border rounded"
          />
        </div>

        <div>
          <label class="block mb-2 text-sm font-bold">Operation</label>
          <input
            v-model="auditForm.operation"
            type="text"
            required
            class="w-full px-3 py-2 border rounded"
          />
        </div>

        <div>
          <label class="block mb-2 text-sm font-bold">Path ID</label>
          <input
            v-model="auditForm.path_id"
            type="text"
            class="w-full px-3 py-2 border rounded"
          />
        </div>

        <button
          type="submit"
          class="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600"
          :disabled="loading"
        >
          {{ loading ? 'Creating...' : 'Create Audit Log' }}
        </button>
      </div>
    </form>

    <h2 class="mb-4 text-2xl font-bold">Audit Logs</h2>

    <!-- Error Alert -->
    <div v-if="error" class="p-4 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center my-4">
      <div class="w-8 h-8 border-b-2 border-gray-900 rounded-full animate-spin"></div>
    </div>

    <!-- Audit List -->
    <div v-else-if="audits.length" class="space-y-4">
      <div v-for="audit in audits" :key="audit.id" class="p-4 border rounded">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="font-bold">Operation: {{ audit.operation }}</h3>
            <p class="text-gray-600">User ID: {{ audit.user_id }}</p>
            <p class="text-gray-600">Timestamp: {{ new Date(audit.timestamp).toLocaleString() }}</p>
            <p class="text-gray-600">Path ID: {{ audit.path_id }}</p>
          </div>
        </div>
      </div>
    </div>
    <p v-else class="text-gray-500">No audit logs found.</p>
  </div>
</template>
