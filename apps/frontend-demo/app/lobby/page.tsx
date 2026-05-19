'use client'

import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import { mockGame, mockPlayers } from '@/lib/mockGameData'
import Link from 'next/link'
import { useSettings, AVATARS, ThemeType } from '@/contexts/SettingsContext'

const THEMES: { id: ThemeType; label: string; colorClass: string }[] = [
  { id: 'indigo', label: 'Classic Indigo', colorClass: 'bg-indigo-500' },
  { id: 'rose', label: 'Vibrant Rose', colorClass: 'bg-rose-500' },
  { id: 'emerald', label: 'Forest Emerald', colorClass: 'bg-emerald-500' },
  { id: 'amber', label: 'Sunset Amber', colorClass: 'bg-amber-500' },
]

export default function LobbyPage() {
  const { avatar, setAvatar, theme, setTheme } = useSettings()

  return (
    <AppShell>
      <TopNav gameCode={mockGame.code} playerName="Sarah Jenkins" role="player" status="Waiting" />
      <main className="flex-1 overflow-y-auto p-4 sm:p-8 flex flex-col items-center">
        <div className="w-full max-w-5xl grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8 mt-4 sm:mt-8">
          
          {/* Status Column */}
          <div className="bg-white rounded-lg p-6 sm:p-10 shadow-xl flex flex-col items-center text-center border border-slate-200 flex-1 lg:col-span-1">
            <div className="w-16 h-16 sm:w-20 sm:h-20 bg-brand-50 rounded-lg flex items-center justify-center mb-6">
              <div className="w-4 h-4 bg-brand-600 rounded-full animate-bounce"></div>
            </div>

            <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400 mb-2">Status</p>
            <h1 className="text-2xl font-black text-slate-900 mb-6">Waiting for Host...</h1>
            
            <div className="w-full bg-slate-50 rounded-2xl p-4 mb-8 border border-slate-200 flex items-center justify-between">
              <div className="text-left">
                <span className="block text-[10px] font-bold uppercase tracking-widest text-slate-400 mb-1">Game Code</span>
                <span className="block text-xl font-black text-brand-600 font-mono tracking-widest">{mockGame.code}</span>
              </div>
              <div className="h-10 w-px bg-slate-200" />
              <div className="text-right">
              <span className="block text-[10px] font-bold uppercase tracking-widest text-slate-400 mb-1">Players</span>
                <span className="block text-xl font-black text-slate-800">{mockGame.connectedPlayers}</span>
              </div>
            </div>

            <Link
               href="/play"
               className="w-full py-4 bg-brand-600 text-white rounded-lg font-bold hover:bg-brand-700 transition-colors shadow-lg shadow-brand-200 flex justify-center mt-auto"
            >
               Enter Game (Demo)
            </Link>
          </div>

          {/* Customization Column */}
          <div className="bg-white rounded-lg p-6 sm:p-8 shadow-sm border border-slate-200 flex flex-col lg:col-span-1">
             <div className="mb-6 pb-4 border-b border-slate-100">
               <h2 className="text-lg font-bold text-slate-800">Profile</h2>
               <p className="text-xs font-medium text-slate-400 mt-1">Choose your avatar and interface color.</p>
             </div>

             <div className="flex flex-col gap-6">
                <div>
                  <p className="text-[10px] font-bold uppercase tracking-widest text-slate-400 mb-3">Your Avatar</p>
                  <div className="grid grid-cols-5 gap-2">
                    {AVATARS.slice(0, 10).map((a) => (
                      <button
                        key={a}
                        onClick={() => setAvatar(a)}
                        className={`aspect-square rounded-lg text-xl flex items-center justify-center transition-all ${
                          avatar === a
                            ? 'bg-brand-100 border-2 border-brand-500 shadow-sm scale-110 z-10'
                            : 'bg-slate-50 border border-slate-200 hover:bg-slate-100 hover:scale-105'
                        }`}
                      >
                        {a}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <p className="text-[10px] font-bold uppercase tracking-widest text-slate-400 mb-3">Game Theme</p>
                  <div className="grid grid-cols-2 gap-2">
                    {THEMES.map((t) => (
                      <button
                        key={t.id}
                        onClick={() => setTheme(t.id)}
                         className={`flex items-center gap-2 p-2 rounded-lg border text-left transition-all ${
                          theme === t.id 
                            ? 'border-brand-500 bg-brand-50 shadow-sm ring-1 ring-brand-500'
                            : 'border-slate-200 hover:bg-slate-50'
                        }`}
                      >
                        <div className={`w-4 h-4 rounded-full ${t.colorClass}`}></div>
                        <span className={`text-xs font-bold ${theme === t.id ? 'text-brand-900' : 'text-slate-600'}`}>{t.label}</span>
                      </button>
                    ))}
                  </div>
                </div>
             </div>
          </div>

          {/* Players Column */}
          <div className="bg-white rounded-lg p-6 sm:p-8 shadow-sm border border-slate-200 flex flex-col max-h-[600px] lg:col-span-1">
            <div className="flex items-center justify-between mb-6 pb-4 border-b border-slate-100">
               <div>
                 <h2 className="text-lg font-bold text-slate-800">Connected Players</h2>
                 <p className="text-xs font-medium text-slate-400 mt-1">Players currently assigned to this session.</p>
               </div>
               <span className="bg-emerald-50 text-emerald-600 border border-emerald-100 px-3 py-1 rounded-full text-xs font-bold uppercase tracking-widest flex items-center gap-2">
                 <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                 Live
               </span>
            </div>
            
            <div className="flex-1 overflow-y-auto pr-2">
               <div className="grid grid-cols-1 gap-3">
                 {mockPlayers.map((p, i) => (
                   <div key={p.id} className="flex items-center gap-3 p-3 bg-slate-50 border border-slate-100 rounded-lg">
                     <div className="w-10 h-10 rounded bg-white border border-slate-200 text-brand-600 text-xl font-black flex items-center justify-center shadow-sm">
                       {i === 0 ? avatar : AVATARS[(i * 3) % AVATARS.length]}
                     </div>
                     <div className="flex flex-col">
                       <span className="text-sm font-bold text-slate-700">{i === 0 ? "You (Sarah)" : p.name}</span>
                       <span className="text-[10px] uppercase font-bold tracking-widest text-slate-400">{p.state}</span>
                     </div>
                   </div>
                 ))}
               </div>
            </div>
          </div>

        </div>
      </main>
    </AppShell>
  )
}
