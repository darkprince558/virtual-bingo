'use client'

import { useState } from 'react'
import { motion } from 'motion/react'
import { DashboardShell } from '@/components/DashboardShell'
import { mockVoiceProfiles } from '@/lib/mockAdminData'
import { VoiceProfile } from '@/types/admin'
import { Mic, Shield, CheckCircle2, XCircle, Clock, AlertTriangle, Mail, Play, Pause } from 'lucide-react'

const STATUS_CONFIG: Record<string, { bg: string; color: string; label: string }> = {
  'Active':           { bg: '#D5F5E6', color: '#116B3F', label: 'Active' },
  'Inactive':         { bg: '#F4F2EF', color: '#78716C', label: 'Inactive' },
  'Pending Consent':  { bg: '#FEF3C7', color: '#B45309', label: 'Pending Consent' },
  'Revoked':          { bg: '#FFE4E6', color: '#BE123C', label: 'Revoked' },
}

export default function VoiceProfilesPage() {
  const [profiles, setProfiles] = useState<VoiceProfile[]>(mockVoiceProfiles)

  const toggleStatus = (id: string) => {
    setProfiles(prev => prev.map(p => {
      if (p.id !== id) return p
      if (p.status === 'Active') return { ...p, status: 'Inactive' as const }
      if (p.status === 'Inactive') return { ...p, status: 'Active' as const }
      return p
    }))
  }

  return (
    <DashboardShell role="admin" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-2xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)', boxShadow: '0 4px 12px rgba(124,92,252,0.25)' }}>
              <Mic className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917' }}>Voice Profile Management</h1>
              <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Manage consented employee voice profiles for AI calling</p>
            </div>
          </div>
        </motion.div>

        {/* Consent warning */}
        <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="rounded-2xl p-4 mb-6 flex items-start gap-3" style={{ background: '#FFFBEB', border: '1.5px solid #FDE68A' }}>
          <AlertTriangle className="w-5 h-5 shrink-0 mt-0.5" style={{ color: '#D97706' }} />
          <div>
            <p className="text-xs font-bold" style={{ color: '#92400E' }}>Voice Consent Required</p>
            <p className="text-[10px] font-semibold" style={{ color: '#B45309' }}>Employee voices can only be replicated with explicit, recorded consent. Azure AI Speech requires a consent statement before creating personal voice profiles.</p>
          </div>
        </motion.div>

        {/* Profile cards */}
        <div className="space-y-4">
          {profiles.map((profile, i) => {
            const cfg = STATUS_CONFIG[profile.status]
            return (
              <motion.div key={profile.id} initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 + i * 0.06 }} className="rounded-3xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
                <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                  {/* Avatar + info */}
                  <div className="flex items-start gap-4 flex-1 min-w-0">
                    <div className="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 text-sm font-black" style={{ background: profile.status === 'Active' ? 'linear-gradient(135deg, #7C5CFC, #9E80FF)' : profile.status === 'Revoked' ? '#FFE4E6' : '#F4F2EF', color: profile.status === 'Active' ? '#FFFFFF' : profile.status === 'Revoked' ? '#BE123C' : '#A8A29E' }}>
                      {profile.employeeName.split(' ').map(n => n[0]).join('').slice(0, 2)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2.5 mb-1 flex-wrap">
                        <h3 className="text-base font-black" style={{ color: '#1C1917' }}>{profile.employeeName}</h3>
                        <span className="text-[9px] font-extrabold px-2.5 py-0.5 rounded-full uppercase tracking-wide" style={{ background: cfg.bg, color: cfg.color }}>{cfg.label}</span>
                      </div>
                      <p className="text-xs font-semibold flex items-center gap-1 mb-3" style={{ color: '#A8A29E' }}>
                        <Mail className="w-3.5 h-3.5" /> {profile.employeeEmail}
                      </p>

                      {/* Details grid */}
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2.5">
                        {profile.consentRecordedAt && (
                          <div className="rounded-xl px-3 py-2" style={{ background: '#EDFAF5', border: '1px solid #A8EBCC' }}>
                            <p className="text-[9px] font-bold uppercase" style={{ color: '#178A53' }}>Consent</p>
                            <p className="text-xs font-bold" style={{ color: '#0D512F' }}>{new Date(profile.consentRecordedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</p>
                          </div>
                        )}
                        <div className="rounded-xl px-3 py-2" style={{ background: '#F5F2FF', border: '1px solid #D9CCFF' }}>
                          <p className="text-[9px] font-bold uppercase" style={{ color: '#6440E8' }}>Usage</p>
                          <p className="text-xs font-bold" style={{ color: '#4F30C2' }}>{profile.usageCount} games</p>
                        </div>
                        {profile.lastUsedAt && (
                          <div className="rounded-xl px-3 py-2" style={{ background: '#F4F2EF', border: '1px solid #E7E5E4' }}>
                            <p className="text-[9px] font-bold uppercase" style={{ color: '#78716C' }}>Last Used</p>
                            <p className="text-xs font-bold" style={{ color: '#57534E' }}>{new Date(profile.lastUsedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex gap-2 shrink-0">
                    {(profile.status === 'Active' || profile.status === 'Inactive') && (
                      <button onClick={() => toggleStatus(profile.id)} className="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-bold transition-all" style={{ background: profile.status === 'Active' ? '#FEF3C7' : '#D5F5E6', color: profile.status === 'Active' ? '#B45309' : '#116B3F', border: `1.5px solid ${profile.status === 'Active' ? '#FDE68A' : '#A8EBCC'}` }}>
                        {profile.status === 'Active' ? <><Pause className="w-3.5 h-3.5" /> Deactivate</> : <><Play className="w-3.5 h-3.5" /> Activate</>}
                      </button>
                    )}
                    {profile.status === 'Pending Consent' && (
                      <span className="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-bold" style={{ background: '#FEF3C7', color: '#B45309' }}>
                        <Clock className="w-3.5 h-3.5" /> Awaiting
                      </span>
                    )}
                    {profile.status === 'Revoked' && (
                      <span className="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-bold" style={{ background: '#FFE4E6', color: '#BE123C' }}>
                        <XCircle className="w-3.5 h-3.5" /> Revoked
                      </span>
                    )}
                  </div>
                </div>
              </motion.div>
            )
          })}
        </div>
      </div>
    </DashboardShell>
  )
}
