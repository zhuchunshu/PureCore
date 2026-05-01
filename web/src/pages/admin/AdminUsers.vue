<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { adminAPI } from '../../services/api'
import { config } from '../../services/config'
import { toastSuccess, toastError } from '../../composables/useToast'
import TechCard from '../../components/TechCard.vue'
import GradientButton from '../../components/GradientButton.vue'

const { t } = useI18n()
useSEO({
  title: t('admin.users_title'),
  description: t('admin.users_description'),
})

const router = useRouter()
const adminPrefix = config.adminRoutePrefix

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const users = ref([])

// Modal state
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const editingUser = ref(null)
const deletingUser = ref(null)

// Search
const searchQuery = ref('')

// Create / Edit form
const userForm = reactive({
  name: '',
  email: '',
  avatar: '',
  bio: '',
  status: 'active',
})

const statusOptions = [
  { value: 'active', label: t('user.active'), color: 'badge-success' },
  { value: 'banned', label: t('user.banned'), color: 'badge-error' },
  { value: 'inactive', label: t('user.inactive'), color: 'badge-warning' },
]

const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  const q = searchQuery.value.toLowerCase()
  return users.value.filter(u =>
    u.name?.toLowerCase().includes(q) ||
    u.email?.toLowerCase().includes(q) ||
    u.status?.toLowerCase().includes(q)
  )
})

onMounted(async () => {
  if (!adminAPI.isLoggedIn()) {
    router.push(`/${adminPrefix}/login`)
    return
  }
  await fetchUsers()
})

