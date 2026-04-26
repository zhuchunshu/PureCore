<script setup>
/**
 * Styled text/password input with icon and glowing focus ring.
 * Props:
 *   - modelValue: v-model binding
 *   - type: 'text' | 'password'
 *   - placeholder: placeholder text
 *   - variant: 'blue' | 'emerald' | 'purple'
 *   - label: optional label text
 *   - icon: optional true to show default icon based on type
 */
defineProps({
  modelValue: { type: String, default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  variant: { type: String, default: 'blue' },
  label: { type: String, default: '' },
  icon: { type: Boolean, default: true },
})

const emit = defineEmits(['update:modelValue'])

const variantClasses = {
  blue: 'focus:ring-blue-500/50 placeholder-blue-200/40',
  emerald: 'focus:ring-emerald-500/50 placeholder-emerald-200/40',
  purple: 'focus:ring-purple-500/50 placeholder-purple-200/40',
}

const iconColors = {
  blue: 'text-blue-300/50',
  emerald: 'text-emerald-300/50',
  purple: 'text-purple-300/50',
}

function onInput(e) {
  emit('update:modelValue', e.target.value)
}
</script>

<template>
  <div>
    <label v-if="label" class="block text-sm font-medium opacity-80 mb-2">{{ label }}</label>
    <div class="relative">
      <div v-if="icon" class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <!-- Person icon for username/name -->
        <svg v-if="type === 'text'" xmlns="http://www.w3.org/2000/svg" :class="['h-5 w-5', iconColors[variant] || iconColors.blue]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <!-- Lock icon for password -->
        <svg v-else xmlns="http://www.w3.org/2000/svg" :class="['h-5 w-5', iconColors[variant] || iconColors.blue]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
      </div>
      <input
        :value="modelValue"
        :type="type"
        :placeholder="placeholder"
        @input="onInput"
        :class="[
          'w-full py-3 bg-white/10 border border-white/10 rounded-xl text-white transition-all',
          variantClasses[variant] || variantClasses.blue,
          icon ? 'pl-10 pr-4' : 'px-4',
          'focus:outline-none focus:ring-2 focus:border-transparent',
        ]"
      />
    </div>
  </div>
</template>
