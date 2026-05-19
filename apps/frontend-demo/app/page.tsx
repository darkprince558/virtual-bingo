import { AppShell } from '@/components/AppShell'
import { TopNav } from '@/components/TopNav'
import Link from 'next/link'

export default function LandingPage() {
  return (
    <AppShell>
      <TopNav role="player" />
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-white rounded-lg p-6 sm:p-10 shadow-xl flex flex-col items-center text-center border border-slate-200">
          
          <div className="w-16 h-16 bg-brand-600 rounded-lg flex items-center justify-center text-white font-bold text-3xl mb-4 sm:mb-6 shadow-md shadow-brand-200">
            B
          </div>
          <h2 className="text-3xl font-black text-slate-900 tracking-tight mb-2">
            Virtual Bingo
          </h2>
          <p className="text-sm font-medium text-slate-500 mb-8">
            Centralized cards, word calls, claim review, and winner tracking for a live team event.
          </p>

          <div className="w-full space-y-4">
            <button
              disabled
              className="w-full py-4 rounded-xl font-bold text-sm bg-slate-100 text-slate-400 cursor-not-allowed border border-slate-200"
            >
              Sign in with Microsoft Entra
            </button>
            
            <div className="relative py-4">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-200" />
              </div>
              <div className="relative flex justify-center">
                <span className="bg-white px-4 text-[10px] uppercase font-bold tracking-widest text-slate-400">Join with Code</span>
              </div>
            </div>

            <div className="flex flex-col space-y-3">
              <div className="flex gap-2">
                <input 
                  type="text" 
                  placeholder="GAME CODE" 
                  aria-label="Game code"
                  className="flex-1 bg-slate-50 border border-slate-200 rounded-lg px-4 py-3.5 text-sm font-bold text-slate-900 uppercase tracking-widest placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:bg-white transition-all text-center sm:text-left"
                />
                <Link
                  href="/lobby"
                  className="flex items-center justify-center px-6 rounded-lg text-sm font-bold bg-brand-600 text-white hover:bg-brand-700 transition-colors shadow-sm"
                >
                  Join
                </Link>
              </div>
              <Link
                href="/host"
                className="flex items-center justify-center py-3.5 px-4 rounded-lg text-[10px] uppercase tracking-widest font-bold text-slate-500 hover:bg-slate-50 transition-colors"
              >
                Host a Game Instead
              </Link>
            </div>
          </div>
        </div>
      </main>
    </AppShell>
  )
}
