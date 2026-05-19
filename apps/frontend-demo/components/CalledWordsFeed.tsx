import { CalledWord } from '@/types/game'
import { Clock } from 'lucide-react'

interface CalledWordsFeedProps {
  words: CalledWord[]
}

function timeAgo(isoString: string): string {
  const diff = Math.floor((Date.now() - new Date(isoString).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

export function CalledWordsFeed({ words }: CalledWordsFeedProps) {
  const reversed = [...words].reverse()

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 mb-3">
        <Clock className="w-4 h-4" style={{ color: '#A8A29E' }} />
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
          Called Words
        </span>
        <span
          className="ml-auto text-xs font-bold px-2 py-0.5 rounded-full"
          style={{ background: '#F4F2EF', color: '#78716C' }}
        >
          {words.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 pr-1">
        {reversed.length === 0 ? (
          <p className="text-sm text-center py-8" style={{ color: '#D6D3D1' }}>No words called yet</p>
        ) : (
          reversed.map((w, i) => (
            <div
              key={w.id}
              className="flex items-center justify-between px-3.5 py-2.5 rounded-2xl transition-all"
              style={{
                background: i === 0 ? '#FFF4F0' : '#FAFAF9',
                border: i === 0 ? '1.5px solid #FFE4D9' : '1.5px solid transparent',
              }}
            >
              <span
                className="text-sm font-bold leading-tight"
                style={{ color: i === 0 ? '#E8440A' : '#44403C' }}
              >
                {w.word}
              </span>
              <span className="text-[10px] font-semibold shrink-0 ml-3" style={{ color: '#A8A29E' }}>
                {timeAgo(w.calledAt)}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
