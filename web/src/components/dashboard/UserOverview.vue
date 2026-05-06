<script setup>
import { useI18n } from '../../i18n'
import AvatarInitials from '../AvatarInitials.vue'
import {
  FileText, MessageSquare, Heart, Users,
  HardDrive, Radio, Cpu,
  ClipboardList, Circle, ArrowUpRight,
  Activity, Gauge, Zap, Clock
} from 'lucide-vue-next'

const props = defineProps({
  profile: { type: Object, default: null }
})

const { t } = useI18n()

const statCards = [
  { label: 'user.posts',     value: '0', icon: FileText,      color: 'text-blue-500',      bg: 'bg-blue-500/10' },
  { label: 'user.comments',  value: '0', icon: MessageSquare, color: 'text-emerald-500',   bg: 'bg-emerald-500/10' },
  { label: 'user.likes',     value: '0', icon: Heart,         color: 'text-purple-500',    bg: 'bg-purple-500/10' },
  { label: 'user.followers', value: '0', icon: Users,         color: 'text-amber-500',     bg: 'bg-amber-500/10' },
]

const resourceCards = [
  { label: 'user.storage',   used: 0, limit: 100, unit: 'MB',  icon: HardDrive, color: 'bg-blue-500',    textColor: 'text-blue-500',    bg: 'bg-blue-500/10' },
  { label: 'user.bandwidth', used: 0, limit: 500, unit: 'MB',  icon: Radio,     color: 'bg-emerald-500', textColor: 'text-emerald-500', bg: 'bg-emerald-500/10' },
  { label: 'user.cpu_usage', used: 0, limit: 100, unit: '%',   icon: Cpu,       color: 'bg-purple-500',  textColor: 'text-purple-500',  bg: 'bg-purple-500/10' },
]

function percent(used, limit) {
  if (!limit) return 0
  return Math.min(Math.round((used / limit) * 100), 100)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Welcome message -->
    <p class="text-sm text-base-content/40">{{ t('user.welcome_back') }}<span v-if="profile?.name">, {{ profile.name }}</span></p>

    <!-- 3-column layout -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- ===== COL 1: Profile + Stats ===== -->
      <div class="space-y-4">
        <!-- User profile card -->
        <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
          <div class="flex items-center gap-4 mb-4">
            <AvatarInitials :name="profile?.name" size="lg" />
            <div class="min-w-0">
              <h2 class="text-lg font-bold text-base-content">{{ profile?.name || '—' }}</h2>
              <p class="text-sm text-base-content/40 truncate">{{ profile?.email || '—' }}</p>
            </div>
          </div>
          <div class="flex items-center justify-between">
            <span v-if="profile?.last_login_at" class="text-xs text-base-content/30">
              {{ t('user.last_login') }}: {{ new Date(profile.last_login_at).toLocaleString() }}
            </span>
            <span v-if="profile?.status === 'active'" class="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-500 bg-emerald-500/10 px-3 py-1.5 rounded-full">
              <Circle :size="6" class="fill-current" /> {{ t('user.active') }}
            </span>
            <span v-else-if="profile?.status === 'banned'" class="inline-flex items-center gap-1.5 text-xs font-medium text-red-400 bg-red-500/10 px-3 py-1.5 rounded-full">
              <Circle :size="6" class="fill-current" /> {{ t('user.banned') }}
            </span>
            <span v-else class="inline-flex items-center gap-1.5 text-xs font-medium text-base-content/40 bg-base-200 px-3 py-1.5 rounded-full">
              <Circle :size="6" class="fill-current" /> {{ t('user.inactive') }}
            </span>
          </div>
        </div>

        <!-- Stat cards (2x2 grid) -->
        <div class="grid grid-cols-2 gap-3">
          <div
            v-for="s in statCards"
            :key="s.label"
            class="bg-base-100 border border-base-300/20 rounded-2xl p-4 shadow-sm"
          >
            <div :class="['w-9 h-9 rounded-lg flex items-center justify-center mb-2', s.bg]">
              <component :is="s.icon" :size="18" :class="s.color" />
            </div>
            <p class="text-xl font-bold text-base-content">{{ s.value }}</p>
            <p class="text-xs text-base-content/50 mt-0.5">{{ t(s.label) }}</p>
          </div>
        </div>
      </div>

      <!-- ===== COL 2: Resource Usage ===== -->
      <div class="space-y-4">
        <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
          <h2 class="text-sm font-bold text-base-content flex items-center gap-2 mb-4">
            <Gauge :size="18" class="text-base-content/40" />
            {{ t('user.resource_usage') }}
          </h2>
          <div class="space-y-4">
            <div v-for="r in resourceCards" :key="r.label" :class="['rounded-xl p-4', r.bg]">
              <div class="flex items-center justify-between mb-2">
                <span :class="['w-8 h-8 rounded-lg bg-base-100 flex items-center justify-center shadow-sm', r.textColor]">
                  <component :is="r.icon" :size="16" />
                </span>
                <span class="text-lg font-bold" :class="r.textColor">{{ percent(r.used, r.limit) }}%</span>
              </div>
              <p class="text-sm font-medium text-base-content/80">{{ t(r.label) }}</p>
              <p class="text-xs text-base-content/50 mt-0.5">{{ r.used }} / {{ r.limit }} {{ r.unit }}</p>
              <div class="h-2 rounded-full bg-base-200 overflow-hidden mt-2">
                <div :class="['h-full rounded-full transition-all duration-700', r.color]" :style="{ width: percent(r.used, r.limit) + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== COL 3: Activity + Quick Actions ===== -->
      <div class="space-y-4">
        <!-- Recent activity -->
        <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-sm font-bold text-base-content flex items-center gap-2">
              <Activity :size="18" class="text-base-content/40" />
              {{ t('user.recent_activity') }}
            </h2>
            <button class="text-xs text-primary hover:text-primary/70 font-medium flex items-center gap-1 transition-colors">
              {{ t('user.view_all') }} <ArrowUpRight :size="12" />
            </button>
          </div>
          <div class="text-center py-8 text-base-content/30 text-sm">
            <Clock :size="32" class="mx-auto mb-2 opacity-30" />
            {{ t('user.no_activity') }}
          </div>
        </div>

        <!-- Quick tips card -->
        <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
          <h2 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
            <Zap :size="18" class="text-amber-400" />
            {{ t('user.quick_tips') || 'Quick Tips' }}
          </h2>
          <ul class="space-y-2 text-xs text-base-content/50">
            <li class="flex items-start gap-2">
              <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
              {{ t('user.tip_profile') || 'Complete your profile to get the most out of PureCore.' }}
            </li>
            <li class="flex items-start gap-2">
              <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
              {{ t('user.tip_security') || 'Enable two-factor authentication for better account security.' }}
            </li>
            <li class="flex items-start gap-2">
              <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
              {{ t('user.tip_api') || 'Create API keys to access PureCore programmatically.' }}
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
