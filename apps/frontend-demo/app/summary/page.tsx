import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { mockGame, mockLeaderboard } from '@/lib/mockGameData'
import Link from 'next/link'

export default function SummaryPage() {
  return (
    <AppShell>
      <TopNav gameCode={mockGame.code} playerName="Admin Team" role="host" status="Finished" />
      <main className="max-w-4xl mx-auto px-4 sm:px-8 py-8 sm:py-12 w-full">
        <div className="text-center mb-8 sm:mb-12">
           <div className="mx-auto w-16 h-16 bg-slate-100 rounded-lg flex items-center justify-center text-xl font-black text-slate-400 mb-4 sm:mb-6 border border-slate-200">
              END
           </div>
           <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400 mb-2">Session Ended</p>
           <h1 className="text-3xl sm:text-4xl font-black text-slate-900 tracking-tight">Game Concluded</h1>
        </div>

        <div className="grid grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6 mb-8 sm:mb-12">
           <div className="bg-white rounded-2xl p-6 shadow-sm border border-slate-200 flex flex-col justify-center items-center text-center">
              <span className="text-4xl font-black text-brand-600 mb-2">{mockGame.connectedPlayers}</span>
              <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Total Players</span>
           </div>
           <div className="bg-white rounded-2xl p-6 shadow-sm border border-slate-200 flex flex-col justify-center items-center text-center">
              <span className="text-4xl font-black text-amber-500 mb-2">{mockGame.calledWords.length}</span>
              <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Words Called</span>
           </div>
           <div className="bg-white rounded-2xl p-6 shadow-sm border border-slate-200 flex flex-col justify-center items-center text-center">
              <span className="text-4xl font-black text-emerald-500 mb-2">45</span>
              <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Minutes</span>
           </div>
        </div>

        <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden mb-8">
           <div className="px-8 py-6 border-b border-slate-100 bg-slate-50">
             <h2 className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400">Final Winners</h2>
           </div>
           <div className="p-4 sm:p-8 space-y-4">
                 {mockLeaderboard.slice(0, 3).map((entry) => (
                    <div key={entry.player.id} className="flex flex-col sm:flex-row sm:items-center p-4 rounded-2xl border border-slate-100 bg-white shadow-sm gap-4 sm:gap-6">
                       <div className="flex items-center gap-4 flex-1">
                         <div className="flex-shrink-0 w-12 h-12 rounded-xl flex items-center justify-center bg-slate-50 border border-slate-100">
                             <span className="text-xl font-black text-slate-300">
                               {entry.placement}
                             </span>
                         </div>
                         <div className="flex-1">
                            <h3 className="text-lg font-bold text-slate-800">{entry.player.name}</h3>
                            <p className="text-[10px] uppercase font-bold text-slate-400 tracking-widest mt-1">Placement: {entry.placement === 1 ? '1st' : entry.placement === 2 ? '2nd' : '3rd'}</p>
                         </div>
                       </div>
                       <div className="sm:text-right">
                          <span className="inline-block px-3 py-1 bg-emerald-50 border border-emerald-100 text-emerald-600 text-[10px] font-bold uppercase tracking-widest rounded-full">
                             Prize Notified
                          </span>
                       </div>
                    </div>
                 ))}
           </div>
        </div>

        <div className="flex flex-col-reverse sm:flex-row justify-between items-center py-4 gap-4">
           <Link href="/host" className="text-xs font-bold uppercase tracking-widest text-slate-500 hover:text-slate-900 transition-colors py-2">
              Back to Dashboard
           </Link>
           <button className="w-full sm:w-auto bg-slate-900 hover:bg-slate-800 text-white px-6 py-3 rounded-xl font-bold text-sm transition-colors shadow-md">
              Export Audit Log
           </button>
        </div>
      </main>
    </AppShell>
  )
}
