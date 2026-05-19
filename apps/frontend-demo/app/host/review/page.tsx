'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { useRouter } from 'next/navigation'
import { DashboardShell } from '@/components/DashboardShell'
import { mockWordSet } from '@/lib/mockAdminData'
import { ChevronLeft, Sparkles, Check, X, RefreshCw, AlertTriangle, Edit3, CheckCircle2 } from 'lucide-react'

export default function ContentReviewPage() {
  const router = useRouter()
  const [words, setWords] = useState(mockWordSet.words)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [approved, setApproved] = useState(false)

  const toggleWord = (id: string) => {
    setWords(prev => prev.map(w => w.id === id ? { ...w, approved: !w.approved } : w))
  }
  const startEdit = (id: string, word: string) => { setEditingId(id); setEditValue(word) }
  const saveEdit = (id: string) => {
    setWords(prev => prev.map(w => w.id === id ? { ...w, word: editValue } : w))
    setEditingId(null)
  }
  const approvedCount = words.filter(w => w.approved).length
  const flaggedCount = words.filter(w => w.flagged).length

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
        <motion.button initial={{ opacity: 0 }} animate={{ opacity: 1 }} onClick={() => router.push('/host')} className="flex items-center gap-1.5 text-sm font-bold mb-6" style={{ color: '#A8A29E' }}>
          <ChevronLeft className="w-4 h-4" /> Back to Dashboard
        </motion.button>

        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-2xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #FBBF24, #F59E0B)', boxShadow: '0 4px 12px rgba(245,158,11,0.25)' }}>
              <Sparkles className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917' }}>AI Content Review</h1>
              <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>{mockWordSet.templateName} · Generated {new Date(mockWordSet.generatedAt).toLocaleTimeString()}</p>
            </div>
          </div>
        </motion.div>

        {/* Prompt card */}
        <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="rounded-2xl p-4 mb-6" style={{ background: '#F5F2FF', border: '1.5px solid #D9CCFF' }}>
          <p className="text-[10px] font-extrabold uppercase tracking-wider mb-1" style={{ color: '#6440E8' }}>Generation Prompt</p>
          <p className="text-sm font-semibold" style={{ color: '#4F30C2' }}>{mockWordSet.prompt}</p>
        </motion.div>

        {/* Stats row */}
        <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="grid grid-cols-3 gap-3 mb-6">
          <div className="rounded-2xl p-3 text-center" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
            <p className="text-xl font-black" style={{ color: '#1C1917' }}>{words.length}</p>
            <p className="text-[10px] font-bold uppercase" style={{ color: '#A8A29E' }}>Total</p>
          </div>
          <div className="rounded-2xl p-3 text-center" style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC' }}>
            <p className="text-xl font-black" style={{ color: '#116B3F' }}>{approvedCount}</p>
            <p className="text-[10px] font-bold uppercase" style={{ color: '#178A53' }}>Approved</p>
          </div>
          <div className="rounded-2xl p-3 text-center" style={{ background: flaggedCount > 0 ? '#FFFBEB' : '#F4F2EF', border: `1.5px solid ${flaggedCount > 0 ? '#FDE68A' : '#E7E5E4'}` }}>
            <p className="text-xl font-black" style={{ color: flaggedCount > 0 ? '#B45309' : '#A8A29E' }}>{flaggedCount}</p>
            <p className="text-[10px] font-bold uppercase" style={{ color: flaggedCount > 0 ? '#D97706' : '#A8A29E' }}>Flagged</p>
          </div>
        </motion.div>

        {/* Word grid */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 }} className="rounded-3xl p-5 sm:p-6 mb-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>Generated Words</h2>
            <button className="flex items-center gap-1.5 text-xs font-bold px-3 py-1.5 rounded-full" style={{ background: '#F5F2FF', color: '#6440E8' }}>
              <RefreshCw className="w-3.5 h-3.5" /> Regenerate All
            </button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
            {words.map((w, i) => (
              <motion.div key={w.id} initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 + i * 0.02 }} className="flex items-center gap-3 px-3.5 py-3 rounded-2xl group" style={{ background: w.flagged ? '#FFFBEB' : w.approved ? '#FAFAF9' : '#FFF1F2', border: `1.5px solid ${w.flagged ? '#FDE68A' : w.approved ? '#F0EDE8' : '#FECDD3'}` }}>
                <button onClick={() => toggleWord(w.id)} className="w-7 h-7 rounded-lg flex items-center justify-center shrink-0 transition-all" style={{ background: w.approved ? '#D5F5E6' : '#FFE4E6', color: w.approved ? '#116B3F' : '#BE123C' }}>
                  {w.approved ? <Check className="w-4 h-4" /> : <X className="w-4 h-4" />}
                </button>
                {editingId === w.id ? (
                  <div className="flex-1 flex gap-2">
                    <input autoFocus className="flex-1 px-2 py-1 rounded-lg text-sm font-bold outline-none" style={{ background: '#FAF8F5', border: '1.5px solid #FF5A1F', color: '#1C1917' }} value={editValue} onChange={e => setEditValue(e.target.value)} onKeyDown={e => e.key === 'Enter' && saveEdit(w.id)} />
                    <button onClick={() => saveEdit(w.id)} className="w-7 h-7 rounded-lg flex items-center justify-center" style={{ background: '#D5F5E6', color: '#116B3F' }}><Check className="w-4 h-4" /></button>
                  </div>
                ) : (
                  <>
                    <span className="flex-1 text-sm font-bold truncate" style={{ color: '#1C1917' }}>{w.word}</span>
                    {w.flagged && <AlertTriangle className="w-4 h-4 shrink-0" style={{ color: '#D97706' }} />}
                    <button onClick={() => startEdit(w.id, w.word)} className="w-7 h-7 rounded-lg flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity" style={{ background: '#F4F2EF', color: '#78716C' }}><Edit3 className="w-3.5 h-3.5" /></button>
                  </>
                )}
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Approve action */}
        {!approved ? (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.3 }} className="flex flex-col sm:flex-row gap-3">
            <button onClick={() => setApproved(true)} className="flex-1 flex items-center justify-center gap-2 py-4 rounded-2xl text-base font-extrabold transition-all" style={{ background: 'linear-gradient(135deg, #3DC484, #22AA6A)', color: '#fff', boxShadow: '0 6px 20px rgba(34,170,106,0.30)' }}>
              <CheckCircle2 className="w-5 h-5" /> Approve Word Set
            </button>
            <button className="flex items-center justify-center gap-2 px-6 py-4 rounded-2xl text-sm font-bold" style={{ background: '#F4F2EF', color: '#78716C', border: '1.5px solid #E7E5E4' }}>
              <RefreshCw className="w-4 h-4" /> Regenerate
            </button>
          </motion.div>
        ) : (
          <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="flex items-center gap-4 px-6 py-5 rounded-3xl" style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC' }}>
            <CheckCircle2 className="w-8 h-8" style={{ color: '#22AA6A' }} />
            <div>
              <p className="text-base font-extrabold" style={{ color: '#0D512F' }}>Word Set Approved ✓</p>
              <p className="text-sm font-semibold" style={{ color: '#178A53' }}>{approvedCount} words locked in for the next game run.</p>
            </div>
          </motion.div>
        )}
      </div>
    </DashboardShell>
  )
}
