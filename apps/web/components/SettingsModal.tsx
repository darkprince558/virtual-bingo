'use client'

import React from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { X, Smile } from 'lucide-react'
import { useSettings, AVATARS } from '@/contexts/SettingsContext'

interface SettingsModalProps {
  isOpen: boolean
  onClose: () => void
}

export function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  const { avatar, setAvatar } = useSettings()

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="absolute inset-0"
            style={{ background: 'rgba(28, 25, 23, 0.50)', backdropFilter: 'blur(8px)' }}
          />

          {/* Modal */}
          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 16 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.92, y: 10 }}
            transition={{ type: 'spring', stiffness: 300, damping: 24 }}
            className="relative w-full max-w-sm overflow-hidden flex flex-col"
            style={{
              background: '#FFFFFF',
              border: '1.5px solid #F0EDE8',
              borderRadius: '1.5rem',
              boxShadow: '0 24px 60px rgba(0,0,0,0.14)',
              maxHeight: '80vh',
            }}
          >
            {/* Header */}
            <div
              className="px-6 py-5 flex items-center justify-between shrink-0"
              style={{ borderBottom: '1px solid #F4F2EF' }}
            >
              <div className="flex items-center gap-2.5">
                <div
                  className="w-8 h-8 rounded-md flex items-center justify-center"
                  style={{ background: '#FFF0F3' }}
                >
                  <Smile className="w-4 h-4" style={{ color: '#E8002D' }} />
                </div>
                <h3 className="font-extrabold" style={{ color: '#1C1917' }}>Your Avatar</h3>
              </div>
              <button
                onClick={onClose}
                aria-label="Close settings"
                className="w-8 h-8 flex items-center justify-center rounded-md transition-all"
                style={{ background: '#F4F2EF', color: '#A8A29E' }}
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Avatar grid */}
            <div className="p-6 overflow-y-auto flex-1">
              <p className="text-xs font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>
                Choose your player avatar
              </p>
              <div className="grid grid-cols-5 gap-2">
                {AVATARS.map((a) => {
                  const isSelected = avatar === a
                  return (
                    <motion.button
                      key={a}
                      whileTap={{ scale: 0.9 }}
                      whileHover={{ scale: 1.1 }}
                      onClick={() => { setAvatar(a); onClose(); }}
                      className="aspect-square rounded-lg text-2xl flex items-center justify-center transition-all"
                      style={{
                        background: isSelected ? '#FFF0F3' : '#FAFAF9',
                        border: isSelected ? '2px solid #E8002D' : '2px solid transparent',
                        boxShadow: isSelected ? '0 4px 12px rgba(232,0,45,0.20)' : 'none',
                        transform: isSelected ? 'scale(1.08)' : 'scale(1)',
                      }}
                      aria-label={`Select avatar ${a}`}
                      aria-pressed={isSelected}
                    >
                      {a}
                    </motion.button>
                  )
                })}
              </div>
            </div>

            {/* Current selection preview */}
            <div
              className="px-6 py-4 flex items-center gap-3 shrink-0"
              style={{ borderTop: '1px solid #F4F2EF', background: '#FAFAF9' }}
            >
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center text-2xl"
                style={{ background: '#FFF0F3', border: '2px solid #FFE4D9' }}
              >
                {avatar}
              </div>
              <div>
                <p className="text-sm font-bold" style={{ color: '#1C1917' }}>Your avatar</p>
                <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Visible to all players in lobby</p>
              </div>
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
