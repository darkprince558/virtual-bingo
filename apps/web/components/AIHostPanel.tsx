'use client'

import { motion } from 'motion/react'
import { Sparkles } from 'lucide-react'

interface AIHostPanelProps {
  message?: string
}

function TypingIndicator() {
  return (
    <div className="flex items-center gap-1.5 px-3 py-2">
      <div className="w-2 h-2 rounded-full animate-typing-dot-1" style={{ background: '#9E80FF' }} />
      <div className="w-2 h-2 rounded-full animate-typing-dot-2" style={{ background: '#9E80FF' }} />
      <div className="w-2 h-2 rounded-full animate-typing-dot-3" style={{ background: '#9E80FF' }} />
    </div>
  )
}

/** AI Host SVG avatar - friendly abstract face, Headspace-style */
function AIAvatar() {
  return (
    <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
      <defs>
        <linearGradient id="aiFaceGrad" x1="4" y1="4" x2="24" y2="24">
          <stop offset="0%" stopColor="#9E80FF" />
          <stop offset="100%" stopColor="#7C5CFC" />
        </linearGradient>
      </defs>
      <circle cx="14" cy="14" r="13" fill="url(#aiFaceGrad)" />
      {/* Eyes */}
      <circle cx="10" cy="12" r="1.8" fill="white" />
      <circle cx="18" cy="12" r="1.8" fill="white" />
      {/* Smile */}
      <path d="M10 17 Q14 21 18 17" stroke="white" strokeWidth="1.8" fill="none" strokeLinecap="round" />
      {/* Sparkle */}
      <path d="M22 4L23 6L22 8L21 6L22 4Z" fill="#FBBF24" opacity="0.8" />
    </svg>
  )
}

export function AIHostPanel({ message }: AIHostPanelProps) {
  if (!message) return null

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ type: 'spring', stiffness: 200, damping: 20 }}
      className="mx-4 mb-4 rounded-xl p-4"
      style={{
        background: 'linear-gradient(135deg, #F5F2FF 0%, #EDE5FF 100%)',
        border: '1.5px solid #D9CCFF',
      }}
    >
      <div className="flex items-center gap-2.5 mb-3">
        {/* Animated AI avatar */}
        <motion.div
          animate={{ y: [0, -2, 0] }}
          transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
        >
          <AIAvatar />
        </motion.div>
        <div>
          <span className="text-xs font-extrabold uppercase tracking-widest block" style={{ color: '#7C5CFC' }}>
            AI Host
          </span>
          <span className="text-[9px] font-semibold" style={{ color: '#BFAAFF' }}>
            Game commentary
          </span>
        </div>
        {/* Sparkle icon */}
        <motion.div
          className="ml-auto"
          animate={{ rotate: [0, 15, -15, 0] }}
          transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut' }}
        >
          <Sparkles className="w-4 h-4" style={{ color: '#BFAAFF' }} />
        </motion.div>
      </div>

      {/* Chat bubble */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ delay: 0.2 }}
        className="rounded-lg px-4 py-3 relative"
        style={{ background: 'rgba(255,255,255,0.65)', border: '1px solid rgba(217,204,255,0.5)' }}
      >
        {/* Chat bubble tail */}
        <div
          className="absolute top-[-6px] left-6 w-3 h-3 rotate-45"
          style={{ background: 'rgba(255,255,255,0.65)', borderLeft: '1px solid rgba(217,204,255,0.5)', borderTop: '1px solid rgba(217,204,255,0.5)' }}
        />
        <p className="text-sm font-semibold leading-snug relative z-10" style={{ color: '#4F30C2' }}>
          {message}
        </p>
      </motion.div>
    </motion.div>
  )
}
