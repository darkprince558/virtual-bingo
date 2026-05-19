import { Player } from '@/types/player'
import { Wifi, WifiOff, Users } from 'lucide-react'

interface PlayerListProps {
  players: Player[]
  totalConnected?: number
}

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

const STATE_CHIP: Record<string, { bg: string; text: string }> = {
  Playing:          { bg: '#EDFAF5', text: '#116B3F' },
  Waiting:          { bg: '#FEF3C7', text: '#D97706' },
  'Claimed Bingo':  { bg: '#FFF4F0', text: '#E8440A' },
  'Confirmed Winner': { bg: '#EDFAF5', text: '#116B3F' },
  'Rejected Claim': { bg: '#FFF1F2', text: '#E11D48' },
  Disconnected:     { bg: '#F4F2EF', text: '#A8A29E' },
}

// Warm color palette for avatars (cycles through)
const AVATAR_COLORS = [
  { bg: '#FFF4F0', text: '#E8440A' },
  { bg: '#F5F2FF', text: '#6440E8' },
  { bg: '#EDFAF5', text: '#116B3F' },
  { bg: '#FEF3C7', text: '#B45309' },
  { bg: '#FFF1F2', text: '#BE123C' },
]

export function PlayerList({ players, totalConnected }: PlayerListProps) {
  return (
    <div
      className="h-full rounded-3xl flex flex-col overflow-hidden"
      style={{
        background: '#FFFFFF',
        border: '1.5px solid #F0EDE8',
        boxShadow: '0 2px 16px rgba(0,0,0,0.04)',
      }}
    >
      {/* Header */}
      <div className="px-5 pt-5 pb-4 flex items-center gap-3 shrink-0" style={{ borderBottom: '1px solid #F4F2EF' }}>
        <div
          className="w-8 h-8 rounded-xl flex items-center justify-center"
          style={{ background: '#F5F2FF', color: '#7C5CFC' }}
        >
          <Users className="w-4 h-4" />
        </div>
        <div>
          <h3 className="text-sm font-extrabold" style={{ color: '#1C1917' }}>Players</h3>
          <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>
            {totalConnected ?? players.length} connected
          </p>
        </div>
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {players.map((player, i) => {
          const avatarColor = AVATAR_COLORS[i % AVATAR_COLORS.length]
          const chip = STATE_CHIP[player.state] ?? { bg: '#F4F2EF', text: '#78716C' }
          const isConnected = player.connectionState === 'Connected'

          return (
            <div
              key={player.id}
              className="flex items-center gap-3 px-3.5 py-2.5 rounded-2xl"
              style={{ background: '#FAFAF9', border: '1.5px solid transparent' }}
            >
              {/* Avatar */}
              <div
                className="w-8 h-8 rounded-xl flex items-center justify-center text-xs font-black shrink-0"
                style={{ background: avatarColor.bg, color: avatarColor.text }}
              >
                {getInitials(player.name)}
              </div>

              {/* Name */}
              <span className="flex-1 text-sm font-bold truncate" style={{ color: '#1C1917' }}>
                {player.name}
              </span>

              {/* State chip */}
              <span
                className="text-[10px] font-extrabold px-2 py-0.5 rounded-full uppercase tracking-wide shrink-0"
                style={{ background: chip.bg, color: chip.text }}
              >
                {player.state}
              </span>

              {/* Connection icon */}
              {isConnected
                ? <Wifi className="w-3.5 h-3.5 shrink-0" style={{ color: '#22AA6A' }} />
                : <WifiOff className="w-3.5 h-3.5 shrink-0" style={{ color: '#F43F5E' }} />
              }
            </div>
          )
        })}
      </div>
    </div>
  )
}
