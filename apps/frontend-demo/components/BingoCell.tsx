'use client'

import { motion } from 'motion/react'
import { cn } from '@/lib/utils'

interface BingoCellProps {
  word: string
  isMarked: boolean
  onClick?: () => void
  disabled?: boolean
  isFreeSpace?: boolean
}

export function BingoCell({ word, isMarked, onClick, disabled, isFreeSpace }: BingoCellProps) {
  return (
    <motion.button
      whileHover={!disabled && !isFreeSpace && !isMarked ? { scale: 1.05, y: -2 } : {}}
      whileTap={!disabled && !isFreeSpace && !isMarked ? { scale: 0.93 } : {}}
      animate={
        isMarked && !isFreeSpace
          ? { scale: [1, 1.12, 0.96, 1.04, 1] }
          : { scale: 1 }
      }
      transition={
        isMarked && !isFreeSpace
          ? { duration: 0.5, times: [0, 0.25, 0.5, 0.75, 1], ease: 'easeInOut' }
          : { type: 'spring', stiffness: 350, damping: 22 }
      }
      onClick={onClick}
      disabled={disabled || isFreeSpace}
      aria-pressed={isMarked}
      aria-label={`${word}${isMarked ? ', marked' : ', unmarked'}`}
      className={cn(
        'group aspect-square rounded-2xl flex flex-col items-center justify-center p-2 sm:p-2.5 text-center outline-none relative transition-all duration-200',
        'focus-visible:ring-4 focus-visible:ring-offset-2',
        isFreeSpace
          ? 'cursor-default'
          : isMarked
            ? 'cursor-default'
            : 'cursor-pointer'
      )}
      style={
        isFreeSpace
          ? {
              background: 'linear-gradient(135deg, #EDE5FF 0%, #D9CCFF 100%)',
              border: '2px solid #BFAAFF',
              boxShadow: '0 2px 10px rgba(124, 92, 252, 0.12)',
            }
          : isMarked
            ? {
                background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
                border: '2px solid #FF5A1F',
                boxShadow: '0 6px 20px rgba(255, 90, 31, 0.35)',
              }
            : {
                background: '#FFFFFF',
                border: '2px solid #F0EDE8',
                boxShadow: '0 2px 8px rgba(0,0,0,0.04)',
              }
      }
    >
      {/* Cell text */}
      <span
        className={cn(
          'break-words w-full leading-tight z-10 transition-colors duration-200',
          isFreeSpace
            ? 'text-[8px] sm:text-[10px] font-black uppercase tracking-wider'
            : isMarked
              ? 'text-[9px] sm:text-[11px] font-extrabold'
              : 'text-[9px] sm:text-[11px] font-bold'
        )}
        style={{
          color: isFreeSpace ? '#6440E8' : isMarked ? '#FFFFFF' : '#44403C',
        }}
      >
        {word}
      </span>

      {/* Check mark on marked cells */}
      {isMarked && !isFreeSpace && (
        <motion.span
          initial={{ scale: 0, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ type: 'spring', stiffness: 500, damping: 22, delay: 0.08 }}
          className="absolute top-1 right-1 sm:top-1.5 sm:right-1.5 w-4 h-4 sm:w-5 sm:h-5 rounded-full flex items-center justify-center text-[8px] sm:text-[10px] font-black z-10"
          style={{ background: 'rgba(255,255,255,0.25)', color: '#FFFFFF' }}
        >
          ✓
        </motion.span>
      )}

      {/* Hover glow for unmarked cells */}
      {!isMarked && !isFreeSpace && !disabled && (
        <span
          className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none"
          style={{ background: 'linear-gradient(135deg, rgba(255,164,112,0.08), rgba(255,90,31,0.06))' }}
        />
      )}
    </motion.button>
  )
}
