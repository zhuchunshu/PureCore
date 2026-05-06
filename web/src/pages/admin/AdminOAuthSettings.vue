<script setup>
/**
 * Admin OAuth Settings Page
 *
 * Renders a per-provider tabbed interface for configuring third-party OAuth login.
 * Each provider has a dedicated template with setup guide, credentials form, and
 * toggle switches — tailored to the provider's official documentation requirements.
 */
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { toastSuccess } from '../../composables/useToast'
import { adminAPI } from '../../services/api'
import GradientButton from '../../components/GradientButton.vue'
import {
  IconBrandGithub,
  IconBrandGoogle,
  IconBrandApple,
  IconBrandTelegram,
  IconBrandDiscord,
  IconExternalLink,
  IconCopy,
  IconCheck,
  IconPlugConnected,
  IconBook,
  IconLink,
  IconKey,
  IconLock,
  IconId,
  IconSettings,
} from '@tabler/icons-vue'

const { t } = useI18n()
useSEO({
  title: t('admin.oauth_settings'),
  description: t('admin.oauth_settings_desc'),
})

const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const copiedCallback = ref('')

const providers = ref([])
const activeProviderIndex = ref(0)
const pendingTabIndex = ref(null)
const providerValues = reactive({})
const originalValues = reactive({})
const dirtyProviders = reactive(new Set())
const showUnsavedDialog = ref(false)

const providerIconMap = {
  github: IconBrandGithub,
  google: IconBrandGoogle,
  apple: IconBrandApple,
  telegram: IconBrandTelegram,
  discord: IconBrandDiscord,
}
function getProviderIcon(name) {
  return providerIconMap[name] || IconPlugConnected
}

const activeProvider = computed(() => providers.value[activeProviderIndex.value] || null)
const activeProviderName = computed(() => activeProvider.value?.name || '')
const isCurrentDirty = computed(() => dirtyProviders.has(activeProviderName.value))

function getValue(key, fallback = '') {
  const name = activeProviderName.value
  if (!name || !providerValues[name]) return fallback
  return providerValues[name][key] !== undefined && providerValues[name][key] !== null
    ? providerValues[name][key] : fallback
}

function setValue(key, value) {
  const name = activeProviderName.value
  if (name && providerValues[name]) {
    providerValues[name][key] = value
    checkDirty(name)
  }
}

function handleToggleChange(key, event) {
  setValue(key, event.target.checked ? '1' : '0')
}

function checkDirty(name) {
  const current = providerValues[name] || {}
  const original = originalValues[name] || {}
  const keys = new Set([...Object.keys(current), ...Object.keys(original)])
  let isDirty = false
  for (const key of keys) {
    const cv = current[key] !== undefined && current[key] !== null ? String(current[key]) : ''
    const ov = original[key] !== undefined && original[key] !== null ? String(original[key]) : ''
    if (cv !== ov) { isDirty = true; break }
  }
  if (isDirty) dirtyProviders.add(name)
  else dirtyProviders.delete(name)
}

function snapshotOriginal(name) {
  const vals = providerValues[name]
  if (!vals) return
  originalValues[name] = { ...vals }
  dirtyProviders.delete(name)
}

function copyCallbackUrl() {
  const vals = providerValues[activeProviderName.value] || {}
  const url = vals._recommended_redirect_url || vals.redirect_url
  if (!url || typeof navigator === 'undefined') return
  navigator.clipboard.writeText(url).then(() => {
    copiedCallback.value = activeProviderName.value
    setTimeout(() => { copiedCallback.value = '' }, 2000)
  }).catch(() => {})
}

function switchTab(index) {
  if (activeProviderIndex.value === index) return
  if (isCurrentDirty.value) {
    pendingTabIndex.value = index
    showUnsavedDialog.value = true
    return
  }
  activeProviderIndex.value = index
}

function confirmDiscardAndSwitch() {
  const name = activeProviderName.value
  if (name) {
    providerValues[name] = { ...originalValues[name] }
    dirtyProviders.delete(name)
  }
  showUnsavedDialog.value = false
  if (pendingTabIndex.value !== null) {
    activeProviderIndex.value = pendingTabIndex.value
    pendingTabIndex.value = null
  }
}

function cancelSwitch() {
  showUnsavedDialog.value = false
  pendingTabIndex.value = null
}

function getConfigFields() {
  return activeProvider.value?.config_fields || []
}

