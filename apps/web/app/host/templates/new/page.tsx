'use client'

import { useState } from 'react'
import { motion } from 'motion/react'
import { useRouter } from 'next/navigation'
import { DashboardShell } from '@/components/DashboardShell'
import { CalendarClock, Users, Sparkles, Mic, Gift, ChevronLeft, Check, AlertCircle } from 'lucide-react'

const STEPS = ['Schedule', 'Content', 'Voice & Prizes']

const inputStyle = {
  background: '#FAF8F5', border: '1.5px solid #E7E5E4', color: '#1C1917',
  fontFamily: 'inherit', borderRadius: '1rem', padding: '0.75rem 1rem',
  fontSize: '0.875rem', fontWeight: 700, outline: 'none', width: '100%',
}
const labelStyle = { color: '#57534E', fontSize: '0.75rem', fontWeight: 800, textTransform: 'uppercase' as const, letterSpacing: '0.08em', marginBottom: '0.375rem', display: 'block' }

const selectStyle = { ...inputStyle, appearance: 'none' as const, cursor: 'pointer' }

export default function NewTemplatePage() {
  const router = useRouter()
  const [step, setStep] = useState(0)
  const [form, setForm] = useState({
    name: '', recurrence: 'Weekly', dayOfWeek: 'Friday', time: '15:00', timezone: 'America/Toronto',
    audienceSize: '25', minPlayers: '6', contentMode: 'AI Generated', voiceMode: 'Default Neural',
    prizeEnabled: false, prizeValue: '',
  })

  const update = (key: string, value: string | boolean) => setForm(prev => ({ ...prev, [key]: value }))
  const createTemplate = () => {
    router.push('/host/templates')
  }

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-2xl mx-auto">
        {/* Back */}
        <motion.button initial={{ opacity: 0 }} animate={{ opacity: 1 }} onClick={() => router.push('/host/templates')} className="flex items-center gap-1.5 text-sm font-bold mb-6 transition-all" style={{ color: '#A8A29E' }}>
          <ChevronLeft className="w-4 h-4" /> Back to Templates
        </motion.button>

        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <h1 className="text-3xl font-black tracking-tight mb-2" style={{ color: '#1C1917' }}>Create Game Template</h1>
          <p className="text-sm font-semibold" style={{ color: '#78716C' }}>Configure once. The system will handle weekly runs automatically.</p>
        </motion.div>

        {/* Step indicator */}
        <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="flex items-center gap-2 mb-8">
          {STEPS.map((s, i) => (
            <div key={s} className="flex items-center gap-2">
              <button onClick={() => setStep(i)} className="flex items-center gap-2 px-3 py-2 rounded-full text-xs font-extrabold transition-all" style={{ background: step === i ? '#FFF4F0' : step > i ? '#D5F5E6' : '#F4F2EF', color: step === i ? '#E8440A' : step > i ? '#116B3F' : '#A8A29E', border: step === i ? '1.5px solid #FFE4D9' : '1.5px solid transparent' }}>
                {step > i ? <Check className="w-3.5 h-3.5" /> : <span className="w-4 h-4 rounded-full flex items-center justify-center text-[10px] font-black" style={{ background: step === i ? '#FF5A1F' : '#E7E5E4', color: step === i ? '#fff' : '#A8A29E' }}>{i + 1}</span>}
                {s}
              </button>
              {i < STEPS.length - 1 && <div className="w-6 h-px" style={{ background: '#E7E5E4' }} />}
            </div>
          ))}
        </motion.div>

        {/* Form card */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="rounded-xl p-6 sm:p-8" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 4px 24px rgba(0,0,0,0.05)' }}>

          {step === 0 && (
            <div className="space-y-5">
              <div><label style={labelStyle}>Template Name</label><input style={inputStyle} placeholder="e.g., Friday Fun Bingo" value={form.name} onChange={e => update('name', e.target.value)} /></div>
              <div className="grid grid-cols-2 gap-4">
                <div><label style={labelStyle}>Recurrence</label><select style={selectStyle} value={form.recurrence} onChange={e => update('recurrence', e.target.value)}><option>Weekly</option><option>Bi-Weekly</option><option>Monthly</option><option>One-Time</option></select></div>
                <div><label style={labelStyle}>Day of Week</label><select style={selectStyle} value={form.dayOfWeek} onChange={e => update('dayOfWeek', e.target.value)}><option>Monday</option><option>Tuesday</option><option>Wednesday</option><option>Thursday</option><option>Friday</option></select></div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div><label style={labelStyle}>Start Time</label><input type="time" style={inputStyle} value={form.time} onChange={e => update('time', e.target.value)} /></div>
                <div><label style={labelStyle}>Timezone</label><select style={selectStyle} value={form.timezone} onChange={e => update('timezone', e.target.value)}><option>America/Toronto</option><option>America/New_York</option><option>America/Chicago</option><option>America/Denver</option><option>America/Los_Angeles</option><option>Europe/London</option><option>Europe/Paris</option></select></div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div><label style={labelStyle}>Audience Size</label><input type="number" style={inputStyle} value={form.audienceSize} onChange={e => update('audienceSize', e.target.value)} /></div>
                <div><label style={labelStyle}>Min Players to Start</label><input type="number" style={inputStyle} value={form.minPlayers} onChange={e => update('minPlayers', e.target.value)} /></div>
              </div>
            </div>
          )}

          {step === 1 && (
            <div className="space-y-5">
              <div><label style={labelStyle}>Content Generation Mode</label>
                <div className="space-y-2.5">
                  {['AI Generated', 'Manual', 'Reuse Word Bank'].map(mode => (
                    <button key={mode} onClick={() => update('contentMode', mode)} className="w-full flex items-center gap-3 px-4 py-3.5 rounded-lg text-left transition-all" style={{ background: form.contentMode === mode ? '#FFF4F0' : '#FAFAF9', border: form.contentMode === mode ? '1.5px solid #FFE4D9' : '1.5px solid #F0EDE8' }}>
                      <div className="w-9 h-9 rounded-md flex items-center justify-center" style={{ background: form.contentMode === mode ? 'linear-gradient(135deg, #FF7A42, #FF5A1F)' : '#F4F2EF' }}>
                        <Sparkles className="w-4 h-4" style={{ color: form.contentMode === mode ? '#fff' : '#A8A29E' }} />
                      </div>
                      <div><p className="text-sm font-bold" style={{ color: '#1C1917' }}>{mode}</p><p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>{mode === 'AI Generated' ? 'AI creates fresh words from a prompt each week' : mode === 'Manual' ? 'You provide the word list manually' : 'Pick from previously saved word banks'}</p></div>
                    </button>
                  ))}
                </div>
              </div>
              {form.contentMode === 'AI Generated' && (
                <div className="flex items-start gap-3 px-4 py-3 rounded-lg" style={{ background: '#F5F2FF', border: '1.5px solid #D9CCFF' }}>
                  <Sparkles className="w-4 h-4 mt-0.5 shrink-0" style={{ color: '#7C5CFC' }} />
                  <div><p className="text-xs font-bold" style={{ color: '#4F30C2' }}>AI will generate words before each run</p><p className="text-[10px] font-semibold" style={{ color: '#7C5CFC' }}>You can review and edit the generated content in the Content Review screen.</p></div>
                </div>
              )}
            </div>
          )}

          {step === 2 && (
            <div className="space-y-5">
              <div><label style={labelStyle}>Voice Caller Mode</label>
                <div className="space-y-2.5">
                  {[{ v: 'Default Neural', d: 'Use Azure default neural voice' }, { v: 'Employee Voice', d: 'Use approved employee voice profile' }, { v: 'Auto Select', d: 'System picks from approved voices' }].map(({ v, d }) => (
                    <button key={v} onClick={() => update('voiceMode', v)} className="w-full flex items-center gap-3 px-4 py-3.5 rounded-lg text-left transition-all" style={{ background: form.voiceMode === v ? '#EDFAF5' : '#FAFAF9', border: form.voiceMode === v ? '1.5px solid #A8EBCC' : '1.5px solid #F0EDE8' }}>
                      <div className="w-9 h-9 rounded-md flex items-center justify-center" style={{ background: form.voiceMode === v ? 'linear-gradient(135deg, #3DC484, #22AA6A)' : '#F4F2EF' }}>
                        <Mic className="w-4 h-4" style={{ color: form.voiceMode === v ? '#fff' : '#A8A29E' }} />
                      </div>
                      <div><p className="text-sm font-bold" style={{ color: '#1C1917' }}>{v}</p><p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>{d}</p></div>
                    </button>
                  ))}
                </div>
              </div>
              <div className="h-px" style={{ background: '#F0EDE8' }} />
              <div>
                <label style={labelStyle}>Prize Configuration</label>
                <button onClick={() => update('prizeEnabled', !form.prizeEnabled)} className="w-full flex items-center justify-between px-4 py-3.5 rounded-lg transition-all" style={{ background: form.prizeEnabled ? '#FFFBEB' : '#FAFAF9', border: form.prizeEnabled ? '1.5px solid #FDE68A' : '1.5px solid #F0EDE8' }}>
                  <div className="flex items-center gap-3"><Gift className="w-5 h-5" style={{ color: form.prizeEnabled ? '#D97706' : '#A8A29E' }} /><span className="text-sm font-bold" style={{ color: '#1C1917' }}>Enable prizes for winners</span></div>
                  <div className="w-10 h-6 rounded-full transition-all flex items-center" style={{ background: form.prizeEnabled ? '#22AA6A' : '#E7E5E4', padding: '2px' }}>
                    <motion.div animate={{ x: form.prizeEnabled ? 16 : 0 }} className="w-5 h-5 rounded-full bg-white" style={{ boxShadow: '0 1px 4px rgba(0,0,0,0.15)' }} />
                  </div>
                </button>
                {form.prizeEnabled && <div className="mt-3"><label style={labelStyle}>Prize Description</label><input style={inputStyle} placeholder="e.g., $25 Gift Card" value={form.prizeValue} onChange={e => update('prizeValue', e.target.value)} /></div>}
              </div>
            </div>
          )}

          {/* Navigation buttons */}
          <div className="flex justify-between mt-8 pt-5" style={{ borderTop: '1.5px solid #F0EDE8' }}>
            <button onClick={() => step > 0 && setStep(step - 1)} disabled={step === 0} className="px-5 py-3 rounded-lg text-sm font-bold transition-all" style={{ background: '#F4F2EF', color: step === 0 ? '#D6D3D1' : '#57534E' }}>
              Previous
            </button>
            {step < STEPS.length - 1 ? (
              <button onClick={() => setStep(step + 1)} className="px-6 py-3 rounded-lg text-sm font-extrabold" style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', color: '#fff', boxShadow: '0 4px 16px rgba(255,90,31,0.30)' }}>
                Continue
              </button>
            ) : (
              <button onClick={createTemplate} className="px-6 py-3 rounded-lg text-sm font-extrabold" style={{ background: 'linear-gradient(135deg, #3DC484, #22AA6A)', color: '#fff', boxShadow: '0 4px 16px rgba(34,170,106,0.30)' }}>
                Create Template
              </button>
            )}
          </div>
        </motion.div>
      </div>
    </DashboardShell>
  )
}
