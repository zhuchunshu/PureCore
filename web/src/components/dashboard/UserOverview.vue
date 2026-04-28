<script setup>
import { useI18n } from '../../i18n'
import TechCard from '../TechCard.vue'

const props = defineProps({
  profile: { type: Object, default: null }
})

const { t } = useI18n()

const statCards = [
  { label: 'user.posts',       value: '0', icon: '📝', gradient: 'from-blue-500/20 to-blue-600/20',   iconBg: 'from-blue-500 to-blue-600',   barColor: 'bg-blue-500',   barWidth: '0%' },
  { label: 'user.comments',    value: '0', icon: '💬', gradient: 'from-emerald-500/20 to-emerald-600/20', iconBg: 'from-emerald-500 to-emerald-600', barColor: 'bg-emerald-500', barWidth: '0%' },
  { label: 'user.likes',       value: '0', icon: '❤️', gradient: 'from-purple-500/20 to-purple-600/20',  iconBg: 'from-purple-500 to-purple-600',  barColor: 'bg-purple-500',  barWidth: '0%' },
  { label: 'user.followers',   value: '0', icon: '👥', gradient: 'from-amber-500/20 to-amber-600/20',   iconBg: 'from-amber-500 to-amber-600',   barColor: 'bg-amber-500',   barWidth: '0%' },
]

const resourceCards = [
  { label: 'user.storage',     used: 0, limit: 100, unit: 'MB',  icon: '💾', color: 'blue' },
  { label: 'user.bandwidth',   used: 0, limit: 500, unit: 'MB',  icon: '📡', color: 'emerald' },
  { label: 'user.cpu_usage',   used: 0, limit: 100, unit: '%',   icon: '⚙️', color: 'purple' },
]

function percent(used, limit) {
  if (!limit) return 0
  return Math.min(Math.round((used / limit) * 100), 100)
}
</script>

<template>
  <div class="space-y-6">
    <!-- User profile card -->
    <TechCard variant="blue" :hover="true" padded>
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <div class="w-16 h-16 rounded-full bg-gradient-to-br from-blue-400 to-emerald-400 flex items-center justify-center text-2xl font-bold text-white shrink-0 shadow-lg shadow-blue-400/20">
          {{ profile?.name?.charAt(0)?.toUpperCase() || '?' }}
        </div>
        <div class="flex-1 min-w-0">
          <h2 class="text-lg font-bold text-base-content/80">{{ profile?.name || '—' }}</h2>
          <p class="text-sm text-base-content/50">{{ profile?.email || '—' }}</p>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="profile?.status === 'active'" class="badge badge-success badge-sm">● {{ t('user.active') }}</span>
          <span v-else-if="profile?.status === 'banned'" class="badge badge-error badge-sm">● {{ t('user.banned') }}</span>
          <span v-else class="badge badge-ghost badge-sm">● {{ t('user.inactive') }}</span>
        </div>
      </div>
      <div v-if="profile?.last_login_at" class="mt-3 text-xs text-base-content/40">
        {{ t('user.last_login') }}: {{ new Date(profile.last_login_at).toLocaleString() }}
      </div>
    </TechCard>

    <!-- Stat cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <TechCard v-for="s in statCards" :key="s.label" variant="blue" :hover="true" :padded="false">
        <div :class="['p-5 rounded-2xl bg-gradient-to-br', s.gradient]">
          <div class="flex items-center justify-between mb-3">
            <span class="text-3xl">{{ s.icon }}</span>
            <span :class="['w-10 h-10 rounded-xl bg-gradient-to-br flex items-center justify-center text-white text-xs font-bold shadow-lg', s.iconBg]">
              {{ s.value }}
            </span>
          </div>
          <p class="text-sm font-medium text-base-content/60">{{ t(s.label) }}</p>
          <div class="mt-2 h-1.5 rounded-full bg-base-300/50 overflow-hidden">
            <div :class="['h-full rounded-full transition-all duration-700', s.barColor]" :style="{ width: s.barWidth }"></div>
          </div>
        </div>
      </TechCard>
    </div>

    <!-- Resource usage -->
    <div>
      <h2 class="text-lg font-bold text-base-content/80 mb-4">📊 {{ t('user.resource_usage') }}</h2>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <TechCard v-for="r in resourceCards" :key="r.label" variant="blue" :hover="true" padded>
          <div class="flex items-center gap-3 mb-3">
            <span :class="['w-10 h-10 rounded-xl bg-gradient-to-br flex items-center justify-center text-lg shadow-md',
              r.color === 'blue' ? 'from-blue-500 to-blue-600' :
              r.color === 'emerald' ? 'from-emerald-500 to-emerald-600' :
              'from-purple-500 to-purple-600']">{{ r.icon }}</span>
            <div>
              <p class="text-sm font-medium text-base-content/80">{{ t(r.label) }}</p>
              <p class="text-xs text-base-content/50">{{ t('user.usage_this_month') }}</p>
            </div>
          </div>
          <div class="h-2.5 rounded-full bg-base-300/50 overflow-hidden">
            <div
              :class="['h-full rounded-full transition-all duration-700',
                r.color === 'blue' ? 'bg-blue-500' :
                r.color === 'emerald' ? 'bg-emerald-500' :
                'bg-purple-500']"
              :style="{ width: percent(r.used, r.limit) + '%' }"
            ></div>
          </div>
          <p class="text-xs text-base-content/40 mt-2">
            {{ r.used }} / {{ r.limit }} {{ r.unit }}
            <span class="ml-1">({{ percent(r.used, r.limit) }}% {{ t('user.of_limit') }})</span>
          </p>
        </TechCard>
      </div>
    </div>

    <!-- Recent activity -->
    <TechCard variant="blue" padded>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-bold text-base-content/80">📋 {{ t('user.recent_activity') }}</h2>
        <button class="text-xs text-primary hover:underline">{{ t('user.view_all') }} →</button>
      </div>
      <div class="text-center py-8 text-base-content/40 text-sm">
        {{ t('user.no_activity') }}
      </div>
    </TechCard>
  </div>
</template>
