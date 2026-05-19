import { BingoClaim } from "@/types/game"
import { AlertCircle, Check, X } from "lucide-react"

interface BingoClaimQueueProps {
  claims: BingoClaim[]
  onApprove?: (id: string) => void
  onReject?: (id: string) => void
}

export function BingoClaimQueue({ claims, onApprove, onReject }: BingoClaimQueueProps) {
  const pendingClaims = claims.filter(c => c.status === 'Pending' || c.status === 'Valid');

  return (
    <div className="bg-white rounded-2xl border border-slate-200 flex flex-col h-full shadow-sm overflow-hidden">
      <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50">
         <h3 className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400">Bingo Claims</h3>
         {pendingClaims.length > 0 && (
           <span className="bg-amber-100 text-amber-700 text-[10px] font-bold px-2 py-0.5 rounded-full uppercase tracking-widest">
             {pendingClaims.length} New
           </span>
         )}
      </div>
      <div className="p-0 overflow-y-auto max-h-[400px]">
        <ul className="divide-y divide-slate-100">
          {pendingClaims.map((claim) => (
            <li key={claim.id} className="p-5 flex flex-col space-y-4 bg-white hover:bg-slate-50 transition-colors">
              <div className="flex justify-between items-start">
                <div>
                  <span className="font-bold text-slate-900 block">{claim.playerName}</span>
                  <span className="text-xs font-semibold text-slate-500 mt-1 block">Pattern: {claim.pattern}</span>
                </div>
                <span className="text-[10px] text-slate-400 font-bold uppercase tracking-widest">
                  Just now
                </span>
              </div>
              <div className="flex space-x-2 w-full">
                <button 
                  onClick={() => onApprove?.(claim.id)}
                  className="flex-1 flex justify-center items-center space-x-1.5 bg-emerald-500 hover:bg-emerald-600 text-white px-3 py-2 rounded-lg text-sm font-bold transition-all shadow-sm active:scale-[0.98]"
                >
                  <Check className="w-4 h-4 stroke-[3]" />
                  <span>Verify Win</span>
                </button>
                <button 
                  onClick={() => onReject?.(claim.id)}
                  className="flex-none flex justify-center items-center bg-white border border-slate-200 hover:bg-slate-100 text-slate-600 px-3 py-2 rounded-lg transition-colors"
                  aria-label="Reject Claim"
                >
                  <X className="w-4 h-4 stroke-[2]" />
                </button>
              </div>
            </li>
          ))}
          {pendingClaims.length === 0 && (
            <li className="p-8 text-center text-slate-400 text-sm italic">
              No pending bingo claims.
            </li>
          )}
        </ul>
      </div>
    </div>
  )
}
