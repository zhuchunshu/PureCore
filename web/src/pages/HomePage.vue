<script setup>
import { ref, onMounted, computed, inject } from 'vue'
import { useI18n } from '../i18n'
import ParticleBackground from '../components/ParticleBackground.vue'
import GridOverlay from '../components/GridOverlay.vue'
import GradientButton from '../components/GradientButton.vue'

const { t, locale } = useI18n()

const features = [
  { svg: 'routing', titleKey: 'home.feature_1_title', descKey: 'home.feature_1_desc' },
  { svg: 'validation', titleKey: 'home.feature_2_title', descKey: 'home.feature_2_desc' },
  { svg: 'i18n', titleKey: 'home.feature_3_title', descKey: 'home.feature_3_desc' },
  { svg: 'response', titleKey: 'home.feature_4_title', descKey: 'home.feature_4_desc' },
]

const stats = [
  { name: 'GoFiber', value: 'v3', descKey: 'home.stats_1_desc' },
  { name: 'Vue', value: '3.x', descKey: 'home.stats_2_desc' },
  { name: 'Tailwind', value: '4.x', descKey: 'home.stats_3_desc' },
  { name: 'DaisyUI', value: '5.x', descKey: 'home.stats_4_desc' },
]

// Use SSR-provided project info if available (eliminates hydration flash)
const ssrProjectInfo = inject('projectInfo', null)
const projectInfo = ref(ssrProjectInfo)
const loading = ref(!ssrProjectInfo)

const releaseTypeLabel = computed(() => {
  if (!projectInfo.value) return t('version.loading')
  const type = projectInfo.value.release_type || projectInfo.value.ReleaseType
  if (!type) return t('version.unknown')
  return t(`version.${type}`) || type
})

const versionText = computed(() => {
  if (!projectInfo.value) return ''
  return projectInfo.value.version || projectInfo.value.Version || ''
})

const authorName = computed(() => {
  if (!projectInfo.value) return ''
  const author = projectInfo.value.author || projectInfo.value.Author
  return author ? author.name : ''
})

const description = computed(() => {
  if (!projectInfo.value) return ''
  const desc = projectInfo.value.description || projectInfo.value.Description
  if (!desc) return ''
  return desc[locale.value] || desc.en || ''
})

onMounted(async () => {
  // Only fetch if SSR didn't provide it (e.g., direct client-side navigation)
  if (projectInfo.value) {
    loading.value = false
    return
  }
  try {
    const resp = await fetch('/api/v1/system/info')
    const json = await resp.json()
    if (json.code === 0) {
      projectInfo.value = json.data
    }
  } catch (err) {
    console.error('Failed to fetch project info:', err)
  } finally {
    loading.value = false
  }
})
const themeParticleColor = computed(() => {
  // Shooting stars look better with a bright, ethereal color
  // Bright blue-white for dark themes, darker blue-gray for light themes
  if (typeof window !== 'undefined') {
    const isDark = document.documentElement.getAttribute('data-theme')?.includes('dark') ||
      window.matchMedia('(prefers-color-scheme: dark)').matches
    return isDark ? '147, 197, 253' : '59, 130, 246' // sky-300 for dark, blue-500 for light
  }
  return '59, 130, 246'
})

</script>

