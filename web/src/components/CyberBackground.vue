<script setup>
/**
 * Pure CSS high-tech animated background.
 * Features: shifting gradient mesh, animated circuit traces,
 * pulsing node points. All GPU-accelerated, SSR-safe.
 */
defineProps({
  intensity: { type: String, default: 'medium' },
})
</script>

<template>
  <div
    class="cb-root"
    :data-intensity="intensity"
    aria-hidden="true"
  >
    <!-- Base gradient layer -->
    <div class="cb-gradient-layer" />

    <!-- Circuit trace lines -->
    <div class="cb-traces">
      <span class="cb-trace cb-trace-h1" />
      <span class="cb-trace cb-trace-h2" />
      <span class="cb-trace cb-trace-h3" />
      <span class="cb-trace cb-trace-v1" />
      <span class="cb-trace cb-trace-v2" />
      <span class="cb-trace cb-trace-d1" />
      <span class="cb-trace cb-trace-d2" />
    </div>

    <!-- Pulsing nodes at trace intersections -->
    <div class="cb-nodes">
      <span class="cb-node" style="top:15%;left:20%" />
      <span class="cb-node" style="top:35%;left:50%" />
      <span class="cb-node" style="top:55%;left:80%" />
      <span class="cb-node" style="top:75%;left:35%" />
      <span class="cb-node" style="top:25%;left:70%" />
      <span class="cb-node" style="top:65%;left:15%" />
    </div>
  </div>
</template>

<style>
/* ============================================
   CyberBackground — Pure CSS Tech Background
   NOT scoped — must survive SSR hydration
   All selectors prefixed with cb- to avoid conflicts
   ============================================ */

.cb-root {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  z-index: 0;
}

/* ---- Base Gradient Layer ---- */
.cb-gradient-layer {
  position: absolute;
  inset: -50%;
  background:
    radial-gradient(ellipse 30% 40% at 20% 30%, rgba(99, 102, 241, 0.06) 0%, transparent 50%),
    radial-gradient(ellipse 25% 35% at 70% 60%, rgba(6, 182, 212, 0.05) 0%, transparent 50%),
    radial-gradient(ellipse 35% 45% at 50% 20%, rgba(168, 85, 247, 0.04) 0%, transparent 50%);
  animation: cbGradientShift 20s ease-in-out infinite;
  will-change: transform;
}

@keyframes cbGradientShift {
  0%, 100% { transform: translate(0, 0) rotate(0deg); }
  33% { transform: translate(2%, -1%) rotate(1deg); }
  66% { transform: translate(-1%, 2%) rotate(-1deg); }
}

/* ---- Circuit Trace Lines ---- */
.cb-traces {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.cb-trace {
  position: absolute;
  opacity: 0.15;
  will-change: transform, opacity;
}

/* Horizontal traces */
.cb-trace-h1,
.cb-trace-h2,
.cb-trace-h3 {
  left: -10%;
  right: -10%;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(99, 102, 241, 0.5) 15%,
    rgba(99, 102, 241, 0.8) 30%,
    rgba(6, 182, 212, 0.4) 60%,
    transparent 100%
  );
  animation: cbTraceHoriz 8s ease-in-out infinite;
}

.cb-trace-h1 { top: 20%; animation-delay: 0s; }
.cb-trace-h2 { top: 40%; animation-delay: -3s; }
.cb-trace-h3 { top: 70%; animation-delay: -6s; }

@keyframes cbTraceHoriz {
  0%, 100% { transform: translateX(0); opacity: 0.12; }
  50% { transform: translateX(2%); opacity: 0.25; }
}

/* Vertical traces */
.cb-trace-v1,
.cb-trace-v2 {
  top: -10%;
  bottom: -10%;
  width: 1px;
  background: linear-gradient(
    0deg,
    transparent 0%,
    rgba(168, 85, 247, 0.5) 20%,
    rgba(99, 102, 241, 0.7) 50%,
    rgba(6, 182, 212, 0.3) 80%,
    transparent 100%
  );
  animation: cbTraceVert 10s ease-in-out infinite;
}

.cb-trace-v1 { left: 25%; animation-delay: 0s; }
.cb-trace-v2 { left: 65%; animation-delay: -5s; }

@keyframes cbTraceVert {
  0%, 100% { transform: translateY(0); opacity: 0.1; }
  50% { transform: translateY(3%); opacity: 0.22; }
}

