'use client'
import { useState } from 'react'
import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { BingoCard } from '@/components/BingoCard'
import { CurrentCallDisplay } from '@/components/CurrentCallDisplay'
import { CalledWordsFeed } from '@/components/CalledWordsFeed'
import { Leaderboard } from '@/components/Leaderboard'
import { AIHostPanel } from '@/components/AIHostPanel'
import { mockGame, mockLeaderboard, mockBingoCard } from '@/lib/mockGameData'
import { WinnerModal } from '@/components/WinnerModal'
import { BingoCellData } from '@/types/player'

export default function PlayPage() {
  const [cells, setCells] = useState<BingoCellData[]>(mockBingoCard)
  const [showWinner, setShowWinner] = useState(false)

  const handleCellClick = (id: string) => {
    setCells(cells.map(c => c.id === id ? { ...c, isMarked: !c.isMarked } : c))
  }

  const handleClaim = () => {
    setShowWinner(true)
  }

  return (
    <AppShell>
      <TopNav gameCode={mockGame.code} playerName="Alice Smith" role="player" status="Live" />
      <main className="flex-1 flex flex-col lg:flex-row overflow-y-auto lg:overflow-hidden">
        
        {/* Main Game Area */}
        <div className="flex-1 p-4 sm:p-8 flex flex-col gap-6 lg:gap-8 overflow-y-visible lg:overflow-y-auto">
          <CurrentCallDisplay 
            word={mockGame.currentWord?.word} 
            aiMessage="Let's leverage our core competencies to create meaningful synergy across departments!" 
          />
          <div className="flex-1 flex items-center justify-center">
            <BingoCard cells={cells} onCellClick={handleCellClick} />
          </div>
          <button 
            onClick={handleClaim}
            className="w-full max-w-2xl mx-auto py-4 sm:py-5 bg-brand-600 text-white rounded-2xl font-bold text-base sm:text-lg shadow-xl shadow-brand-100 hover:bg-brand-700 active:scale-[0.98] transition-all"
          >
            CLAIM BINGO
          </button>
        </div>

        {/* Sidebar */}
        <aside className="w-full lg:w-80 bg-white border-t lg:border-t-0 lg:border-l border-slate-200 flex flex-col shrink-0 shrink-0 lg:overflow-y-auto transition-all">
          <div className="p-4 sm:p-6 border-b border-slate-100">
            <Leaderboard entries={mockLeaderboard} />
          </div>
          <div className="flex-1 p-4 sm:p-6 overflow-hidden flex flex-col lg:min-h-[200px]">
             <CalledWordsFeed words={mockGame.calledWords} />
          </div>
          <div className="mt-auto shrink-0 pt-4 sm:pt-6">
             <AIHostPanel message="Alice is only 1 mark away from a horizontal bingo! The tension is building..." />
          </div>
        </aside>

      </main>

      <WinnerModal 
         isOpen={showWinner}
         winner={{ id: '1', name: 'David Chen', state: 'Playing', connectionState: 'Connected' }}
         placement={1}
         pattern="Horizontal Row (Middle)"
         onClose={() => setShowWinner(false)}
      />
    </AppShell>
  )
}