<template>
  <div>
    <!-- Hero -->
    <div class="relative hero min-h-[90vh] bg-base-100 overflow-hidden">
      <GridOverlay :opacity="0.04" />
      <ParticleBackground :particle-color="themeParticleColor" :particle-count="80" :speed="1.0" />
      <div class="hero-content text-center relative z-10">
        <div class="max-w-2xl">
          <div class="flex flex-wrap gap-2 justify-center mb-8">
            <span class="badge badge-primary badge-lg gap-1">
              <span class="inline-block w-2 h-2 rounded-full bg-primary-content animate-pulse"></span>
              {{ releaseTypeLabel }} {{ versionText }}
            </span>
            <span class="badge badge-ghost badge-lg">{{ t('home.badge_license') }}</span>
          </div>
          <h1 class="text-5xl md:text-7xl font-black tracking-tight">
            <span class="bg-gradient-to-r from-primary via-secondary to-accent bg-clip-text text-transparent">PureCore</span>
          </h1>
          <p class="py-6 text-xl md:text-2xl font-light opacity-80">{{ t('home.subtitle') }}</p>
          <p class="mb-10 max-w-xl mx-auto leading-relaxed text-base-content/60">{{ t('home.description') }}</p>
          <div class="flex flex-col sm:flex-row gap-4 justify-center">
            <GradientButton href="/docs/zh/README.md" variant="blue" size="lg">
              <template #icon>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
              </template>
              {{ t('home.get_started') }}
            </GradientButton>
            <GradientButton href="https://github.com/zhuchunshu/PureCore" variant="purple" size="lg">
              <template #icon>
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              </template>
              {{ t('home.view_github') }}
            </GradientButton>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="relative z-20 -mt-20">
      <div class="max-w-4xl mx-auto px-4">
        <div class="stats stats-vertical lg:stats-horizontal shadow-2xl w-full bg-base-100/60 backdrop-blur-xl rounded-box border border-primary/20 shadow-primary/10">
          <div v-for="s in stats" :key="s.name" class="stat place-items-center">
            <div class="stat-title text-sm opacity-60">{{ s.name }}</div>
            <div class="stat-value text-2xl font-black text-primary">{{ s.value }}</div>
            <div class="stat-desc text-xs opacity-40">{{ t(s.descKey) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Features -->
    <div class="bg-base-100 pt-32 pb-24">
      <div class="max-w-6xl mx-auto px-4">
        <div class="text-center mb-16">
          <h2 class="text-4xl font-black">{{ t('home.features_title') }}</h2>
          <div class="divider w-16 mx-auto my-4"></div>
          <p class="text-lg opacity-50 max-w-lg mx-auto">{{ t('home.features_subtitle') }}</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div v-for="f in features" :key="f.titleKey" class="relative group card bg-base-200/30 backdrop-blur-sm hover:bg-base-200/60 hover:shadow-xl hover:-translate-y-2 transition-all duration-500 border border-primary/10 hover:border-primary/30 cursor-default overflow-hidden">
            <div class="absolute inset-0 bg-gradient-to-br from-primary/5 to-secondary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
            <div class="card-body items-center text-center p-8 relative z-10">
              <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform duration-500">
                <svg v-if="f.svg === 'routing'" xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"/></svg>
                <svg v-else-if="f.svg === 'validation'" xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/></svg>
                <svg v-else-if="f.svg === 'i18n'" xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418"/></svg>
                <svg v-else-if="f.svg === 'response'" xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17.25 6.75L22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3l-4.5 16.5"/></svg>
              </div>
              <h3 class="card-title text-lg font-bold">{{ t(f.titleKey) }}</h3>
              <p class="text-sm opacity-50 leading-relaxed">{{ t(f.descKey) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- CTA -->
    <div class="relative py-24 px-4">
      <div class="absolute inset-0 bg-gradient-to-b from-transparent via-primary/5 to-transparent opacity-30 pointer-events-none"></div>
      <div class="max-w-3xl mx-auto relative">
        <div class="card bg-gradient-to-br from-primary to-secondary text-primary-content shadow-2xl shadow-primary/30 ring-1 ring-primary/20">
          <div class="card-body text-center p-12">
            <h2 class="text-3xl md:text-4xl font-black mb-4">{{ t('greeting') }}</h2>
            <p class="text-primary-content/80 text-lg mb-8 max-w-md mx-auto">{{ t('cta.text') }}</p>
            <div class="flex flex-col sm:flex-row gap-4 justify-center">
              <GradientButton href="/docs/zh/README.md" variant="blue" size="lg">
                {{ t('home.get_started') }}
              </GradientButton>
              <GradientButton href="https://github.com/zhuchunshu/PureCore" variant="purple" size="lg">
                <template #icon>
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" viewBox="0 0 24 24" class="mr-1"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
                </template>
                {{ t('home.star_github') }}
              </GradientButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
