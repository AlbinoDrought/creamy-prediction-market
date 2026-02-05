import { ref } from 'vue'

const isActive = ref(false)

const colors = ['#FACC15', '#22C55E', '#7C3AED', '#EF4444', '#3B82F6', '#EC4899']

interface Particle {
  element: HTMLDivElement
  x: number
  y: number
  vx: number
  vy: number
  rotation: number
  rotationSpeed: number
  size: number
}

let particles: Particle[] = []
let animationId: number | null = null
let container: HTMLDivElement | null = null

function createParticle(): Particle {
  const element = document.createElement('div')
  const size = Math.random() * 10 + 5
  const color = colors[Math.floor(Math.random() * colors.length)]

  element.style.cssText = `
    position: fixed;
    width: ${size}px;
    height: ${size}px;
    background: ${color};
    pointer-events: none;
    z-index: 9999;
    border-radius: ${Math.random() > 0.5 ? '50%' : '0'};
  `

  container?.appendChild(element)

  return {
    element,
    x: Math.random() * window.innerWidth,
    y: -20,
    vx: (Math.random() - 0.5) * 6,
    vy: Math.random() * 2 + 1,
    rotation: Math.random() * 360,
    rotationSpeed: (Math.random() - 0.5) * 6,
    size,
  }
}

function animate() {
  let allDone = true

  for (const particle of particles) {
    particle.vy += 0.06 // gravity
    particle.vx *= 0.995 // slight air resistance
    particle.x += particle.vx
    particle.y += particle.vy
    particle.rotation += particle.rotationSpeed

    particle.element.style.transform = `translate(${particle.x}px, ${particle.y}px) rotate(${particle.rotation}deg)`

    if (particle.y < window.innerHeight + 50) {
      allDone = false
    }
  }

  if (allDone) {
    cleanup()
  } else {
    animationId = requestAnimationFrame(animate)
  }
}

function cleanup() {
  if (animationId) {
    cancelAnimationFrame(animationId)
    animationId = null
  }

  for (const particle of particles) {
    particle.element.remove()
  }
  particles = []

  container?.remove()
  container = null
  isActive.value = false
}

export function useConfetti() {
  function fire(particleCount = 100) {
    if (isActive.value) return
    isActive.value = true

    // Create container
    container = document.createElement('div')
    container.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      pointer-events: none;
      z-index: 9999;
      overflow: hidden;
    `
    document.body.appendChild(container)

    // Create particles
    for (let i = 0; i < particleCount; i++) {
      particles.push(createParticle())
    }

    // Start animation
    animate()

    // Auto cleanup after 5 seconds
    setTimeout(cleanup, 5000)
  }

  return {
    fire,
    isActive,
  }
}
