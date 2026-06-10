'use client'

import { motion } from 'motion/react'
import { BingoCellData } from '@/types/player'
import { BingoCell } from './BingoCell'

interface BingoCardProps {
  cells: BingoCellData[]
  onCellClick?: (id: string) => void
  disabled?: boolean
  currentWord?: string
}

const BINGO_LETTERS = ['B', 'I', 'N', 'G', 'O']
const LETTER_COLORS = ['#E8002D', '#7C5CFC', '#22AA6A', '#FBBF24', '#F43F5E']

export function BingoCard({ cells, onCellClick, disabled, currentWord }: BingoCardProps) {
  const markedCount = cells.filter(c => c.isMarked).length
  const totalCells = cells.length
  const progress = totalCells > 0 ? (markedCount / totalCells) * 100 : 0

  return (
    <div className="flex flex-col items-center justify-center w-full">
      {/* Card wrapper with warm shadow */}
      <div
        className="w-full max-w-[min(42rem,calc(100vw-2rem))] lg:max-w-[min(36rem,calc(100vw-28rem))] rounded-2xl p-2 sm:p-4 xl:p-5 relative overflow-hidden"
        style={{
          background: '#FFFFFF',
          border: '2px solid #F0EDE8',
          boxShadow: '0 8px 40px rgba(232, 0, 45, 0.08), 0 2px 12px rgba(0,0,0,0.04)',
        }}
      >
        {/* Decorative corner blobs */}
        <div
          className="absolute top-0 right-0 w-32 h-32 rounded-full pointer-events-none"
          style={{ background: 'radial-gradient(circle, rgba(124,92,252,0.06) 0%, transparent 70%)', transform: 'translate(30%, -30%)' }}
        />
        <div
          className="absolute bottom-0 left-0 w-28 h-28 rounded-full pointer-events-none"
          style={{ background: 'radial-gradient(circle, rgba(232,0,45,0.06) 0%, transparent 70%)', transform: 'translate(-30%, 30%)' }}
        />

        {/* B I N G O header letters with individual colors */}
        <div className="grid grid-cols-5 gap-1.5 sm:gap-2 lg:gap-2.5 mb-2 sm:mb-3">
          {BINGO_LETTERS.map((letter, i) => (
            <motion.div
              key={letter}
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.08, type: 'spring', stiffness: 300 }}
              className="text-center text-lg sm:text-2xl font-black pb-1 relative"
              style={{ color: LETTER_COLORS[i], letterSpacing: '0.05em' }}
            >
              {letter}
              {/* Small decorative dot under each letter */}
              <motion.div
                className="absolute bottom-0 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full"
                style={{ background: LETTER_COLORS[i], opacity: 0.4 }}
                animate={{ scale: [1, 1.5, 1] }}
                transition={{ duration: 2, repeat: Infinity, delay: i * 0.3 }}
              />
            </motion.div>
          ))}
        </div>

        {/* Grid */}
        <div className="grid grid-cols-5 gap-1.5 sm:gap-2 lg:gap-2.5">
          {cells.map((cell, i) => (
            <motion.div
              key={cell.id}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.02, type: 'spring', stiffness: 300, damping: 20 }}
              className="w-full"
            >
              <BingoCell
                word={cell.word}
                isMarked={cell.isMarked}
                isFreeSpace={cell.isFreeSpace}
                disabled={disabled}
                isCurrentWord={currentWord ? cell.word.toLowerCase() === currentWord.toLowerCase() : false}
                onClick={() => onCellClick?.(cell.id)}
              />
            </motion.div>
          ))}
        </div>

        {/* Progress bar */}
        <div className="mt-3 sm:mt-4">
          <div className="flex items-center justify-between mb-1.5">
            <span className="text-[10px] font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
              Progress
            </span>
            <span className="text-[10px] font-black" style={{ color: markedCount >= 5 ? '#E8002D' : '#A8A29E' }}>
              {markedCount} / {totalCells}
            </span>
          </div>
          <div
            className="h-2 rounded-full overflow-hidden"
            style={{ background: '#F4F2EF' }}
          >
            <motion.div
              className="h-full rounded-full"
              initial={{ width: 0 }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 0.5, ease: 'easeOut' }}
              style={{
                background: progress > 60
                  ? 'linear-gradient(90deg, #C0003D, #E8002D, #C40026)'
                  : progress > 30
                    ? 'linear-gradient(90deg, #FFB0C0, #C0003D)'
                    : 'linear-gradient(90deg, #FFE4D9, #FFB0C0)',
                boxShadow: progress > 60 ? '0 0 12px rgba(232, 0, 45, 0.4)' : 'none',
              }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