async function fetchUsers() {
  loading.value = true
  error.value = ''
  try {
    const resp = await adminAPI.get(`${config.adminApiPath}/users`)
    const json = await resp.json()
    if (json.code === 0) {
      users.value = json.data || []
    } else {
      error.value = json.message || t('admin.network_error')
    }
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  userForm.name = ''
  userForm.email = ''
  userForm.avatar = ''
  userForm.bio = ''
  userForm.status = 'active'
  showCreateModal.value = true
}

function openEditModal(user) {
  editingUser.value = user
  userForm.name = user.name || ''
  userForm.email = user.email || ''
  userForm.avatar = user.avatar || ''
  userForm.bio = user.bio || ''
  userForm.status = user.status || 'active'
  showEditModal.value = true
}

function openDeleteModal(user) {
  deletingUser.value = user
  showDeleteModal.value = true
}

function closeModals() {
  showCreateModal.value = false
  showEditModal.value = false
  showDeleteModal.value = false
  editingUser.value = null
  deletingUser.value = null
}

async function handleCreate() {
  saving.value = true
  try {
    const resp = await adminAPI.post(`${config.adminApiPath}/users`, {
      name: userForm.name,
      email: userForm.email,
    })
    const json = await resp.json()
    if (json.code === 0) {
      toastSuccess(t('admin.users_create_success'))
      closeModals()
      await fetchUsers()
    } else {
      toastError(json.message || t('admin.network_error'))
    }
  } catch (err) {
    toastError(t('admin.network_error'))
  } finally {
    saving.value = false
  }
}

async function handleUpdate() {
  if (!editingUser.value) return
  saving.value = true
  try {
    const resp = await adminAPI.put(`${config.adminApiPath}/users/${editingUser.value.id}`, {
      name: userForm.name,
      email: userForm.email,
      avatar: userForm.avatar,
      bio: userForm.bio,
      status: userForm.status,
    })
    const json = await resp.json()
    if (json.code === 0) {
      toastSuccess(t('admin.users_update_success'))
      closeModals()
      await fetchUsers()
    } else {
      toastError(json.message || t('admin.network_error'))
    }
  } catch (err) {
    toastError(t('admin.network_error'))
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!deletingUser.value) return
  saving.value = true
  try {
    const resp = await adminAPI.delete(`${config.adminApiPath}/users/${deletingUser.value.id}`)
    const json = await resp.json()
    if (json.code === 0) {
      toastSuccess(t('admin.users_delete_success'))
      closeModals()
      await fetchUsers()
    } else {
      toastError(json.message || t('admin.network_error'))
    }
  } catch (err) {
    toastError(t('admin.network_error'))
  } finally {
    saving.value = false
  }
}

function getStatusBadge(status) {
  const opt = statusOptions.find(o => o.value === status)
  return opt || { label: status, color: 'badge-ghost' }
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}
</script>

<template>
  <div class="space-y-8">
    <!-- Loading spinner -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <span class="loading loading-spinner loading-lg text-primary"></span>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="flex items-center justify-center py-20">
      <div class="p-6 bg-error/10 border border-error/20 rounded-2xl text-error max-w-lg flex items-center gap-3 backdrop-blur-sm">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span class="font-medium">{{ error }}</span>
      </div>
    </div>

    <!-- Users content -->
    <template v-else>
      <!-- Header section -->
      <div class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-blue-500/10 via-cyan-500/10 to-teal-500/10 border border-blue-500/10 p-6 md:p-8">
        <div class="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/20 to-cyan-500/20 rounded-full blur-3xl -translate-y-1/2 translate-x-1/4 pointer-events-none"></div>
        <div class="absolute bottom-0 left-1/3 w-48 h-48 bg-gradient-to-tr from-teal-500/15 to-emerald-500/15 rounded-full blur-3xl translate-y-1/2 pointer-events-none"></div>
        <div class="relative flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 class="text-2xl md:text-3xl font-black tracking-tight">
              <span class="bg-gradient-to-r from-blue-400 via-cyan-400 to-teal-400 bg-clip-text text-transparent">{{ t('admin.users_title') }}</span>
            </h1>
            <p class="text-base-content/50 mt-2 max-w-lg text-sm md:text-base">{{ t('admin.users_description') }}</p>
          </div>
          <div class="flex items-center gap-3">
            <GradientButton variant="blue" size="sm" @click="openCreateModal">
              <span class="flex items-center gap-2">➕ {{ t('admin.users_add_user') }}</span>
            </GradientButton>
          </div>
        </div>
      </div>

      <!-- Search bar -->
      <div class="flex items-center gap-3">
        <div class="relative flex-1 max-w-md">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
          </div>
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('admin.users_search_placeholder')"
            class="input input-bordered w-full pl-10 bg-base-100/80 backdrop-blur-sm border-base-300/30 focus:border-blue-400 focus:ring-2 focus:ring-blue-400/20 transition-all rounded-xl"
          />
        </div>
        <span class="text-sm text-base-content/40">{{ filteredUsers.length }} / {{ users.length }} {{ t('admin.users') }}</span>
      </div>

      <!-- Users table -->
      <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="table w-full">
            <thead>
              <tr class="bg-base-200/50">
                <th class="font-semibold text-xs uppercase tracking-wider text-base-content/50">ID</th>
                <th class="font-semibold text-xs uppercase tracking-wider text-base-content/50">{{ t('user.name') }}</th>
                <th class="font-semibold text-xs uppercase tracking-wider text-base-content/50">{{ t('user.email') }}</th>
                <th class="font-semibold text-xs uppercase tracking-wider text-base-content/50">{{ t('admin.users_status') }}</th>
                <th class="font-semibold text-xs uppercase tracking-wider text-base-content/50">{{ t('admin.users_created_at') }}</th>
                <th class="font-semibold text-xs uppercase tracking-wider text-base-content/50">{{ t('admin.users_actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="filteredUsers.length === 0">
                <td colspan="6" class="text-center py-10 text-base-content/40">
                  <div class="flex flex-col items-center gap-2">
                    <span class="text-4xl">👥</span>
                    <span>{{ t('admin.users_no_users') }}</span>
                  </div>
                </td>
              </tr>
              <tr
                v-for="user in filteredUsers"
                :key="user.id"
                class="hover:bg-base-200/30 transition-colors"
              >
                <td class="font-mono text-xs text-base-content/40">#{{ user.id }}</td>
                <td>
                  <div class="flex items-center gap-3">
                    <div class="avatar placeholder">
                      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center text-xs font-bold text-primary">
                        {{ (user.name || 'U')[0].toUpperCase() }}
                      </div>
                    </div>
                    <span class="font-medium text-sm">{{ user.name }}</span>
                  </div>
                </td>
                <td class="text-sm text-base-content/60">{{ user.email }}</td>
                <td>
                  <span :class="['badge badge-sm', getStatusBadge(user.status).color]">
                    {{ getStatusBadge(user.status).label }}
                  </span>
                </td>
                <td class="text-xs text-base-content/40">{{ formatDate(user.created_at) }}</td>
                <td>
                  <div class="flex items-center gap-1">
                    <button class="btn btn-ghost btn-xs text-info hover:text-info/80" @click="openEditModal(user)" :title="t('admin.users_edit_user')">
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                    </button>
                    <button class="btn btn-ghost btn-xs text-error hover:text-error/80" @click="openDeleteModal(user)" :title="t('admin.users_delete_user')">
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- Create User Modal -->
    <div v-if="showCreateModal" class="modal modal-open">
      <div class="modal-box bg-base-100/95 backdrop-blur-xl border border-base-300/20 shadow-2xl max-w-lg">
        <h3 class="text-lg font-bold mb-4 flex items-center gap-2">
          <span class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center">➕</span>
          {{ t('admin.users_add_user') }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('user.name') }}</span></label>
            <input v-model="userForm.name" type="text" :placeholder="t('user.name_placeholder')" class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
          </div>
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('user.email') }}</span></label>
            <input v-model="userForm.email" type="email" :placeholder="t('user.email_placeholder')" class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
          </div>
        </div>
        <div class="modal-action">
          <button class="btn btn-ghost rounded-xl" @click="closeModals" :disabled="saving">{{ t('admin.users_cancel') }}</button>
          <GradientButton variant="blue" :loading="saving" :disabled="saving || !userForm.name || !userForm.email" @click="handleCreate">
            {{ t('admin.users_create') }}
          </GradientButton>
        </div>
      </div>
      <div class="modal-backdrop" @click="closeModals"></div>
    </div>

    <!-- Edit User Modal -->
    <div v-if="showEditModal" class="modal modal-open">
      <div class="modal-box bg-base-100/95 backdrop-blur-xl border border-base-300/20 shadow-2xl max-w-lg">
        <h3 class="text-lg font-bold mb-4 flex items-center gap-2">
          <span class="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center">✏️</span>
          {{ t('admin.users_edit_user') }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('user.name') }}</span></label>
            <input v-model="userForm.name" type="text" :placeholder="t('user.name_placeholder')" class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
          </div>
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('user.email') }}</span></label>
            <input v-model="userForm.email" type="email" :placeholder="t('user.email_placeholder')" class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
          </div>
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('admin.users_avatar') }}</span></label>
            <input v-model="userForm.avatar" type="text" placeholder="https://..." class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
          </div>
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('user.bio') }}</span></label>
            <textarea v-model="userForm.bio" rows="2" :placeholder="t('user.bio_placeholder')" class="textarea textarea-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl"></textarea>
          </div>
          <div>
            <label class="label"><span class="label-text font-medium">{{ t('admin.users_status') }}</span></label>
            <select v-model="userForm.status" class="select select-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl">
              <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn btn-ghost rounded-xl" @click="closeModals" :disabled="saving">{{ t('admin.users_cancel') }}</button>
          <GradientButton variant="purple" :loading="saving" :disabled="saving || !userForm.name || !userForm.email" @click="handleUpdate">
            {{ t('admin.users_save') }}
          </GradientButton>
        </div>
      </div>
      <div class="modal-backdrop" @click="closeModals"></div>
    </div>

    <!-- Delete User Modal -->
    <div v-if="showDeleteModal" class="modal modal-open">
      <div class="modal-box bg-base-100/95 backdrop-blur-xl border border-base-300/20 shadow-2xl max-w-md">
        <h3 class="text-lg font-bold mb-4 flex items-center gap-2 text-error">
          <span class="w-8 h-8 rounded-lg bg-error/10 flex items-center justify-center">⚠️</span>
          {{ t('admin.users_delete_user') }}
        </h3>
        <p class="text-sm text-base-content/60 mb-2">
          {{ t('admin.users_confirm_delete') }}
        </p>
        <p class="font-semibold text-sm mb-4">
          {{ deletingUser?.name }} ({{ deletingUser?.email }})
        </p>
        <div class="modal-action">
          <button class="btn btn-ghost rounded-xl" @click="closeModals" :disabled="saving">{{ t('admin.users_cancel') }}</button>
          <button class="btn btn-error rounded-xl" :disabled="saving" @click="handleDelete">
            <span v-if="saving" class="loading loading-spinner loading-xs"></span>
            {{ t('admin.users_delete') }}
          </button>
        </div>
      </div>
      <div class="modal-backdrop" @click="closeModals"></div>
    </div>
  </div>
</template>
