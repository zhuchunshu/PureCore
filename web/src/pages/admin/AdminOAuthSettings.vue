<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
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

// Providers list from backend
const providers = ref([])
const activeProviderIndex = ref(0)
const pendingTabIndex = ref(null) // used when user tries to switch tabs with unsaved changes

// Settings values for each provider (keyed by provider name)
const providerValues = reactive({})
// Original values snapshot for dirty detection
const originalValues = reactive({})
// Track which providers have unsaved changes
const dirtyProviders = reactive(new Set())

// Confirm dialog state
const showUnsavedDialog = ref(false)

// Provider icon map using Tabler icons
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

const activeProvider = computed(() => {
  return providers.value[activeProviderIndex.value] || null
})

const activeValues = computed(() => {
  const name = activeProvider.value?.name
  return name ? (providerValues[name] || {}) : {}
})

const activeConfigFields = computed(() => {
  return activeProvider.value?.config_fields || []
})

// Separate toggles from credential fields
const toggleFields = computed(() => {
  return activeConfigFields.value.filter(f => f.type === 'toggle')
})

const credentialFields = computed(() => {
  return activeConfigFields.value.filter(f => f.type !== 'toggle')
})

const redirectUrlField = computed(() => {
  return credentialFields.value.find(f => f.key === 'redirect_url')
})

const otherCredentialFields = computed(() => {
  return credentialFields.value.filter(f => f.key !== 'redirect_url')
})

// Check if current active provider has unsaved changes
const isCurrentDirty = computed(() => {
  const name = activeProvider.value?.name
  return name ? dirtyProviders.has(name) : false
})

function getValue(key, fallback = '') {
  const vals = activeValues.value
  return vals[key] !== undefined && vals[key] !== null ? vals[key] : fallback
}

function setValue(key, value) {
  const name = activeProvider.value?.name
  if (name && providerValues[name]) {
    providerValues[name][key] = value
    // Mark as dirty by comparing with original
    checkDirty(name)
  }
}

function handleToggleChange(key, event) {
  setValue(key, event.target.checked ? '1' : '0')
}

// Compare current values with originals to determine dirty state
function checkDirty(name) {
  const current = providerValues[name] || {}
  const original = originalValues[name] || {}
  const keys = new Set([...Object.keys(current), ...Object.keys(original)])
  let isDirty = false
  for (const key of keys) {
    const cv = current[key] !== undefined && current[key] !== null ? String(current[key]) : ''
    const ov = original[key] !== undefined && original[key] !== null ? String(original[key]) : ''
    if (cv !== ov) {
      isDirty = true
      break
    }
  }
  if (isDirty) {
    dirtyProviders.add(name)
  } else {
    dirtyProviders.delete(name)
  }
}

// Snapshot current values as originals (after load or save)
function snapshotOriginal(name) {
  const vals = providerValues[name]
  if (!vals) return
  originalValues[name] = { ...vals }
  dirtyProviders.delete(name)
}

function copyCallbackUrl() {
  const url = activeValues.value._recommended_redirect_url || getValue('redirect_url')
  if (!url || typeof navigator === 'undefined') return
  navigator.clipboard.writeText(url).then(() => {
    copiedCallback.value = activeProvider.value?.name || ''
    setTimeout(() => {
      copiedCallback.value = ''
    }, 2000)
  }).catch(() => {})
}

// Handle tab switching with unsaved changes warning
function switchTab(index) {
  if (activeProviderIndex.value === index) return

  // If current tab has unsaved changes, show confirmation
  if (isCurrentDirty.value) {
    pendingTabIndex.value = index
    showUnsavedDialog.value = true
    return
  }

  activeProviderIndex.value = index
}

