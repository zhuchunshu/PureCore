<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import { accessToken } from '../../composables/useAuth'
import { adminOption, adminOptionSet, refreshOptions } from '../../composables/useAdminOption'
import { toastSuccess } from '../../composables/useToast'
import GradientButton from '../../components/GradientButton.vue'
import { IconBrandGithub, IconBrandGoogle, IconBrandDiscord, IconBrandApple, IconBrandTelegram } from '@tabler/icons-vue'

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

// All 5 providers
const providerKeys = ['github', 'google', 'discord', 'apple', 'telegram']

const providerMeta = {
  github: {
    iconBg: 'bg-neutral-800',
    iconColor: 'text-white',
    toggleClass: 'toggle-neutral',
    gradient: 'from-neutral-500/10 via-stone-500/10 to-neutral-500/10',
    textGradient: 'from-neutral-400 via-stone-400 to-neutral-400',
    guideLink: 'https://github.com/settings/developers',
    guideLinkText: 'GitHub Developer Settings',
    fields: ['client_id', 'client_secret'],
  },
  google: {
    iconBg: 'bg-white',
    iconColor: 'text-red-500',
    toggleClass: 'toggle-primary',
    gradient: 'from-blue-500/10 via-red-500/10 to-yellow-500/10',
    textGradient: 'from-blue-400 via-red-400 to-yellow-400',
    guideLink: 'https://console.cloud.google.com/apis/credentials',
    guideLinkText: 'Google Cloud Console',
    fields: ['client_id', 'client_secret'],
  },
  discord: {
    iconBg: 'bg-indigo-600',
    iconColor: 'text-white',
    toggleClass: 'toggle-secondary',
    gradient: 'from-indigo-500/10 via-purple-500/10 to-indigo-500/10',
    textGradient: 'from-indigo-400 via-purple-400 to-indigo-400',
    guideLink: 'https://discord.com/developers/applications',
    guideLinkText: 'Discord Developer Portal',
    fields: ['client_id', 'client_secret'],
  },
  apple: {
    iconBg: 'bg-black',
    iconColor: 'text-white',
    toggleClass: 'toggle-accent',
    gradient: 'from-slate-500/10 via-zinc-500/10 to-slate-500/10',
    textGradient: 'from-slate-400 via-zinc-400 to-slate-400',
    guideLink: 'https://developer.apple.com/account/resources/identifiers/list',
    guideLinkText: 'Apple Developer Console',
    fields: ['client_id', 'team_id', 'key_id', 'private_key'],
  },
  telegram: {
    iconBg: 'bg-sky-500',
    iconColor: 'text-white',
    toggleClass: 'toggle-info',
    gradient: 'from-sky-500/10 via-cyan-500/10 to-sky-500/10',
    textGradient: 'from-sky-400 via-cyan-400 to-sky-400',
    guideLink: 'https://core.telegram.org/bots#how-do-i-create-a-bot',
    guideLinkText: 'Telegram BotFather',
    fields: ['bot_token'],
  },
}

const callbackUrls = reactive({})

// Dynamic forms: one per provider
const forms = reactive({})

for (const key of providerKeys) {
  forms[key] = reactive({
    allow_login: '0',
    allow_register: '0',
    callback_url: '',
    client_id: '',
    client_secret: '',
    team_id: '',
    key_id: '',
    private_key: '',
    bot_token: '',
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
      forms[key].allow_login = await adminOption(`oauth_${key}_allow_login`, '0')
      forms[key].allow_register = await adminOption(`oauth_${key}_allow_register`, '0')
      forms[key].callback_url = await adminOption(`oauth_${key}_callback`, '')
      forms[key].client_id = await adminOption(`oauth_${key}_client_id`, '')
      forms[key].client_secret = await adminOption(`oauth_${key}_client_secret`, '')
    }
    // Apple-specific fields
    forms.apple.team_id = await adminOption('oauth_apple_team_id', '')
    forms.apple.key_id = await adminOption('oauth_apple_key_id', '')
    forms.apple.private_key = await adminOption('oauth_apple_private_key', '')
    // Telegram-specific field
    forms.telegram.bot_token = await adminOption('oauth_telegram_bot_token', '')
  } catch (err) {
    error.value = t('admin.network_error')
  } finally {
    loading.value = false
  }
})

