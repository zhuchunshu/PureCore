<script setup>
/**
 * Glass card with gradient border, suitable for tech-forward UI.
 * Props:
 *   - variant: 'blue' | 'purple' | 'emerald' - accent color
 *   - padded: boolean - whether to add padding
 *   - hover: boolean - whether to add hover lift effect
 */
defineProps({
  variant: { type: String, default: 'blue' },
  padded: { type: Boolean, default: true },
  hover: { type: Boolean, default: false },
})

const borderColors = {
  blue: 'border-blue-500/20 hover:border-blue-400/40',
  purple: 'border-purple-500/20 hover:border-purple-400/40',
  emerald: 'border-emerald-500/20 hover:border-emerald-400/40',
}
</script>

<template>
  <div
    :class="[
      'backdrop-blur-xl bg-white/5 rounded-2xl shadow-2xl border transition-all duration-300',
      borderColors[variant] || borderColors.blue,
      { 'p-6 md:p-8': padded, 'hover:-translate-y-1 hover:shadow-3xl': hover },
    ]"
  >
    <slot />
  </div>
</template>
