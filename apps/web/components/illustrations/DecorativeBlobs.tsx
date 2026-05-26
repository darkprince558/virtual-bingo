'use client'

import { motion } from 'motion/react'

interface DecorativeBlobsProps {
  variant?: 'landing' | 'lobby' | 'play' | 'celebration'
}

/** Headspace-style organic blob shapes - warm, morphing, ambient */
export function DecorativeBlobs({ variant = 'landing' }: DecorativeBlobsProps) {
  const configs = {
    landing: [
      { cx: '80%', cy: '15%', r: 260, color: '#FF7A42', opacity: 0.12, dur: 18, dx: 30, dy: 20 },
      { cx: '10%', cy: '70%', r: 200, color: '#7C5CFC', opacity: 0.09, dur: 22, dx: -20, dy: 25 },
      { cx: '50%', cy: '45%', r: 150, color: '#22AA6A', opacity: 0.06, dur: 25, dx: 15, dy: -15 },
      { cx: '30%', cy: '20%', r: 120, color: '#FBBF24', opacity: 0.07, dur: 20, dx: -10, dy: 20 },
    ],
    lobby: [
      { cx: '70%', cy: '30%', r: 220, color: '#FF7A42', opacity: 0.10, dur: 20, dx: 25, dy: 15 },
      { cx: '20%', cy: '60%', r: 180, color: '#7C5CFC', opacity: 0.08, dur: 24, dx: -15, dy: 20 },
      { cx: '85%', cy: '75%', r: 140, color: '#22AA6A', opacity: 0.07, dur: 18, dx: 20, dy: -10 },
    ],
    play: [
      { cx: '90%', cy: '10%', r: 200, color: '#FF7A42', opacity: 0.08, dur: 22, dx: 20, dy: 15 },
      { cx: '5%', cy: '85%', r: 160, color: '#7C5CFC', opacity: 0.06, dur: 26, dx: -15, dy: 10 },
    ],
    celebration: [
      { cx: '50%', cy: '30%', r: 300, color: '#FBBF24', opacity: 0.15, dur: 15, dx: 30, dy: 20 },
      { cx: '20%', cy: '50%', r: 250, color: '#FF7A42', opacity: 0.12, dur: 18, dx: -25, dy: 25 },
      { cx: '80%', cy: '60%', r: 220, color: '#7C5CFC', opacity: 0.10, dur: 20, dx: 20, dy: -15 },
      { cx: '60%', cy: '80%', r: 180, color: '#22AA6A', opacity: 0.08, dur: 22, dx: -10, dy: 20 },
    ],
  }

  const blobs = configs[variant]

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none" aria-hidden="true">
      {blobs.map((blob, i) => (
        <motion.div
          key={i}
          className="absolute rounded-full"
          initial={{
            left: blob.cx,
            top: blob.cy,
            width: blob.r * 2,
            height: blob.r * 2,
            x: '-50%',
            y: '-50%',
          }}
          animate={{
            x: ['-50%', `calc(-50% + ${blob.dx}px)`, '-50%'],
            y: ['-50%', `calc(-50% + ${blob.dy}px)`, '-50%'],
            scale: [1, 1.08, 0.95, 1],
          }}
          transition={{
            duration: blob.dur,
            repeat: Infinity,
            ease: 'easeInOut',
          }}
          style={{
            background: `radial-gradient(circle, ${blob.color} 0%, transparent 70%)`,
            opacity: blob.opacity,
            filter: 'blur(60px)',
          }}
        />
      ))}
    </div>
  )
}
