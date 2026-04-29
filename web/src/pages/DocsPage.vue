<script setup>
import { ref, watch, onMounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useSEO } from '../composables/useSEO'
import { config as appConfig } from '../services/config'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css'

const props = defineProps({
  locale: { type: String, default: 'en' },
  page: { type: String, default: 'README' },
})

const router = useRouter()

useSEO({
  title: 'Documentation',
  description: 'PureCore Framework Documentation',
  keywords: 'purecore, documentation, go, vue, framework',
})

const locale = ref(props.locale)
const pages = ref([])
const currentPage = ref(props.page)
const content = ref('')
const loading = ref(false)
const error = ref('')
const showBackToTop = ref(false)
const mobileSidebarOpen = ref(false)

const apiBase = appConfig.apiBaseUrl

const md = new MarkdownIt({
  html: false,
  breaks: true,
  linkify: true,
  typographer: true,
  highlight: function (code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value
      } catch (e) { /* ignore */ }
    }
    return hljs.highlightAuto(code).value
  },
})

function renderMarkdown(raw) {
  if (!raw) return ''
  return md.render(raw)
}

async function fetchPages() {
  try {
    const resp = await fetch(`${apiBase}/docs/list?locale=${locale.value}`)
    const json = await resp.json()
    if (json.code === 0) {
      pages.value = json.data || []
      if (pages.value.length > 0 && !pages.value.find(p => p.page === currentPage.value)) {
        currentPage.value = pages.value[0].page
      }
    }
  } catch (e) {
    console.error('Failed to fetch docs list:', e)
  }
}

async function fetchDoc(page) {
  loading.value = true
  error.value = ''
  try {
    const resp = await fetch(`${apiBase}/docs?locale=${locale.value}&page=${page}`)
    const json = await resp.json()
    if (json.code === 0) {
      content.value = json.data.content
      document.title = `${json.data.page} - PureCore Docs`
      await nextTick()
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } else {
      error.value = json.message || 'Failed to load documentation'
    }
  } catch (e) {
    error.value = 'Failed to connect to backend'
    console.error('Failed to fetch doc:', e)
  } finally {
    loading.value = false
  }
}

function switchLocale(newLocale) {
  locale.value = newLocale
  if (pages.value.length > 0) {
    currentPage.value = pages.value[0].page || 'README'
  }
  router.replace({ params: { locale: newLocale, page: currentPage.value } })
}

