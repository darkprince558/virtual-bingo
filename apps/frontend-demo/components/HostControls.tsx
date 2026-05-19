import { Play, Square, Pause, Mic } from "lucide-react"

interface HostControlsProps {
  onStart?: () => void
  onPause?: () => void
  onNextWord?: () => void
  onEnd?: () => void
  status: string
}

export function HostControls({ onStart, onPause, onNextWord, onEnd, status }: HostControlsProps) {
  const isLive = status === 'Live'

  return (
    <div className="bg-white rounded-2xl border border-slate-200 p-6 flex flex-col space-y-4 shadow-sm">
      <h3 className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400 mb-2">Game Controls</h3>
      <button 
        onClick={onNextWord}
        disabled={!isLive}
        className="w-full flex items-center justify-center space-x-2 bg-brand-600 hover:bg-brand-700 disabled:bg-brand-300 text-white py-4 rounded-xl text-lg font-bold transition-all shadow-md active:scale-[0.98]"
      >
        <Mic className="w-5 h-5" />
        <span>Call Next Word</span>
      </button>

      <div className="grid grid-cols-2 gap-3 pt-4 border-t border-slate-100">
        {!isLive ? (
           <button 
             onClick={onStart}
             className="flex items-center justify-center space-x-2 bg-slate-50 text-emerald-600 hover:bg-slate-100 border border-slate-200 py-3 rounded-xl text-sm font-bold transition-colors"
           >
             <Play className="w-4 h-4" />
             <span>Start Game</span>
           </button>
        ) : (
           <button 
             onClick={onPause}
             className="flex items-center justify-center space-x-2 bg-slate-50 text-amber-600 hover:bg-slate-100 border border-slate-200 py-3 rounded-xl text-sm font-bold transition-colors"
           >
             <Pause className="w-4 h-4" />
             <span>Pause Game</span>
           </button>
        )}
        <button 
          onClick={onEnd}
          className="flex items-center justify-center space-x-2 bg-slate-50 text-rose-600 hover:bg-slate-100 border border-slate-200 py-3 rounded-xl text-sm font-bold transition-colors"
        >
          <Square className="w-4 h-4" />
          <span>End Game</span>
        </button>
      </div>
    </div>
  )
}
