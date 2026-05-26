'use client'

import { motion } from 'motion/react'

interface BingoCharacterProps {
  mood?: 'happy' | 'excited' | 'waiting' | 'celebrating' | 'thinking'
  size?: number
  className?: string
}

/**
 * Headspace-inspired abstract character - organic rounded shapes,
 * warm gradients, friendly face, no sharp edges.
 */
export function BingoCharacter({ mood = 'happy', size = 120, className }: BingoCharacterProps) {
  const scale = size / 120

  return (
    <div className={className} style={{ width: size, height: size }}>
      <svg
        viewBox="0 0 120 120"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style={{ width: '100%', height: '100%' }}
      >
        {/* Body - warm orange organic blob */}
        <motion.ellipse
          cx="60"
          cy="68"
          rx="38"
          ry="40"
          fill="url(#bodyGrad)"
          animate={
            mood === 'celebrating'
              ? { ry: [40, 38, 42, 40], rx: [38, 40, 36, 38] }
              : mood === 'waiting'
                ? { ry: [40, 41, 40], rx: [38, 37, 38] }
                : {}
          }
          transition={{ duration: mood === 'celebrating' ? 0.8 : 3, repeat: Infinity, ease: 'easeInOut' }}
        />

        {/* Head - slightly offset organic circle */}
        <motion.circle
          cx="60"
          cy="35"
          r="26"
          fill="url(#headGrad)"
          animate={
            mood === 'excited'
              ? { cy: [35, 32, 35] }
              : mood === 'celebrating'
                ? { cy: [35, 30, 35], r: [26, 27, 26] }
                : mood === 'thinking'
                  ? { cx: [60, 62, 60] }
                  : {}
          }
          transition={{ duration: mood === 'celebrating' ? 1 : 2.5, repeat: Infinity, ease: 'easeInOut' }}
        />

        {/* Eyes */}
        <motion.circle
          cx="50"
          cy="33"
          r={mood === 'excited' || mood === 'celebrating' ? 3 : 2.5}
          fill="#FFFFFF"
          animate={mood === 'celebrating' ? { r: [3, 2, 3] } : {}}
          transition={{ duration: 0.4, repeat: Infinity, repeatDelay: 2 }}
        />
        <motion.circle
          cx="70"
          cy="33"
          r={mood === 'excited' || mood === 'celebrating' ? 3 : 2.5}
          fill="#FFFFFF"
          animate={mood === 'celebrating' ? { r: [3, 2, 3] } : {}}
          transition={{ duration: 0.4, delay: 0.05, repeat: Infinity, repeatDelay: 2 }}
        />

        {/* Smile / expression */}
        {(mood === 'happy' || mood === 'excited' || mood === 'celebrating') && (
          <motion.path
            d="M52 40 Q60 48 68 40"
            stroke="#FFFFFF"
            strokeWidth="2.5"
            strokeLinecap="round"
            fill="none"
            animate={
              mood === 'celebrating'
                ? { d: ['M52 40 Q60 48 68 40', 'M50 39 Q60 50 70 39', 'M52 40 Q60 48 68 40'] }
                : {}
            }
            transition={{ duration: 1.2, repeat: Infinity, ease: 'easeInOut' }}
          />
        )}
        {mood === 'waiting' && (
          <circle cx="60" cy="42" r="3" fill="#FFFFFF" opacity="0.8" />
        )}
        {mood === 'thinking' && (
          <path d="M54 41 L66 41" stroke="#FFFFFF" strokeWidth="2.5" strokeLinecap="round" />
        )}

        {/* Arms for celebrating */}
        {mood === 'celebrating' && (
          <>
            <motion.path
              d="M28 65 Q20 50 15 40"
              stroke="url(#bodyGrad)"
              strokeWidth="8"
              strokeLinecap="round"
              fill="none"
              animate={{ d: ['M28 65 Q20 50 15 40', 'M28 65 Q18 45 12 32', 'M28 65 Q20 50 15 40'] }}
              transition={{ duration: 0.8, repeat: Infinity, ease: 'easeInOut' }}
            />
            <motion.path
              d="M92 65 Q100 50 105 40"
              stroke="url(#bodyGrad)"
              strokeWidth="8"
              strokeLinecap="round"
              fill="none"
              animate={{ d: ['M92 65 Q100 50 105 40', 'M92 65 Q102 45 108 32', 'M92 65 Q100 50 105 40'] }}
              transition={{ duration: 0.8, repeat: Infinity, ease: 'easeInOut', delay: 0.1 }}
            />
          </>
        )}

        {/* Tiny bingo card in hand for happy/excited */}
        {(mood === 'happy' || mood === 'excited') && (
          <g transform="translate(82, 55) rotate(-12)">
            <rect x="0" y="0" width="18" height="22" rx="3" fill="#FFFFFF" opacity="0.9" />
            {/* Mini grid */}
            {[0, 1, 2].map(row =>
              [0, 1, 2].map(col => (
                <rect
                  key={`${row}-${col}`}
                  x={2 + col * 5}
                  y={5 + row * 5}
                  width="4"
                  height="4"
                  rx="1"
                  fill={
                    (row === 1 && col === 1) ? '#7C5CFC'
                    : (row + col) % 2 === 0 ? '#FF5A1F'
                    : '#FFE4D9'
                  }
                  opacity="0.8"
                />
              ))
            )}
          </g>
        )}

        {/* Sparkles for excited/celebrating */}
        {(mood === 'excited' || mood === 'celebrating') && (
          <>
            <motion.circle
              cx="20"
              cy="25"
              r="2.5"
              fill="#FBBF24"
              animate={{ scale: [0, 1.2, 0], opacity: [0, 1, 0] }}
              transition={{ duration: 1.5, repeat: Infinity, delay: 0 }}
            />
            <motion.circle
              cx="100"
              cy="20"
              r="2"
              fill="#FF7A42"
              animate={{ scale: [0, 1.2, 0], opacity: [0, 1, 0] }}
              transition={{ duration: 1.5, repeat: Infinity, delay: 0.5 }}
            />
            <motion.circle
              cx="15"
              cy="55"
              r="1.5"
              fill="#7C5CFC"
              animate={{ scale: [0, 1.2, 0], opacity: [0, 1, 0] }}
              transition={{ duration: 1.5, repeat: Infinity, delay: 1 }}
            />
            <motion.path
              d="M98 48 L100 44 L102 48 L98 48"
              fill="#22AA6A"
              animate={{ scale: [0, 1.2, 0], opacity: [0, 1, 0], rotate: [0, 180, 360] }}
              transition={{ duration: 2, repeat: Infinity, delay: 0.3 }}
              style={{ transformOrigin: '100px 46px' }}
            />
          </>
        )}

        {/* Thinking bubble */}
        {mood === 'thinking' && (
          <>
            <circle cx="85" cy="18" r="8" fill="#F5F2FF" opacity="0.9" />
            <circle cx="78" cy="26" r="4" fill="#F5F2FF" opacity="0.7" />
            <circle cx="75" cy="32" r="2" fill="#F5F2FF" opacity="0.5" />
            <text x="81" y="21" fontSize="8" fill="#7C5CFC" fontWeight="bold" textAnchor="middle">?</text>
          </>
        )}

        <defs>
          <linearGradient id="bodyGrad" x1="30" y1="40" x2="90" y2="110">
            <stop offset="0%" stopColor="#FF7A42" />
            <stop offset="100%" stopColor="#FF5A1F" />
          </linearGradient>
          <linearGradient id="headGrad" x1="40" y1="15" x2="80" y2="55">
            <stop offset="0%" stopColor="#FFA070" />
            <stop offset="100%" stopColor="#FF7A42" />
          </linearGradient>
        </defs>
      </svg>
    </div>
  )
}
