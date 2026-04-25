<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'

const { t } = useI18n()
const router = useRouter()
const username = ref('')
const password = ref('')
const errMsg = ref('')
const loading = ref(false)
const checkingAdmins = ref(true)
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

onMounted(async () => {
  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/check`)
    const json = await resp.json()
    if (json.code === 0 && !json.data.exists) {
      router.push(`/${adminPrefix}/register`)
    }
  } catch (err) {
    // If check fails, stay on login page (might be network issue)
  } finally {
    checkingAdmins.value = false
  }
})

async function login() {
  if (!username.value || !password.value) {
    errMsg.value = t('admin.enter_credentials')
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    const resp = await fetch(`/api/v1/${adminPrefix}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })
    const json = await resp.json()
    if (json.code === 0) {
      localStorage.setItem('admin_token', json.data.token)
      localStorage.setItem('admin_user', JSON.stringify(json.data))
      router.push(`/${adminPrefix}`)
    } else {
      errMsg.value = json.message || t('admin.login_failed')
    }
  } catch (err) {
    errMsg.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-start justify-center bg-base-200 pt-24">
    <div class="card w-full max-w-md bg-base-100 shadow-xl h-fit">
      <div class="card-body">
        <div class="text-center mb-6">
          <span class="text-4xl font-black">
            Pure<span class="text-primary">Core</span>
          </span>
          <p class="text-base-content/50 mt-2">{{ t('admin.title') }}</p>
        </div>

        <form @submit.prevent="login" class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">{{ t('admin.username') }}</span>
            </label>
            <input
              v-model="username"
              type="text"
              :placeholder="t('admin.username_placeholder')"
              class="input input-bordered w-full"
              autocomplete="username"
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">{{ t('admin.password') }}</span>
            </label>
            <input
              v-model="password"
              type="password"
              :placeholder="t('admin.password_placeholder')"
              class="input input-bordered w-full"
              autocomplete="current-password"
            />
          </div>

          <div v-if="errMsg" class="alert alert-error text-sm">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <span>{{ errMsg }}</span>
          </div>

          <button type="submit" class="btn btn-primary w-full" :disabled="loading">
            <span v-if="loading" class="loading loading-spinner"></span>
            {{ t('admin.sign_in') }}
          </button>
        </form>

        <div class="text-center mt-4">
          <a href="/" class="link link-hover text-sm text-base-content/50">← {{ t('admin.back_home') }}</a>
        </div>
      </div>
    </div>
  </div>
</template>
