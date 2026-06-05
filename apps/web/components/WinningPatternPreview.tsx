'use client'

import { useState } from 'react'
import { motion } from 'motion/react'
import {
  BOARD_SIZE,
  WINNING_PATTERN_LIST,
  isCellInPattern,
  resolveWinningPattern,
  type WinningPatternKey,
  type WinningPatternShape,
} from '@/lib/winningPatterns'

interface WinningPatternPreviewProps {
  /** The game's configured winning pattern (backend key). Pre-selected and badged. */
  gamePattern?: string | null
  /** Fired when the host selects a pattern to preview. */
  onSelect?: (key: WinningPatternKey) => void
}

const BINGO_LETTERS = ['B', 'I', 'N', 'G', 'O']
const LETTER_COLORS = ['#FF5A1F', '#7C5CFC', '#22AA6A', '#FBBF24', '#F43F5E']

export function WinningPatternPreview({ gamePattern, onSelect }: WinningPatternPreviewProps) {
  const gameShape = resolveWinningPattern(gamePattern)
  const [selectedKey, setSelectedKey] = useState<WinningPatternKey>(gameShape.key)

  const selected: WinningPatternShape =
    WINNING_PATTERN_LIST.find(p => p.key === selectedKey) ?? gameShape

  const handleSelect = (key: WinningPatternKey) => {
    setSelectedKey(key)
    onSelect?.(key)
  }

  return (
    <div
      className="rounded-xl p-4 sm:p-5"
      style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
    >
      <div className="mb-4 flex items-center justify-between gap-3">
        <div>
          <p className="text-[10px] font-extrabold uppercase tracking-[0.2em]" style={{ color: '#A8A29E' }}>
            Winning Pattern
          </p>
          <h2 className="text-base font-black" style={{ color: '#1C1917' }}>
            Pattern preview
          </h2>
        </div>
      </div>

      {/* Pattern selector chips */}
      <div className="flex flex-wrap gap-2 mb-5">
        {WINNING_PATTERN_LIST.map(pattern => {
          const isSelected = pattern.key === selectedKey
          const isGamePattern = pattern.key === gameShape.key
          return (
            <button
              key={pattern.key}
              onClick={() => handleSelect(pattern.key)}
              aria-pressed={isSelected}
              className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-xs font-extrabold transition-all active:scale-95"
              style={{
                background: isSelected ? '#FFF4F0' : '#FAFAF9',
                color: isSelected ? '#E8440A' : '#78716C',
                border: isSelected ? '1.5px solid #FFC5A8' : '1.5px solid #F0EDE8',
              }}
            >
              {pattern.label}
              {isGamePattern && (
                <span
                  className="rounded-full px-1.5 py-0.5 text-[8px] font-black uppercase tracking-wider"
                  style={{ background: '#EDFAF5', color: '#116B3F' }}
                >
                  In play
                </span>
              )}
            </button>
          )
        })}
      </div>

      <div className="flex flex-col gap-4 lg:flex-row lg:items-start">
        {/* Highlighted board */}
        <div className="w-full max-w-[18rem] mx-auto lg:mx-0">
          {/* B I N G O header */}
          <div className="grid grid-cols-5 gap-1.5 mb-1.5">
            {BINGO_LETTERS.map((letter, i) => (
              <div
                key={letter}
                className="text-center text-sm font-black"
                style={{ color: LETTER_COLORS[i], letterSpacing: '0.05em' }}
              >
                {letter}
              </div>
            ))}
          </div>

          {/* 5x5 grid */}
          <div className="grid grid-cols-5 gap-1.5">
            {Array.from({ length: BOARD_SIZE * BOARD_SIZE }, (_, index) => {
              const rowIndex = Math.floor(index / BOARD_SIZE)
              const colIndex = index % BOARD_SIZE
              const active = isCellInPattern(selected, rowIndex, colIndex)
              return (
                <motion.div
                  key={`${rowIndex}-${colIndex}`}
                  className="aspect-square rounded-md"
                  initial={false}
                  animate={{
                    scale: active ? 1 : 0.94,
                    opacity: active ? 1 : 0.6,
                  }}
                  transition={{ type: 'spring', stiffness: 320, damping: 22 }}
                  style={
                    active
                      ? {
                          background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 60%, #E8440A 100%)',
                          border: '1.5px solid #FF5A1F',
                          boxShadow: '0 4px 14px rgba(255, 90, 31, 0.35)',
                        }
                      : {
                          background: '#FAFAF9',
                          border: '1.5px solid #F0EDE8',
                        }
                  }
                />
              )
            })}
          </div>
        </div>

        {/* Description */}
        <div className="flex-1">
          <p className="text-sm font-black mb-1" style={{ color: '#1C1917' }}>
            {selected.label}
          </p>
          <p className="text-sm font-semibold leading-relaxed" style={{ color: '#78716C' }}>
            {selected.description}
          </p>
          {selected.exampleNote && (
            <p
              className="mt-3 rounded-lg px-3 py-2 text-xs font-bold leading-relaxed"
              style={{ background: '#FFFBEB', color: '#B45309', border: '1.5px solid #FDE68A' }}
            >
              {selected.exampleNote}
            </p>
          )}
          <p className="mt-3 text-xs font-bold" style={{ color: '#A8A29E' }}>
            {selected.cells.length} cell{selected.cells.length === 1 ? '' : 's'} to complete
          </p>
        </div>
      </div>
    </div>
  )
}
