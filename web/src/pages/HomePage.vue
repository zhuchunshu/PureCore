<script setup>
import { ref, onMounted, computed, inject } from 'vue'
import { useI18n } from '../i18n'
import { useSEO } from '../composables/useSEO'
import CyberBackground from '../components/CyberBackground.vue'
import GridOverlay from '../components/GridOverlay.vue'
import GradientButton from '../components/GradientButton.vue'
import {
  Route, Shield, Database, Settings, Monitor, Zap,
  ArrowDown, Server, Layers, Cpu, Globe
} from 'lucide-vue-next'

const { t, locale } = useI18n()
useSEO({
  title: t('home.title'),
  description: t('home.description'),
  keywords: 'purecore, go, gofiber, vue, full-stack, framework',
})

const features = [
  {
    icon: Route,
    titleKey: 'home.feature_1_title',
    descKey: 'home.feature_1_desc',
    gradient: 'from-blue-500/20 to-cyan-500/20',
    iconColor: 'text-blue-400',
  },
  {
    icon: Shield,
    titleKey: 'home.feature_2_title',
    descKey: 'home.feature_2_desc',
    gradient: 'from-emerald-500/20 to-teal-500/20',
    iconColor: 'text-emerald-400',
  },
  {
    icon: Database,
    titleKey: 'home.feature_3_title',
    descKey: 'home.feature_3_desc',
    gradient: 'from-purple-500/20 to-pink-500/20',
    iconColor: 'text-purple-400',
  },
  {
    icon: Settings,
    titleKey: 'home.feature_4_title',
    descKey: 'home.feature_4_desc',
    gradient: 'from-orange-500/20 to-amber-500/20',
    iconColor: 'text-orange-400',
  },
  {
    icon: Monitor,
    titleKey: 'home.feature_5_title',
    descKey: 'home.feature_5_desc',
    gradient: 'from-rose-500/20 to-red-500/20',
    iconColor: 'text-rose-400',
  },
  {
    icon: Zap,
    titleKey: 'home.feature_6_title',
    descKey: 'home.feature_6_desc',
    gradient: 'from-yellow-500/20 to-lime-500/20',
    iconColor: 'text-yellow-400',
  },
]

const techCategories = [
  {
    label: t('home.tech_backend'),
    icon: Server,
    items: ['Go', 'GoFiber v3', 'GORM', 'PostgreSQL'],
  },
  {
    label: t('home.tech_frontend'),
    icon: Layers,
    items: ['Vue 3', 'Vite', 'Bun', 'Tailwind CSS 4'],
  },
  {
    label: t('home.tech_auth'),
    icon: Shield,
    items: ['JWT', 'OAuth 2.0', 'Turnstile', 'RBAC'],
  },
  {
    label: t('home.tech_deploy'),
    icon: Globe,
    items: ['Docker', 'Systemd', 'Nginx', 'CI/CD'],
  },
]

// SSR-provided project info
const ssrProjectInfo = inject('projectInfo', null)
const projectInfo = ref(ssrProjectInfo)
const loading = ref(!ssrProjectInfo)

