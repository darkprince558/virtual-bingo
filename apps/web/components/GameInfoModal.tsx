'use client'

import React, { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { X, Info, Sparkles, BookOpen, Settings } from 'lucide-react'
import { apiClient } from '@/lib/apiClient'
import type { GameContentResponse, GameSettingsResponse } from '@/types/api'
import { mapBingoPattern } from '@/lib/uiMappers'

interface GameInfoModalProps {
  isOpen: boolean
  onClose: () => void
  gameId?: string
}

export function GameInfoModal({ isOpen, onClose, gameId }: GameInfoModalProps) {
  const [content, setContent] = useState<GameContentResponse | null>(null)
  const [settings, setSettings] = useState<GameSettingsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'content' | 'settings' | 'words'>('content')

  useEffect(() => {
    if (!isOpen || !gameId) return

    let isCancelled = false

    async function fetchData() {
      setIsLoading(true)
      setError(null)
      try {
        const [fetchedContent, fetchedSettings] = await Promise.all([
          apiClient<GameContentResponse>(`/games/${gameId}/content`),
          apiClient<GameSettingsResponse>(`/games/${gameId}/settings`)
        ])

        if (!isCancelled) {
          setContent(fetchedContent)
          setSettings(fetchedSettings)
        }
      } catch (err) {
        console.error('Failed to load game info details:', err)
        if (!isCancelled) {
          setError('Could not retrieve game settings and topic details.')
        }
      } finally {
        if (!isCancelled) {
          setIsLoading(false)
        }
      }
    }

    void fetchData()

    return () => {
      isCancelled = true
    }
  }, [isOpen, gameId])

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
          {/* Glassmorphic Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="absolute inset-0"
            style={{ background: 'rgba(28, 25, 23, 0.45)', backdropFilter: 'blur(10px)' }}
          />

          {/* Premium Modal Dialog */}
          <motion.div
            initial={{ opacity: 0, scale: 0.93, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 15 }}
            transition={{ type: 'spring', stiffness: 350, damping: 26 }}
            className="relative w-full max-w-lg overflow-hidden flex flex-col"
            style={{
              background: 'rgba(255, 255, 255, 0.95)',
              backdropFilter: 'blur(16px)',
              border: '1.5px solid rgba(240, 237, 232, 0.9)',
              borderRadius: '1.5rem',
              boxShadow: '0 24px 60px rgba(0, 0, 0, 0.15)',
              maxHeight: '85vh',
            }}
          >
            {/* Header */}
            <div
              className="px-6 py-5 flex items-center justify-between shrink-0"
              style={{ borderBottom: '1px solid rgba(244, 242, 239, 0.9)' }}
            >
              <div className="flex items-center gap-3">
                <div
                  className="w-9 h-9 rounded-lg flex items-center justify-center text-white"
                  style={{
                    background: 'linear-gradient(135deg, #C0003D 0%, #E8002D 100%)',
                    boxShadow: '0 4px 12px rgba(232, 0, 45, 0.25)',
                  }}
                >
                  <Info className="w-4.5 h-4.5" />
                </div>
                <div>
                  <h3 className="font-black text-lg leading-tight" style={{ color: '#1C1917' }}>Game Details</h3>
                  <p className="text-xs font-bold uppercase tracking-wider mt-0.5" style={{ color: '#A8A29E' }}>Lobby & Play Info</p>
                </div>
              </div>
              <button
                onClick={onClose}
                aria-label="Close game info"
                className="w-8 h-8 flex items-center justify-center rounded-lg transition-all"
                style={{ background: '#F4F2EF', color: '#A8A29E' }}
                onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.background = '#E7E5E4'; (e.currentTarget as HTMLButtonElement).style.color = '#78716C'; }}
                onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.background = '#F4F2EF'; (e.currentTarget as HTMLButtonElement).style.color = '#A8A29E'; }}
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Tab navigation */}
            <div className="flex px-4 pt-3 gap-1 border-b border-[#F4F2EF] bg-[#FCFAF7] shrink-0">
              {(['content', 'settings', 'words'] as const).map(tab => {
                const isActive = activeTab === tab
                const labels = {
                  content: { label: 'Topic & Style', icon: <Sparkles className="w-4 h-4" /> },
                  settings: { label: 'Rules & Settings', icon: <Settings className="w-4 h-4" /> },
                  words: { label: 'Bingo Deck', icon: <BookOpen className="w-4 h-4" /> }
                }
                return (
                  <button
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className="flex-1 py-3 px-2 flex items-center justify-center gap-1.5 text-xs font-black uppercase tracking-wider transition-all border-b-2 rounded-t-lg"
                    style={{
                      borderColor: isActive ? '#E8002D' : 'transparent',
                      color: isActive ? '#E8002D' : '#878684',
                    }}
                  >
                    {labels[tab].icon}
                    {labels[tab].label}
                  </button>
                )
              })}
            </div>

            {/* Scrollable Content View */}
            <div className="flex-1 p-6 overflow-y-auto" style={{ overscrollBehavior: 'contain' }}>
              {isLoading && (
                <div className="flex flex-col items-center justify-center py-16 gap-3">
                  <div className="w-8 h-8 rounded-full border-4 border-stone-200 border-t-[#E8002D] animate-spin" />
                  <p className="text-sm font-bold text-[#A8A29E]">Fetching game settings...</p>
                </div>
              )}

              {error && (
                <div className="rounded-xl p-4 text-center text-sm font-semibold" style={{ background: '#FFF1F2', border: '1px solid #FECDD3', color: '#BE123C' }}>
                  {error}
                </div>
              )}

              {!isLoading && !error && (
                <AnimatePresence mode="wait">
                  {activeTab === 'content' && (
                    <motion.div
                      key="tab-content"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      transition={{ duration: 0.2 }}
                      className="space-y-5"
                    >
                      <div>
                        <span className="text-[10px] font-extrabold uppercase tracking-widest text-[#A8A29E]">AI Topic</span>
                        <h4 className="text-xl font-black mt-1" style={{ color: '#1C1917', lineHeight: 1.2 }}>
                          {content?.topic || 'Standard Workplace Bingo'}
                        </h4>
                      </div>

                      <div className="rounded-xl p-4 bg-[#FAFAF9] border border-[#F0EDE8]">
                        <span className="text-[10px] font-extrabold uppercase tracking-widest text-[#A8A29E]">Summary</span>
                        <p className="text-sm font-medium mt-1 leading-relaxed" style={{ color: '#44403C' }}>
                          {content?.summary || 'A lighthearted bingo game to celebrate workplace synergy, active habits, and collaboration.'}
                        </p>
                      </div>

                      {content?.callerStyle && (
                        <div className="rounded-xl p-4 bg-[#FFF0F3] border border-[#FFE4D9]">
                          <span className="text-[10px] font-extrabold uppercase tracking-widest text-[#E8002D]">AI Host Personality</span>
                          <p className="text-sm font-bold mt-1 text-[#C40026]">
                            {content.callerStyle}
                          </p>
                        </div>
                      )}
                    </motion.div>
                  )}

                  {activeTab === 'settings' && (
                    <motion.div
                      key="tab-settings"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      transition={{ duration: 0.2 }}
                      className="space-y-4"
                    >
                      <div className="grid grid-cols-2 gap-3">
                        <div className="rounded-xl p-4 border border-[#F0EDE8] bg-white">
                          <span className="text-[10px] font-extrabold uppercase tracking-widest text-[#A8A29E]">Marking Mode</span>
                          <p className="text-sm font-extrabold mt-1.5 capitalize" style={{ color: '#1C1917' }}>
                            {settings?.markingMode === 'auto_mark' ? '🚀 Automatic' : '🎯 Manual'}
                          </p>
                          <p className="text-[10px] font-semibold text-[#A8A29E] mt-1 leading-normal">
                            {settings?.markingMode === 'auto_mark' 
                              ? 'Words marked automatically as they are called.' 
                              : 'You must click cells to mark called words.'}
                          </p>
                        </div>

                        <div className="rounded-xl p-4 border border-[#F0EDE8] bg-white">
                          <span className="text-[10px] font-extrabold uppercase tracking-widest text-[#A8A29E]">Winning Pattern</span>
                          <p className="text-sm font-extrabold mt-1.5 capitalize" style={{ color: '#1C1917' }}>
                            {content?.lockedWordSetId ? 'Single Line' : 'Single Line'}
                          </p>
                          <p className="text-[10px] font-semibold text-[#A8A29E] mt-1 leading-normal">
                            First player to match 5 in a row, column, or diagonal claims BINGO.
                          </p>
                        </div>
                      </div>

                      <div className="rounded-xl p-4 border border-[#F0EDE8] bg-[#FAFAF9] space-y-2">
                        <div className="flex justify-between items-center text-xs font-bold border-b border-[#F0EDE8] pb-2">
                          <span style={{ color: '#78716C' }}>Player marking choice allowed:</span>
                          <span style={{ color: settings?.allowPlayerMarkingModeChoice ? '#22AA6A' : '#F43F5E' }}>
                            {settings?.allowPlayerMarkingModeChoice ? 'Yes' : 'No'}
                          </span>
                        </div>
                        <div className="flex justify-between items-center text-xs font-bold border-b border-[#F0EDE8] py-2">
                          <span style={{ color: '#78716C' }}>Show claim readiness status:</span>
                          <span style={{ color: settings?.showClaimReadiness ? '#22AA6A' : '#F43F5E' }}>
                            {settings?.showClaimReadiness ? 'Yes' : 'No'}
                          </span>
                        </div>
                        <div className="flex justify-between items-center text-xs font-bold pt-2">
                          <span style={{ color: '#78716C' }}>Voice announcements:</span>
                          <span style={{ color: settings?.callerMode === 'text_only' ? '#A8A29E' : '#22AA6A' }}>
                            {settings?.callerMode === 'text_only' ? 'Text Only' : 'Text & Audio'}
                          </span>
                        </div>
                      </div>
                    </motion.div>
                  )}

                  {activeTab === 'words' && (
                    <motion.div
                      key="tab-words"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      transition={{ duration: 0.2 }}
                      className="space-y-4"
                    >
                      <p className="text-[10px] font-extrabold uppercase tracking-widest text-[#A8A29E] mb-2">
                        Active vocabulary set ({content?.words?.length || 0} words)
                      </p>
                      
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-[300px] overflow-y-auto pr-1">
                        {content?.words?.map((word, i) => (
                          <div
                            key={i}
                            className="px-3 py-2 rounded-lg text-xs font-bold text-center border border-[#F0EDE8] bg-[#FAFAF9]"
                            style={{ color: '#57534E' }}
                          >
                            {word}
                          </div>
                        ))}
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              )}
            </div>

            {/* Footer */}
            <div
              className="px-6 py-4 flex items-center justify-between shrink-0"
              style={{ borderTop: '1px solid rgba(244, 242, 239, 0.9)', background: '#FCFAF7' }}
            >
              <div className="flex items-center gap-1.5 text-[10px] font-bold text-[#A8A29E]">
                <Sparkles className="w-3.5 h-3.5 text-[#E8002D] animate-pulse" />
                <span>POWERED BY AI CONDUCTOR</span>
              </div>
              <button
                onClick={onClose}
                className="px-4 py-2 rounded-lg text-xs font-black uppercase tracking-wider text-white"
                style={{
                  background: 'linear-gradient(135deg, #7C5CFC 0%, #6440E8 100%)',
                  boxShadow: '0 4px 12px rgba(100, 64, 232, 0.25)',
                }}
              >
                Got It
              </button>
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
