import { GameState } from '@/types/game'

interface GameStatusBadgeProps {
  status: GameState
  size?: 'sm' | 'md'
}

const STATUS_CONFIG: Record<GameState, { bg: string; text: string; dot: string; label: string; pulse: boolean }> = {
  'Waiting': {
    bg: '#FEF3C7', text: '#B45309', dot: '#F59E0B', label: 'Waiting', pulse: false,
  },
  'Starting Soon': {
    bg: '#FEF3C7', text: '#D97706', dot: '#F59E0B', label: 'Starting Soon', pulse: true,
  },
  'Live': {
    bg: '#EDFAF5', text: '#116B3F', dot: '#22AA6A', label: 'Live', pulse: true,
  },
  'Paused': {
    bg: '#FEF3C7', text: '#D97706', dot: '#F59E0B', label: 'Paused', pulse: false,
  },
  'Finished': {
    bg: '#F4F2EF', text: '#78716C', dot: '#A8A29E', label: 'Finished', pulse: false,
  },
}

export function GameStatusBadge({ status, size = 'md' }: GameStatusBadgeProps) {
  const config = STATUS_CONFIG[status]
  const padding = size === 'sm' ? 'px-2.5 py-1' : 'px-4 py-1.5'
  const textSize = size === 'sm' ? 'text-[10px]' : 'text-xs'

  return (
    <span
      className={`inline-flex items-center gap-1.5 ${padding} ${textSize} rounded-full font-extrabold uppercase tracking-wider`}
      style={{ background: config.bg, color: config.text }}
      role="status"
      aria-label={`Game status: ${config.label}`}
    >
      <span
        className={`w-1.5 h-1.5 rounded-full shrink-0 ${config.pulse ? 'animate-pulse' : ''}`}
        style={{ background: config.dot }}
      />
      {config.label}
    </span>
  )
}
