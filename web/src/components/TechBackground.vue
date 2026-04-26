<script setup>
/**
 * Reusable animated tech background component.
 * Features a dot grid pattern with floating glowing orbs.
 * Props:
 *   - variant: 'blue' | 'purple' | 'emerald' - color theme for the orbs
 *   - opacity: number - background opacity (0-1)
 */
defineProps({
  variant: { type: String, default: 'blue' },
  opacity: { type: Number, default: 0.1 },
})

const orbColors = {
  blue: ['from-blue-500/20', 'to-purple-500/20', 'from-cyan-500/20', 'to-blue-500/20'],
  purple: ['from-purple-500/20', 'to-pink-500/20', 'from-indigo-500/20', 'to-purple-500/20'],
  emerald: ['from-emerald-500/20', 'to-teal-500/20', 'from-cyan-500/20', 'to-blue-500/20'],
}
</script>

<template>
  <div class="absolute inset-0 overflow-hidden pointer-events-none">
    <div
      class="absolute inset-0"
      :style="{
        backgroundImage: `radial-gradient(circle at 1px 1px, rgba(255,255,255,${opacity * 1.5}) 1px, transparent 0)`,
        backgroundSize: '40px 40px',
      }"
    ></div>
    <div
      v-for="(color, i) in orbColors[variant] || orbColors.blue"
      :key="i"
      :class="[
        'absolute w-96 h-96 bg-gradient-to-r rounded-full blur-3xl animate-pulse',
        color,
      ]"
      :style="{
        top: i % 2 === 0 ? '25%' : '60%',
        left: i % 2 === 0 ? '-12%' : '70%',
        animationDelay: `${i * 2}s`,
      }"
    ></div>
  </div>
</template>
