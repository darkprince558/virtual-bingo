'use client'

import { useEffect, useCallback } from 'react'
import confetti from 'canvas-confetti'

interface ConfettiCelebrationProps {
  /** Fire the confetti burst when this becomes true */
  trigger: boolean
  /** Intensity: 'small' for cell marks, 'big' for winners */
  intensity?: 'small' | 'big'
}

const BRAND_COLORS = [
  '#FF5A1F', // brand orange
  '#FF7A42', // brand light orange
  '#FFC5A8', // brand peach
  '#7C5CFC', // lavender
  '#BFAAFF', // lavender light
  '#22AA6A', // mint
  '#FBBF24', // warm yellow
  '#F43F5E', // coral
]

export function ConfettiCelebration({ trigger, intensity = 'big' }: ConfettiCelebrationProps) {
  const fire = useCallback(() => {
    let cancelled = false
    let rafId: number | null = null
    const cancel = () => {
      cancelled = true
      if (rafId) cancelAnimationFrame(rafId)
    }

    if (intensity === 'big') {
      // Multi-burst celebration
      const duration = 2500
      const end = Date.now() + duration

      const frame = () => {
        confetti({
          particleCount: 3,
          angle: 60,
          spread: 55,
          origin: { x: 0, y: 0.65 },
          colors: BRAND_COLORS,
          shapes: ['circle', 'square'],
          scalar: 1.1,
          drift: 0.5,
          ticks: 120,
        })
        confetti({
          particleCount: 3,
          angle: 120,
          spread: 55,
          origin: { x: 1, y: 0.65 },
          colors: BRAND_COLORS,
          shapes: ['circle', 'square'],
          scalar: 1.1,
          drift: -0.5,
          ticks: 120,
        })

        if (Date.now() < end && !cancelled) {
          rafId = requestAnimationFrame(frame)
        }
      }

      // Initial burst from center
      confetti({
        particleCount: 80,
        spread: 100,
        origin: { x: 0.5, y: 0.4 },
        colors: BRAND_COLORS,
        shapes: ['circle', 'square'],
        scalar: 1.3,
        ticks: 150,
        gravity: 0.8,
      })

      // Side bursts
      rafId = requestAnimationFrame(frame)
    } else {
      // Small celebration burst
      confetti({
        particleCount: 30,
        spread: 60,
        origin: { x: 0.5, y: 0.5 },
        colors: BRAND_COLORS.slice(0, 4),
        shapes: ['circle'],
        scalar: 0.8,
        ticks: 80,
        gravity: 1.2,
      })
    }
    return cancel
  }, [intensity])

  useEffect(() => {
    if (trigger) {
      return fire()
    }
  }, [trigger, fire])

  return null // canvas-confetti creates its own canvas
}
