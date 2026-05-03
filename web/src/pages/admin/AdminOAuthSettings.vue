<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { adminOption, adminOptionSetMany, refreshOptions } from '../../composables/useAdminOption'
import { toastSuccess } from '../../composables/useToast'
import { adminAPI } from '../../services/api'
import GradientButton from '../../components/GradientButton.vue'

const { t } = useI18n()
useSEO({
  title: t('admin.oauth_settings'),
  description: t('admin.oauth_settings_desc'),
})

const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const loading = ref(true)
const saving = ref(false)
const error = ref('')

// OAuth providers fetched from backend
const providers = ref([])
const activeProvider = ref(0)

// Settings form for each provider: { providerName: { enabled, login_enabled, register_enabled, client_id, client_secret, redirect_url } }
const providerSettings = reactive({})

const providerIcon = (name) => {
  if (name === 'github') return '🐙'
  if (name === 'google') return '🔵'
  return '🔌'
}

const providerGradient = (name) => {
  if (name === 'github') return 'from-gray-500 to-gray-700'
  if (name === 'google') return 'from-blue-500 to-red-500'
  return 'from-purple-500 to-pink-500'
}

onMounted(async () => {
  try {
    // Fetch provider list and settings from admin API (authenticated)
    const resp = await adminAPI.get(`/api/v1/${adminPrefix}/oauth/settings`)
    const json = await resp.json()
    if (json.code === 0) {
      const settings = json.data || {}
      // settings is { github: { enabled, login_enabled, ... }, google: { ... } }
      for (const [name, config] of Object.entries(settings)) {
        providers.value.push({ name, display_name: config.display_name || name })
        providerSettings[name] = { ...config }
      }
    } else {
      error.value = json.message || t('admin.network_error')
    }
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
})

const activeProviderName = computed(() => {
  if (providers.value.length === 0) return ''
  return providers.value[activeProvider.value]?.name || ''
})

const activeSettings = computed(() => {
  const name = activeProviderName.value
  return providerSettings[name] || {}
})

function getSetting(key, fallback = '') {
  const s = activeSettings.value
  return s[key] !== undefined ? s[key] : fallback
}

function setSetting(key, value) {
  const name = activeProviderName.value
  if (providerSettings[name]) {
    providerSettings[name][key] = value
  }
}

async function handleSave() {
  saving.value = true
  error.value = ''
  try {
    const name = activeProviderName.value
    const settings = providerSettings[name]
    if (!name || !settings) return

    // Build the payload matching AdminSetSettings endpoint
    const payload = {
      providers: [
        {
          provider: name,
          settings: {
            enabled: settings.enabled || '0',
            login_enabled: settings.login_enabled || '0',
            register_enabled: settings.register_enabled || '0',
            client_id: settings.client_id || '',
            client_secret: settings.client_secret || '',
            redirect_url: settings.redirect_url || '',
          },
        },
      ],
    }

    const resp = await adminAPI.post(`/api/v1/${adminPrefix}/oauth/settings`, payload)
    const json = await resp.json()
    if (json.code === 0) {
      await refreshOptions()
      toastSuccess(t('admin.settings_saved'))
    } else {
      error.value = json.message || t('admin.settings_save_failed')
    }
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
      <div class="skeleton h-28 rounded-2xl"></div>
      <div class="flex gap-2 p-1.5">
        <div class="skeleton h-10 w-32 rounded-xl"></div>
        <div class="skeleton h-10 w-32 rounded-xl"></div>
      </div>
      <div class="grid gap-5 sm:grid-cols-2">
        <div class="skeleton h-44 rounded-2xl"></div>
        <div class="skeleton h-44 rounded-2xl"></div>
        <div class="skeleton h-44 rounded-2xl sm:col-span-2"></div>
      </div>
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

    <!-- Empty state: no providers registered -->
    <div v-else-if="providers.length === 0" class="flex items-center justify-center py-20">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-base-content/30 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
        </svg>
        <h3 class="text-lg font-semibold text-base-content/50">{{ t('admin.oauth_no_providers') }}</h3>
        <p class="text-sm text-base-content/40 mt-2">{{ t('admin.oauth_no_providers_desc') }}</p>
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
              <span class="bg-gradient-to-r from-purple-400 via-fuchsia-400 to-pink-400 bg-clip-text text-transparent">{{ t('admin.oauth_settings') }}</span>
            </h1>
            <p class="text-base-content/50 mt-2 max-w-lg text-sm md:text-base">{{ t('admin.oauth_settings_desc') }}</p>
          </div>
        </div>
      </div>

      <!-- Provider tabs -->
      <div class="tabs tabs-boxed bg-base-100/80 backdrop-blur-sm border border-base-300/20 rounded-xl p-1.5 shadow-sm">
        <button
          v-for="(provider, index) in providers"
          :key="provider.name"
          :class="['tab tab-lg gap-2.5 transition-all duration-200 font-medium', { 'tab-active': activeProvider === index }]"
          @click="activeProvider = index"
        >
          <span class="text-lg">{{ providerIcon(provider.name) }}</span>
          <span>{{ provider.display_name }}</span>
        </button>
      </div>

      <!-- Provider settings form -->
      <div v-if="activeProviderName" class="grid gap-6">
        <!-- Toggle switches -->
        <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
          <div class="card-body p-5">
            <div class="flex items-center gap-3 mb-5">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center text-xl shadow-sm">
                ⚡
              </div>
              <div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
                <p class="text-xs text-base-content/40">{{ t('admin.oauth_toggle_section_desc') }}</p>
              </div>
            </div>
            <div class="divide-y divide-base-300/10">
              <!-- Enabled -->
              <div class="flex items-center justify-between py-4 first:pt-0">
                <div class="flex items-center gap-3">
                  <span class="text-lg">🔌</span>
                  <div>
                    <p class="font-medium text-sm">{{ t('admin.oauth_enabled') }}</p>
                    <p class="text-xs text-base-content/40">{{ t('admin.oauth_enabled_desc') }}</p>
                  </div>
                </div>
                <input
                  :checked="getSetting('enabled') === '1'"
                  @change="setSetting('enabled', ($event.target.checked ? '1' : '0'))"
                  type="checkbox"
                  class="toggle toggle-primary"
                />
              </div>
              <!-- Login enabled -->
              <div class="flex items-center justify-between py-4">
                <div class="flex items-center gap-3">
                  <span class="text-lg">🔑</span>
                  <div>
                    <p class="font-medium text-sm">{{ t('admin.oauth_login_enabled') }}</p>
                    <p class="text-xs text-base-content/40">{{ t('admin.oauth_login_enabled_desc') }}</p>
                  </div>
                </div>
                <input
                  :checked="getSetting('login_enabled') === '1'"
                  @change="setSetting('login_enabled', ($event.target.checked ? '1' : '0'))"
                  type="checkbox"
                  class="toggle toggle-secondary"
                />
              </div>
              <!-- Register enabled -->
              <div class="flex items-center justify-between py-4 last:pb-0">
                <div class="flex items-center gap-3">
                  <span class="text-lg">📝</span>
                  <div>
                    <p class="font-medium text-sm">{{ t('admin.oauth_register_enabled') }}</p>
                    <p class="text-xs text-base-content/40">{{ t('admin.oauth_register_enabled_desc') }}</p>
                  </div>
                </div>
                <input
                  :checked="getSetting('register_enabled') === '1'"
                  @change="setSetting('register_enabled', ($event.target.checked ? '1' : '0'))"
                  type="checkbox"
                  class="toggle toggle-accent"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- API Credentials -->
        <div class="grid gap-5 sm:grid-cols-2">
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm hover:shadow-md transition-shadow duration-300 overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center text-xl shadow-sm">
                  🆔
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_client_id') }}</h3>
                </div>
              </div>
              <input
                :value="getSetting('client_id')"
                @input="setSetting('client_id', $event.target.value)"
                type="text"
                :placeholder="t('admin.oauth_client_id_placeholder')"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-blue-400 focus:ring-2 focus:ring-blue-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm hover:shadow-md transition-shadow duration-300 overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500/20 to-orange-500/20 flex items-center justify-center text-xl shadow-sm">
                  🔒
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_client_secret') }}</h3>
                </div>
              </div>
              <input
                :value="getSetting('client_secret')"
                @input="setSetting('client_secret', $event.target.value)"
                type="password"
                :placeholder="t('admin.oauth_client_secret_placeholder')"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-red-400 focus:ring-2 focus:ring-red-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm hover:shadow-md transition-shadow duration-300 overflow-hidden sm:col-span-2">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center text-xl shadow-sm">
                  🔗
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_redirect_url') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>
              <input
                :value="getSetting('redirect_url')"
                @input="setSetting('redirect_url', $event.target.value)"
                type="text"
                :placeholder="t('admin.oauth_redirect_url_placeholder')"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
              <p class="text-xs text-base-content/40 mt-2">
                {{ t('admin.oauth_redirect_url_hint', { provider: activeProviderName }) }}
              </p>
            </div>
          </div>
        </div>
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
