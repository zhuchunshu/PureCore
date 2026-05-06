<script setup>
/**
 * PureCore branded loading component.
 *
 * A centered loading state with gradient logo text, animated spinner ring,
 * and a customizable status message. Designed for transitional pages like
 * OAuth callbacks where skeleton blocks cause jarring layout shifts.
 *
 * Props:
 *   text      — Status message shown below the spinner
 *   fullscreen — When true (default), centers in full viewport height
 */
defineProps({
  text: { type: String, default: 'Loading...' },
  fullscreen: { type: Boolean, default: true },
})
</script>

<template>
  <div
    :class="fullscreen ? 'min-h-[calc(100vh-4rem)]' : ''"
    class="flex flex-col items-center justify-center gap-8 px-4"
  >
    <!-- PureCore gradient logo -->
    <div class="flex flex-col items-center gap-2 animate-pulse">
      <h1 class="text-4xl font-black tracking-tight">
        <span class="bg-gradient-to-r from-primary via-secondary to-accent bg-clip-text text-transparent">
          PureCore
        </span>
      </h1>
    </div>

    <!-- Animated ring spinner -->
    <div class="relative w-14 h-14">
      <!-- Track -->
      <div class="absolute inset-0 rounded-full border-4 border-base-300/20" />
      <!-- Spinning arc -->
      <div
        class="absolute inset-0 rounded-full border-4 border-transparent animate-spin"
        style="border-top-color: oklch(var(--p)); border-right-color: oklch(var(--s)/0.6); border-bottom-color: oklch(var(--a)/0.3);"
      />
    </div>

    <!-- Status text -->
    <p class="text-sm text-base-content/40 tracking-wide">{{ text }}</p>
  </div>
</template>
