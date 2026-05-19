'use client'
import { useState } from 'react'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { HostControls } from '@/components/HostControls'
import { CurrentCallDisplay } from '@/components/CurrentCallDisplay'
import { BingoClaimQueue } from '@/components/BingoClaimQueue'
import { PlayerList } from '@/components/PlayerList'
import { CalledWordsFeed } from '@/components/CalledWordsFeed'
import { Leaderboard } from '@/components/Leaderboard'
import { ActivityFeed } from '@/components/ActivityFeed'
import { mockGame, mockPlayers, mockLeaderboard } from '@/lib/mockGameData'
import { GameState } from '@/types/game'
import Link from 'next/link'

export default function HostPage() {
  const [status, setStatus] = useState<GameState>('Live')

  return (
    <AppShell>
      <TopNav gameCode={mockGame.code} playerName={mockGame.hostName} role="host" status={status} />
      <main className="flex-1 overflow-y-auto p-4 sm:p-8">
        
        <div className="mb-6 sm:mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
           <div>
             <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400 mb-1">Administration</p>
             <h1 className="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight">Host Dashboard</h1>
             <p className="mt-2 text-sm font-medium text-slate-500">Run the session, review claims, and keep the winner order auditable.</p>
           </div>
           <Link href="/summary" className="bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-widest transition-colors shadow-sm self-start sm:self-auto">
              View Summary
           </Link>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 w-full">
          
          {/* Left Column: Primary Controls & Current Status */}
          <div className="col-span-1 lg:col-span-5 flex flex-col space-y-6">
             <CurrentCallDisplay word={mockGame.currentWord?.word} />
             <HostControls 
                status={status}
                onStart={() => setStatus('Live')}
                onPause={() => setStatus('Paused')}
                onEnd={() => setStatus('Finished')}
                onNextWord={() => console.log('Next Word')}
             />
             <div className="flex-1 bg-white rounded-2xl border border-slate-200 p-6 flex flex-col min-h-[250px] shadow-sm">
                <CalledWordsFeed words={mockGame.calledWords} />
             </div>
             <ActivityFeed events={mockGame.activityEvents} />
          </div>

          {/* Right Column (Stacked grids): Claims & Players */}
          <div className="col-span-1 lg:col-span-7 flex flex-col gap-6">
             <div className="h-[400px]">
               <BingoClaimQueue 
                 claims={mockGame.claims} 
                 onApprove={(id) => console.log('Approve', id)}
                 onReject={(id) => console.log('Reject', id)}
               />
             </div>
             <div className="flex-1 min-h-[300px]">
                <PlayerList players={mockPlayers} totalConnected={mockGame.connectedPlayers} />
             </div>
             <div className="bg-white rounded-lg border border-slate-200 p-5 shadow-sm">
               <Leaderboard entries={mockLeaderboard.slice(0, 3)} />
             </div>
          </div>

        </div>
      </main>
    </AppShell>
  )
}
