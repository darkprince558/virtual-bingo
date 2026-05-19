'use client'

import { GameState } from '@/types/game'
import { Play, SkipForward, Pause, Square, ChevronRight } from 'lucide-react'
import { motion } from 'motion/react'

interface HostControlsProps {
  status: GameState
  onStart: () => void
  onNextWord: () => void
  onPause: () => void
  onEnd: () => void
}

interface ControlButton {
  label: string
  action: () => void
  icon: React.ReactNode
  variant: 'primary' | 'secondary' | 'danger' | 'warning'
  disabled?: boolean
}

export function HostControls({ status, onStart, onNextWord, onPause, onEnd }: HostControlsProps) {
  const isLive = status === 'Live'
  const isPaused = status === 'Paused'
  const isFinished = status === 'Finished'
  const isWaiting = status === 'Waiting' || status === 'Starting Soon'

  const BUTTON_STYLES = {
    primary: {
      background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
      color: '#FFFFFF',
      border: 'none',
      boxShadow: '0 4px 16px rgba(255, 90, 31, 0.35)',
    },
    secondary: {
      background: '#FFFFFF',
      color: '#44403C',
      border: '2px solid #E7E5E4',
      boxShadow: '0 2px 8px rgba(0,0,0,0.04)',
    },
    warning: {
      background: '#FEF3C7',
      color: '#D97706',
      border: '2px solid #FCD34D',
      boxShadow: '0 2px 8px rgba(245, 158, 11, 0.12)',
    },
    danger: {
      background: '#FFF1F2',
      color: '#E11D48',
      border: '2px solid #FECDD3',
      boxShadow: '0 2px 8px rgba(244, 63, 94, 0.10)',
    },
  }

  const controls: ControlButton[] = [
    {
      label: isWaiting ? 'Start Game' : isPaused ? 'Resume Game' : 'Game Running',
      action: onStart,
      icon: <Play className="w-4 h-4" />,
      variant: 'primary',
      disabled: isLive || isFinished,
    },
    {
      label: 'Call Next Word',
      action: onNextWord,
      icon: <SkipForward className="w-4 h-4" />,
      variant: 'secondary',
      disabled: !isLive,
    },
    {
      label: isPaused ? 'Paused' : 'Pause Game',
      action: onPause,
      icon: <Pause className="w-4 h-4" />,
      variant: 'warning',
      disabled: !isLive,
    },
    {
      label: 'End Game',
      action: onEnd,
      icon: <Square className="w-4 h-4" />,
      variant: 'danger',
      disabled: isFinished,
    },
  ]

  return (
    <div
      className="rounded-3xl p-5"
      style={{
        background: '#FFFFFF',
        border: '1.5px solid #F0EDE8',
        boxShadow: '0 2px 16px rgba(0,0,0,0.04)',
      }}
    >
      <div className="flex items-center gap-2 mb-4">
        <ChevronRight className="w-4 h-4" style={{ color: '#FF5A1F' }} />
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
          Game Controls
        </span>
      </div>

      <div className="grid grid-cols-2 gap-3">
        {controls.map((btn) => {
          const s = BUTTON_STYLES[btn.variant]
          return (
            <motion.button
              key={btn.label}
              whileTap={!btn.disabled ? { scale: 0.96 } : {}}
              whileHover={!btn.disabled ? { y: -1 } : {}}
              onClick={btn.action}
              disabled={btn.disabled}
              className="flex items-center justify-center gap-2 px-4 py-3 rounded-2xl font-bold text-sm transition-all"
              style={{
                ...s,
                opacity: btn.disabled ? 0.45 : 1,
                cursor: btn.disabled ? 'not-allowed' : 'pointer',
              }}
            >
              {btn.icon}
              <span>{btn.label}</span>
            </motion.button>
          )
        })}
      </div>
    </div>
  )
}