function confirmDiscardAndSwitch() {
  const name = activeProvider.value?.name
  if (name) {
    // Restore original values
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

async function handleSave() {
  saving.value = true
  error.value = ''
  const provider = activeProvider.value
  if (!provider) return

  try {
    // Build settings payload from all config fields
    const settings = {}
    const vals = providerValues[provider.name] || {}
    for (const field of provider.config_fields) {
      settings[field.key] = vals[field.key] !== undefined && vals[field.key] !== null ? String(vals[field.key]) : ''
    }

    const payload = {
      providers: [
        {
          provider: provider.name,
          settings,
        },
      ],
    }

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
        // Snapshot originals for dirty detection
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

// Translate a label key or return as-is
function fieldLabel(field) {
  const label = field.label || ''
  // If it's a translation key (contains dots), try translating
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

      <!-- Provider tabs with Tabler icons and unsaved indicators -->
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
          <!-- Unsaved changes indicator dot -->
          <span
            v-if="dirtyProviders.has(provider.name)"
            class="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-warning animate-pulse"
            :title="t('admin.unsaved_changes')"
          ></span>
        </button>
      </div>

      <!-- Unsaved changes confirmation dialog -->
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

      <!-- Active provider settings -->
      <div v-if="activeProvider" class="grid gap-6">
        <!-- Usage instructions card -->
        <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
          <div class="card-body p-5">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 flex items-center justify-center shadow-sm">
                <IconBook :size="20" class="text-cyan-400" />
              </div>
              <div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_usage_title') }}</h3>
                <p class="text-xs text-base-content/40">{{ t('admin.oauth_usage_instructions', { provider: activeProvider.display_name }) }}</p>
              </div>
            </div>

            <!-- Flow type hint -->
            <div class="flex items-center gap-2 mb-4 px-3 py-2 bg-base-200/50 rounded-lg">
              <IconPlugConnected :size="16" class="text-base-content/50 shrink-0" />
              <span class="text-xs text-base-content/60">
                {{ activeProvider.is_oauth2 ? t('admin.oauth_is_oauth2_hint') : t('admin.oauth_non_oauth2_hint') }}
              </span>
            </div>

            <!-- Step-by-step instructions -->
            <div class="space-y-3 text-sm text-base-content/70">
              <div class="flex items-start gap-3">
                <span class="flex items-center justify-center w-6 h-6 rounded-full bg-primary/10 text-primary font-semibold text-xs shrink-0 mt-0.5">1</span>
                <span>
                  {{ t('admin.oauth_usage_instructions', { provider: activeProvider.display_name }) }}
                  <a v-if="activeProvider.apply_url" :href="activeProvider.apply_url" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 text-primary hover:underline ml-1">
                    {{ t('admin.oauth_apply') }}
                    <IconExternalLink :size="12" />
                  </a>
                </span>
              </div>
              <div class="flex items-start gap-3">
                <span class="flex items-center justify-center w-6 h-6 rounded-full bg-primary/10 text-primary font-semibold text-xs shrink-0 mt-0.5">2</span>
                <span>
                  {{ t('admin.oauth_redirect_url_desc') }}
                </span>
              </div>
              <div class="flex items-start gap-3">
                <span class="flex items-center justify-center w-6 h-6 rounded-full bg-primary/10 text-primary font-semibold text-xs shrink-0 mt-0.5">3</span>
                <span>
                  {{ t('admin.oauth_redirect_url_hint', { provider: activeProvider.display_name }) }}
                </span>
              </div>
            </div>

            <!-- Documentation and Apply links -->
            <div class="flex flex-wrap gap-3 mt-5">
              <a
                v-if="activeProvider.doc_url"
                :href="activeProvider.doc_url"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary"
              >
                <IconBook :size="16" />
                {{ t('admin.oauth_documentation') }}
                <IconExternalLink :size="12" />
              </a>
              <a
                v-if="activeProvider.apply_url"
                :href="activeProvider.apply_url"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-sm btn-ghost gap-2 text-base-content/70 hover:text-primary"
              >
                <IconExternalLink :size="16" />
                {{ t('admin.oauth_apply') }}
              </a>
            </div>
          </div>
        </div>

        <!-- Toggle switches -->
        <div v-if="toggleFields.length > 0" class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
          <div class="card-body p-5">
            <div class="flex items-center gap-3 mb-5">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center shadow-sm">
                <IconSettings :size="20" class="text-emerald-400" />
              </div>
              <div>
                <h3 class="font-semibold text-sm">{{ t('admin.oauth_toggle_section') }}</h3>
                <p class="text-xs text-base-content/40">{{ t('admin.oauth_toggle_section_desc') }}</p>
              </div>
            </div>
            <div class="divide-y divide-base-300/10">
              <div
                v-for="(field, idx) in toggleFields"
                :key="field.key"
                class="flex items-center justify-between py-4"
                :class="{ 'pt-0': idx === 0, 'pb-0': idx === toggleFields.length - 1 }"
              >
                <div class="flex items-center gap-3">
                  <span class="text-lg">
                    <IconPlugConnected v-if="field.key === 'enabled'" :size="20" />
                    <IconKey v-else-if="field.key === 'login_enabled'" :size="20" />
                    <IconId v-else-if="field.key === 'register_enabled'" :size="20" />
                    <IconSettings v-else :size="20" />
                  </span>
                  <div>
                    <p class="font-medium text-sm">{{ fieldLabel(field) }}</p>
                    <p class="text-xs text-base-content/40">{{ fieldHelp(field) }}</p>
                  </div>
                </div>
                <input
                  :checked="getValue(field.key) === '1'"
                  @change="handleToggleChange(field.key, $event)"
                  type="checkbox"
                  class="toggle"
                  :class="{
                    'toggle-primary': field.key === 'enabled',
                    'toggle-secondary': field.key === 'login_enabled',
                    'toggle-accent': field.key === 'register_enabled'
                  }"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- API Credentials -->
        <div v-if="credentialFields.length > 0">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center">
              <IconLock :size="16" class="text-blue-400" />
            </div>
            <h3 class="font-semibold text-sm">{{ t('admin.oauth_credentials_section') }}</h3>
          </div>

          <!-- Redirect URL (always shown first if exists) -->
          <div v-if="redirectUrlField" class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm mb-5 overflow-hidden">
            <div class="card-body p-5">
              <div class="flex items-center gap-3 mb-4">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center shadow-sm">
                  <IconLink :size="20" class="text-green-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-sm">{{ fieldLabel(redirectUrlField) }}</h3>
                  <p class="text-xs text-base-content/40">{{ fieldHelp(redirectUrlField) || t('admin.oauth_redirect_url_desc') }}</p>
                </div>
              </div>

              <!-- Recommended callback URL -->
              <div class="flex items-center gap-2 mb-3 px-3 py-2 bg-success/5 border border-success/10 rounded-lg">
                <span class="text-xs text-success font-medium shrink-0">{{ t('admin.oauth_recommended_callback') }}:</span>
                <code class="text-xs text-success/80 break-all">{{ activeValues._recommended_redirect_url || getValue('redirect_url') }}</code>
                <button
                  class="btn btn-ghost btn-xs shrink-0 ml-auto"
                  @click="copyCallbackUrl"
                  :title="t('user.copy')"
                >
                  <IconCheck v-if="copiedCallback === activeProvider.name" :size="14" class="text-success" />
                  <IconCopy v-else :size="14" />
                </button>
              </div>

              <input
                :value="getValue('redirect_url')"
                @input="setValue('redirect_url', $event.target.value)"
                type="text"
                :placeholder="redirectUrlField.placeholder || t('admin.oauth_redirect_url_placeholder')"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-green-400 focus:ring-2 focus:ring-green-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>
          </div>

          <!-- Other credential fields in grid -->
          <div class="grid gap-5" :class="otherCredentialFields.length > 1 ? 'sm:grid-cols-2' : ''">
            <div
              v-for="field in otherCredentialFields"
              :key="field.key"
              class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm hover:shadow-md transition-shadow duration-300 overflow-hidden"
            >
              <div class="card-body p-5">
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center shadow-sm">
                    <IconKey v-if="field.type === 'password'" :size="18" class="text-orange-400" />
                    <IconId v-else :size="18" class="text-blue-400" />
                  </div>
                  <div>
                    <h3 class="font-semibold text-sm">{{ fieldLabel(field) }}</h3>
                    <p v-if="fieldHelp(field)" class="text-xs text-base-content/40">{{ fieldHelp(field) }}</p>
                  </div>
                </div>
                <input
                  :value="getValue(field.key)"
                  @input="setValue(field.key, $event.target.value)"
                  :type="field.type === 'password' ? 'password' : 'text'"
                  :placeholder="field.placeholder || ''"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-primary/40 focus:ring-2 focus:ring-primary/20 transition-all rounded-xl font-mono text-sm"
                />
                <div v-if="field.required" class="flex items-center gap-1 mt-2">
                  <span class="text-xs text-error/60">* {{ fieldLabel(field) }} is required</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Error message -->
        <div v-if="error" class="p-4 bg-error/10 border border-error/20 rounded-xl text-error text-sm flex items-center gap-2">
          <IconPlugConnected :size="18" class="shrink-0" />
          {{ error }}
        </div>

        <!-- Save button with unsaved indicator -->
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
