<script setup>
defineProps({
  to: { type: String, default: null },
  href: { type: String, default: null },
  size: { type: String, default: 'md' },
  variant: { type: String, default: 'blue' }, // 'blue', 'emerald', 'purple'
  loading: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
})

const variantClasses = {
  blue: 'from-blue-500 to-purple-600 shadow-purple-500/25 hover:shadow-purple-500/50 hover:shadow-purple-500/30 hover:from-blue-400 hover:to-purple-500 ring-1 ring-purple-500/30 hover:ring-purple-400/50',
  emerald: 'from-emerald-500 to-teal-600 shadow-teal-500/25 hover:shadow-teal-500/50 hover:shadow-teal-500/30 hover:from-emerald-400 hover:to-teal-500 ring-1 ring-teal-500/30 hover:ring-teal-400/50',
  purple: 'from-purple-500 to-pink-600 shadow-pink-500/25 hover:shadow-pink-500/50 hover:shadow-pink-500/30 hover:from-purple-400 hover:to-pink-500 ring-1 ring-pink-500/30 hover:ring-pink-400/50',
}

const sizeClasses = {
  sm: 'py-2 px-3 text-sm',
  md: 'py-3 px-4',
  lg: 'py-4 px-6 text-lg',
}
</script>

<template>
  <component
    :is="to ? 'router-link' : href ? 'a' : 'button'"
    :to="to"
    :href="href"
    :target="href ? '_blank' : undefined"
    :disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center gap-2 font-semibold rounded-xl shadow-lg transition-all duration-300 bg-gradient-to-r text-white hover:scale-[1.02]',
      variantClasses[variant] || variantClasses.blue,
      sizeClasses[size] || sizeClasses.md,
      (disabled || loading) ? 'opacity-50 cursor-not-allowed' : '',
    ]"
  >
    <span v-if="loading" class="loading loading-spinner loading-sm"></span>
    <slot v-else name="icon" />
    <slot />
  </component>
</template>