async function handleSave() {
  saving.value = true
  error.value = ''
  const provider = activeProvider.value
  if (!provider) return
  try {
    const vals = providerValues[provider.name] || {}
    const fields = getConfigFields()
    for (const field of fields) {
      if (field.required) {
        const val = vals[field.key]
        if (val === undefined || val === null || String(val).trim() === '') {
          error.value = `${fieldLabel(field)} ${t('admin.is_required') || 'is required'}`
          saving.value = false
          return
        }
      }
    }
    const settings = {}
    for (const field of fields) {
      settings[field.key] = vals[field.key] !== undefined && vals[field.key] !== null ? String(vals[field.key]) : ''
    }
    const payload = { providers: [{ provider: provider.name, settings }] }
    const resp = await adminAPI.post(`/api/v1/${adminPrefix}/oauth/settings`, payload)
    const json = await resp.json()
    if (json.code === 0) {
      snapshotOriginal(provider.name)
      toastSuccess(t('admin.settings_saved'))
    } else {
      error.value = json.message || t('admin.oauth_settings_save_failed')
    }
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    const resp = await adminAPI.get(`/api/v1/${adminPrefix}/oauth/settings`)
    const json = await resp.json()
    if (json.code === 0) {
      const list = json.data || []
      for (const item of list) {
        providers.value.push(item)
        providerValues[item.name] = { ...item.values }
        originalValues[item.name] = { ...item.values }
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

function fieldLabel(field) {
  const label = field.label || ''
  if (label.includes('.')) {
    const translated = t(label)
    if (translated !== label) return translated
  }
  return label
}

function fieldHelp(field) {
  const help = field.help || ''
  if (!help) return ''
  if (help.includes('.')) {
    const translated = t(help)
    if (translated !== help) return translated
  }
  return help
}

// Get recommended callback URL for the active provider
function recommendedCallback() {
  return getValue('_recommended_redirect_url') || getValue('redirect_url')
}
</script>

<template>
  <div class="space-y-8">
    <!-- Loading skeleton -->
    <template v-if="loading">
      <div class="skeleton h-28 rounded-2xl"></div>
      <div class="flex gap-2 p-1.5">
        <div v-for="i in 5" :key="i" class="skeleton h-10 w-32 rounded-xl"></div>
      </div>
      <div class="skeleton h-48 rounded-2xl"></div>
      <div class="grid gap-5 sm:grid-cols-2">
        <div v-for="i in 4" :key="i" class="skeleton h-32 rounded-2xl"></div>
      </div>
      <div class="flex justify-end pt-2">
        <div class="skeleton h-12 w-36 rounded-xl"></div>
      </div>
    </template>

    <!-- Error state -->
    <div v-else-if="error && providers.length === 0" class="flex items-center justify-center py-20">
      <div class="p-6 bg-error/10 border border-error/20 rounded-2xl text-error max-w-lg flex items-center gap-3 backdrop-blur-sm">
        <IconPlugConnected :size="24" class="shrink-0" />
        <span class="font-medium">{{ error }}</span>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="providers.length === 0" class="flex items-center justify-center py-20">
      <div class="text-center">
        <IconPlugConnected :size="64" class="mx-auto text-base-content/30 mb-4" />
        <h3 class="text-lg font-semibold text-base-content/50">{{ t('admin.oauth_no_providers') }}</h3>
        <p class="text-sm text-base-content/40 mt-2">{{ t('admin.oauth_no_providers_desc') }}</p>
      </div>
    </div>

    <!-- Settings content -->
    <template v-else>
      <!-- Header -->
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
      <div class="tabs tabs-boxed bg-base-100/80 backdrop-blur-sm border border-base-300/20 rounded-xl p-1.5 shadow-sm overflow-x-auto flex-nowrap">
        <button
          v-for="(provider, index) in providers"
          :key="provider.name"
          :class="[
            'tab tab-lg gap-2.5 transition-all duration-300 font-medium flex-shrink-0 relative rounded-lg',
            {
              'tab-active !bg-primary/15 !text-primary shadow-sm ring-1 ring-primary/30 scale-[1.02]': activeProviderIndex === index,
              'hover:bg-base-200/50': activeProviderIndex !== index
            }
          ]"
          @click="switchTab(index)"
        >
          <component :is="getProviderIcon(provider.name)" :size="20" />
          <span>{{ provider.display_name }}</span>
          <span
            v-if="dirtyProviders.has(provider.name)"
            class="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-warning animate-pulse"
            :title="t('admin.unsaved_changes')"
          ></span>
        </button>
      </div>

      <!-- Unsaved changes dialog -->
      <div v-if="showUnsavedDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
        <div class="card bg-base-100 border border-base-300/20 shadow-2xl p-6 max-w-sm w-full mx-4 animate-in fade-in zoom-in duration-200">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-xl bg-warning/20 flex items-center justify-center">
              <IconSettings :size="20" class="text-warning" />
            </div>
            <div>
              <h3 class="font-semibold text-base-content">{{ t('admin.unsaved_changes_title') }}</h3>
              <p class="text-sm text-base-content/50">{{ t('admin.unsaved_changes_desc') }}</p>
            </div>
          </div>
          <div class="flex gap-3 justify-end">
            <button class="btn btn-ghost btn-sm" @click="cancelSwitch">{{ t('admin.cancel') }}</button>
            <button class="btn btn-warning btn-sm" @click="confirmDiscardAndSwitch">{{ t('admin.discard_changes') }}</button>
          </div>
        </div>
      </div>

      <!-- ====== PER-PROVIDER SETTINGS ====== -->
      <div v-if="activeProvider" class="space-y-6">

        <!-- ========== GITHUB ========== -->
        <template v-if="activeProviderName === 'github'">
          <!-- Setup guide -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gray-800 flex items-center justify-center shadow-sm">
                  <IconBrandGithub :size="20" class="text-white" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_setup_title') }} — GitHub</h3>
              </div>
              <div class="space-y-3 text-sm text-base-content/70">
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">1</span>
                  <span>
                    Go to
                    <a v-if="activeProvider.apply_url" :href="activeProvider.apply_url" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">GitHub Developer Settings</a>
                    to create a new OAuth App
                  </span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">2</span>
                  <span>Set <strong>Homepage URL</strong> to your site URL</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">3</span>
                  <span>Set <strong>Authorization callback URL</strong> to the recommended callback below</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">4</span>
                  <span>Copy the <strong>Client ID</strong> and generate a <strong>Client Secret</strong></span>
                </div>
              </div>
              <div class="mt-4 flex items-center gap-2 px-3 py-2 bg-base-200/50 rounded-lg text-xs text-base-content/60">
                <IconKey :size="14" />
                <span>Scopes:</span>
                <code class="text-primary/70">read:user, user:email</code>
              </div>
              <div class="flex gap-3 mt-4">
                <a v-if="activeProvider.doc_url" :href="activeProvider.doc_url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary">
                  <IconBook :size="16" /> {{ t('admin.oauth_documentation') }} <IconExternalLink :size="12" />
                </a>
                <a v-if="activeProvider.apply_url" :href="activeProvider.apply_url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary">
                  <IconExternalLink :size="16" /> {{ t('admin.oauth_apply') }}
                </a>
              </div>
            </div>
          </div>

          <!-- Redirect URL -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center shadow-sm">
                  <IconLink :size="20" class="text-green-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_redirect_url') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2 mb-3 px-3 py-2 bg-success/5 border border-success/10 rounded-lg">
                <span class="text-xs text-success font-medium shrink-0">{{ t('admin.oauth_recommended_callback') }}:</span>
                <code class="text-xs text-success/80 break-all">{{ recommendedCallback() }}</code>
                <button class="btn btn-ghost btn-xs shrink-0 ml-auto" @click="copyCallbackUrl" :title="t('user.copy')">
                  <IconCheck v-if="copiedCallback === 'github'" :size="14" class="text-success" />
                  <IconCopy v-else :size="14" />
                </button>
              </div>
              <input
                :value="getValue('redirect_url')"
                @input="setValue('redirect_url', $event.target.value)"
                type="text"
                :placeholder="'https://your-domain.com/oauth/github/callback'"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <!-- Credentials -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center shadow-sm">
                  <IconLock :size="20" class="text-blue-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_credentials_section') }}</h3>
                  <p class="text-xs text-base-content/40">GitHub OAuth App credentials</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_id') }} <span class="text-error">*</span>
                  </label>
                  <input
                    :value="getValue('client_id')"
                    @input="setValue('client_id', $event.target.value)"
                    type="text"
                    placeholder="Iv23li..."
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm"
                  />
                </div>
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_secret') }} <span class="text-error">*</span>
                  </label>
                  <input
                    :value="getValue('client_secret')"
                    @input="setValue('client_secret', $event.target.value)"
                    type="password"
                    placeholder="••••••••••••••••"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Toggles -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center shadow-sm">
                  <IconSettings :size="20" class="text-emerald-400" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
              </div>
              <div class="divide-y divide-base-300/10">
                <div v-for="(field, idx) in getConfigFields().filter(f => f.type === 'toggle')" :key="field.key"
                  class="flex items-center justify-between py-4" :class="{ 'pt-0': idx === 0, 'pb-0': idx === getConfigFields().filter(f => f.type === 'toggle').length - 1 }">
                  <div class="flex items-center gap-3">
                    <span class="text-lg">
                      <IconPlugConnected v-if="field.key === 'enabled'" :size="20" />
                      <IconKey v-else-if="field.key === 'login_enabled'" :size="20" />
                      <IconId v-else-if="field.key === 'register_enabled'" :size="20" />
                      <IconSettings v-else :size="20" />
                    </span>
                    <div>
                      <p class="font-medium text-sm">{{ fieldLabel(field) }}</p>
                      <p class="text-xs text-base-content/40" v-if="fieldHelp(field)">{{ fieldHelp(field) }}</p>
                    </div>
                  </div>
                  <input
                    :checked="getValue(field.key) === '1'"
                    @change="handleToggleChange(field.key, $event)"
                    type="checkbox"
                    :class="['toggle', {
                      'toggle-primary': field.key === 'enabled',
                      'toggle-secondary': field.key === 'login_enabled',
                      'toggle-accent': field.key === 'register_enabled'
                    }]"
                  />
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ========== GOOGLE ========== -->
        <template v-if="activeProviderName === 'google'">
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-white flex items-center justify-center shadow-sm border">
                  <IconBrandGoogle :size="20" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_setup_title') }} — Google</h3>
              </div>
              <div class="space-y-3 text-sm text-base-content/70">
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">1</span>
                  <span>Go to <a v-if="activeProvider.apply_url" :href="activeProvider.apply_url" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">Google Cloud Console</a> and create an OAuth 2.0 Client ID</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">2</span>
                  <span>Configure the <strong>OAuth consent screen</strong> with your app details</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">3</span>
                  <span>Add the recommended callback below as an <strong>Authorized redirect URI</strong></span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">4</span>
                  <span>Copy the <strong>Client ID</strong> and <strong>Client Secret</strong></span>
                </div>
              </div>
              <div class="mt-4 flex items-center gap-2 px-3 py-2 bg-base-200/50 rounded-lg text-xs text-base-content/60">
                <IconKey :size="14" />
                <span>Scopes:</span>
                <code class="text-primary/70">openid, profile, email</code>
              </div>
              <div class="flex gap-3 mt-4">
                <a v-if="activeProvider.doc_url" :href="activeProvider.doc_url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary">
                  <IconBook :size="16" /> {{ t('admin.oauth_documentation') }} <IconExternalLink :size="12" />
                </a>
              </div>
            </div>
          </div>

          <!-- Redirect URL -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center shadow-sm">
                  <IconLink :size="20" class="text-green-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_redirect_url') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2 mb-3 px-3 py-2 bg-success/5 border border-success/10 rounded-lg">
                <span class="text-xs text-success font-medium shrink-0">{{ t('admin.oauth_recommended_callback') }}:</span>
                <code class="text-xs text-success/80 break-all">{{ recommendedCallback() }}</code>
                <button class="btn btn-ghost btn-xs shrink-0 ml-auto" @click="copyCallbackUrl" :title="t('user.copy')">
                  <IconCheck v-if="copiedCallback === 'google'" :size="14" class="text-success" />
                  <IconCopy v-else :size="14" />
                </button>
              </div>
              <input
                :value="getValue('redirect_url')"
                @input="setValue('redirect_url', $event.target.value)"
                type="text"
                :placeholder="'https://your-domain.com/oauth/google/callback'"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <!-- Credentials -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center shadow-sm">
                  <IconLock :size="20" class="text-blue-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_credentials_section') }}</h3>
                  <p class="text-xs text-base-content/40">Google OAuth 2.0 credentials</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_id') }} <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('client_id')" @input="setValue('client_id', $event.target.value)" type="text"
                    placeholder="123456789-xxxxx.apps.googleusercontent.com"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm" />
                </div>
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_secret') }} <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('client_secret')" @input="setValue('client_secret', $event.target.value)" type="password"
                    placeholder="GOCSPX-••••••••••••••••"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm" />
                </div>
              </div>
            </div>
          </div>

          <!-- Toggles -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center shadow-sm">
                  <IconSettings :size="20" class="text-emerald-400" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
              </div>
              <div class="divide-y divide-base-300/10">
                <div v-for="(field, idx) in getConfigFields().filter(f => f.type === 'toggle')" :key="field.key"
                  class="flex items-center justify-between py-4" :class="{ 'pt-0': idx === 0, 'pb-0': idx === getConfigFields().filter(f => f.type === 'toggle').length - 1 }">
                  <div class="flex items-center gap-3">
                    <span class="text-lg">
                      <IconPlugConnected v-if="field.key === 'enabled'" :size="20" />
                      <IconKey v-else-if="field.key === 'login_enabled'" :size="20" />
                      <IconId v-else-if="field.key === 'register_enabled'" :size="20" />
                      <IconSettings v-else :size="20" />
                    </span>
                    <div>
                      <p class="font-medium text-sm">{{ fieldLabel(field) }}</p>
                      <p class="text-xs text-base-content/40" v-if="fieldHelp(field)">{{ fieldHelp(field) }}</p>
                    </div>
                  </div>
                  <input :checked="getValue(field.key) === '1'" @change="handleToggleChange(field.key, $event)" type="checkbox"
                    :class="['toggle', {
                      'toggle-primary': field.key === 'enabled',
                      'toggle-secondary': field.key === 'login_enabled',
                      'toggle-accent': field.key === 'register_enabled'
                    }]" />
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ========== DISCORD ========== -->
        <template v-if="activeProviderName === 'discord'">
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-[#5865F2] flex items-center justify-center shadow-sm">
                  <IconBrandDiscord :size="20" class="text-white" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_setup_title') }} — Discord</h3>
              </div>
              <div class="space-y-3 text-sm text-base-content/70">
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">1</span>
                  <span>Go to <a v-if="activeProvider.apply_url" :href="activeProvider.apply_url" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">Discord Developer Portal</a> and create a new application</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">2</span>
                  <span>Navigate to the <strong>OAuth2</strong> settings page</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">3</span>
                  <span>Add the recommended callback below under <strong>Redirects</strong></span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">4</span>
                  <span>Copy the <strong>Client ID</strong> and <strong>Client Secret</strong></span>
                </div>
              </div>
              <div class="mt-4 flex items-center gap-2 px-3 py-2 bg-base-200/50 rounded-lg text-xs text-base-content/60">
                <IconKey :size="14" />
                <span>Scopes:</span>
                <code class="text-primary/70">identify, email</code>
              </div>
              <div class="flex gap-3 mt-4">
                <a v-if="activeProvider.doc_url" :href="activeProvider.doc_url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary">
                  <IconBook :size="16" /> {{ t('admin.oauth_documentation') }} <IconExternalLink :size="12" />
                </a>
              </div>
            </div>
          </div>

          <!-- Redirect URL -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center shadow-sm">
                  <IconLink :size="20" class="text-green-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_redirect_url') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2 mb-3 px-3 py-2 bg-success/5 border border-success/10 rounded-lg">
                <span class="text-xs text-success font-medium shrink-0">{{ t('admin.oauth_recommended_callback') }}:</span>
                <code class="text-xs text-success/80 break-all">{{ recommendedCallback() }}</code>
                <button class="btn btn-ghost btn-xs shrink-0 ml-auto" @click="copyCallbackUrl" :title="t('user.copy')">
                  <IconCheck v-if="copiedCallback === 'discord'" :size="14" class="text-success" />
                  <IconCopy v-else :size="14" />
                </button>
              </div>
              <input
                :value="getValue('redirect_url')"
                @input="setValue('redirect_url', $event.target.value)"
                type="text"
                :placeholder="'https://your-domain.com/oauth/discord/callback'"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <!-- Credentials -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center shadow-sm">
                  <IconLock :size="20" class="text-blue-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_credentials_section') }}</h3>
                  <p class="text-xs text-base-content/40">Discord OAuth2 credentials</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_id') }} <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('client_id')" @input="setValue('client_id', $event.target.value)" type="text"
                    placeholder="123456789012345678"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm" />
                </div>
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_secret') }} <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('client_secret')" @input="setValue('client_secret', $event.target.value)" type="password"
                    placeholder="••••••••••••••••••••••••"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm" />
                </div>
              </div>
            </div>
          </div>

          <!-- Toggles -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center shadow-sm">
                  <IconSettings :size="20" class="text-emerald-400" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
              </div>
              <div class="divide-y divide-base-300/10">
                <div v-for="(field, idx) in getConfigFields().filter(f => f.type === 'toggle')" :key="field.key"
                  class="flex items-center justify-between py-4" :class="{ 'pt-0': idx === 0, 'pb-0': idx === getConfigFields().filter(f => f.type === 'toggle').length - 1 }">
                  <div class="flex items-center gap-3">
                    <span class="text-lg">
                      <IconPlugConnected v-if="field.key === 'enabled'" :size="20" />
                      <IconKey v-else-if="field.key === 'login_enabled'" :size="20" />
                      <IconId v-else-if="field.key === 'register_enabled'" :size="20" />
                      <IconSettings v-else :size="20" />
                    </span>
                    <div>
                      <p class="font-medium text-sm">{{ fieldLabel(field) }}</p>
                      <p class="text-xs text-base-content/40" v-if="fieldHelp(field)">{{ fieldHelp(field) }}</p>
                    </div>
                  </div>
                  <input :checked="getValue(field.key) === '1'" @change="handleToggleChange(field.key, $event)" type="checkbox"
                    :class="['toggle', {
                      'toggle-primary': field.key === 'enabled',
                      'toggle-secondary': field.key === 'login_enabled',
                      'toggle-accent': field.key === 'register_enabled'
                    }]" />
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ========== APPLE ========== -->
        <template v-if="activeProviderName === 'apple'">
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-black flex items-center justify-center shadow-sm">
                  <IconBrandApple :size="20" class="text-white" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_setup_title') }} — Sign in with Apple</h3>
              </div>
              <div class="mb-4 px-3 py-2 bg-info/5 border border-info/10 rounded-lg text-xs text-base-content/70 flex items-start gap-2">
                <IconKey :size="14" class="text-info shrink-0 mt-0.5" />
                <span>Apple does not use a traditional Client Secret. The system auto-generates a JWT client secret from your private key (.p8).</span>
              </div>
              <div class="space-y-3 text-sm text-base-content/70">
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">1</span>
                  <span>Go to <a href="https://developer.apple.com/account/resources/identifiers/list/serviceId" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">Apple Developer</a> and create a <strong>Services ID</strong></span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">2</span>
                  <span>Enable <strong>Sign in with Apple</strong> on the Services ID and configure the callback domain</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">3</span>
                  <span>Go to <a href="https://developer.apple.com/account/resources/authkeys/list" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">Keys</a> to create a private key and download the <strong>.p8</strong> file</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">4</span>
                  <span>Get your <strong>Team ID</strong> from <a href="https://developer.apple.com/account" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">Membership</a></span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">5</span>
                  <span>Fill in <strong>Services ID</strong> (as Client ID), <strong>Team ID</strong>, <strong>Key ID</strong>, and <strong>Private Key</strong></span>
                </div>
              </div>
              <div class="mt-4 flex items-center gap-2 px-3 py-2 bg-base-200/50 rounded-lg text-xs text-base-content/60">
                <IconKey :size="14" />
                <span>Scopes:</span>
                <code class="text-primary/70">name, email</code>
              </div>
              <div class="flex gap-3 mt-4">
                <a v-if="activeProvider.doc_url" :href="activeProvider.doc_url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary">
                  <IconBook :size="16" /> {{ t('admin.oauth_documentation') }} <IconExternalLink :size="12" />
                </a>
              </div>
            </div>
          </div>

          <!-- Redirect URL -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center shadow-sm">
                  <IconLink :size="20" class="text-green-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_redirect_url') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2 mb-3 px-3 py-2 bg-success/5 border border-success/10 rounded-lg">
                <span class="text-xs text-success font-medium shrink-0">{{ t('admin.oauth_recommended_callback') }}:</span>
                <code class="text-xs text-success/80 break-all">{{ recommendedCallback() }}</code>
                <button class="btn btn-ghost btn-xs shrink-0 ml-auto" @click="copyCallbackUrl" :title="t('user.copy')">
                  <IconCheck v-if="copiedCallback === 'apple'" :size="14" class="text-success" />
                  <IconCopy v-else :size="14" />
                </button>
              </div>
              <input
                :value="getValue('redirect_url')"
                @input="setValue('redirect_url', $event.target.value)"
                type="text"
                :placeholder="'https://your-domain.com/oauth/apple/callback'"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <!-- Apple-specific credentials -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center shadow-sm">
                  <IconLock :size="20" class="text-blue-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_credentials_section') }}</h3>
                  <p class="text-xs text-base-content/40">Apple Sign In credentials</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_client_id') }} (Services ID) <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('client_id')" @input="setValue('client_id', $event.target.value)" type="text"
                    placeholder="com.example.app"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 transition-all rounded-xl font-mono text-sm" />
                </div>
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_apple_team_id') }} <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('team_id')" @input="setValue('team_id', $event.target.value)" type="text"
                    placeholder="ABCDEF1234"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 transition-all rounded-xl font-mono text-sm" />
                </div>
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_apple_key_id') }} <span class="text-error">*</span>
                  </label>
                  <input :value="getValue('key_id')" @input="setValue('key_id', $event.target.value)" type="text"
                    placeholder="ABCDEF1234"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 transition-all rounded-xl font-mono text-sm" />
                </div>
              </div>
              <!-- Private key textarea -->
              <div class="mt-4">
                <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                  {{ t('admin.oauth_apple_private_key') }} <span class="text-error">*</span>
                </label>
                <p class="text-xs text-base-content/40 mb-2">{{ t('admin.oauth_apple_private_key_help') }}</p>
                <textarea
                  :value="getValue('private_key')"
                  @input="setValue('private_key', $event.target.value)"
                  class="textarea textarea-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-xs min-h-[150px]"
                  placeholder="-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdw...
