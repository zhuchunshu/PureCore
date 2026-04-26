<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  particleCount: { type: Number, default: 60 },
  particleColor: { type: String, default: '99, 102, 241' },
  speed: { type: Number, default: 1.2 },
})

const canvas = ref(null)
let animationId = null
let ctx = null
let resizeObserver = null
let stars = []

class Star {
  constructor(w, h) {
    this.reset(w, h, true)
  }

  reset(w, h, initial = false) {
    this.x = Math.random() * w
    this.y = Math.random() * h * (initial ? 1 : 0.3)
    this.len = Math.random() * 80 + 40
    this.vx = (Math.random() - 0.5) * 0.3
    this.vy = Math.random() * props.speed + 0.5
    this.opacity = Math.random() * 0.6 + 0.2
    this.thickness = Math.random() * 1.5 + 0.3
  }

  update(w, h) {
    this.x += this.vx
    this.y += this.vy

    if (this.y > h + this.len || this.x < -this.len || this.x > w + this.len) {
      this.reset(w, h)
    }
  }

  draw(ctx, color) {
    const gradient = ctx.createLinearGradient(
      this.x, this.y,
      this.x - this.vx * 10, this.y - this.vy * 10
    )
    gradient.addColorStop(0, `rgba(${color}, ${this.opacity})`)
    gradient.addColorStop(1, `rgba(${color}, 0)`)
    ctx.beginPath()
    ctx.moveTo(this.x, this.y)
    ctx.lineTo(this.x - this.vx * 20, this.y - this.vy * 20)
    ctx.strokeStyle = gradient
    ctx.lineWidth = this.thickness
    ctx.lineCap = 'round'
    ctx.stroke()
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

  stars = Array.from({ length: props.particleCount }, () => new Star(w, h))
}

function animate() {
  if (!canvas.value || !ctx) return
  const w = canvas.value.width
  const h = canvas.value.height

  ctx.fillStyle = 'transparent'
  ctx.clearRect(0, 0, w, h)

  const color = props.particleColor

  for (const star of stars) {
    star.update(w, h)
    star.draw(ctx, color)
  }

  animationId = requestAnimationFrame(animate)
}

onMounted(() => {
  initCanvas()
  animate()

  resizeObserver = new ResizeObserver(() => {
    initCanvas()
  })
  if (canvas.value?.parentElement) {
    resizeObserver.observe(canvas.value.parentElement)
  }
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (resizeObserver) resizeObserver.disconnect()
})
</script>

<template>
  <canvas ref="canvas" class="absolute inset-0 pointer-events-none"></canvas>
</template>