const releaseTypeLabel = computed(() => {
  if (!projectInfo.value) return ''
  const type = projectInfo.value.release_type || projectInfo.value.ReleaseType
  if (!type) return ''
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

const goVersion = computed(() => {
  if (!projectInfo.value) return ''
  return projectInfo.value.go_version || projectInfo.value.GoVersion || ''
})

onMounted(async () => {
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

function scrollToFeatures() {
  document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<template>
  <div>
    <!-- ============================================
         Hero Section
         ============================================ -->
    <section class="relative min-h-screen flex items-center justify-center overflow-hidden">
      <!-- Background layers -->
      <GridOverlay :opacity="0.04" />
      <CyberBackground intensity="medium" />

      <!-- Floating geometric shapes -->
      <div class="absolute inset-0 pointer-events-none overflow-hidden" aria-hidden="true">
        <div class="absolute top-1/4 left-[15%] w-40 h-40 rounded-full border border-primary/10 animate-float-slow" />
        <div class="absolute top-1/3 right-[20%] w-24 h-24 rounded-2xl border border-secondary/10 rotate-12 animate-float-medium" />
        <div class="absolute bottom-1/3 left-[25%] w-20 h-20 rounded-lg border border-accent/10 -rotate-6 animate-float-fast" />
        <div class="absolute top-[20%] right-[30%] w-3 h-3 rounded-full bg-primary/20 animate-float-fast" style="animation-delay: 1s" />
        <div class="absolute bottom-[35%] right-[15%] w-4 h-4 rounded-full bg-secondary/20 animate-float-slow" style="animation-delay: 2s" />
        <div class="absolute top-[40%] left-[10%] w-2 h-2 rounded-full bg-accent/20 animate-float-medium" style="animation-delay: 0.5s" />
      </div>

      <!-- Hero content -->
      <div class="relative z-10 text-center px-4 max-w-4xl mx-auto">
        <!-- Badge row -->
        <div v-if="projectInfo" class="flex flex-wrap gap-2 justify-center mb-8">
          <span class="badge badge-primary badge-lg gap-2 px-4 py-4 shadow-lg shadow-primary/20">
            <span class="inline-block w-2 h-2 rounded-full bg-primary-content animate-pulse" />
            <span class="font-semibold">{{ releaseTypeLabel }} {{ versionText }}</span>
          </span>
          <span v-if="projectInfo.license || projectInfo.License" class="badge badge-ghost badge-lg px-4 py-4">
            {{ projectInfo.license || projectInfo.License }}
          </span>
          <span v-if="goVersion" class="badge badge-outline badge-lg px-4 py-4 gap-1">
            <Cpu :size="14" />
            {{ goVersion }}
          </span>
        </div>

        <!-- Skeleton for badge row -->
        <div v-else class="flex flex-wrap gap-2 justify-center mb-8">
          <div class="skeleton h-10 w-36 rounded-full" />
          <div class="skeleton h-10 w-24 rounded-full" />
        </div>

        <!-- Title -->
        <h1 class="text-6xl md:text-8xl font-black tracking-tight mb-6">
          <span class="bg-gradient-to-r from-primary via-secondary to-accent bg-clip-text text-transparent">
            PureCore
          </span>
        </h1>

        <!-- Subtitle -->
        <p class="text-xl md:text-2xl font-light text-base-content/70 mb-4 max-w-2xl mx-auto leading-relaxed">
          {{ t('home.subtitle') }}
        </p>

        <!-- Description -->
        <p class="text-base md:text-lg text-base-content/40 mb-10 max-w-xl mx-auto leading-relaxed">
          {{ t('home.description') }}
        </p>

        <!-- CTA buttons -->
        <div class="flex flex-col sm:flex-row gap-4 justify-center mb-16">
          <GradientButton href="/docs/zh/README" variant="blue" size="lg">
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

        <!-- Scroll indicator -->
        <button
          class="flex flex-col items-center gap-2 text-base-content/20 hover:text-primary transition-colors duration-500 mx-auto animate-bounce cursor-pointer"
          @click="scrollToFeatures"
          aria-label="Scroll to features"
        >
          <span class="text-xs uppercase tracking-widest">{{ t('home.scroll_down') }}</span>
          <ArrowDown :size="20" class="animate-bounce" />
        </button>
      </div>
    </section>

    <!-- ============================================
         Features Section
         ============================================ -->
    <section id="features" class="py-24 lg:py-32 bg-base-100">
      <div class="max-w-6xl mx-auto px-4">
        <!-- Section header -->
        <div class="text-center mb-16">
          <span class="badge badge-primary badge-sm mb-4">{{ t('home.features_badge') }}</span>
          <h2 class="text-4xl md:text-5xl font-black tracking-tight mb-4">
            <span class="bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
              {{ t('home.features_title') }}
            </span>
          </h2>
          <p class="text-lg text-base-content/40 max-w-lg mx-auto">
            {{ t('home.features_subtitle') }}
          </p>
        </div>

        <!-- Feature cards grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div
            v-for="f in features"
            :key="f.titleKey"
            class="group relative bg-base-200/40 backdrop-blur-sm rounded-2xl p-6 border border-base-300/20 hover:border-primary/30 hover:shadow-xl hover:shadow-primary/5 hover:-translate-y-1 transition-all duration-500 overflow-hidden cursor-default"
          >
            <!-- Hover gradient -->
            <div
              :class="['absolute inset-0 bg-gradient-to-br opacity-0 group-hover:opacity-100 transition-opacity duration-500', f.gradient]"
            />
            <!-- Icon -->
            <div class="relative z-10 w-12 h-12 rounded-xl bg-base-100 border border-base-300/20 flex items-center justify-center mb-4 group-hover:scale-110 group-hover:border-primary/20 transition-all duration-300">
              <component :is="f.icon" :size="24" :class="['transition-colors duration-300', f.iconColor]" />
            </div>
            <!-- Content -->
            <div class="relative z-10">
              <h3 class="text-lg font-bold mb-2">{{ t(f.titleKey) }}</h3>
              <p class="text-sm text-base-content/50 leading-relaxed">{{ t(f.descKey) }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============================================
         Tech Stack Section
         ============================================ -->
    <section class="py-24 bg-base-200/50">
      <div class="max-w-6xl mx-auto px-4">
        <div class="text-center mb-16">
          <span class="badge badge-secondary badge-sm mb-4">{{ t('home.tech_badge') }}</span>
          <h2 class="text-3xl md:text-4xl font-black tracking-tight mb-4">
            {{ t('home.tech_title') }}
          </h2>
          <p class="text-lg text-base-content/40 max-w-lg mx-auto">
            {{ t('home.tech_subtitle') }}
          </p>
        </div>

        <!-- Tech grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div
            v-for="cat in techCategories"
            :key="cat.label"
            class="bg-base-100/80 backdrop-blur-sm rounded-2xl p-6 border border-base-300/20 hover:border-primary/10 hover:shadow-md transition-all duration-300"
          >
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
                <component :is="cat.icon" :size="18" class="text-primary" />
              </div>
              <span class="font-bold text-sm">{{ cat.label }}</span>
            </div>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="item in cat.items"
                :key="item"
                class="badge badge-ghost badge-sm"
              >{{ item }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============================================
         Stats Bar
         ============================================ -->
    <section class="relative z-10 -mt-12 mb-24">
      <div class="max-w-4xl mx-auto px-4">
        <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-base-100 via-base-100 to-primary/5 border border-primary/10 shadow-2xl shadow-primary/5">
          <div class="absolute inset-0 bg-grid-pattern opacity-[0.03] pointer-events-none" />
          <div class="relative grid grid-cols-2 lg:grid-cols-4 divide-x divide-base-300/20">
            <div class="p-6 text-center">
              <div class="text-2xl md:text-3xl font-black text-primary mb-1">
                {{ versionText || '—' }}
              </div>
              <div class="text-xs text-base-content/40 uppercase tracking-wider">{{ t('home.stats_version') }}</div>
            </div>
            <div class="p-6 text-center">
              <div class="text-2xl md:text-3xl font-black text-secondary mb-1">
                {{ goVersion || '—' }}
              </div>
              <div class="text-xs text-base-content/40 uppercase tracking-wider">{{ t('home.stats_go') }}</div>
            </div>
            <div class="p-6 text-center">
              <div class="text-2xl md:text-3xl font-black text-accent mb-1">
                v3
              </div>
              <div class="text-xs text-base-content/40 uppercase tracking-wider">GoFiber</div>
            </div>
            <div class="p-6 text-center">
              <div class="text-2xl md:text-3xl font-black text-success mb-1">
                5
              </div>
              <div class="text-xs text-base-content/40 uppercase tracking-wider">{{ t('home.stats_providers') }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============================================
         CTA Section
         ============================================ -->
    <section class="py-24 lg:py-32">
      <div class="max-w-3xl mx-auto px-4 text-center">
        <h2 class="text-3xl md:text-5xl font-black tracking-tight mb-4">
          {{ t('home.cta_title') }}
        </h2>
        <p class="text-lg text-base-content/40 mb-10 max-w-lg mx-auto leading-relaxed">
          {{ t('home.cta_subtitle') }}
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <GradientButton href="/docs/zh/README" variant="blue" size="lg">
            {{ t('home.get_started') }}
          </GradientButton>
          <GradientButton href="https://github.com/zhuchunshu/PureCore" variant="purple" size="lg">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
            </template>
            {{ t('home.view_github') }}
          </GradientButton>
        </div>
        <p class="mt-6 text-sm text-base-content/30">
          <a href="/docs/zh/README" class="hover:text-primary transition-colors underline underline-offset-4 decoration-base-content/10 hover:decoration-primary/30">
            {{ t('home.cta_docs_link') }}
          </a>
        </p>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* Floating shape animations */
@keyframes float-slow {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(5deg); }
}
@keyframes float-medium {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-15px) rotate(-3deg); }
}
@keyframes float-fast {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-10px) scale(1.1); }
}
.animate-float-slow { animation: float-slow 8s ease-in-out infinite; }
.animate-float-medium { animation: float-medium 6s ease-in-out infinite; }
.animate-float-fast { animation: float-fast 4s ease-in-out infinite; }

/* Grid pattern for stats bar */
.bg-grid-pattern {
  background-image:
    linear-gradient(rgba(128,128,128,0.1) 1px, transparent 1px),
    linear-gradient(90deg, rgba(128,128,128,0.1) 1px, transparent 1px);
  background-size: 20px 20px;
}
</style>
