import { Player } from "@/types/player"

interface PlayerListProps {
  players: Player[]
  totalConnected?: number
}

export function PlayerList({ players, totalConnected }: PlayerListProps) {
  return (
    <div className="bg-white rounded-2xl border border-slate-200 flex flex-col h-full shadow-sm overflow-hidden">
      <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50">
        <h3 className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400">Players</h3>
        <span className="text-[10px] font-bold uppercase tracking-widest bg-brand-50 text-brand-600 px-2 py-0.5 rounded-full border border-brand-100">
            {totalConnected || players.length} Online
        </span>
      </div>
      <div className="p-0 overflow-y-auto max-h-[400px]">
        <ul className="divide-y divide-slate-100">
          {players.map((player) => (
            <li key={player.id} className="px-6 py-4 flex items-center justify-between hover:bg-slate-50 transition-colors">
              <span className="font-semibold text-slate-700">{player.name}</span>
              <span className="flex items-center space-x-2">
                  <span className={`w-2 h-2 rounded-full ${player.connectionState === 'Connected' ? 'bg-emerald-500' : 'bg-slate-300'}`} />
                  <span className="text-[10px] text-slate-400 font-bold uppercase tracking-widest">{player.state}</span>
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
