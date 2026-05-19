'use client'

import { BingoClaim } from '@/types/game'
import { Check, X, AlertCircle } from 'lucide-react'
import { motion, AnimatePresence } from 'motion/react'

interface BingoClaimQueueProps {
  claims: BingoClaim[]
  onApprove: (id: string) => void
  onReject: (id: string) => void
}

const STATUS_CHIP: Record<string, { bg: string; text: string; label: string }> = {
  Pending:   { bg: '#FEF3C7', text: '#D97706', label: 'Pending Review' },
  Valid:     { bg: '#EDFAF5', text: '#116B3F', label: 'Valid' },
  Invalid:   { bg: '#FFF1F2', text: '#E11D48', label: 'Invalid' },
  Confirmed: { bg: '#EDFAF5', text: '#116B3F', label: 'Confirmed' },
}

function getInitials(name: string) {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

export function BingoClaimQueue({ claims, onApprove, onReject }: BingoClaimQueueProps) {
  const pendingClaims = claims.filter(c => c.status === 'Pending')
  const otherClaims = claims.filter(c => c.status !== 'Pending')

  return (
    <div
      className="h-full rounded-3xl flex flex-col overflow-hidden"
      style={{
        background: '#FFFFFF',
        border: '1.5px solid #F0EDE8',
        boxShadow: '0 2px 16px rgba(0,0,0,0.04)',
      }}
    >
      {/* Header */}
      <div className="px-5 pt-5 pb-4 flex items-center gap-3 shrink-0" style={{ borderBottom: '1px solid #F4F2EF' }}>
        <div
          className="w-8 h-8 rounded-xl flex items-center justify-center"
          style={{ background: '#FEF3C7', color: '#D97706' }}
        >
          <AlertCircle className="w-4 h-4" />
        </div>
        <div>
          <h3 className="text-sm font-extrabold" style={{ color: '#1C1917' }}>Claim Queue</h3>
          <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>
            {pendingClaims.length} pending review
          </p>
        </div>
        {pendingClaims.length > 0 && (
          <div
            className="ml-auto w-6 h-6 rounded-full flex items-center justify-center text-xs font-black"
            style={{ background: '#FF5A1F', color: '#FFFFFF' }}
          >
            {pendingClaims.length}
          </div>
        )}
      </div>

      {/* Claim list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        <AnimatePresence initial={false}>
          {claims.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-40 gap-3">
              <div className="w-12 h-12 rounded-2xl flex items-center justify-center" style={{ background: '#F4F2EF' }}>
                <Check className="w-6 h-6" style={{ color: '#D6D3D1' }} />
              </div>
              <p className="text-sm font-semibold" style={{ color: '#D6D3D1' }}>No claims yet</p>
            </div>
          ) : (
            claims.map((claim) => {
              const chip = STATUS_CHIP[claim.status] ?? STATUS_CHIP['Pending']
              return (
                <motion.div
                  key={claim.id}
                  layout
                  initial={{ opacity: 0, y: 12 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ type: 'spring', stiffness: 300, damping: 22 }}
                  className="rounded-2xl p-4"
                  style={{
                    background: claim.status === 'Pending' ? '#FFFBF0' : '#FAFAF9',
                    border: `1.5px solid ${claim.status === 'Pending' ? '#FCD34D' : '#F0EDE8'}`,
                  }}
                >
                  <div className="flex items-start gap-3 mb-3">
                    {/* Avatar */}
                    <div
                      className="w-9 h-9 rounded-xl flex items-center justify-center text-sm font-black shrink-0"
                      style={{ background: '#FFF4F0', color: '#E8440A', border: '1.5px solid #FFE4D9' }}
                    >
                      {getInitials(claim.playerName)}
                    </div>

                    {/* Info */}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-bold truncate" style={{ color: '#1C1917' }}>{claim.playerName}</p>
                      <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Pattern: {claim.pattern}</p>
                    </div>

                    {/* Status chip */}
                    <span
                      className="shrink-0 text-[10px] font-extrabold px-2.5 py-1 rounded-full uppercase tracking-wide"
                      style={{ background: chip.bg, color: chip.text }}
                    >
                      {chip.label}
                    </span>
                  </div>

                  {/* Action buttons (Pending only) */}
                  {claim.status === 'Pending' && (
                    <div className="flex gap-2 mt-2">
                      <button
                        onClick={() => onApprove(claim.id)}
                        className="flex-1 flex items-center justify-center gap-1.5 py-2.5 rounded-xl text-xs font-extrabold transition-all hover:opacity-90 active:scale-95"
                        style={{ background: 'linear-gradient(135deg, #3DC484, #22AA6A)', color: '#FFFFFF', boxShadow: '0 4px 12px rgba(34, 170, 106, 0.25)' }}
                      >
                        <Check className="w-3.5 h-3.5" />
                        Approve
                      </button>
                      <button
                        onClick={() => onReject(claim.id)}
                        className="flex-1 flex items-center justify-center gap-1.5 py-2.5 rounded-xl text-xs font-extrabold transition-all hover:opacity-90 active:scale-95"
                        style={{ background: '#FFF1F2', color: '#E11D48', border: '1.5px solid #FECDD3' }}
                      >
                        <X className="w-3.5 h-3.5" />
                        Reject
                      </button>
                    </div>
                  )}
                </motion.div>
              )
            })
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}
