'use client'

import { motion, AnimatePresence } from 'motion/react'
import { Player } from '@/types/player'
import { BingoPattern } from '@/types/game'
import { X, Trophy, Medal } from 'lucide-react'
import { ConfettiCelebration } from './effects/ConfettiCelebration'

interface WinnerModalProps {
  isOpen: boolean
  winner: Player | null
  placement: 1 | 2 | 3
  pattern: BingoPattern | string
  onClose: () => void
}

const PLACEMENT_CONFIG = {
  1: {
    emoji: <Trophy size={48} color="#F59E0B" />,
    label: '1st Place',
    gradient: 'linear-gradient(135deg, #FBBF24 0%, #F59E0B 100%)',
    glow: 'rgba(245, 158, 11, 0.30)',
    accent: '#B45309',
    bg: '#FFFBEB',
    chipBg: '#FEF3C7',
    chipText: '#B45309',
    ring: '#FBBF24',
  },
  2: {
    emoji: <Medal size={48} color="#A8A29E" />,
    label: '2nd Place',
    gradient: 'linear-gradient(135deg, #D6D3D1 0%, #A8A29E 100%)',
    glow: 'rgba(168, 162, 158, 0.25)',
    accent: '#57534E',
    bg: '#FAFAF9',
    chipBg: '#F4F2EF',
    chipText: '#57534E',
    ring: '#D6D3D1',
  },
  3: {
    emoji: <Medal size={48} color="#FF7A42" />,
    label: '3rd Place',
    gradient: 'linear-gradient(135deg, #FFA070 0%, #FF7A42 100%)',
    glow: 'rgba(255, 90, 31, 0.25)',
    accent: '#C23208',
    bg: '#FFF4F0',
    chipBg: '#FFE4D9',
    chipText: '#C23208',
    ring: '#FFA070',
  },
}