function getSecretPlaceholder(form) {
  const current = form.client_secret
  if (current && current !== '••••••••' && current !== '0' && current !== '') {
    return '••••••••'
  }
  return ''
}

function getPrivateKeyPlaceholder(form) {
  const current = form.private_key
  if (current && current !== '••••••••' && current !== '0' && current !== '') {
    return '••••••••'
  }
  return ''
}

function getBotTokenPlaceholder(form) {
  const current = form.bot_token
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
      await adminOptionSet(`oauth_${key}_allow_login`, form.allow_login)
      await adminOptionSet(`oauth_${key}_allow_register`, form.allow_register)
      await adminOptionSet(`oauth_${key}_callback`, form.callback_url)
      if (key !== 'telegram') {
        await adminOptionSet(`oauth_${key}_client_id`, form.client_id)
      }
      if (key === 'github' || key === 'google' || key === 'discord') {
        if (form.client_secret && form.client_secret !== '••••••••') {
          await adminOptionSet(`oauth_${key}_client_secret`, form.client_secret)
        }
      }
    }
    // Apple-specific
    await adminOptionSet('oauth_apple_team_id', forms.apple.team_id)
    await adminOptionSet('oauth_apple_key_id', forms.apple.key_id)
    if (forms.apple.private_key && forms.apple.private_key !== '••••••••') {
      await adminOptionSet('oauth_apple_private_key', forms.apple.private_key)
    }
    // Telegram-specific
    if (forms.telegram.bot_token && forms.telegram.bot_token !== '••••••••') {
      await adminOptionSet('oauth_telegram_bot_token', forms.telegram.bot_token)
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
  apple: IconBrandApple,
  telegram: IconBrandTelegram,
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
        <div class="skeleton h-10 w-32 rounded-xl"></div>
        <div class="skeleton h-10 w-32 rounded-xl"></div>
        <div class="skeleton h-10 w-32 rounded-xl"></div>
      </div>
      <div class="skeleton h-96 rounded-2xl"></div>
    </template>

    <!-- Error state -->
    <div v-else-if="error" class="flex items-center justify-center py-20">
      <div class="p-6 bg-error/10 border border-error/20 rounded-2xl text-error max-w-lg flex items-center gap-3 backdrop-blur-sm">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span class="font-medium">{{ error }}</span>
      </div>
    </div>

    <template v-else>
      <!-- Header -->
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
      <div class="tabs tabs-boxed bg-base-100/80 backdrop-blur-sm border border-base-300/20 rounded-xl p-1.5 shadow-sm overflow-x-auto flex-nowrap">
        <button
          v-for="key in providerKeys"
          :key="key"
          :class="['tab tab-lg gap-2.5 transition-all duration-200 font-medium whitespace-nowrap', { 'tab-active': activeProvider === key }]"
          @click="activeProvider = key"
        >
          <component :is="providerIcons[key]" :size="20" />
          <span>{{ t(`admin.oauth.${key}.title`) }}</span>
        </button>
      </div>

      <!-- Active provider form -->
      <div class="card bg-base-100/80 backdrop-blur-sm border border-base-300/20 shadow-sm overflow-hidden">
        <div class="card-body p-6">
          <!-- Provider header with icon and toggles -->
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
          </div>

          <div class="space-y-4">
            <!-- Enable/Disable toggles -->
            <div class="grid grid-cols-2 gap-4 p-4 bg-base-200/40 rounded-xl border border-base-300/15">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-base-content/80">{{ t('admin.oauth.allow_login') }}</p>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth.allow_login_hint') }}</p>
                </div>
                <input
                  v-model="forms[activeProvider].allow_login"
                  type="checkbox"
                  true-value="1"
                  false-value="0"
                  :class="['toggle', providerMeta[activeProvider].toggleClass]"
                />
              </div>
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-base-content/80">{{ t('admin.oauth.allow_register') }}</p>
                  <p class="text-xs text-base-content/40">{{ t('admin.oauth.allow_register_hint') }}</p>
                </div>
                <input
                  v-model="forms[activeProvider].allow_register"
                  type="checkbox"
                  true-value="1"
                  false-value="0"
                  :class="['toggle', providerMeta[activeProvider].toggleClass]"
                />
              </div>
            </div>

            <!-- Callback URL (editable) -->
            <div>
              <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t('admin.oauth.callback_url') }}</label>
              <input
                v-model="forms[activeProvider].callback_url"
                type="text"
                :placeholder="callbackUrls[activeProvider] || ''"
                class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-xs"
              />
              <p class="text-xs text-base-content/30 mt-1 ml-1">{{ t('admin.oauth.callback_url_hint') }} <code class="text-xs bg-base-300/50 px-1 rounded">{{ callbackUrls[activeProvider] || '' }}</code></p>
            </div>

            <!-- GitHub/Google/Discord: Client ID + Client Secret -->
            <template v-if="['github', 'google', 'discord'].includes(activeProvider)">
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t(`admin.oauth.${activeProvider}.client_id`) }}</label>
                <input
                  v-model="forms[activeProvider].client_id"
                  type="text"
                  :placeholder="t(`admin.oauth.${activeProvider}.client_id_placeholder`)"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t(`admin.oauth.${activeProvider}.client_secret`) }}</label>
                <input
                  v-model="forms[activeProvider].client_secret"
                  type="password"
                  :placeholder="getSecretPlaceholder(forms[activeProvider]) || t(`admin.oauth.${activeProvider}.client_secret_placeholder`)"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                  @focus="($event) => { if (getSecretPlaceholder(forms[activeProvider])) { forms[activeProvider].client_secret = '' } }"
                />
              </div>
            </template>

            <!-- Apple: Client ID (Service ID), Team ID, Key ID, Private Key -->
            <template v-if="activeProvider === 'apple'">
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t('admin.oauth.apple.client_id') }}</label>
                <input
                  v-model="forms.apple.client_id"
                  type="text"
                  :placeholder="t('admin.oauth.apple.client_id_placeholder')"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t('admin.oauth.apple.team_id') }}</label>
                <input
                  v-model="forms.apple.team_id"
                  type="text"
                  :placeholder="t('admin.oauth.apple.team_id_placeholder')"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t('admin.oauth.apple.key_id') }}</label>
                <input
                  v-model="forms.apple.key_id"
                  type="text"
                  :placeholder="t('admin.oauth.apple.key_id_placeholder')"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t('admin.oauth.apple.private_key') }}</label>
                <textarea
                  v-model="forms.apple.private_key"
                  :placeholder="getPrivateKeyPlaceholder(forms.apple) || t('admin.oauth.apple.private_key_placeholder')"
                  rows="4"
                  class="textarea textarea-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                  @focus="($event) => { if (getPrivateKeyPlaceholder(forms.apple)) { forms.apple.private_key = '' } }"
                ></textarea>
              </div>
            </template>

            <!-- Telegram: Bot Token -->
            <template v-if="activeProvider === 'telegram'">
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1.5 ml-1">{{ t('admin.oauth.telegram.bot_token') }}</label>
                <input
                  v-model="forms.telegram.bot_token"
                  type="password"
                  :placeholder="getBotTokenPlaceholder(forms.telegram) || t('admin.oauth.telegram.bot_token_placeholder')"
                  class="input input-bordered w-full bg-base-200/50 border-base-300/30 focus:border-neutral-400 focus:ring-2 focus:ring-neutral-400/20 transition-all rounded-xl font-mono text-sm"
                  @focus="($event) => { if (getBotTokenPlaceholder(forms.telegram)) { forms.telegram.bot_token = '' } }"
                />
              </div>
              <p class="text-xs text-info/80 mt-1 ml-1">{{ t('admin.oauth.telegram.bot_token_hint') }}</p>
            </template>

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
          <span v-else class="flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" /></svg>
            {{ t('admin.settings_save') }}
          </span>
        </GradientButton>
      </div>
    </template>
  </div>
</template>
