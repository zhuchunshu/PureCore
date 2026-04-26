<script setup>
/**
 * Animated tech grid overlay.
 * Creates a subtle perspective grid that animates vertically,
 * giving a "cyberpunk/data center" feel.
 */
defineProps({
  opacity: { type: Number, default: 0.05 },
  speed: { type: Number, default: 20 },
  lineColor: { type: String, default: 'var(--fallback-bc,oklch(var(--bc)/0.1))' },
})
</script>

<template>
  <div class="absolute inset-0 overflow-hidden pointer-events-none" :style="{ opacity }">
    <div class="absolute inset-0 animate-grid-scroll" :style="{
      backgroundImage: `
        linear-gradient(0deg, transparent 24%, ${lineColor} 25%, ${lineColor} 26%, transparent 27%, transparent 74%, ${lineColor} 75%, ${lineColor} 76%, transparent 77%, transparent),
        linear-gradient(90deg, transparent 24%, ${lineColor} 25%, ${lineColor} 26%, transparent 27%, transparent 74%, ${lineColor} 75%, ${lineColor} 76%, transparent 77%, transparent)
      `,
      backgroundSize: '80px 80px',
    }">
    </div>
  </div>
</template>

<style scoped>
@keyframes grid-scroll {
  0% { transform: translateY(0); }
  100% { transform: translateY(80px); }
}

.animate-grid-scroll {
  animation: grid-scroll 20s linear infinite;
}
</style>
