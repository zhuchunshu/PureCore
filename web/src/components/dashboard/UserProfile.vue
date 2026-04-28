<script setup>
import { ref } from 'vue'
import { useI18n } from '../../i18n'
import { accessToken } from '../../composables/useUserAuth'
import TechCard from '../TechCard.vue'

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
  <TechCard variant="blue" padded>
    <h2 class="text-lg font-bold text-base-content/80 mb-5">👤 {{ t('user.profile') }}</h2>
    <div v-if="profileMsg" :class="['p-3 rounded-xl mb-4 text-sm font-medium', profileMsgType === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20']">
      {{ profileMsg }}
    </div>
    <form @submit.prevent="saveProfile" class="space-y-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.name') }}</label>
          <input v-model="profileForm.name" type="text" :placeholder="t('user.name_placeholder')"
            class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
        </div>
        <div>
          <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.email') }}</label>
          <input v-model="profileForm.email" type="email" :placeholder="t('user.email_placeholder')"
            class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
        </div>
        <div class="sm:col-span-2">
          <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.bio') }}</label>
          <textarea v-model="profileForm.bio" :placeholder="t('user.bio_placeholder')" rows="3"
            class="textarea textarea-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl resize-none"></textarea>
        </div>
        <div>
          <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.website') }}</label>
          <input v-model="profileForm.website" type="url" :placeholder="t('user.website_placeholder')"
            class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
        </div>
        <div>
          <label class="label text-sm font-medium text-base-content/70 pb-1">{{ t('user.location') }}</label>
          <input v-model="profileForm.location" type="text" :placeholder="t('user.location_placeholder')"
            class="input input-bordered w-full bg-base-200/50 border-base-300/30 rounded-xl" />
        </div>
      </div>
      <div class="pt-2">
        <button type="submit" :disabled="profileSaving" class="btn btn-primary rounded-xl">
          <span v-if="profileSaving" class="loading loading-spinner loading-xs"></span>
          <span v-else>💾</span>
          {{ profileSaving ? t('admin.settings_saving') : t('user.save_changes') }}
        </button>
      </div>
    </form>
  </TechCard>
</template>
