<script setup>
import { useTheme } from '../composables/useTheme'

const { theme, setThemeMode, resolvedTheme } = useTheme()

const modes = [
  { key: 'auto', label: 'Auto', icon: '💻' },
  { key: 'light', label: 'Light', icon: '☀️' },
  { key: 'dark', label: 'Dark', icon: '🌙' },
]

// All DaisyUI v5 themes available in theme/ directory
const themePresets = [
  { key: 'sunset', label: 'Sunset', color: 'bg-orange-500' },
  { key: 'cyberpunk', label: 'Cyberpunk', color: 'bg-yellow-400' },
  { key: 'dracula', label: 'Dracula', color: 'bg-purple-600' },
  { key: 'nord', label: 'Nord', color: 'bg-blue-400' },
  { key: 'emerald', label: 'Emerald', color: 'bg-emerald-500' },
  { key: 'corporate', label: 'Corporate', color: 'bg-blue-600' },
  { key: 'synthwave', label: 'Synthwave', color: 'bg-pink-500' },
  { key: 'retro', label: 'Retro', color: 'bg-amber-500' },
  { key: 'garden', label: 'Garden', color: 'bg-green-500' },
  { key: 'forest', label: 'Forest', color: 'bg-green-700' },
  { key: 'aqua', label: 'Aqua', color: 'bg-cyan-400' },
  { key: 'pastel', label: 'Pastel', color: 'bg-pink-300' },
  { key: 'cupcake', label: 'Cupcake', color: 'bg-pink-400' },
  { key: 'bumblebee', label: 'Bumblebee', color: 'bg-yellow-500' },
  { key: 'fantasy', label: 'Fantasy', color: 'bg-purple-400' },
  { key: 'valentine', label: 'Valentine', color: 'bg-red-400' },
  { key: 'halloween', label: 'Halloween', color: 'bg-orange-600' },
  { key: 'black', label: 'Black', color: 'bg-gray-900' },
  { key: 'luxury', label: 'Luxury', color: 'bg-yellow-600' },
  { key: 'cmyk', label: 'CMYK', color: 'bg-cyan-500' },
  { key: 'autumn', label: 'Autumn', color: 'bg-orange-700' },
  { key: 'business', label: 'Business', color: 'bg-slate-700' },
  { key: 'acid', label: 'Acid', color: 'bg-lime-400' },
  { key: 'lemonade', label: 'Lemonade', color: 'bg-yellow-300' },
  { key: 'night', label: 'Night', color: 'bg-indigo-900' },
  { key: 'coffee', label: 'Coffee', color: 'bg-amber-800' },
  { key: 'winter', label: 'Winter', color: 'bg-blue-200' },
  { key: 'dim', label: 'Dim', color: 'bg-gray-600' },
  { key: 'wireframe', label: 'Wireframe', color: 'bg-gray-300' },
  { key: 'lofi', label: 'Lofi', color: 'bg-gray-400' },
  { key: 'caramellatte', label: 'Caramellatte', color: 'bg-amber-400' },
  { key: 'abyss', label: 'Abyss', color: 'bg-blue-950' },
  { key: 'silk', label: 'Silk', color: 'bg-stone-300' },
]
</script>

<template>
  <div class="dropdown dropdown-end">
    <label tabindex="0" class="btn btn-ghost btn-sm gap-1">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
      </svg>
      <span class="hidden sm:inline text-xs opacity-70">{{ resolvedTheme() }}</span>
    </label>
    <div tabindex="0" class="dropdown-content mt-2 z-50 p-3 shadow-xl bg-base-100 rounded-box w-72 max-h-80 overflow-y-auto">
      <!-- Mode selector -->
      <div class="text-xs font-semibold text-base-content/50 mb-2 uppercase tracking-wider">Mode</div>
      <div class="grid grid-cols-3 gap-1 mb-3">
        <button
          v-for="mode in modes"
          :key="mode.key"
          @click="setThemeMode(mode.key)"
          :class="[
            'btn btn-xs gap-1',
            theme === mode.key ? 'btn-primary' : 'btn-ghost'
          ]"
        >
          <span>{{ mode.icon }}</span>
          <span>{{ mode.label }}</span>
        </button>
      </div>

      <!-- Theme presets -->
      <div class="text-xs font-semibold text-base-content/50 mb-2 uppercase tracking-wider">Themes</div>
      <div class="grid grid-cols-2 gap-1">
        <button
          v-for="preset in themePresets"
          :key="preset.key"
          @click="setThemeMode(preset.key)"
          :class="[
            'btn btn-xs gap-2 justify-start',
            resolvedTheme() === preset.key ? 'btn-primary' : 'btn-ghost'
          ]"
        >
          <span :class="['w-3 h-3 rounded-full', preset.color]"></span>
          <span>{{ preset.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>
