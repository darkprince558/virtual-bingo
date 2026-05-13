'use client'

import React from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { X, Palette, Smile } from 'lucide-react'
import { useSettings, ThemeType, AVATARS } from '@/contexts/SettingsContext'

interface SettingsModalProps {
  isOpen: boolean
  onClose: () => void
}

const THEMES: { id: ThemeType; label: string; colorClass: string }[] = [
  { id: 'indigo', label: 'Classic Indigo', colorClass: 'bg-indigo-500' },
  { id: 'rose', label: 'Vibrant Rose', colorClass: 'bg-rose-500' },
  { id: 'emerald', label: 'Forest Emerald', colorClass: 'bg-emerald-500' },
  { id: 'amber', label: 'Sunset Amber', colorClass: 'bg-amber-500' },
]

export function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  const { theme, setTheme, avatar, setAvatar } = useSettings()

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
          />
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 10 }}
            className="relative w-full max-w-sm bg-white rounded-lg shadow-xl border border-slate-100 overflow-hidden flex flex-col max-h-[80vh]"
          >
            <div className="p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50 shrink-0">
              <div className="flex items-center gap-2">
                <Palette className="w-5 h-5 text-slate-500" />
              <h3 className="font-bold text-slate-800">Appearance</h3>
              </div>
              <button 
                onClick={onClose}
                aria-label="Close appearance settings"
                className="w-8 h-8 flex items-center justify-center bg-white rounded hover:bg-slate-200 transition-colors border border-slate-200 text-slate-500"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            
            <div className="p-6 overflow-y-auto flex-1">
              <div className="mb-8">
                <div className="flex items-center gap-2 mb-4">
                  <Palette className="w-4 h-4 text-brand-500" />
                  <p className="text-xs font-bold uppercase tracking-widest text-slate-400">Choose Theme</p>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  {THEMES.map((t) => (
                    <button
                      key={t.id}
                      onClick={() => setTheme(t.id)}
                      className={`flex items-center gap-3 p-3 rounded-lg border text-left transition-all ${
                        theme === t.id 
                          ? 'border-brand-500 bg-brand-50 shadow-sm ring-1 ring-brand-500'
                          : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50 bg-white'
                      }`}
                    >
                      <div className={`w-5 h-5 rounded-full ${t.colorClass} shadow-inner`}></div>
                      <span className={`text-sm font-semibold ${theme === t.id ? 'text-brand-900' : 'text-slate-700'}`}>
                        {t.label}
                      </span>
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <div className="flex items-center gap-2 mb-4">
                  <Smile className="w-4 h-4 text-brand-500" />
                  <p className="text-xs font-bold uppercase tracking-widest text-slate-400">Choose Avatar</p>
                </div>
                <div className="grid grid-cols-5 gap-2">
                  {AVATARS.map((a) => (
                    <button
                      key={a}
                      onClick={() => setAvatar(a)}
                      className={`aspect-square rounded-lg text-2xl flex items-center justify-center transition-all ${
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
            </div>
            
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
