'use client'
import { motion, AnimatePresence } from 'motion/react'
import { Player } from '@/types/player'
import { BingoPattern } from '@/types/game'

interface WinnerModalProps {
  isOpen: boolean
  winner: Player | null
  placement: 1 | 2 | 3
  pattern: BingoPattern | string
  onClose: () => void
}

export function WinnerModal({ isOpen, winner, placement, pattern, onClose }: WinnerModalProps) {
  if (!isOpen || !winner) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
        <motion.div
           initial={{ opacity: 0, scale: 0.95, y: 20 }}
           animate={{ opacity: 1, scale: 1, y: 0 }}
           exit={{ opacity: 0, scale: 0.95, y: 20 }}
           className="bg-white w-full max-w-[480px] rounded-lg p-6 sm:p-10 shadow-2xl flex flex-col items-center text-center relative overflow-hidden"
        >
          <div className="absolute top-0 inset-x-0 h-2 bg-brand-600"></div>
          
          <div className="w-16 h-16 sm:w-20 sm:h-20 bg-brand-100 rounded-full flex items-center justify-center text-brand-600 mb-4 sm:mb-6">
            <svg className="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 5a3 3 0 013 3v1h-3V8a1 1 0 011-1zM5 8a3 3 0 013-3h7a3 3 0 013 3v10a3 3 0 01-3 3H8a3 3 0 01-3-3V8z"></path>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4"></path>
            </svg>
          </div>
          
          <p className="text-xs font-bold uppercase tracking-[0.3em] text-brand-600 mb-2">BINGO CONFIRMED</p>
          <h2 className="text-3xl font-black text-slate-900 mb-2">{winner.name}</h2>
          
          <div className="inline-block px-4 py-1 bg-amber-100 text-amber-700 text-xs font-bold rounded-full mb-6 uppercase tracking-wider">
            {placement === 1 ? '1st' : placement === 2 ? '2nd' : '3rd'} Place Winner
          </div>
          
          <div className="w-full p-4 bg-slate-50 rounded-2xl border border-slate-100 flex items-center justify-between mb-8 text-left">
            <div>
              <p className="text-[10px] uppercase font-bold text-slate-400 tracking-widest mb-1">Winning Pattern</p>
              <p className="text-sm font-semibold text-slate-700">{pattern}</p>
            </div>
            <div className="w-10 h-10 bg-white border border-slate-200 rounded p-1.5 grid grid-cols-3 gap-0.5 opacity-60">
              <div className="bg-slate-100"></div><div className="bg-slate-100"></div><div className="bg-slate-100"></div>
              <div className="bg-brand-500"></div><div className="bg-brand-500"></div><div className="bg-brand-500"></div>
              <div className="bg-slate-100"></div><div className="bg-slate-100"></div><div className="bg-slate-100"></div>
            </div>
          </div>
          
          <button 
             onClick={onClose}
             className="w-full py-4 bg-slate-900 text-white rounded-xl font-bold hover:bg-slate-800 transition-colors shadow-lg shadow-slate-200"
          >
             Continue Game
          </button>
          
          <p className="mt-6 text-xs text-slate-400 italic">Validation complete • Activity #4298</p>
          
        </motion.div>
      </div>
    </AnimatePresence>
  )
}
