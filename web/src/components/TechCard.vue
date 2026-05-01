<script setup>
/**
 * Tech card with gradient border, corner accents, and scanline effect.
 * Suitable for cyberpunk / tech-forward UI.
 * Props:
 *   - variant: 'blue' | 'purple' | 'emerald' - accent color
 *   - padded: boolean - whether to add padding
 *   - hover: boolean - whether to add hover lift + glow effect
 *   - scanline: boolean - whether to show subtle scanline overlay
 */
defineProps({
  variant: { type: String, default: 'blue' },
  padded: { type: Boolean, default: true },
  hover: { type: Boolean, default: false },
  scanline: { type: Boolean, default: false },
})

const accentMap = {
  blue: {
    border: 'border-blue-500/20 hover:border-blue-400/40',
    glow: 'hover:shadow-blue-500/10',
    topBar: 'from-blue-400 via-blue-500 to-blue-400',
    corner: 'bg-blue-400',
    cornerGlow: 'shadow-blue-400/50',
  },
  purple: {
    border: 'border-purple-500/20 hover:border-purple-400/40',
    glow: 'hover:shadow-purple-500/10',
    topBar: 'from-purple-400 via-purple-500 to-purple-400',
    corner: 'bg-purple-400',
    cornerGlow: 'shadow-purple-400/50',
  },
  emerald: {
    border: 'border-emerald-500/20 hover:border-emerald-400/40',
    glow: 'hover:shadow-emerald-500/10',
    topBar: 'from-emerald-400 via-emerald-500 to-emerald-400',
    corner: 'bg-emerald-400',
    cornerGlow: 'shadow-emerald-400/50',
  },
}

const accent = (v) => accentMap[v] || accentMap.blue
</script>

<template>
  <div
    :class="[
      'relative backdrop-blur-xl bg-white/[0.03] rounded-2xl border transition-all duration-500 overflow-hidden group',
      accent(variant).border,
      accent(variant).glow,
      {
        'p-6 md:p-8': padded,
        'hover:-translate-y-1 hover:shadow-[0_0_30px_rgba(59,130,246,0.08)]': hover,
      },
    ]"
  >
    <!-- Top accent gradient bar -->
    <div
      :class="[
        'absolute top-0 left-0 right-0 h-[1px] bg-gradient-to-r opacity-60',
        accent(variant).topBar,
      ]"
    />

    <!-- Corner accent dots -->
    <span
      :class="[
        'absolute top-0 left-0 w-2 h-2 rounded-full -translate-x-1/2 -translate-y-1/2 shadow-[0_0_6px]',
        accent(variant).corner,
        accent(variant).cornerGlow,
      ]"
    />
    <span
      :class="[
        'absolute top-0 right-0 w-2 h-2 rounded-full translate-x-1/2 -translate-y-1/2 shadow-[0_0_6px]',
        accent(variant).corner,
        accent(variant).cornerGlow,
      ]"
    />

    <!-- Subtle scanline overlay -->
    <div
      v-if="scanline"
      class="pointer-events-none absolute inset-0 rounded-2xl overflow-hidden opacity-[0.03]"
      aria-hidden="true"
    >
      <div
        class="absolute inset-0"
        style="background: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(255,255,255,0.4) 2px, rgba(255,255,255,0.4) 3px);"
      />
    </div>

    <!-- Content -->
    <div class="relative z-10">
      <slot />
    </div>
  </div>
</template>
