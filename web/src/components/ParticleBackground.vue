<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  particleCount: { type: Number, default: 120 },
  particleColor: { type: String, default: '99, 102, 241' },
  connectDistance: { type: Number, default: 150 },
  speed: { type: Number, default: 0.8 },
})

const canvas = ref(null)
let animationId = null
let particles = []
let ctx = null
let resizeObserver = null
let mouse = { x: -1000, y: -1000 }
const mouseInfluence = 120

class Particle {
  constructor(w, h) {
    this.reset(w, h, true)
  }

  reset(w, h, initial = false) {
    this.x = initial ? Math.random() * w : Math.random() * w
    this.y = initial ? Math.random() * h : Math.random() * h
    const angle = Math.random() * Math.PI * 2
    const spd = props.speed * (0.3 + Math.random() * 0.7)
    this.vx = Math.cos(angle) * spd
    this.vy = Math.sin(angle) * spd
    this.radius = Math.random() * 1.5 + 0.5
    this.opacity = Math.random() * 0.4 + 0.2
    this.pulse = Math.random() * Math.PI * 2
    this.pulseSpeed = 0.02 + Math.random() * 0.03
  }

  update(w, h) {
    // Gentle mouse repulsion
    const dx = mouse.x - this.x
    const dy = mouse.y - this.y
    const distToMouse = Math.sqrt(dx * dx + dy * dy)

    if (distToMouse < mouseInfluence) {
      const force = (1 - distToMouse / mouseInfluence) * 0.3
      this.vx -= (dx / distToMouse) * force * 0.02
      this.vy -= (dy / distToMouse) * force * 0.02
    }

    // Damping
    this.vx *= 0.998
    this.vy *= 0.998

    this.x += this.vx
    this.y += this.vy

    // Wrap around edges
    if (this.x < -20) this.x = w + 20
    if (this.x > w + 20) this.x = -20
    if (this.y < -20) this.y = h + 20
    if (this.y > h + 20) this.y = -20

    this.pulse += this.pulseSpeed
  }

  draw(ctx, color) {
    const pulseAlpha = Math.sin(this.pulse) * 0.15 + 0.35
    const alpha = this.opacity * (0.6 + pulseAlpha)

    // Outer glow
    const gradient = ctx.createRadialGradient(this.x, this.y, 0, this.x, this.y, this.radius * 4)
    gradient.addColorStop(0, `rgba(${color}, ${alpha * 0.6})`)
    gradient.addColorStop(0.5, `rgba(${color}, ${alpha * 0.2})`)
    gradient.addColorStop(1, `rgba(${color}, 0)`)
    ctx.fillStyle = gradient
    ctx.beginPath()
    ctx.arc(this.x, this.y, this.radius * 4, 0, Math.PI * 2)
    ctx.fill()

    // Core dot
    ctx.beginPath()
    ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2)
    ctx.fillStyle = `rgba(${color}, ${alpha})`
    ctx.fill()
  }
}

function initCanvas() {
  if (!canvas.value) return
  const parent = canvas.value.parentElement
  const w = parent.clientWidth
  const h = parent.clientHeight
  canvas.value.width = w
  canvas.value.height = h
  ctx = canvas.value.getContext('2d')

  particles = Array.from({ length: props.particleCount }, () => new Particle(w, h))
}

function animate() {
  if (!canvas.value || !ctx) return
  const w = canvas.value.width
  const h = canvas.value.height

  ctx.clearRect(0, 0, w, h)

  const color = props.particleColor

  // Draw connections
  for (let i = 0; i < particles.length; i++) {
    for (let j = i + 1; j < particles.length; j++) {
      const dx = particles[i].x - particles[j].x
      const dy = particles[i].y - particles[j].y
      const dist = Math.sqrt(dx * dx + dy * dy)

      if (dist < props.connectDistance) {
        const opacity = (1 - dist / props.connectDistance) * 1
        ctx.beginPath()
        ctx.moveTo(particles[i].x, particles[i].y)
        ctx.lineTo(particles[j].x, particles[j].y)
        ctx.strokeStyle = `rgba(${color}, ${opacity})`
        ctx.lineWidth = 0.3
        ctx.stroke()
      }
    }
  }

  // Draw particles on top
  for (const particle of particles) {
    particle.update(w, h)
    particle.draw(ctx, color)
  }

  animationId = requestAnimationFrame(animate)
}

function handleMouseMove(e) {
  const rect = canvas.value?.getBoundingClientRect()
  if (rect) {
    mouse.x = e.clientX - rect.left
    mouse.y = e.clientY - rect.top
  }
}

function handleMouseLeave() {
  mouse.x = -1000
  mouse.y = -1000
}

onMounted(() => {
  initCanvas()
  animate()

  resizeObserver = new ResizeObserver(() => {
    initCanvas()
  })
  if (canvas.value?.parentElement) {
    resizeObserver.observe(canvas.value.parentElement)
    canvas.value.parentElement.addEventListener('mousemove', handleMouseMove)
    canvas.value.parentElement.addEventListener('mouseleave', handleMouseLeave)
  }
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (resizeObserver) resizeObserver.disconnect()
  if (canvas.value?.parentElement) {
    canvas.value.parentElement.removeEventListener('mousemove', handleMouseMove)
    canvas.value.parentElement.removeEventListener('mouseleave', handleMouseLeave)
  }
})
</script>

<template>
  <canvas ref="canvas" class="absolute inset-0 pointer-events-none"></canvas>
</template>
