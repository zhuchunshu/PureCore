<script setup>
import { computed } from 'vue'
import { createAvatar } from '@dicebear/core'
import { initials } from '@dicebear/collection'

const props = defineProps({
  name: { type: String, default: '' },
  size: { type: String, default: 'md', validator: (v) => ['xs', 'sm', 'md', 'lg'].includes(v) },
  rounded: { type: Boolean, default: true },
})

const sizeClasses = {
  xs: 'w-7 h-7',
  sm: 'w-10 h-10',
  md: 'w-14 h-14',
  lg: 'w-20 h-20',
}

const roundedClass = computed(() => props.rounded ? 'rounded-full' : 'rounded-lg')
const sizeClass = computed(() => sizeClasses[props.size] || sizeClasses.sm)

const avatarDataUri = computed(() => {
  const seed = props.name || '?'
  const avatar = createAvatar(initials, { seed })
  return avatar.toDataUri()
})
</script>

<template>
  <div class="avatar">
    <div
      :class="[
        sizeClass,
        roundedClass,
        'overflow-hidden',
        'shadow-md hover:shadow-lg transition-shadow duration-200',
      ]"
    >
      <img
        :src="avatarDataUri"
        :alt="props.name || 'Avatar'"
        class="w-full h-full object-cover"
      />
    </div>
  </div>
</template>
