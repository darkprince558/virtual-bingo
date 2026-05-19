'use client'

import { motion, AnimatePresence } from 'motion/react'
import { Sparkles } from 'lucide-react'

interface CurrentCallDisplayProps {
  word?: string
  aiMessage?: string
}

export function CurrentCallDisplay({ word, aiMessage }: CurrentCallDisplayProps) {
  return (
    <div
      className="w-full rounded-3xl overflow-hidden"
      style={{
        background: 'linear-gradient(135deg, #FFF4F0 0%, #FFFFFF 60%, #F5F2FF 100%)',
        border: '2px solid #FFE4D9',
        boxShadow: '0 4px 24px rgba(255, 90, 31, 0.10)',
      }}
    >
      {/* Header label */}
      <div className="px-5 pt-5 pb-0 flex items-center gap-2">
        <div
          className="w-2 h-2 rounded-full animate-pulse"
          style={{ background: '#FF5A1F' }}
        />
        <span
          className="text-xs font-extrabold uppercase tracking-widest"
          style={{ color: '#FF5A1F', letterSpacing: '0.18em' }}
        >
          Current Word
        </span>
      </div>

      {/* Word display */}
      <div className="px-5 py-4 flex items-center justify-center min-h-[100px]">
        <AnimatePresence mode="wait">
          {word ? (
            <motion.div
              key={word}
              initial={{ opacity: 0, scale: 0.75, y: 12 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.85, y: -8 }}
              transition={{ type: 'spring', stiffness: 300, damping: 22 }}
              className="text-center"
            >
              <h2
                className="font-black leading-tight break-words"
                style={{
                  color: '#1C1917',
                  fontSize: 'clamp(1.75rem, 4vw, 3rem)',
                  textShadow: '0 2px 12px rgba(255, 90, 31, 0.08)',
                }}
              >
                {word}
              </h2>
            </motion.div>
          ) : (
            <motion.p
              key="empty"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-lg font-semibold"
              style={{ color: '#D6D3D1' }}
            >
              Waiting for first call…
            </motion.p>
          )}
        </AnimatePresence>
      </div>

      {/* AI message */}
      {aiMessage && (
        <div
          className="mx-4 mb-4 rounded-2xl px-4 py-3 flex items-start gap-3"
          style={{ background: 'linear-gradient(135deg, #F5F2FF, #EDE5FF)', border: '1px solid #D9CCFF' }}
        >
          <Sparkles className="w-4 h-4 shrink-0 mt-0.5" style={{ color: '#7C5CFC' }} />
          <p className="text-sm font-semibold leading-snug" style={{ color: '#4F30C2' }}>
            {aiMessage}
          </p>
        </div>
      )}
    </div>
  )
}
