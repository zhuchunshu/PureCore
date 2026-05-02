<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { clearTokens, accessToken } from '../../composables/useAuth'
import { adminOption, adminOptionSet, refreshOptions } from '../../composables/useAdminOption'
import { toastSuccess } from '../../composables/useToast'
import TechCard from '../../components/TechCard.vue'
import GradientButton from '../../components/GradientButton.vue'

const { t } = useI18n()
useSEO({
  title: t('admin.settings'),
  description: t('admin.settings_description'),
})
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const loading = ref(true)
const saving = ref(false)
const error = ref('')

const activeTab = ref(0)

const tabs = computed(() => [
  { key: 'general', label: t('admin.settings_tab_general'), icon: '⚙️', gradient: 'from-purple-500 to-pink-500' },
  { key: 'captcha', label: t('admin.settings_tab_captcha'), icon: '🛡️', gradient: 'from-blue-500 to-cyan-500' },
])

// ===== General settings =====
const generalForm = reactive({
  site_name: '',
  site_description: '',
  site_keywords: '',
  site_logo: '',
  footer_text: '',
})

const generalFields = [
  { key: 'site_name', label: t('admin.settings_site_name'), placeholder: t('admin.settings_site_name_placeholder'), type: 'text', icon: '🏷️', desc: t('admin.settings_site_name_desc') },
  { key: 'site_description', label: t('admin.settings_site_description'), placeholder: t('admin.settings_site_description_placeholder'), type: 'textarea', icon: '📝', desc: t('admin.settings_site_description_desc') },
  { key: 'site_keywords', label: t('admin.settings_site_keywords'), placeholder: t('admin.settings_site_keywords_placeholder'), type: 'text', icon: '🔑', desc: t('admin.settings_site_keywords_desc') },
  { key: 'site_logo', label: t('admin.settings_site_logo'), placeholder: t('admin.settings_site_logo_placeholder'), type: 'text', icon: '🖼️', desc: t('admin.settings_site_logo_desc') },
  { key: 'footer_text', label: t('admin.settings_footer_text'), placeholder: t('admin.settings_footer_text_placeholder'), type: 'textarea', icon: '📄', desc: t('admin.settings_footer_text_desc') },
]

// ===== Captcha settings =====
const captchaForm = reactive({
  turnstile_site_key: '',
  turnstile_secret_key: '',
  turnstile_admin_login: '',
  turnstile_admin_register: '',
  turnstile_public_login: '',
})

const captchaFields = [
  { key: 'turnstile_site_key', label: t('admin.settings_turnstile_site_key'), placeholder: t('admin.settings_turnstile_site_key_placeholder'), type: 'text', icon: '🔑', desc: '' },
  { key: 'turnstile_secret_key', label: t('admin.settings_turnstile_secret_key'), placeholder: t('admin.settings_turnstile_secret_key_placeholder'), type: 'password', icon: '🔒', desc: '' },
]

const captchaToggles = [
  { key: 'turnstile_admin_login', label: t('admin.settings_turnstile_admin_login'), hint: t('admin.settings_turnstile_toggle_hint'), icon: '👤', desc: t('admin.settings_turnstile_admin_login_desc') },
  { key: 'turnstile_admin_register', label: t('admin.settings_turnstile_admin_register'), hint: t('admin.settings_turnstile_toggle_hint'), icon: '📝', desc: t('admin.settings_turnstile_admin_register_desc') },
  { key: 'turnstile_public_login', label: t('admin.settings_turnstile_public_login'), hint: t('admin.settings_turnstile_toggle_hint'), icon: '🌐', desc: t('admin.settings_turnstile_public_login_desc') },
]

