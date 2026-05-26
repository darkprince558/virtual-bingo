'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import Link from 'next/link'
import { DashboardShell } from '@/components/DashboardShell'
import { mockGameTemplates } from '@/lib/mockAdminData'
import { GameTemplate } from '@/types/admin'
import { Plus, Users, Sparkles, Mic, Gift, Repeat, Pause, Play, Settings2, X, Save } from 'lucide-react'

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<GameTemplate[]>(mockGameTemplates)
  const [editingTemplateId, setEditingTemplateId] = useState<string | null>(null)
  const [editForm, setEditForm] = useState<Partial<GameTemplate>>({})

  const toggleActive = (id: string) => {
    setTemplates(prev => prev.map(t => t.id === id ? { ...t, isActive: !t.isActive } : t))
  }

  const openEdit = (tmpl: GameTemplate) => {
    setEditingTemplateId(tmpl.id)
    setEditForm(tmpl)
  }

  const saveEdit = () => {
    if (!editingTemplateId) return
    setTemplates(prev => {
      return prev.map(t => t.id === editingTemplateId ? { ...t, ...editForm } as GameTemplate : t)
    })
    setEditingTemplateId(null)
  }

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 mb-8">
          <div>
            <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1.5" style={{ color: '#A8A29E' }}>Game Templates</p>
            <h1 className="text-3xl sm:text-4xl font-black tracking-tight mb-2" style={{ color: '#1C1917' }}>Recurring Games</h1>
            <p className="text-sm font-semibold" style={{ color: '#78716C' }}>Set up a template once and the system creates weekly runs automatically.</p>
          </div>
          <Link href="/host/templates/new" className="flex items-center justify-center gap-2 px-5 py-3 rounded-lg text-sm font-extrabold shrink-0" style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', color: '#FFFFFF', boxShadow: '0 4px 16px rgba(255,90,31,0.30)' }}>
            <Plus className="w-4 h-4" /> New Template
          </Link>
        </motion.div>

        <div className="space-y-4">
          {templates.map((tmpl, i) => (
            <motion.div key={tmpl.id} initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 + i * 0.06 }} className="rounded-xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)', opacity: tmpl.isActive ? 1 : 0.7 }}>
              <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                <div className="flex items-start gap-4 flex-1 min-w-0">
                  <div className="w-12 h-12 rounded-lg flex items-center justify-center shrink-0 text-xl font-black" style={{ background: tmpl.isActive ? 'linear-gradient(135deg, #FF7A42, #FF5A1F)' : '#E7E5E4', color: tmpl.isActive ? '#FFFFFF' : '#A8A29E', boxShadow: tmpl.isActive ? '0 4px 12px rgba(255,90,31,0.25)' : 'none' }}>
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
                  <button onClick={() => toggleActive(tmpl.id)} className="flex items-center gap-2 px-3 py-2 rounded-md text-xs font-bold transition-all" style={{ background: tmpl.isActive ? '#EDFAF5' : '#F4F2EF', color: tmpl.isActive ? '#116B3F' : '#78716C', border: `1.5px solid ${tmpl.isActive ? '#A8EBCC' : '#E7E5E4'}` }}>
                    {tmpl.isActive ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
                    {tmpl.isActive ? 'Pause' : 'Resume'}
                  </button>
                  <button onClick={() => openEdit(tmpl)} className="flex items-center gap-2 px-3 py-2 rounded-md text-xs font-bold transition-all bg-white hover:bg-gray-50" style={{ color: '#78716C', border: '1.5px solid #E7E5E4' }}>
                    <Settings2 className="w-3.5 h-3.5" /> Edit
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

      <AnimatePresence>
        {editingTemplateId && (
          <>
            {/* Backdrop */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setEditingTemplateId(null)}
              className="fixed inset-0 bg-black/20 backdrop-blur-sm z-40"
            />
            
            {/* Sliding Drawer */}
            <motion.div
              initial={{ x: '100%', opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              exit={{ x: '100%', opacity: 0 }}
              transition={{ type: 'spring', damping: 25, stiffness: 200 }}
              className="fixed top-0 right-0 h-full w-full sm:w-[480px] bg-white shadow-2xl z-50 flex flex-col border-l border-[#F0EDE8]"
            >
              {/* Header */}
              <div className="flex items-center justify-between p-6 border-b border-[#F0EDE8] bg-[#FAFAF9]">
                <div>
                  <h2 className="text-xl font-black text-[#1C1917]">Edit Template</h2>
                  <p className="text-xs font-semibold text-[#78716C] mt-1">Update rules and dynamics for future runs</p>
                </div>
                <button onClick={() => setEditingTemplateId(null)} className="p-2 rounded-full hover:bg-[#F0EDE8] transition-colors text-[#78716C]">
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Body Form */}
              <div className="flex-1 overflow-y-auto p-6 space-y-8">
                {/* General */}
                <section>
                  <h3 className="text-xs font-black uppercase tracking-wider text-[#A8A29E] mb-4">Identity</h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Template Name</label>
                      <input
                        value={editForm.name || ''}
                        onChange={e => setEditForm(prev => ({ ...prev, name: e.target.value }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      />
                    </div>
                  </div>
                </section>

                {/* Schedule */}
                <section>
                  <h3 className="text-xs font-black uppercase tracking-wider text-[#A8A29E] mb-4">Schedule & Reach</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Recurrence</label>
                      <select
                        value={editForm.recurrence || 'Weekly'}
                        onChange={e => setEditForm(prev => ({ ...prev, recurrence: e.target.value as any }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      >
                        <option>Weekly</option>
                        <option>Bi-Weekly</option>
                        <option>Monthly</option>
                        <option>One-Time</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Day of Week</label>
                      <select
                        value={editForm.dayOfWeek || 'Friday'}
                        onChange={e => setEditForm(prev => ({ ...prev, dayOfWeek: e.target.value }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      >
                        <option>Monday</option>
                        <option>Tuesday</option>
                        <option>Wednesday</option>
                        <option>Thursday</option>
                        <option>Friday</option>
                        <option>Saturday</option>
                        <option>Sunday</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Time</label>
                      <input
                        value={editForm.time || ''}
                        onChange={e => setEditForm(prev => ({ ...prev, time: e.target.value }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      />
                    </div>
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Audience Size</label>
                      <input
                        type="number"
                        value={editForm.audienceSize || 0}
                        onChange={e => setEditForm(prev => ({ ...prev, audienceSize: parseInt(e.target.value, 10) }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      />
                    </div>
                  </div>
                </section>

                {/* Dynamics */}
                <section>
                  <h3 className="text-xs font-black uppercase tracking-wider text-[#A8A29E] mb-4">Game Dynamics</h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Content Generation</label>
                      <select
                        value={editForm.contentMode || 'AI Generated'}
                        onChange={e => setEditForm(prev => ({ ...prev, contentMode: e.target.value as any }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      >
                        <option>AI Generated</option>
                        <option>Manual</option>
                        <option>Reuse Word Bank</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5">Voice Profile</label>
                      <select
                        value={editForm.voiceMode || 'Default Neural'}
                        onChange={e => setEditForm(prev => ({ ...prev, voiceMode: e.target.value as any }))}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                      >
                        <option>Default Neural</option>
                        <option>Employee Voice</option>
                        <option>Auto Select</option>
                      </select>
                    </div>
                    
                    <div className="flex items-center gap-3 pt-2">
                      <input
                        type="checkbox"
                        id="prizeEnabled"
                        checked={editForm.prizeEnabled || false}
                        onChange={e => setEditForm(prev => ({ ...prev, prizeEnabled: e.target.checked }))}
                        className="w-4 h-4 rounded text-orange-500 focus:ring-orange-500 border-gray-300"
                      />
                      <label htmlFor="prizeEnabled" className="text-sm font-bold text-[#1C1917]">Enable Prize & Fulfillment</label>
                    </div>
                    
                    <AnimatePresence>
                      {editForm.prizeEnabled && (
                        <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: 'auto' }} exit={{ opacity: 0, height: 0 }}>
                          <label className="block text-[10px] font-extrabold uppercase tracking-wider text-[#A8A29E] mb-1.5 mt-2">Prize Value (e.g. $50 Gift Card)</label>
                          <input
                            value={editForm.prizeValue || ''}
                            onChange={e => setEditForm(prev => ({ ...prev, prizeValue: e.target.value }))}
                            className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20 bg-[#FAFAF9] border-[1.5px] border-[#F0EDE8] text-[#1C1917]"
                          />
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                </section>
              </div>

              {/* Footer Actions */}
              <div className="p-6 border-t border-[#F0EDE8] bg-[#FAFAF9] flex gap-3">
                <button onClick={() => setEditingTemplateId(null)} className="flex-1 py-3 rounded-lg text-sm font-bold text-[#78716C] bg-white border-[1.5px] border-[#E7E5E4] hover:bg-gray-50 transition-colors">
                  Cancel
                </button>
                <button onClick={saveEdit} className="flex-1 py-3 rounded-lg text-sm font-bold text-white flex items-center justify-center gap-2 transition-all hover:scale-[1.02]" style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', boxShadow: '0 4px 12px rgba(255,90,31,0.25)' }}>
                  <Save className="w-4 h-4" /> Save Changes
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </DashboardShell>
  )
}
