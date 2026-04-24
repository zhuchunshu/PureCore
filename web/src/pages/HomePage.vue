<script setup>
import { ref, onMounted, computed } from 'vue'
import { useI18n } from '../i18n'

const { t, locale } = useI18n()

const features = [
  { icon: '🔀', titleKey: 'home.feature_1_title', descKey: 'home.feature_1_desc' },
  { icon: '✅', titleKey: 'home.feature_2_title', descKey: 'home.feature_2_desc' },
  { icon: '🌐', titleKey: 'home.feature_3_title', descKey: 'home.feature_3_desc' },
  { icon: '📦', titleKey: 'home.feature_4_title', descKey: 'home.feature_4_desc' },
]

const stats = [
  { name: 'GoFiber', value: 'v3', descKey: 'home.stats_1_desc' },
  { name: 'Vue', value: '3.x', descKey: 'home.stats_2_desc' },
  { name: 'Tailwind', value: '4.x', descKey: 'home.stats_3_desc' },
  { name: 'DaisyUI', value: '5.x', descKey: 'home.stats_4_desc' },
]

const projectInfo = ref(null)
const loading = ref(true)

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
</script>

<template>
  <div>
    <!-- Hero -->
    <div class="hero min-h-[90vh]" style="background: linear-gradient(135deg, oklch(var(--b1)) 0%, oklch(var(--b2)) 50%, oklch(var(--b3)) 100%);">
      <div class="hero-content text-center">
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
            <a href="/docs/zh/README.md" class="btn btn-primary btn-lg btn-wide shadow-lg shadow-primary/20 hover:shadow-primary/40 transition-all duration-300">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
              {{ t('home.get_started') }}
            </a>
            <a href="https://github.com/zhuchunshu/PureCore" target="_blank" class="btn btn-ghost btn-lg">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
              {{ t('home.view_github') }}
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="relative z-20 -mt-20">
      <div class="max-w-4xl mx-auto px-4">
        <div class="stats stats-vertical lg:stats-horizontal shadow-2xl w-full bg-base-100/80 backdrop-blur-lg rounded-box border border-base-300/20">
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
          <div v-for="f in features" :key="f.titleKey" class="card bg-base-200/50 hover:bg-base-200 hover:shadow-lg hover:-translate-y-1 transition-all duration-300 border border-base-300/20 cursor-default">
            <div class="card-body items-center text-center p-8">
              <div class="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mb-4 text-3xl">{{ f.icon }}</div>
              <h3 class="card-title text-lg font-bold">{{ t(f.titleKey) }}</h3>
              <p class="text-sm opacity-50 leading-relaxed">{{ t(f.descKey) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- CTA -->
    <div class="py-24 px-4">
      <div class="max-w-3xl mx-auto">
        <div class="card bg-gradient-to-br from-primary to-secondary text-primary-content shadow-2xl">
          <div class="card-body text-center p-12">
            <h2 class="text-3xl md:text-4xl font-black mb-4">{{ t('greeting') }}</h2>
            <p class="text-primary-content/80 text-lg mb-8 max-w-md mx-auto">{{ t('cta.text') }}</p>
            <div class="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="/docs/zh/README.md" class="btn btn-lg bg-primary-content text-primary hover:bg-primary-content/90 border-0">{{ t('home.get_started') }}</a>
              <a href="https://github.com/zhuchunshu/PureCore" target="_blank" class="btn btn-ghost btn-lg text-primary-content hover:bg-primary-content/10">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" viewBox="0 0 24 24" class="mr-1"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
                {{ t('home.star_github') }}
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
