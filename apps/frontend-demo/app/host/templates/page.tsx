'use client'

import { useState } from 'react'
import { motion } from 'motion/react'
import Link from 'next/link'
import { DashboardShell } from '@/components/DashboardShell'
import { mockGameTemplates } from '@/lib/mockAdminData'
import { GameTemplate } from '@/types/admin'
import { Plus, Users, Sparkles, Mic, Gift, Repeat, Pause, Play } from 'lucide-react'

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<GameTemplate[]>(mockGameTemplates)
  const toggleActive = (id: string) => {
    setTemplates(prev => prev.map(t => t.id === id ? { ...t, isActive: !t.isActive } : t))
  }

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 mb-8">
          <div>
            <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1.5" style={{ color: '#A8A29E' }}>Game Templates</p>
            <h1 className="text-3xl sm:text-4xl font-black tracking-tight mb-2" style={{ color: '#1C1917' }}>Recurring Games</h1>
            <p className="text-sm font-semibold" style={{ color: '#78716C' }}>Set up a template once — the system creates weekly runs automatically.</p>
          </div>
          <Link href="/host/templates/new" className="flex items-center justify-center gap-2 px-5 py-3 rounded-2xl text-sm font-extrabold shrink-0" style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', color: '#FFFFFF', boxShadow: '0 4px 16px rgba(255,90,31,0.30)' }}>
            <Plus className="w-4 h-4" /> New Template
          </Link>
        </motion.div>

        <div className="space-y-4">
          {templates.map((tmpl, i) => (
            <motion.div key={tmpl.id} initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 + i * 0.06 }} className="rounded-3xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)', opacity: tmpl.isActive ? 1 : 0.7 }}>
              <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                <div className="flex items-start gap-4 flex-1 min-w-0">
                  <div className="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 text-xl font-black" style={{ background: tmpl.isActive ? 'linear-gradient(135deg, #FF7A42, #FF5A1F)' : '#E7E5E4', color: tmpl.isActive ? '#FFFFFF' : '#A8A29E', boxShadow: tmpl.isActive ? '0 4px 12px rgba(255,90,31,0.25)' : 'none' }}>
                    {tmpl.name.charAt(0)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2.5 mb-1.5 flex-wrap">
                      <h3 className="text-lg font-black truncate" style={{ color: '#1C1917' }}>{tmpl.name}</h3>
                      <span className="text-[9px] font-extrabold px-2.5 py-0.5 rounded-full uppercase tracking-wide" style={{ background: tmpl.isActive ? '#D5F5E6' : '#F4F2EF', color: tmpl.isActive ? '#116B3F' : '#A8A29E' }}>
                        {tmpl.isActive ? 'Active' : 'Paused'}
                      </span>
                    </div>
                    <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 mb-3">
                      <span className="flex items-center gap-1.5 text-xs font-semibold" style={{ color: '#78716C' }}><Repeat className="w-3.5 h-3.5" /> {tmpl.recurrence} · {tmpl.dayOfWeek} · {tmpl.time}</span>
                      <span className="flex items-center gap-1.5 text-xs font-semibold" style={{ color: '#78716C' }}><Users className="w-3.5 h-3.5" /> {tmpl.audienceSize} audience</span>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <span className="flex items-center gap-1.5 text-[10px] font-bold px-2.5 py-1 rounded-full" style={{ background: '#F5F2FF', color: '#6440E8' }}><Sparkles className="w-3 h-3" /> {tmpl.contentMode}</span>
                      <span className="flex items-center gap-1.5 text-[10px] font-bold px-2.5 py-1 rounded-full" style={{ background: '#EDFAF5', color: '#116B3F' }}><Mic className="w-3 h-3" /> {tmpl.voiceMode}</span>
                      {tmpl.prizeEnabled && <span className="flex items-center gap-1.5 text-[10px] font-bold px-2.5 py-1 rounded-full" style={{ background: '#FEF3C7', color: '#B45309' }}><Gift className="w-3 h-3" /> {tmpl.prizeValue || 'Prize'}</span>}
                      <span className="text-[10px] font-bold px-2.5 py-1 rounded-full" style={{ background: '#F4F2EF', color: '#78716C' }}>Min {tmpl.minPlayers} players</span>
                    </div>
                  </div>
                </div>
                <div className="flex sm:flex-col items-center sm:items-end gap-3 sm:gap-2 shrink-0">
                  <button onClick={() => toggleActive(tmpl.id)} className="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-bold transition-all" style={{ background: tmpl.isActive ? '#EDFAF5' : '#F4F2EF', color: tmpl.isActive ? '#116B3F' : '#78716C', border: `1.5px solid ${tmpl.isActive ? '#A8EBCC' : '#E7E5E4'}` }}>
                    {tmpl.isActive ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
                    {tmpl.isActive ? 'Pause' : 'Resume'}
                  </button>
                  <div className="text-right">
                    <p className="text-xs font-bold" style={{ color: '#1C1917' }}>{tmpl.totalRuns} runs</p>
                    {tmpl.nextRunAt && tmpl.isActive && <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>Next: {new Date(tmpl.nextRunAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>}
                  </div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </DashboardShell>
  )
}
