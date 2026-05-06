<script setup>
import { ref } from 'vue'
import { useI18n } from '../../i18n'
import { accessToken } from '../../composables/useUserAuth'
import AvatarInitials from '../AvatarInitials.vue'
import { User, Save, CircleCheck, CircleAlert, Shield, Globe, MapPin } from 'lucide-vue-next'

const props = defineProps({
  profile: { type: Object, default: null }
})

const emit = defineEmits(['profile-updated'])

const { t } = useI18n()

const profileForm = ref({
  name: '',
  email: '',
  bio: '',
  website: '',
  location: '',
})
const profileSaving = ref(false)
const profileMsg = ref('')
const profileMsgType = ref('success')

function initProfileForm() {
  if (props.profile) {
    profileForm.value.name = props.profile.name || ''
    profileForm.value.email = props.profile.email || ''
    profileForm.value.bio = props.profile.bio || ''
    profileForm.value.website = props.profile.website || ''
    profileForm.value.location = props.profile.location || ''
  }
}

async function saveProfile() {
  profileSaving.value = true
  profileMsg.value = ''
  try {
    const resp = await fetch('/api/v1/user/profile', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken.value}`,
      },
      body: JSON.stringify(profileForm.value),
    })
    const json = await resp.json()
    if (json.code === 0) {
      emit('profile-updated', profileForm.value)
      profileMsg.value = t('user.profile_updated')
      profileMsgType.value = 'success'
    } else {
      profileMsg.value = json.message || t('user.profile_update_failed')
      profileMsgType.value = 'error'
    }
  } catch {
    profileMsg.value = t('user.profile_update_failed')
    profileMsgType.value = 'error'
  } finally {
    profileSaving.value = false
  }
}

defineExpose({ initProfileForm })
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- ===== COL 1-2: Form ===== -->
    <div class="lg:col-span-2">
      <div class="bg-base-100 border border-base-300/20 rounded-2xl shadow-sm overflow-hidden">
        <div class="p-6">
          <div v-if="profileMsg" :class="[
            'flex items-center gap-2 px-4 py-3 rounded-xl mb-5 text-sm font-medium',
            profileMsgType === 'success'
              ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20'
              : 'bg-red-500/10 text-red-400 border border-red-500/20'
          ]">
            <CircleCheck v-if="profileMsgType === 'success'" :size="16" />
            <CircleAlert v-else :size="16" />
            <span>{{ profileMsg }}</span>
          </div>

          <form @submit.prevent="saveProfile" class="space-y-5">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
              <div>
                <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.name') }}</label>
                <input
                  v-model="profileForm.name"
                  type="text"
                  :placeholder="t('user.name_placeholder')"
                  class="w-full px-4 py-3 bg-base-200 border border-base-300/20 rounded-xl text-sm text-base-content placeholder:text-base-content/25 focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10 focus:outline-none transition-all duration-200"
                  autocomplete="name"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.email') }}</label>
                <input
                  v-model="profileForm.email"
                  type="email"
                  :placeholder="t('user.email_placeholder')"
                  class="w-full px-4 py-3 bg-base-200 border border-base-300/20 rounded-xl text-sm text-base-content placeholder:text-base-content/25 focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10 focus:outline-none transition-all duration-200"
                  autocomplete="email"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.website') }}</label>
                <input
                  v-model="profileForm.website"
                  type="url"
                  :placeholder="t('user.website_placeholder')"
                  class="w-full px-4 py-3 bg-base-200 border border-base-300/20 rounded-xl text-sm text-base-content placeholder:text-base-content/25 focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10 focus:outline-none transition-all duration-200"
                  autocomplete="url"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.location') }}</label>
                <input
                  v-model="profileForm.location"
                  type="text"
                  :placeholder="t('user.location_placeholder')"
                  class="w-full px-4 py-3 bg-base-200 border border-base-300/20 rounded-xl text-sm text-base-content placeholder:text-base-content/25 focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10 focus:outline-none transition-all duration-200"
                  autocomplete="address-level2"
                />
              </div>
              <div class="sm:col-span-2">
                <label class="block text-xs font-semibold text-base-content/50 uppercase tracking-wider mb-1.5">{{ t('user.bio') }}</label>
                <textarea
                  v-model="profileForm.bio"
                  :placeholder="t('user.bio_placeholder')"
                  rows="3"
                  class="w-full px-4 py-3 bg-base-200 border border-base-300/20 rounded-xl text-sm text-base-content placeholder:text-base-content/25 focus:bg-base-100 focus:border-primary/40 focus:ring-2 focus:ring-primary/10 focus:outline-none transition-all duration-200 resize-none"
                ></textarea>
              </div>
            </div>
            <div class="flex pt-2">
              <button
                type="submit"
                :disabled="profileSaving"
                class="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-content rounded-xl px-6 py-3 font-semibold text-sm shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:hover:shadow-primary/20"
              >
                <span v-if="profileSaving" class="loading loading-spinner loading-xs"></span>
                <Save v-else :size="16" />
                {{ profileSaving ? t('admin.settings_saving') : t('user.save_changes') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- ===== COL 3: Sidebar ===== -->
    <div class="space-y-4">
      <!-- Avatar preview -->
      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm text-center">
        <div class="flex justify-center mb-3">
          <AvatarInitials :name="profileForm.name || 'You'" size="lg" />
        </div>
        <p class="text-sm font-semibold text-base-content">{{ profileForm.name || '—' }}</p>
        <p class="text-xs text-base-content/40 truncate">{{ profileForm.email || '—' }}</p>
        <p v-if="profileForm.location" class="text-xs text-base-content/30 mt-1 flex items-center justify-center gap-1">
          <MapPin :size="11" /> {{ profileForm.location }}
        </p>
        <p v-if="profileForm.website" class="text-xs text-primary mt-1 flex items-center justify-center gap-1 truncate">
          <Globe :size="11" /> {{ profileForm.website }}
        </p>
      </div>

      <!-- Tips card -->
      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
          <Shield :size="16" class="text-primary/60" />
          {{ t('user.profile_tips') || 'Profile Tips' }}
        </h3>
        <ul class="space-y-2 text-xs text-base-content/50">
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.tip_name') || 'Use your real name to help others recognize you.' }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.tip_bio') || 'A good bio helps people know what you do.' }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.tip_link') || 'Add your website to showcase your work.' }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
