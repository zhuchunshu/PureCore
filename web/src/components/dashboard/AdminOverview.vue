<script setup>
import { useI18n } from '../../i18n'
import TechCard from '../TechCard.vue'

const props = defineProps({
  profile: { type: Object, default: null }
})

const { t } = useI18n()

const statCards = [
  { label: 'admin.users', value: '—', icon: '👥', gradient: 'from-blue-500/20 to-blue-600/20', iconBg: 'from-blue-500 to-blue-600' },
  { label: 'admin.active_sessions', value: '1', icon: '🔑', gradient: 'from-emerald-500/20 to-emerald-600/20', iconBg: 'from-emerald-500 to-emerald-600' },
  { label: 'admin.database', value: 'PostgreSQL', icon: '🗄️', gradient: 'from-purple-500/20 to-purple-600/20', iconBg: 'from-purple-500 to-purple-600' },
  { label: 'admin.framework', value: 'PureCore', icon: '⚡', gradient: 'from-amber-500/20 to-amber-600/20', iconBg: 'from-amber-500 to-amber-600' },
]

const actions = [
  { label: 'admin.add_user', icon: '👤', variant: 'blue' },
  { label: 'admin.view_logs', icon: '📋', variant: 'emerald' },
  { label: 'admin.backup_db', icon: '💾', variant: 'purple' },
  { label: 'admin.clear_cache', icon: '🧹', variant: 'blue' },
]
</script>

<template>
  <div class="space-y-6">
    <!-- Welcome header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl md:text-3xl font-black tracking-tight">
          <span class="bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">{{ t('admin.dashboard') }}</span>
        </h1>
        <p v-if="profile" class="text-base-content/50 mt-1">👋 {{ t('admin.welcome') }}, <span class="font-semibold text-base-content/80">{{ profile.name }}</span></p>
      </div>
    </div>

    <!-- Stats grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <TechCard v-for="s in statCards" :key="s.label" variant="blue" :hover="true" :padded="false">
        <div :class="['p-5 rounded-2xl bg-gradient-to-br', s.gradient]">
          <div class="flex items-center justify-between">
            <span class="text-3xl">{{ s.icon }}</span>
            <span :class="['w-10 h-10 rounded-xl bg-gradient-to-br flex items-center justify-center text-white text-xs font-bold shadow-lg', s.iconBg]">
              {{ s.value }}
            </span>
          </div>
          <p class="text-sm font-medium text-base-content/60 mt-3">{{ t(s.label) }}</p>
        </div>
      </TechCard>
    </div>

    <!-- Quick actions -->
    <TechCard variant="emerald" padded>
      <h2 class="text-lg font-bold text-base-content/80 mb-4">⚡ {{ t('admin.quick_actions') }}</h2>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="action in actions"
          :key="action.label"
          class="btn btn-sm gap-2 bg-base-200/80 border border-base-300/30 hover:bg-base-300/50 hover:border-base-300/50 transition-colors rounded-xl"
        >
          <span>{{ action.icon }}</span>
          <span>{{ t(action.label) }}</span>
        </button>
      </div>
    </TechCard>
  </div>
</template>
