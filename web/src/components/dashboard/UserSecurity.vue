<script setup>
import { ref, computed } from 'vue'
import { useI18n } from '../../i18n'
import { accessToken } from '../../composables/useUserAuth'
import { Lock, Save, Eye, EyeOff, CircleCheck, CircleAlert, ShieldCheck, Key, AlertTriangle } from 'lucide-vue-next'

const { t } = useI18n()

const passwordForm = ref({
  current_password: '',
  new_password: '',
  confirm_password: '',
})
const passwordSaving = ref(false)
const passwordMsg = ref('')
const passwordMsgType = ref('success')

const showCurrent = ref(false)
const showNew = ref(false)
const showConfirm = ref(false)

const strengthLevel = computed(() => {
  const len = passwordForm.value.new_password.length
  if (len < 8) return 0
  if (len < 10) return 1
  if (len < 12) return 2
  return 3
})

const strengthLabel = computed(() => {
  const labels = ['', 'Weak', 'Fair', 'Strong']
  return labels[strengthLevel.value] || ''
})

const strengthColor = computed(() => {
  const colors = ['', 'text-red-400', 'text-amber-400', 'text-emerald-400']
  return colors[strengthLevel.value] || ''
})

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

const inputClass = 'w-full px-4 py-3 pr-12 bg-base-200 border border-base-300/20 rounded-xl text-sm text-base-content placeholder:text-base-content/25 focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10 focus:outline-none transition-all duration-200'
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- ===== COL 1-2: Form ===== -->
    <div class="lg:col-span-2">
      <div class="bg-base-100 border border-base-300/20 rounded-2xl shadow-sm overflow-hidden">
        <div class="p-6">
          <div v-if="passwordMsg" :class="[
            'flex items-center gap-2 px-4 py-3 rounded-xl mb-5 text-sm font-medium',
            passwordMsgType === 'success'
              ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20'
              : 'bg-red-500/10 text-red-400 border border-red-500/20'
          ]">
            <CircleCheck v-if="passwordMsgType === 'success'" :size="16" />
            <CircleAlert v-else :size="16" />
            <span>{{ passwordMsg }}</span>
          </div>

          <form @submit.prevent="changePassword" class="space-y-5 max-w-lg">
            <div>
              <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.current_password') }}</label>
              <div class="relative">
                <input
                  v-model="passwordForm.current_password"
                  :type="showCurrent ? 'text' : 'password'"
                  :placeholder="t('user.current_password_placeholder')"
                  :class="inputClass"
                  required
                  autocomplete="current-password"
                />
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-base-content/25 hover:text-base-content/50 transition-colors cursor-pointer"
                  @click="showCurrent = !showCurrent"
                >
                  <EyeOff v-if="showCurrent" :size="18" />
                  <Eye v-else :size="18" />
                </button>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.new_password') }}</label>
              <div class="relative">
                <input
                  v-model="passwordForm.new_password"
                  :type="showNew ? 'text' : 'password'"
                  :placeholder="t('user.new_password_placeholder')"
                  :class="inputClass"
                  required
                  minlength="8"
                  autocomplete="new-password"
                />
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-base-content/25 hover:text-base-content/50 transition-colors cursor-pointer"
                  @click="showNew = !showNew"
                >
                  <EyeOff v-if="showNew" :size="18" />
                  <Eye v-else :size="18" />
                </button>
              </div>
              <div v-if="passwordForm.new_password.length > 0" class="mt-2 flex items-center gap-2">
                <div class="flex gap-1.5 flex-1 max-w-[200px]">
                  <div
                    v-for="i in 3"
                    :key="i"
                    :class="[
                      'h-1.5 flex-1 rounded-full transition-all duration-300',
                      i <= strengthLevel
                        ? i === 1 ? 'bg-red-400' : i === 2 ? 'bg-amber-400' : 'bg-emerald-400'
                        : 'bg-base-300/40'
                    ]"
                  />
                </div>
                <span :class="['text-xs font-medium', strengthColor]">{{ strengthLabel }}</span>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.confirm_password') }}</label>
              <div class="relative">
                <input
                  v-model="passwordForm.confirm_password"
                  :type="showConfirm ? 'text' : 'password'"
                  :placeholder="t('user.confirm_password_placeholder')"
                  :class="inputClass"
                  required
                  autocomplete="new-password"
                />
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-base-content/25 hover:text-base-content/50 transition-colors cursor-pointer"
                  @click="showConfirm = !showConfirm"
                >
                  <EyeOff v-if="showConfirm" :size="18" />
                  <Eye v-else :size="18" />
                </button>
              </div>
            </div>

            <div class="flex pt-2">
              <button
                type="submit"
                :disabled="passwordSaving"
                class="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-content rounded-xl px-6 py-3 font-semibold text-sm shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:hover:shadow-primary/20"
              >
                <span v-if="passwordSaving" class="loading loading-spinner loading-xs"></span>
                <Lock v-else :size="16" />
                {{ passwordSaving ? t('admin.settings_saving') : t('user.change_password') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- ===== COL 3: Sidebar ===== -->
    <div class="space-y-4">
      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
          <ShieldCheck :size="16" class="text-emerald-500" />
          {{ t('user.password_tips_title') }}
        </h3>
        <ul class="space-y-2 text-xs text-base-content/50">
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-emerald-500/60 mt-1.5 shrink-0"></span>
            {{ t('user.password_tip_length') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-emerald-500/60 mt-1.5 shrink-0"></span>
            {{ t('user.password_tip_mix') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-emerald-500/60 mt-1.5 shrink-0"></span>
            {{ t('user.password_tip_reuse') }}
          </li>
        </ul>
      </div>

      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
          <AlertTriangle :size="16" class="text-amber-400" />
          {{ t('user.session_security_title') }}
        </h3>
        <p class="text-xs text-base-content/50 leading-relaxed">
          {{ t('user.session_security_desc') }}
        </p>
      </div>
    </div>
  </div>
</template>
