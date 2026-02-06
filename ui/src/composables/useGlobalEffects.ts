const snowflakeChars = ['❄', '❅', '❆', '•']
const fireworkColors = ['#FACC15', '#EF4444', '#22C55E', '#7C3AED', '#3B82F6', '#EC4899', '#F97316']

interface Particle {
  element: HTMLDivElement
  x: number
  y: number
  vx: number
  vy: number
  life: number
}

let container: HTMLDivElement | null = null

function ensureContainer(): HTMLDivElement {
  if (container) return container
  container = document.createElement('div')
  container.style.cssText = `
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 9998;
    overflow: hidden;
  `
  document.body.appendChild(container)
  return container
}

function cleanupContainer() {
  if (container && container.childElementCount === 0) {
    container.remove()
    container = null
  }
}

function addBanner(c: HTMLDivElement, text: string, color: string) {
  const banner = document.createElement('div')
  banner.textContent = text
  banner.style.cssText = `
    position: absolute;
    top: 60px;
    left: 50%;
    transform: translateX(-50%);
    color: ${color};
    font-size: 14px;
    font-weight: 600;
    text-shadow: 0 1px 4px rgba(0,0,0,0.5);
    white-space: nowrap;
    opacity: 0;
    animation: globalBannerFade 4s ease-in-out forwards;
  `
  c.appendChild(banner)
  setTimeout(() => banner.remove(), 4000)
}

function triggerSnowflakes(actorName?: string) {
  const c = ensureContainer()
  const particles: Particle[] = []
  const count = 60
  let stopped = false

  if (actorName) {
    addBanner(c, `${actorName} made it snow!`, '#93C5FD')
  }

  for (let i = 0; i < count; i++) {
    setTimeout(() => {
      if (stopped) return
      const el = document.createElement('div')
      const char = snowflakeChars[Math.floor(Math.random() * snowflakeChars.length)]
      const size = Math.random() * 14 + 10
      const startX = Math.random() * window.innerWidth
      el.textContent = char !== undefined ? char : null
      el.style.cssText = `
        position: absolute;
        top: 0;
        left: 0;
        font-size: ${size}px;
        color: rgba(200, 220, 255, ${0.5 + Math.random() * 0.5});
        pointer-events: none;
        user-select: none;
        will-change: transform;
        transform: translate(${startX}px, -20px);
      `
      c.appendChild(el)

      particles.push({
        element: el,
        x: startX,
        y: -20,
        vx: (Math.random() - 0.5) * 1.5,
        vy: Math.random() * 1.5 + 0.5,
        life: 0,
      })
    }, Math.random() * 3000)
  }

  function animate() {
    if (stopped) return
    for (const p of particles) {
      p.x += p.vx + Math.sin(p.life * 0.02) * 0.5
      p.y += p.vy
      p.life++
      p.element.style.transform = `translate(${p.x}px, ${p.y}px)`

      if (p.y > window.innerHeight + 30) {
        p.element.remove()
      }
    }
    requestAnimationFrame(animate)
  }
  requestAnimationFrame(animate)

  setTimeout(() => {
    stopped = true
    for (const p of particles) p.element.remove()
    particles.length = 0
    cleanupContainer()
  }, 8000)
}

function triggerFireworks(actorName?: string) {
  const c = ensureContainer()
  let stopped = false
  const allParticles: Particle[] = []

  if (actorName) {
    addBanner(c, `${actorName} launched fireworks!`, '#FACC15')
  }

  const burstCount = 5

  for (let b = 0; b < burstCount; b++) {
    setTimeout(() => {
      if (stopped) return
      const cx = Math.random() * window.innerWidth * 0.6 + window.innerWidth * 0.2
      const cy = Math.random() * window.innerHeight * 0.4 + window.innerHeight * 0.1
      const color = fireworkColors[Math.floor(Math.random() * fireworkColors.length)]
      const particleCount = 30

      for (let i = 0; i < particleCount; i++) {
        const angle = (Math.PI * 2 * i) / particleCount + (Math.random() - 0.5) * 0.3
        const speed = Math.random() * 4 + 2
        const el = document.createElement('div')
        const size = Math.random() * 4 + 2
        el.style.cssText = `
          position: absolute;
          top: 0;
          left: 0;
          width: ${size}px;
          height: ${size}px;
          background: ${color};
          border-radius: 50%;
          pointer-events: none;
          box-shadow: 0 0 6px ${color};
          will-change: transform, opacity;
          transform: translate(${cx}px, ${cy}px);
        `
        c.appendChild(el)

        allParticles.push({
          element: el,
          x: cx,
          y: cy,
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          life: 0,
        })
      }
    }, b * 600)
  }

  function animate() {
    if (stopped) return
    for (const p of allParticles) {
      p.vy += 0.05
      p.vx *= 0.98
      p.x += p.vx
      p.y += p.vy
      p.life++

      const opacity = Math.max(0, 1 - p.life / 80)
      p.element.style.transform = `translate(${p.x}px, ${p.y}px)`
      p.element.style.opacity = String(opacity)

      if (opacity <= 0) {
        p.element.remove()
      }
    }
    requestAnimationFrame(animate)
  }
  requestAnimationFrame(animate)

  setTimeout(() => {
    stopped = true
    for (const p of allParticles) p.element.remove()
    allParticles.length = 0
    cleanupContainer()
  }, 6000)
}

export function useGlobalEffects() {
  function trigger(actionType: string, actorName?: string) {
    switch (actionType) {
      case 'snowflakes':
        triggerSnowflakes(actorName)
        break
      case 'fireworks':
        triggerFireworks(actorName)
        break
    }
  }

  return { trigger, triggerSnowflakes, triggerFireworks }
}
