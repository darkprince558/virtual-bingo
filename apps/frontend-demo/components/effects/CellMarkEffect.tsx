'use client'

import { motion, AnimatePresence } from 'motion/react'
import { useMemo } from 'react'

interface CellMarkEffectProps {
  trigger: boolean
  color?: string
}

const PARTICLE_COUNT = 8
const PARTICLE_COLORS = ['#FF5A1F', '#FF7A42', '#FFC5A8', '#FBBF24', '#7C5CFC', '#22AA6A']

function getParticleProps(index: number) {
  const angle = (index / PARTICLE_COUNT) * Math.PI * 2
  const distance = 28 + (index * 7) % 16
  return {
    x: Math.cos(angle) * distance,
    y: Math.sin(angle) * distance,
    color: PARTICLE_COLORS[index % PARTICLE_COLORS.length],
    size: 4 + (index * 3) % 4,
    delay: index * 0.02,
  }
}

const PARTICLES = Array.from({ length: PARTICLE_COUNT }, (_, i) => getParticleProps(i))

export function CellMarkEffect({ trigger, color }: CellMarkEffectProps) {
  return (
    <AnimatePresence>
      {trigger && (
        <div
          className="absolute inset-0 pointer-events-none z-20 flex items-center justify-center"
        >
          {/* Ripple ring */}
          <motion.div
            initial={{ scale: 0.3, opacity: 0.8 }}
            animate={{ scale: 2.2, opacity: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            className="absolute rounded-full"
            style={{
              width: 40,
              height: 40,
              border: `3px solid ${color || '#FF5A1F'}`,
            }}
          />

          {/* Particles */}
          {PARTICLES.map((p, i) => (
            <motion.div
              key={i}
              initial={{ x: 0, y: 0, scale: 1, opacity: 1 }}
              animate={{ x: p.x, y: p.y, scale: 0, opacity: 0 }}
              exit={{ opacity: 0 }}
              transition={{
                duration: 0.45,
                delay: p.delay,
                ease: [0.25, 0.46, 0.45, 0.94],
              }}
              className="absolute rounded-full"
              style={{
                width: p.size,
                height: p.size,
                background: p.color,
              }}
            />
          ))}
        </div>
      )}
    </AnimatePresence>
  )
}
