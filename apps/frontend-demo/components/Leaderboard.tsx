import { LeaderboardEntry } from '@/types/player'
import { Trophy } from 'lucide-react'

interface LeaderboardProps {
  entries: LeaderboardEntry[]
}

const PLACE_STYLES: Record<number, { bg: string; text: string; border: string; emoji: string }> = {
  1: { bg: '#FEF3C7', text: '#B45309', border: '#FCD34D', emoji: '🥇' },
  2: { bg: '#F4F2EF', text: '#57534E', border: '#D6D3D1', emoji: '🥈' },
  3: { bg: '#FFF4F0', text: '#C23208', border: '#FFC5A8', emoji: '🥉' },
}

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

export function Leaderboard({ entries }: LeaderboardProps) {
  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-2 mb-4">
        <Trophy className="w-4 h-4" style={{ color: '#F59E0B' }} />
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
          Leaderboard
        </span>
      </div>

      <div className="space-y-2.5">
        {entries.map((entry) => {
          const style = PLACE_STYLES[entry.placement] ?? { bg: '#FAFAF9', text: '#78716C', border: 'transparent', emoji: '' }
          return (
            <div
              key={entry.player.id}
              className="flex items-center gap-3 px-3.5 py-3 rounded-2xl transition-all"
              style={{
                background: style.bg,
                border: `1.5px solid ${style.border}`,
              }}
            >
              {/* Placement emoji */}
              <span className="text-lg shrink-0 w-7 text-center select-none">{style.emoji || `#${entry.placement}`}</span>

              {/* Avatar */}
              <div
                className="w-8 h-8 rounded-xl flex items-center justify-center text-xs font-black shrink-0"
                style={{
                  background: entry.placement === 1 ? 'linear-gradient(135deg, #FBBF24, #F59E0B)' : '#FFFFFF',
                  color: entry.placement === 1 ? '#FFFFFF' : '#78716C',
                  border: `1.5px solid ${style.border}`,
                }}
              >
                {getInitials(entry.player.name)}
              </div>

              {/* Name & match count */}
              <div className="flex-1 min-w-0">
                <p className="text-sm font-bold truncate" style={{ color: style.text }}>
                  {entry.player.name}
                </p>
                <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>
                  {entry.wordsMatched} matched
                </p>
              </div>
            </div>
          )
        })}

        {entries.length === 0 && (
          <p className="text-sm text-center py-6" style={{ color: '#D6D3D1' }}>
            No scores yet
          </p>
        )}
      </div>
    </div>
  )
}
