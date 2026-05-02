<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { accessToken } from '../../composables/useAuth'
import { adminOption, adminOptionSet, refreshOptions } from '../../composables/useAdminOption'
import { toastSuccess } from '../../composables/useToast'
import GradientButton from '../../components/GradientButton.vue'
import { IconBrandGithub, IconBrandGoogle, IconBrandDiscord } from '@tabler/icons-vue'

const { t } = useI18n()
useSEO({
  title: t('admin.oauth.title'),
  description: t('admin.oauth.description'),
})
const router = useRouter()
const adminPrefix = import.meta.env.VITE_ADMIN_ROUTE_PREFIX || 'control-panel'

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const activeProvider = ref('github')

// Supported providers
const providerKeys = ['github', 'google', 'discord']

// Provider-specific styling and callback URL segments
const providerMeta = {
  github: {
    iconBg: 'bg-neutral-800',
    iconColor: 'text-white',
    toggleClass: 'toggle-neutral',
    gradient: 'from-neutral-500/10 via-stone-500/10 to-neutral-500/10',
    textGradient: 'from-neutral-400 via-stone-400 to-neutral-400',
    btnVariant: 'neutral',
    guideLink: 'https://github.com/settings/developers',
    guideLinkText: 'GitHub Developer Settings',
  },
  google: {
    iconBg: 'bg-white',
    iconColor: 'text-red-500',
    toggleClass: 'toggle-primary',
    gradient: 'from-blue-500/10 via-red-500/10 to-yellow-500/10',
    textGradient: 'from-blue-400 via-red-400 to-yellow-400',
    btnVariant: 'purple',
    guideLink: 'https://console.cloud.google.com/apis/credentials',
    guideLinkText: 'Google Cloud Console',
  },
  discord: {
    iconBg: 'bg-indigo-600',
    iconColor: 'text-white',
    toggleClass: 'toggle-secondary',
    gradient: 'from-indigo-500/10 via-purple-500/10 to-indigo-500/10',
    textGradient: 'from-indigo-400 via-purple-400 to-indigo-400',
    btnVariant: 'purple',
    guideLink: 'https://discord.com/developers/applications',
    guideLinkText: 'Discord Developer Portal',
  },
}

// Build callback URLs safely on client
const callbackUrls = reactive({})

// Dynamic forms: one per provider
const forms = reactive({})

for (const key of providerKeys) {
  forms[key] = reactive({
    oauth_enabled: '',
    oauth_client_id: '',
    oauth_client_secret: '',
  })
}