onMounted(async () => {
  if (!accessToken.value) {
    router.push(`/${adminPrefix}/login`)
    return
  }

  try {
    generalForm.site_name = await adminOption('site_name', 'PureCore')
    generalForm.site_description = await adminOption('site_description', '')
    generalForm.site_keywords = await adminOption('site_keywords', '')
    generalForm.site_logo = await adminOption('site_logo', '')
    generalForm.footer_text = await adminOption('footer_text', '')

    captchaForm.turnstile_site_key = await adminOption('turnstile_site_key', '')
    captchaForm.turnstile_secret_key = await adminOption('turnstile_secret_key', '')
    captchaForm.turnstile_admin_login = await adminOption('turnstile_admin_login', '0')
    captchaForm.turnstile_admin_register = await adminOption('turnstile_admin_register', '0')
    captchaForm.turnstile_public_login = await adminOption('turnstile_public_login', '0')
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
})

function activeForm() {
  if (activeTab.value === 0) return generalForm
  if (activeTab.value === 1) return captchaForm
  return generalForm
}

function activeFields() {
  if (activeTab.value === 0) return generalFields
  if (activeTab.value === 1) return captchaFields
  return generalFields
}

function activeToggles() {
  if (activeTab.value === 1) return captchaToggles
  return []
}

async function handleSave() {
  saving.value = true
  error.value = ''

  try {
    const form = activeForm()
    const fields = activeFields()
    for (const field of fields) {
      await adminOptionSet(field.key, form[field.key])
    }
    const toggles = activeToggles()
    for (const toggle of toggles) {
      await adminOptionSet(toggle.key, form[toggle.key])
    }
    await refreshOptions()
    toastSuccess(t('admin.settings_saved'))
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-8">
    <!-- Loading skeleton -->
    <template v-if="loading">
      <!-- Header skeleton -->
      <div class="skeleton h-28 rounded-2xl"></div>
      <!-- Tabs skeleton -->
      <div class="flex gap-2 p-1.5">
        <div class="skeleton h-10 w-32 rounded-xl"></div>
        <div class="skeleton h-10 w-32 rounded-xl"></div>
      </div>
      <!-- Form cards skeleton -->
      <div class="grid gap-5 sm:grid-cols-2">
        <div class="skeleton h-44 rounded-2xl"></div>
        <div class="skeleton h-44 rounded-2xl"></div>
        <div class="skeleton h-44 rounded-2xl sm:col-span-2"></div>
        <div class="skeleton h-44 rounded-2xl"></div>
        <div class="skeleton h-44 rounded-2xl sm:col-span-2"></div>
      </div>
      <!-- Save button skeleton -->
      <div class="flex justify-end pt-2">
        <div class="skeleton h-12 w-36 rounded-xl"></div>
      </div>
    </template>

    <!-- Error state -->
    <div v-else-if="error" class="flex items-center justify-center py-20">
      <div class="p-6 bg-error/10 border border-error/20 rounded-2xl text-error max-w-lg flex items-center gap-3 backdrop-blur-sm">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span class="font-medium">{{ error }}</span>
      </div>
    </div>

    <!-- Settings content -->
    <template v-else>
      <!-- Header section -->
      <div class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-purple-500/10 via-fuchsia-500/10 to-pink-500/10 border border-purple-500/10 p-6 md:p-8">
        <div class="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-purple-500/20 to-pink-500/20 rounded-full blur-3xl -translate-y-1/2 translate-x-1/4 pointer-events-none"></div>
        <div class="absolute bottom-0 left-1/3 w-48 h-48 bg-gradient-to-tr from-cyan-500/15 to-blue-500/15 rounded-full blur-3xl translate-y-1/2 pointer-events-none"></div>
        <div class="relative flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 class="text-2xl md:text-3xl font-black tracking-tight">
              <span class="bg-gradient-to-r from-purple-400 via-fuchsia-400 to-pink-400 bg-clip-text text-transparent">{{ t('admin.settings') }}</span>
            </h1>
            <p class="text-base-content/50 mt-2 max-w-lg text-sm md:text-base">{{ t('admin.settings_description') }}</p>
          </div>
          <div class="flex items-center gap-3">
            <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-base-100/50 border border-base-300/20 text-xs text-base-content/50">
              <span class="w-2 h-2 rounded-full bg-success animate-pulse"></span>
              <span v-if="loading">Loading...</span>
              <span v-else>Online</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="tabs tabs-boxed bg-base-100/80 backdrop-blur-sm border border-base-300/20 rounded-xl p-1.5 shadow-sm">
        <button
          v-for="(tab, index) in tabs"
          :key="tab.key"
          :class="['tab tab-lg gap-2.5 transition-all duration-200 font-medium', { 'tab-active': activeTab === index }]"
          @click="activeTab = index"
        >
          <span class="text-lg">{{ tab.icon }}</span>
          <span>{{ tab.label }}</span>
        </button>
      </div>

      <!-- Captcha hint -->
      <div v-if="activeTab === 1" class="flex items-start gap-4 p-5 bg-info/5 border border-info/15 rounded-xl backdrop-blur-sm">
        <div class="flex-shrink-0 w-10 h-10 rounded-xl bg-info/10 flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-info" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        </div>
        <p class="text-sm text-info/80 leading-relaxed">{{ t('admin.settings_turnstile_hint') }}</p>
      </div>

      <!-- Tab content -->
      <div class="grid gap-6">
        <!-- API Keys section (captcha tab) -->
        <template v-if="activeTab === 1">
          <div class="grid gap-5 sm:grid-cols-2">
            <div
              v-for="field in activeFields()"
              :key="field.key"
              class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm hover:shadow-md transition-shadow duration-300 overflow-hidden"
            >
              <div class="card-body p-5">
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center text-xl shadow-sm">
                    {{ field.icon }}
                  </div>
                  <div>
                    <h3 class="font-semibold text-sm">{{ field.label }}</h3>
                  </div>
                </div>
                <input
                  v-model="activeForm()[field.key]"
                  :type="field.type"
                  :placeholder="field.placeholder"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-blue-400 focus:ring-2 focus:ring-blue-400/20 transition-all rounded-xl font-mono text-sm"
                />
              </div>
            </div>
          </div>

          <!-- Toggle switches section -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center text-xl shadow-sm">
                  ⚡
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.settings_turnstile_toggle_section') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.settings_turnstile_toggle_section_desc') }}</p>
                </div>
              </div>
              <div class="divide-y divide-base-300/10">
                <div
                  v-for="toggle in activeToggles()"
                  :key="toggle.key"
                  class="flex items-center justify-between py-4 first:pt-0 last:pb-0"
                >
                  <div class="flex items-center gap-3">
                    <span class="text-lg">{{ toggle.icon }}</span>
                    <div>
                      <p class="font-medium text-sm">{{ toggle.label }}</p>
                      <p v-if="toggle.desc" class="text-xs text-base-content/40">{{ toggle.desc }}</p>
                    </div>
                  </div>
                  <input
                    v-model="activeForm()[toggle.key]"
                    type="checkbox"
                    true-value="1"
                    false-value="0"
                    class="toggle toggle-primary"
                  />
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- General settings -->
        <template v-else>
          <div class="grid gap-5 sm:grid-cols-2">
            <div
              v-for="field in activeFields()"
              :key="field.key"
              :class="[
                'card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm hover:shadow-md transition-all duration-300 overflow-hidden',
                field.type === 'textarea' ? 'sm:col-span-2' : ''
              ]"
            >
              <div class="card-body p-5">
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center text-xl shadow-sm">
                    {{ field.icon }}
                  </div>
                  <div>
                    <h3 class="font-semibold text-sm">{{ field.label }}</h3>
                    <p v-if="field.desc" class="text-xs text-base-content/40">{{ field.desc }}</p>
                  </div>
                </div>
                <textarea
                  v-if="field.type === 'textarea'"
                  v-model="activeForm()[field.key]"
                  :placeholder="field.placeholder"
                  rows="3"
                  class="textarea textarea-bordered w-full bg-base-200/50 border-base-300/30 focus:border-purple-400 focus:ring-2 focus:ring-purple-400/20 transition-all rounded-xl"
                ></textarea>
                <input
                  v-else
                  v-model="activeForm()[field.key]"
                  type="text"
                  :placeholder="field.placeholder"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-purple-400 focus:ring-2 focus:ring-purple-400/20 transition-all rounded-xl"
                />
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Save button -->
      <div class="flex justify-end pt-2">
        <GradientButton
          variant="purple"
          size="lg"
          :loading="saving"
          :disabled="saving"
          @click="handleSave"
          class="shadow-lg shadow-purple-500/20"
        >
          <span v-if="saving">{{ t('admin.settings_saving') }}</span>
          <span v-else class="flex items-center gap-2">💾 {{ t('admin.settings_save') }}</span>
        </GradientButton>
      </div>
    </template>
  </div>
</template>