-----END PRIVATE KEY-----"
                ></textarea>
              </div>
            </div>
          </div>

          <!-- Toggles -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center shadow-sm">
                  <IconSettings :size="20" class="text-emerald-400" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
              </div>
              <div class="divide-y divide-base-300/10">
                <div v-for="(field, idx) in getConfigFields().filter(f => f.type === 'toggle')" :key="field.key"
                  class="flex items-center justify-between py-4" :class="{ 'pt-0': idx === 0, 'pb-0': idx === getConfigFields().filter(f => f.type === 'toggle').length - 1 }">
                  <div class="flex items-center gap-3">
                    <span class="text-lg">
                      <IconPlugConnected v-if="field.key === 'enabled'" :size="20" />
                      <IconKey v-else-if="field.key === 'login_enabled'" :size="20" />
                      <IconId v-else-if="field.key === 'register_enabled'" :size="20" />
                      <IconSettings v-else :size="20" />
                    </span>
                    <div>
                      <p class="font-medium text-sm">{{ fieldLabel(field) }}</p>
                      <p class="text-xs text-base-content/40" v-if="fieldHelp(field)">{{ fieldHelp(field) }}</p>
                    </div>
                  </div>
                  <input :checked="getValue(field.key) === '1'" @change="handleToggleChange(field.key, $event)" type="checkbox"
                    :class="['toggle', {
                      'toggle-primary': field.key === 'enabled',
                      'toggle-secondary': field.key === 'login_enabled',
                      'toggle-accent': field.key === 'register_enabled'
                    }]" />
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ========== TELEGRAM ========== -->
        <template v-if="activeProviderName === 'telegram'">
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-[#2AABEE] flex items-center justify-center shadow-sm">
                  <IconBrandTelegram :size="20" class="text-white" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_setup_title') }} — Telegram</h3>
              </div>
              <div class="mb-4 px-3 py-2 bg-warning/5 border border-warning/10 rounded-lg text-xs text-base-content/70 flex items-start gap-2">
                <IconPlugConnected :size="14" class="text-warning shrink-0 mt-0.5" />
                <span>Telegram uses a login widget, not OAuth2. It does not provide user email, so auto-registration is disabled — users must bind manually.</span>
              </div>
              <div class="space-y-3 text-sm text-base-content/70">
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">1</span>
                  <span>Contact <a href="https://t.me/BotFather" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">@BotFather</a> on Telegram to create a bot</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">2</span>
                  <span>Use the <code>/newbot</code> command to create a bot and set its username</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">3</span>
                  <span>Save the <strong>Bot Token</strong> returned by @BotFather</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">4</span>
                  <span>Use <code>/setdomain</code> to set your site URL as the bot domain</span>
                </div>
                <div class="flex items-start gap-3">
                  <span class="flex items-center justify-center w-6 h-6 rounded-full bg-base-300/50 font-semibold text-xs shrink-0 mt-0.5">5</span>
                  <span>Fill in the <strong>Bot Token</strong>, <strong>Bot Username</strong>, and callback URL</span>
                </div>
              </div>
              <div class="flex gap-3 mt-4">
                <a v-if="activeProvider.doc_url" :href="activeProvider.doc_url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary">
                  <IconBook :size="16" /> {{ t('admin.oauth_documentation') }} <IconExternalLink :size="12" />
                </a>
              </div>
            </div>
          </div>

          <!-- Redirect URL -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center shadow-sm">
                  <IconLink :size="20" class="text-green-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_redirect_url') }}</h3>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2 mb-3 px-3 py-2 bg-success/5 border border-success/10 rounded-lg">
                <span class="text-xs text-success font-medium shrink-0">{{ t('admin.oauth_recommended_callback') }}:</span>
                <code class="text-xs text-success/80 break-all">{{ recommendedCallback() }}</code>
                <button class="btn btn-ghost btn-xs shrink-0 ml-auto" @click="copyCallbackUrl" :title="t('user.copy')">
                  <IconCheck v-if="copiedCallback === 'telegram'" :size="14" class="text-success" />
                  <IconCopy v-else :size="14" />
                </button>
              </div>
              <input
                :value="getValue('redirect_url')"
                @input="setValue('redirect_url', $event.target.value)"
                type="text"
                :placeholder="'https://your-domain.com/oauth/telegram/callback'"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <!-- Telegram credentials -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center shadow-sm">
                  <IconLock :size="20" class="text-blue-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ t('admin.oauth_credentials_section') }}</h3>
                  <p class="text-xs text-base-content/40">Telegram Bot credentials</p>
                </div>
              </div>
              <div class="grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_telegram_bot_token') }} <span class="text-error">*</span>
                  </label>
                  <p class="text-xs text-base-content/40 mb-1.5">{{ t('admin.oauth_telegram_bot_token_help') }}</p>
                  <input :value="getValue('bot_token')" @input="setValue('bot_token', $event.target.value)" type="password"
                    placeholder="123456:ABC-DEF1234ghIkl..."
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 transition-all rounded-xl font-mono text-sm" />
                </div>
                <div>
                  <label class="text-sm font-medium text-base-content/80 mb-1.5 block">
                    {{ t('admin.oauth_telegram_bot_username') }} <span class="text-error">*</span>
                  </label>
                  <p class="text-xs text-base-content/40 mb-1.5">{{ t('admin.oauth_telegram_bot_username_help') }}</p>
                  <input :value="getValue('bot_username')" @input="setValue('bot_username', $event.target.value)" type="text"
                    placeholder="@YourBot"
                    class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 transition-all rounded-xl font-mono text-sm" />
                </div>
              </div>
            </div>
          </div>

          <!-- Toggles -->
          <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center shadow-sm">
                  <IconSettings :size="20" class="text-emerald-400" />
                </div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
              </div>
              <div class="divide-y divide-base-300/10">
                <div v-for="(field, idx) in getConfigFields().filter(f => f.type === 'toggle')" :key="field.key"
                  class="flex items-center justify-between py-4" :class="{ 'pt-0': idx === 0, 'pb-0': idx === getConfigFields().filter(f => f.type === 'toggle').length - 1 }">
                  <div class="flex items-center gap-3">
                    <span class="text-lg">
                      <IconPlugConnected v-if="field.key === 'enabled'" :size="20" />
                      <IconKey v-else-if="field.key === 'login_enabled'" :size="20" />
                      <IconId v-else-if="field.key === 'register_enabled'" :size="20" />
                      <IconSettings v-else :size="20" />
                    </span>
                    <div>
                      <p class="font-medium text-sm">{{ fieldLabel(field) }}</p>
                      <p class="text-xs text-base-content/40" v-if="fieldHelp(field)">{{ fieldHelp(field) }}</p>
                    </div>
                  </div>
                  <input :checked="getValue(field.key) === '1'" @change="handleToggleChange(field.key, $event)" type="checkbox"
                    :class="['toggle', {
                      'toggle-primary': field.key === 'enabled',
                      'toggle-secondary': field.key === 'login_enabled',
                      'toggle-accent': field.key === 'register_enabled'
                    }]" />
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- Error -->
        <div v-if="error" class="p-4 bg-error/10 border border-error/20 rounded-xl text-error text-sm flex items-center gap-2">
          <IconPlugConnected :size="18" class="shrink-0" />
          {{ error }}
        </div>

        <!-- Save button -->
        <div class="flex justify-end pt-2 items-center gap-3">
          <span v-if="isCurrentDirty" class="text-xs text-warning/80 flex items-center gap-1.5 px-3 py-1.5 bg-warning/10 rounded-lg">
            <span class="w-1.5 h-1.5 rounded-full bg-warning animate-pulse"></span>
            {{ t('admin.unsaved_changes') }}
          </span>
          <GradientButton
            variant="purple"
            size="lg"
            :loading="saving"
            :disabled="saving"
            @click="handleSave"
            class="shadow-lg shadow-purple-500/20"
          >
            <span v-if="saving">{{ t('admin.settings_saving') }}</span>
            <span v-else class="flex items-center gap-2">
              <IconSettings :size="18" />
              {{ t('admin.settings_save') }}
            </span>
          </GradientButton>
        </div>
      </div>
    </template>
  </div>
</template>