onMounted(async () => {
  if (typeof window !== 'undefined') {
    for (const key of providerKeys) {
      callbackUrls[key] = `${window.location.protocol}//${window.location.host}/api/v1/oauth/${key}/callback`
    }
  }

  if (!accessToken.value) {
    loading.value = false
    router.push(`/${adminPrefix}/login`)
    return
  }

  try {
    for (const key of providerKeys) {
      forms[key].oauth_enabled = await adminOption(`oauth_${key}_enabled`, '0')
      forms[key].oauth_client_id = await adminOption(`oauth_${key}_client_id`, '')
      forms[key].oauth_client_secret = await adminOption(`oauth_${key}_client_secret`, '')
    }
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
})

function getSecretPlaceholder(form) {
  const current = form.oauth_client_secret
  if (current && current !== '••••••••' && current !== '0' && current !== '') {
    return '••••••••'
  }
  return ''
}

async function handleSave() {
  saving.value = true
  error.value = ''

  try {
    for (const key of providerKeys) {
      const form = forms[key]
      await adminOptionSet(`oauth_${key}_enabled`, form.oauth_enabled)
      await adminOptionSet(`oauth_${key}_client_id`, form.oauth_client_id)
      if (form.oauth_client_secret && form.oauth_client_secret !== '••••••••') {
        await adminOptionSet(`oauth_${key}_client_secret`, form.oauth_client_secret)
      }
    }
    await refreshOptions()
    toastSuccess(t('admin.settings_saved'))
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    saving.value = false
  }
}

const providerIcons = {
  github: IconBrandGithub,
  google: IconBrandGoogle,
  discord: IconBrandDiscord,
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
        <div class="skeleton h-10 w-32 rounded-xl"></div>
      </div>
      <!-- Form card skeleton -->
      <div class="skeleton h-96 rounded-2xl"></div>
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
      <div class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-neutral-500/10 via-stone-500/10 to-neutral-500/10 border border-neutral-500/10 p-6 md:p-8">
        <div class="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-neutral-500/15 to-stone-500/15 rounded-full blur-3xl -translate-y-1/2 translate-x-1/4 pointer-events-none"></div>
        <div class="relative flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 class="text-2xl md:text-3xl font-black tracking-tight">
              <span class="bg-gradient-to-r from-neutral-400 via-stone-400 to-neutral-400 bg-clip-text text-transparent">{{ t('admin.oauth.title') }}</span>
            </h1>
            <p class="text-base-content/50 mt-2 max-w-lg text-sm md:text-base">{{ t('admin.oauth.description') }}</p>
          </div>
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="tabs tabs-boxed bg-base-100/80 backdrop-blur-sm border border-base-300/20 rounded-xl p-1.5 shadow-sm">
        <button
          v-for="key in providerKeys"
          :key="key"
          :class="['tab tab-lg gap-2.5 transition-all duration-200 font-medium', { 'tab-active': activeProvider === key }]"
          @click="activeProvider = key"
        >
          <component :is="providerIcons[key]" :size="20" />
          <span>{{ t(`admin.oauth.${key}.title`) }}</span>
        </button>
      </div>

      <!-- Active provider section -->
      <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
        <div class="card-body p-6">
          <!-- Provider header -->
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div :class="['w-10 h-10 rounded-xl flex items-center justify-center shadow-sm', providerMeta[activeProvider].iconBg]">
                <component :is="providerIcons[activeProvider]" :size="20" :class="providerMeta[activeProvider].iconColor" />
              </div>
              <div>
                <h3 class="font-semibold text-sm">{{ t(`admin.oauth.${activeProvider}.title`) }}</h3>
                <p class="text-xs text-base-content/40">{{ t(`admin.oauth.${activeProvider}.desc`) }}</p>
              </div>
            </div>
            <!-- Enable toggle -->
            <input
              v-model="forms[activeProvider].oauth_enabled"
              type="checkbox"
              true-value="1"
              false-value="0"
              :class="['toggle', providerMeta[activeProvider].toggleClass]"
            />
          </div>

          <div class="space-y-4">
            <!-- Callback URL -->
            <div>
              <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t(`admin.oauth.${activeProvider}.callback_url`) }}</label>
              <div class="relative">
                <input
                  type="text"
                  readonly
                  :value="callbackUrls[activeProvider] || ''"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl font-mono text-xs text-base-content/50 cursor-default"
                />
              </div>
              <p class="text-xs text-base-content/30 mt-1 ml-1">{{ t(`admin.oauth.${activeProvider}.callback_url_hint`) }}</p>
            </div>

            <!-- Client ID -->
            <div>
              <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t(`admin.oauth.${activeProvider}.client_id`) }}</label>
              <input
                v-model="forms[activeProvider].oauth_client_id"
                type="text"
                :placeholder="t(`admin.oauth.${activeProvider}.client_id_placeholder`)"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
              />
            </div>

            <!-- Client Secret -->
            <div>
              <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t(`admin.oauth.${activeProvider}.client_secret`) }}</label>
              <input
                v-model="forms[activeProvider].oauth_client_secret"
                type="password"
                :placeholder="getSecretPlaceholder(forms[activeProvider]) || t(`admin.oauth.${activeProvider}.client_secret_placeholder`)"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                @focus="($event) => { if (getSecretPlaceholder(forms[activeProvider])) { forms[activeProvider].oauth_client_secret = '' } }"
              />
            </div>

            <!-- Setup guide -->
            <div class="mt-4 p-4 bg-info/5 border border-info/15 rounded-xl">
              <div class="flex items-start gap-3">
                <div class="flex-shrink-0 w-8 h-8 rounded-lg bg-info/10 flex items-center justify-center mt-0.5">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-info" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                </div>
                <div class="text-sm text-info/80">
                  <p class="font-medium mb-1">{{ t(`admin.oauth.${activeProvider}.setup_guide`) }}</p>
                  <ol class="list-decimal list-inside space-y-1 text-xs">
                    <li>
                      <a v-if="providerMeta[activeProvider].guideLink" :href="providerMeta[activeProvider].guideLink" target="_blank" class="underline">
                        {{ t(`admin.oauth.${activeProvider}.setup_step1`) }}
                      </a>
                      <span v-else>{{ t(`admin.oauth.${activeProvider}.setup_step1`) }}</span>
                    </li>
                    <li>{{ t(`admin.oauth.${activeProvider}.setup_step2`) }}</li>
                    <li>{{ t(`admin.oauth.${activeProvider}.setup_step3`) }}</li>
                    <li>{{ t(`admin.oauth.${activeProvider}.setup_step4`) }}</li>
                  </ol>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Save button -->
      <div class="flex justify-end pt-2">
        <GradientButton
          variant="neutral"
          size="lg"
          :loading="saving"
          :disabled="saving"
          @click="handleSave"
          class="shadow-lg shadow-neutral-500/20"
        >
          <span v-if="saving">{{ t('admin.settings_saving') }}</span>
          <span v-else class="flex items-center gap-2">💾 {{ t('admin.settings_save') }}</span>
        </GradientButton>
      </div>
    </template>
  </div>
</template>
