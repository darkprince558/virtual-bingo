import { LeaderboardEntry } from "@/types/player"

interface LeaderboardProps {
  entries: LeaderboardEntry[]
}

export function Leaderboard({ entries }: LeaderboardProps) {
  return (
    <div className="flex flex-col">
      <h3 className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-4">Live Leaderboard</h3>
      <div className="space-y-3">
        {entries.map((entry) => {
          let rankColor = "text-slate-400";
          if (entry.placement === 1) rankColor = "text-amber-500";
          if (entry.placement === 3) rankColor = "text-slate-300";

          return (
            <div key={entry.player.id} className={`flex items-center justify-between p-3 rounded-xl ${entry.placement === 1 ? 'bg-slate-50' : 'bg-white border border-slate-100'}`}>
              <div className="flex items-center gap-3">
                <span className={`text-sm font-bold ${rankColor}`}>
                  {entry.placement}{entry.placement === 1 ? 'st' : entry.placement === 2 ? 'nd' : entry.placement === 3 ? 'rd' : 'th'}
                </span>
                <span className="text-sm font-medium text-slate-700">{entry.player.name}</span>
              </div>
              <span className="text-xs font-bold text-slate-400">{entry.wordsMatched} Marks</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