/* Diagonal traces */
.cb-trace-d1 {
  left: 0;
  right: 0;
  top: 30%;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(6, 182, 212, 0.3) 20%,
    rgba(99, 102, 241, 0.5) 50%,
    rgba(6, 182, 212, 0.3) 80%,
    transparent 100%
  );
  transform: rotate(-8deg);
  animation: cbTraceDiag 14s ease-in-out infinite;
}

.cb-trace-d2 {
  left: 0;
  right: 0;
  top: 60%;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(168, 85, 247, 0.3) 25%,
    rgba(99, 102, 241, 0.5) 55%,
    rgba(168, 85, 247, 0.3) 85%,
    transparent 100%
  );
  transform: rotate(6deg);
  animation: cbTraceDiag 18s ease-in-out infinite;
  animation-delay: -7s;
}

@keyframes cbTraceDiag {
  0%, 100% { opacity: 0.08; transform: rotate(-8deg) translateY(0); }
  50% { opacity: 0.18; transform: rotate(-8deg) translateY(5%); }
}

.cb-trace-d2 {
  animation-name: cbTraceDiag2;
}

@keyframes cbTraceDiag2 {
  0%, 100% { opacity: 0.08; transform: rotate(6deg) translateY(0); }
  50% { opacity: 0.18; transform: rotate(6deg) translateY(-5%); }
}

/* ---- Pulsing Nodes ---- */
.cb-nodes {
  position: absolute;
  inset: 0;
}

.cb-node {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(99, 102, 241, 0.7);
  box-shadow:
    0 0 6px rgba(99, 102, 241, 0.5),
    0 0 18px rgba(99, 102, 241, 0.2);
  animation: cbNodePulse 3s ease-in-out infinite;
  opacity: 0.5;
}

.cb-node:nth-child(2) { animation-delay: -0.8s; }
.cb-node:nth-child(3) { animation-delay: -1.6s; }
.cb-node:nth-child(4) { animation-delay: -2.4s; }
.cb-node:nth-child(5) { animation-delay: -1.0s; }
.cb-node:nth-child(6) { animation-delay: -2.0s; }

@keyframes cbNodePulse {
  0%, 100% { transform: scale(1); opacity: 0.35; }
  50% { transform: scale(1.8); opacity: 0.8; }
}

/* ---- Light theme ---- */
[data-theme*="light"] .cb-trace,
:root[data-theme*="light"] .cb-trace {
  opacity: 0.08;
}

[data-theme*="light"] .cb-node,
:root[data-theme*="light"] .cb-node {
  background: rgba(79, 70, 229, 0.5);
  box-shadow:
    0 0 4px rgba(79, 70, 229, 0.3),
    0 0 12px rgba(79, 70, 229, 0.1);
}

/* ---- Mobile ---- */
@media (max-width: 768px) {
  .cb-gradient-layer {
    inset: -30%;
  }

  .cb-trace {
    opacity: 0.08;
  }

  .cb-node {
    width: 4px;
    height: 4px;
  }

  @keyframes cbNodePulse {
    0%, 100% { transform: scale(1); opacity: 0.25; }
    50% { transform: scale(1.5); opacity: 0.5; }
  }
}

/* ---- Reduced motion ---- */
@media (prefers-reduced-motion: reduce) {
  .cb-gradient-layer,
  .cb-trace,
  .cb-node {
    animation: none;
  }
  .cb-node { opacity: 0.4; }
}

/* ---- Intensity ---- */
[data-intensity="low"] .cb-gradient-layer { animation-duration: 30s; }
[data-intensity="low"] .cb-trace-h1,
[data-intensity="low"] .cb-trace-h2,
[data-intensity="low"] .cb-trace-h3 { animation-duration: 14s; }
[data-intensity="low"] .cb-trace-v1,
[data-intensity="low"] .cb-trace-v2 { animation-duration: 16s; }
[data-intensity="low"] .cb-node { animation-duration: 5s; }

[data-intensity="high"] .cb-gradient-layer { animation-duration: 10s; }
[data-intensity="high"] .cb-trace-h1,
[data-intensity="high"] .cb-trace-h2,
[data-intensity="high"] .cb-trace-h3 { animation-duration: 4s; }
[data-intensity="high"] .cb-trace-v1,
[data-intensity="high"] .cb-trace-v2 { animation-duration: 5s; }
[data-intensity="high"] .cb-node { animation-duration: 1.5s; }
</style>
