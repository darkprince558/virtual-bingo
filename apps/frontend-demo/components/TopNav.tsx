'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { Shield, Settings } from 'lucide-react'
import { GameState } from '@/types/game'
import { SettingsModal } from './SettingsModal'
import { useSettings } from '@/contexts/SettingsContext'

interface TopNavProps {
  gameCode?: string
  playerName?: string
  status?: GameState
  role: 'player' | 'host'
}

export function TopNav({ gameCode, playerName, status, role }: TopNavProps) {
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)
  const { avatar } = useSettings()

  return (
    <>
      <nav className="h-16 bg-white border-b border-slate-200 px-4 sm:px-8 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2 sm:gap-6">
          <Link href="/" className="flex items-center gap-2">
            <div className="w-8 h-8 bg-brand-600 rounded flex items-center justify-center text-white font-bold shrink-0">B</div>
            <span className="text-base sm:text-lg font-semibold tracking-tight text-slate-800 hidden sm:inline-block">Virtual Bingo</span>
          </Link>
          {gameCode && (
            <>
              <div className="hidden sm:block h-6 w-px bg-slate-200"></div>
              <div className="flex items-center gap-2 sm:gap-3">
                <span className="hidden sm:inline-block text-xs font-semibold uppercase tracking-widest text-slate-400">Game Code:</span>
                <span className="bg-slate-100 px-2 py-1 rounded font-mono text-sm font-bold text-brand-600">{gameCode}</span>
              </div>
            </>
          )}
        </div>

        <div className="flex items-center gap-3 sm:gap-6">
           {status && (
             <div className="flex items-center gap-1.5 sm:gap-2">
               {status === 'Live' && <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>}
               <span className="text-[10px] sm:text-xs font-semibold uppercase tracking-wider text-slate-500">{status}</span>
             </div>
           )}
           {playerName && (
            <div className="flex items-center gap-2 sm:gap-3 sm:border-l sm:pl-6 border-slate-200">
              <div className="text-right hidden sm:block">
                <p className="text-sm font-semibold text-slate-800">{playerName}</p>
                <p className="text-[10px] text-slate-400 uppercase tracking-tighter">{role === 'host' ? 'Host' : 'Player'}</p>
              </div>
              <div className="w-8 h-8 sm:w-9 sm:h-9 rounded bg-slate-100 border border-slate-200 shadow-sm flex items-center justify-center shrink-0 text-xl overflow-hidden leading-none select-none">
                {role === 'host' ? <Shield className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-slate-500" /> : avatar}
              </div>
            </div>
          )}
          
          <div className="h-6 w-px bg-slate-200"></div>
          <button 
            onClick={() => setIsSettingsOpen(true)}
            aria-label="Open appearance settings"
            className="w-8 h-8 flex items-center justify-center text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-full transition-colors"
          >
            <Settings className="w-4 h-4 sm:w-5 sm:h-5" />
          </button>
        </div>
      </nav>
      <SettingsModal isOpen={isSettingsOpen} onClose={() => setIsSettingsOpen(false)} />
    </>
  )
}
