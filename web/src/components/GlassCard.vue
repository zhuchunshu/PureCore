<script setup>
defineProps({
  as: { type: String, default: 'div' },
  padded: { type: Boolean, default: true },
})
</script>

<template>
  <component
    :is="as"
    class="backdrop-blur-xl bg-white/5 rounded-2xl shadow-2xl border border-white/10"
    :class="{ 'p-8 md:p-10': padded }"
  >
    <slot />
  </component>
</template>
