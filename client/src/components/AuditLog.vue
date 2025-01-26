<script setup>
import { ref, onMounted } from 'vue'
import { auditService, userService } from '@/services/api'

const audits = ref([])
const loading = ref(false)
const error = ref(null)
const currentUser = ref(null)

const fetchAudits = async () => {
  loading.value = true
  error.value = null
  try {
    currentUser.value = userService.getCurrentUser()
    if (!currentUser.value) {
      error.value = 'Not authenticated'
      return
    }

    // Fetch audits for current user
    const response = await auditService.getAuditsByUser(currentUser.value.id)
    audits.value = Array.isArray(response) ? response : []
  } catch (err) {
    error.value = 'Error fetching audit logs'
    console.error('Error:', err)
  } finally {
    loading.value = false
  }
}

onMounted(fetchAudits)
</script>

<template>
  <div class="max-w-4xl p-6 mx-auto mt-8 rounded shadow bg-slate-600">
    <h2 class="mb-4 text-2xl font-bold text-white">Audit Logs</h2>

    <!-- Error Alert -->
    <div v-if="error" class="p-4 mb-4 text-red-700 bg-red-100 rounded">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center my-4">
      <div class="w-8 h-8 border-b-2 border-white rounded-full animate-spin"></div>
    </div>

    <!-- Audit List -->
    <div v-else-if="audits.length" class="space-y-4">
      <div v-for="audit in audits" :key="audit.id"
           class="p-4 rounded bg-slate-700">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="font-bold text-white">Operation: {{ audit.operation }}</h3>
            <p class="text-gray-300">
              {{ new Date(audit.timestamp).toLocaleString() }}
            </p>
          </div>
        </div>
      </div>
    </div>
    <p v-else class="text-gray-300">No audit logs found.</p>
  </div>
</template>