function navigateTo(page) {
  currentPage.value = page
  router.push({ params: { locale: locale.value, page } })
  mobileSidebarOpen.value = false
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function handleScroll() {
  showBackToTop.value = window.scrollY > 300
}

onMounted(async () => {
  window.addEventListener('scroll', handleScroll)
  await fetchPages()
  if (pages.value.length > 0) {
    await fetchDoc(currentPage.value)
  }
})

import { onUnmounted } from 'vue'
onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

watch(locale, async () => {
  await fetchPages()
  if (pages.value.length > 0) {
    const found = pages.value.find(p => p.page === currentPage.value)
    if (!found) currentPage.value = pages.value[0].page
    router.replace({ params: { locale: locale.value, page: currentPage.value } })
    await fetchDoc(currentPage.value)
  }
})

const navItems = computed(() => {
  const order = ['README', 'API', 'CLI', 'DATABASE', 'DEVELOPMENT', 'SSR']
  return [...pages.value].sort((a, b) => {
    const ia = order.indexOf(a.page)
    const ib = order.indexOf(b.page)
    if (ia === -1 && ib === -1) return a.page.localeCompare(b.page)
    if (ia === -1) return 1
    if (ib === -1) return -1
    return ia - ib
  })
})

const currentPageTitle = computed(() => {
  const page = navItems.value.find(p => p.page === currentPage.value)
  return page ? page.page : currentPage.value
})
</script>

<template>
  <div class="min-h-screen bg-base-200">
    <!-- Mobile header -->
    <div class="sticky top-0 z-30 bg-base-100/80 backdrop-blur-md border-b border-base-300/50 lg:hidden">
      <div class="flex items-center justify-between px-4 py-3">
        <button
          @click="mobileSidebarOpen = !mobileSidebarOpen"
          class="btn btn-ghost btn-sm"
          aria-label="Toggle sidebar"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
          </svg>
        </button>
        <span class="text-sm font-semibold text-base-content/70">Documentation</span>
      </div>
    </div>

    <div class="flex">
      <!-- Mobile sidebar overlay -->
      <div
        v-if="mobileSidebarOpen"
        class="fixed inset-0 z-20 bg-black/50 lg:hidden"
        @click="mobileSidebarOpen = false"
      ></div>

      <!-- Sidebar -->
      <aside
        :class="[
          'fixed lg:sticky top-0 lg:top-16 z-20 h-screen lg:h-[calc(100vh-4rem)] overflow-y-auto border-r transition-all duration-300 flex-shrink-0',
          'bg-base-100 border-base-300/30',
          mobileSidebarOpen ? 'translate-x-0 w-72' : '-translate-x-full lg:translate-x-0 lg:w-64'
        ]"
      >
        <div class="p-6 pb-24">
          <div class="mb-8">
            <p class="text-xs font-semibold text-base-content/40 uppercase tracking-wider mb-3">Language</p>
            <div class="flex gap-2">
              <button
                @click="switchLocale('en')"
                :class="[
                  'btn btn-sm flex-1 transition-all duration-200',
                  locale === 'en' ? 'btn-primary shadow-lg shadow-primary/25' : 'btn-ghost hover:bg-base-200'
                ]"
              >English</button>
              <button
                @click="switchLocale('zh')"
                :class="[
                  'btn btn-sm flex-1 transition-all duration-200',
                  locale === 'zh' ? 'btn-primary shadow-lg shadow-primary/25' : 'btn-ghost hover:bg-base-200'
                ]"
              >中文</button>
            </div>
          </div>

          <div>
            <p class="text-xs font-semibold text-base-content/40 uppercase tracking-wider mb-3">Pages</p>
            <nav class="space-y-0.5">
              <button
                v-for="item in navItems"
                :key="item.page"
                @click="navigateTo(item.page)"
                :class="[
                  'w-full text-left px-3 py-2.5 rounded-lg text-sm transition-all duration-200',
                  currentPage === item.page
                    ? 'bg-primary/10 text-primary font-medium shadow-sm'
                    : 'text-base-content/60 hover:bg-base-200 hover:text-base-content'
                ]"
              >
                <div class="flex items-center gap-2">
                  <span
                    :class="[
                      'w-1.5 h-1.5 rounded-full transition-all duration-200',
                      currentPage === item.page ? 'bg-primary scale-100' : 'bg-base-300 scale-75'
                    ]"
                  ></span>
                  {{ item.page }}
                </div>
              </button>
            </nav>
          </div>
        </div>
      </aside>

      <!-- Main content -->
      <main class="flex-1 min-w-0">
        <div class="px-4 sm:px-6 lg:px-8 py-8 lg:py-12">
          <!-- Loading state -->
          <div v-if="loading" class="flex flex-col items-center justify-center py-32">
            <span class="loading loading-spinner loading-lg text-primary mb-4"></span>
            <p class="text-sm text-base-content/50">Loading documentation...</p>
          </div>

          <!-- Error state -->
          <div v-else-if="error" class="max-w-2xl mx-auto mt-8">
            <div class="alert alert-error shadow-lg">
              <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{{ error }}</span>
            </div>
          </div>

          <!-- Content -->
          <div v-else class="animate-fade-in">
            <!-- Breadcrumb -->
            <div class="flex items-center gap-2 mb-8 text-sm text-base-content/50">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
              <span>Documentation</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              <span class="text-base-content/70 font-medium">{{ currentPageTitle }}</span>
            </div>

            <!-- Document -->
            <article
              class="prose prose-base lg:prose-lg max-w-none bg-base-100 rounded-2xl shadow-sm border border-base-300/20 p-6 sm:px-6 sm:py-8 lg:px-6 lg:py-10"
              v-html="renderMarkdown(content)"
            ></article>
          </div>

          <!-- Back to top -->
          <Transition name="fade">
            <button
              v-if="showBackToTop"
              @click="scrollToTop"
              class="fixed bottom-8 right-8 btn btn-circle btn-primary shadow-xl shadow-primary/30 hover:shadow-primary/40 z-50 transition-all duration-300"
              aria-label="Back to top"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 10l7-7m0 0l7 7m-7-7v18" />
              </svg>
            </button>
          </Transition>
        </div>
      </main>
    </div>
  </div>
</template>

<style>
.animate-fade-in {
  animation: docsFadeIn 0.35s ease-out;
}

@keyframes docsFadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0; transform: translateY(12px);
}
</style>
