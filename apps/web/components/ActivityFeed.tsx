'use client'

import { ActivityEvent } from '@/types/game'
import { Activity } from 'lucide-react'

interface ActivityFeedProps {
  events: ActivityEvent[]
}

const TONE_STYLES: Record<string, { dot: string; bg: string; text: string }> = {
  success: { dot: '#22AA6A', bg: '#EDFAF5', text: '#116B3F' },
  warning: { dot: '#F59E0B', bg: '#FEF3C7', text: '#B45309' },
  danger:  { dot: '#F43F5E', bg: '#FFF1F2', text: '#BE123C' },
  neutral: { dot: '#A8A29E', bg: '#F4F2EF', text: '#57534E' },
}

function timeAgo(isoString: string): string {
  const diff = Math.floor((Date.now() - new Date(isoString).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

export function ActivityFeed({ events }: ActivityFeedProps) {
  const reversed = [...events].reverse()

  return (
    <div
      className="rounded-xl p-5 flex flex-col"
      style={{
        background: '#FFFFFF',
        border: '1.5px solid #F0EDE8',
        boxShadow: '0 2px 16px rgba(0,0,0,0.04)',
      }}
    >
      <div className="flex items-center gap-2 mb-4">
        <Activity className="w-4 h-4" style={{ color: '#A8A29E' }} />
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
          Activity Log
        </span>
      </div>

      <div className="space-y-3 max-h-48 overflow-y-auto pr-1">
        {reversed.length === 0 ? (
          <p className="text-sm text-center py-4" style={{ color: '#D6D3D1' }}>No activity yet</p>
        ) : (
          reversed.map((event) => {
            const tone = TONE_STYLES[event.tone ?? 'neutral']
            return (
              <div key={event.id} className="flex items-start gap-3">
                {/* Timeline dot */}
                <div className="flex flex-col items-center shrink-0 mt-1.5">
                  <div
                    className="w-2 h-2 rounded-full shrink-0"
                    style={{ background: tone.dot }}
                  />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-baseline gap-2 flex-wrap">
                    <span
                      className="text-[10px] font-extrabold px-2 py-0.5 rounded-full uppercase tracking-wide"
                      style={{ background: tone.bg, color: tone.text }}
                    >
                      {event.label}
                    </span>
                    <span className="text-[10px] font-semibold" style={{ color: '#D6D3D1' }}>
                      {timeAgo(event.createdAt)}
                    </span>
                  </div>
                  <p className="text-xs font-semibold mt-0.5 leading-snug" style={{ color: '#78716C' }}>
                    {event.detail}
                  </p>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
