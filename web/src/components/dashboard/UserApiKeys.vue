<script setup>
import { useI18n } from '../../i18n'
import { Key, Plus, ShieldCheck } from 'lucide-vue-next'

const { t } = useI18n()
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- ===== COL 1-2: Main ===== -->
    <div class="lg:col-span-2 space-y-4">
      <p class="text-sm text-base-content/40">{{ t('user.api_keys_description') }}</p>

      <div class="bg-base-100 border border-base-300/20 rounded-2xl shadow-sm">
        <!-- Header with create button -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-base-200">
          <h2 class="text-base font-bold text-base-content">{{ t('user.api_keys') }}</h2>
          <button class="inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-content rounded-xl px-5 py-2.5 font-semibold text-sm shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all duration-200 cursor-pointer">
            <Plus :size="16" />
            {{ t('user.create_api_key') }}
          </button>
        </div>
        <!-- Empty state -->
        <div class="text-center py-16 px-6">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-base-200 flex items-center justify-center">
            <Key :size="32" class="text-base-content/20" />
          </div>
          <p class="text-sm font-medium text-base-content/30">{{ t('user.no_api_keys') }}</p>
        </div>
      </div>
    </div>

    <!-- ===== COL 3: Sidebar ===== -->
    <div class="space-y-4">
      <div class="bg-base-100 border border-base-300/20 rounded-2xl p-5 shadow-sm">
        <h3 class="text-sm font-bold text-base-content flex items-center gap-2 mb-3">
          <ShieldCheck :size="16" class="text-primary/60" />
          {{ t('user.api_keys_info_title') }}
        </h3>
        <ul class="space-y-2 text-xs text-base-content/50">
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.api_keys_info_secret') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.api_keys_info_share') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="w-1 h-1 rounded-full bg-primary/40 mt-1.5 shrink-0"></span>
            {{ t('user.api_keys_info_rotate') }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
