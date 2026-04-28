<script setup>
import { ref } from 'vue'
import { useI18n } from '../../i18n'
import { accessToken } from '../../composables/useUserAuth'
import TechCard from '../TechCard.vue'

const { t } = useI18n()

const passwordForm = ref({
  current_password: '',
  new_password: '',
  confirm_password: '',
})
const passwordSaving = ref(false)
const passwordMsg = ref('')
const passwordMsgType = ref('success')

async function changePassword() {
  if (passwordForm.value.new_password.length < 8) {
    passwordMsg.value = t('user.password_min_length')
    passwordMsgType.value = 'error'
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    passwordMsg.value = t('user.passwords_not_match')
    passwordMsgType.value = 'error'
    return
  }
  passwordSaving.value = true
  passwordMsg.value = ''
  try {
    const resp = await fetch('/api/v1/user/auth/password', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken.value}`,
      },
      body: JSON.stringify(passwordForm.value),
    })
    const json = await resp.json()
    if (json.code === 0) {
      passwordMsg.value = t('user.password_changed')
      passwordMsgType.value = 'success'
      passwordForm.value = { current_password: '', new_password: '', confirm_password: '' }
    } else {
      passwordMsg.value = json.message || t('user.password_change_failed')
      passwordMsgType.value = 'error'
    }
  } catch {
    passwordMsg.value = t('user.password_change_failed')
    passwordMsgType.value = 'error'
  } finally {
    passwordSaving.value = false
  }
}
</script>

<template>
  <TechCard variant="blue" padded>
    <h2 class="text-lg font-bold text-base-content/80 mb-5">🔒 {{ t('user.change_password') }}</h2>
    <div v-if="passwordMsg" :class="['p-3 rounded-xl mb-4 text-sm font-medium', passwordMsgType === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20']">
      {{ passwordMsg }}
    </div>
    <form @submit.prevent="changePassword" class="space-y-4 max-w-md">
      <div>
        <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.current_password') }}</label>
        <input v-model="passwordForm.current_password" type="password" :placeholder="t('user.current_password_placeholder')"
          class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" required />
      </div>
      <div>
        <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.new_password') }}</label>
        <input v-model="passwordForm.new_password" type="password" :placeholder="t('user.new_password_placeholder')"
          class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" required minlength="8" />
      </div>
      <div>
        <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.confirm_password') }}</label>
        <input v-model="passwordForm.confirm_password" type="password" :placeholder="t('user.confirm_password_placeholder')"
          class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" required />
      </div>
      <div class="pt-2">
        <button type="submit" :disabled="passwordSaving" class="btn btn-primary rounded-xl">
          <span v-if="passwordSaving" class="loading loading-spinner loading-xs"></span>
          <span v-else>🔒</span>
          {{ passwordSaving ? t('admin.settings_saving') : t('user.change_password') }}
        </button>
      </div>
    </form>
  </TechCard>
</template>
