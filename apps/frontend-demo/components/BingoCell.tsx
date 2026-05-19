'use client'

import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { cn } from '@/lib/utils'
import { CellMarkEffect } from './effects/CellMarkEffect'

interface BingoCellProps {
  word: string
  isMarked: boolean
  onClick?: () => void
  disabled?: boolean
  isFreeSpace?: boolean
  /** Whether this cell's word was recently called */
  isCurrentWord?: boolean
}

export function BingoCell({ word, isMarked, onClick, disabled, isFreeSpace, isCurrentWord }: BingoCellProps) {
  const [justMarked, setJustMarked] = useState(false)
  const prevMarked = useRef(isMarked)

  useEffect(() => {
    if (isMarked && !prevMarked.current) {
      setJustMarked(true)
      const t = setTimeout(() => setJustMarked(false), 600)
      return () => clearTimeout(t)
    }
    prevMarked.current = isMarked
  }, [isMarked])

  return (
    <motion.button
      whileHover={!disabled && !isFreeSpace && !isMarked ? { scale: 1.07, y: -3 } : {}}
      whileTap={!disabled && !isFreeSpace && !isMarked ? { scale: 0.90 } : {}}
      animate={
        isMarked && !isFreeSpace
          ? { scale: [1, 1.15, 0.94, 1.06, 1] }
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
        'group aspect-square rounded-lg flex flex-col items-center justify-center p-2 sm:p-2.5 text-center outline-none relative transition-all duration-200',
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
              boxShadow: '0 2px 10px rgba(124, 92, 252, 0.15)',
            }
          : isMarked
            ? {
                background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 60%, #E8440A 100%)',
                border: '2px solid #FF5A1F',
                boxShadow: '0 6px 20px rgba(255, 90, 31, 0.40), inset 0 1px 0 rgba(255,255,255,0.15)',
              }
            : isCurrentWord
              ? {
                  background: 'linear-gradient(135deg, #FFF4F0 0%, #FFE4D9 100%)',
                  border: '2px solid #FFC5A8',
                  boxShadow: '0 2px 12px rgba(255, 90, 31, 0.15)',
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
          color: isFreeSpace ? '#6440E8' : isMarked ? '#FFFFFF' : isCurrentWord ? '#E8440A' : '#44403C',
        }}
      >
        {word}
      </span>

      {/* Animated check/star on marked cells */}
      <AnimatePresence>
        {isMarked && !isFreeSpace && (
          <motion.span
            initial={{ scale: 0, opacity: 0, rotate: -180 }}
            animate={{ scale: 1, opacity: 1, rotate: 0 }}
            exit={{ scale: 0, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 500, damping: 18, delay: 0.05 }}
            className="absolute top-0.5 right-0.5 sm:top-1 sm:right-1 z-10 flex items-center justify-center"
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <circle cx="10" cy="10" r="9" fill="rgba(255,255,255,0.25)" />
              <path
                d="M6 10.5L9 13.5L14.5 7"
                stroke="white"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </motion.span>
        )}
      </AnimatePresence>

      {/* Free space sparkle icon */}
      {isFreeSpace && (
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 8, repeat: Infinity, ease: 'linear' }}
          className="absolute top-1 right-1 z-10"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M7 0L8.5 5.5L14 7L8.5 8.5L7 14L5.5 8.5L0 7L5.5 5.5L7 0Z" fill="#BFAAFF" opacity="0.6" />
          </svg>
        </motion.div>
      )}

      {/* Current word highlight pulse */}
      {isCurrentWord && !isMarked && (
        <motion.span
          className="absolute inset-0 rounded-lg pointer-events-none z-0"
          animate={{ opacity: [0.3, 0.6, 0.3] }}
          transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
          style={{
            background: 'linear-gradient(135deg, rgba(255,164,112,0.12), rgba(255,90,31,0.08))',
          }}
        />
      )}

      {/* Hover glow for unmarked cells */}
      {!isMarked && !isFreeSpace && !disabled && (
        <span
          className="absolute inset-0 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none z-0"
          style={{ background: 'linear-gradient(135deg, rgba(255,164,112,0.12), rgba(255,90,31,0.10))' }}
        />
      )}

      {/* Particle burst effect on mark */}
      <CellMarkEffect trigger={justMarked} />
    </motion.button>
  )
}