export function WinnerModal({ isOpen, winner, placement, pattern, onClose }: WinnerModalProps) {
  const config = PLACEMENT_CONFIG[placement] ?? PLACEMENT_CONFIG[1]

  return (
    <>
      <ConfettiCelebration trigger={isOpen && placement === 1} intensity="big" />
      <ConfettiCelebration trigger={isOpen && placement !== 1} intensity="small" />

      <AnimatePresence>
        {isOpen && winner && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4"
            style={{ background: 'rgba(28, 25, 23, 0.65)', backdropFilter: 'blur(12px)' }}
            onClick={onClose}
          >
            <motion.div
              initial={{ opacity: 0, scale: 0.7, y: 40, rotate: -3 }}
              animate={{ opacity: 1, scale: 1, y: 0, rotate: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              transition={{ type: 'spring', stiffness: 260, damping: 20 }}
              className="w-full max-w-[440px] rounded-xl overflow-hidden relative"
              style={{
                background: '#FFFFFF',
                boxShadow: `0 24px 60px ${config.glow}, 0 8px 24px rgba(0,0,0,0.12)`,
              }}
              onClick={e => e.stopPropagation()}
            >
              {/* Close button */}
              <button
                onClick={onClose}
                className="absolute top-4 right-4 w-8 h-8 rounded-md flex items-center justify-center z-10 transition-all hover:bg-stone-200"
                style={{ background: '#F4F2EF', color: '#A8A29E' }}
                aria-label="Close"
              >
                <X className="w-4 h-4" />
              </button>

              {/* Colored top section */}
              <div
                className="px-8 pt-10 pb-8 flex flex-col items-center text-center relative overflow-hidden"
                style={{ background: config.bg }}
              >
                {/* Animated rings behind trophy */}
                <motion.div
                  className="absolute rounded-full"
                  style={{
                    width: 200, height: 200,
                    border: `3px solid ${config.ring}`,
                    opacity: 0.15,
                  }}
                  animate={{ scale: [0.5, 1.5], opacity: [0.3, 0] }}
                  transition={{ duration: 2, repeat: Infinity, ease: 'easeOut' }}
                />
                <motion.div
                  className="absolute rounded-full"
                  style={{
                    width: 200, height: 200,
                    border: `3px solid ${config.ring}`,
                    opacity: 0.15,
                  }}
                  animate={{ scale: [0.5, 1.5], opacity: [0.3, 0] }}
                  transition={{ duration: 2, repeat: Infinity, ease: 'easeOut', delay: 0.7 }}
                />

                {/* Floating sparkles */}
                {[
                  { x: -60, y: -20, size: 6, delay: 0 },
                  { x: 70, y: -30, size: 5, delay: 0.5 },
                  { x: -40, y: 30, size: 4, delay: 1 },
                  { x: 55, y: 25, size: 5, delay: 0.3 },
                  { x: -15, y: -40, size: 3, delay: 0.8 },
                  { x: 30, y: -35, size: 4, delay: 1.2 },
                ].map((s, i) => (
                  <motion.div
                    key={i}
                    className="absolute"
                    style={{ left: '50%', top: '50%' }}
                    animate={{
                      x: [s.x * 0.5, s.x, s.x * 0.5],
                      y: [s.y * 0.5, s.y, s.y * 0.5],
                      scale: [0, 1, 0],
                      opacity: [0, 1, 0],
                    }}
                    transition={{ duration: 2, repeat: Infinity, delay: s.delay, ease: 'easeInOut' }}
                  >
                    <svg width={s.size * 3} height={s.size * 3} viewBox="0 0 12 12" fill="none">
                      <path d="M6 0L7.5 4.5L12 6L7.5 7.5L6 12L4.5 7.5L0 6L4.5 4.5L6 0Z" fill={config.ring} opacity="0.7" />
                    </svg>
                  </motion.div>
                ))}

                {/* Trophy / medal */}
                <motion.div
                  initial={{ scale: 0, rotate: -20 }}
                  animate={{ scale: 1, rotate: 0, y: [0, -8, 0] }}
                  transition={{
                    scale: { type: 'spring', stiffness: 300, damping: 15, delay: 0.2 },
                    y: { duration: 3, repeat: Infinity, ease: 'easeInOut' },
                  }}
                >
                  <div className="flex justify-center mb-4 relative z-10">{config.emoji}</div>
                </motion.div>

                {/* BINGO CONFIRMED label */}
                <motion.p
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.3 }}
                  className="text-xs font-extrabold uppercase tracking-[0.2em] mb-2"
                  style={{ color: config.accent }}
                >
                  Bingo Confirmed
                </motion.p>

                {/* Winner name */}
                <motion.h2
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.4 }}
                  className="text-3xl font-black mb-3 leading-tight"
                  style={{ color: '#1C1917' }}
                >
                  {winner.name}
                </motion.h2>

                {/* Placement chip */}
                <motion.div
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.5, type: 'spring' }}
                  className="px-5 py-1.5 rounded-full text-sm font-extrabold uppercase tracking-wider"
                  style={{ background: config.chipBg, color: config.chipText }}
                >
                  {config.label}
                </motion.div>
              </div>

              {/* Details */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.5 }}
                className="px-6 py-6"
              >
                <div
                  className="flex items-center justify-between px-4 py-3.5 rounded-lg mb-6"
                  style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                >
                  <div>
                    <p className="text-[10px] font-extrabold uppercase tracking-widest mb-0.5" style={{ color: '#A8A29E' }}>
                      Winning Pattern
                    </p>
                    <p className="text-sm font-bold" style={{ color: '#44403C' }}>{pattern}</p>
                  </div>
                  {/* Mini bingo icon with animated highlight */}
                  <div
                    className="w-10 h-10 rounded-md p-1.5 grid grid-cols-3 gap-0.5"
                    style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}
                  >
                    {[1,2,3,4,5,6,7,8,9].map(i => (
                      <motion.div
                        key={i}
                        className="rounded-sm"
                        initial={{ background: '#F0EDE8' }}
                        animate={{
                          background: i >= 4 && i <= 6
                            ? ['#FF5A1F', '#FF7A42', '#FF5A1F']
                            : '#F0EDE8',
                        }}
                        transition={i >= 4 && i <= 6 ? { duration: 1.5, repeat: Infinity } : {}}
                      />
                    ))}
                  </div>
                </div>

                <motion.button
                  whileTap={{ scale: 0.97 }}
                  whileHover={{ y: -2 }}
                  onClick={onClose}
                  className="w-full py-4 rounded-lg font-extrabold text-base transition-all"
                  style={{
                    background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
                    color: '#FFFFFF',
                    boxShadow: '0 6px 20px rgba(255, 90, 31, 0.35)',
                  }}
                >
                  Continue Game
                </motion.button>

                <p className="text-center text-[10px] font-semibold mt-4" style={{ color: '#D6D3D1' }}>
                  Validation complete · Activity logged
                </p>
              </motion.div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
