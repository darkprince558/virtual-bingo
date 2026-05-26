'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { DashboardShell } from '@/components/DashboardShell'
import { mockVoiceProfiles } from '@/lib/mockAdminData'
import { VoiceProfile } from '@/types/admin'
import { Mic, Shield, CheckCircle2, XCircle, Clock, AlertTriangle, Mail, Play, Pause, Trash2, Info, X, FileAudio, UserCheck, Calendar } from 'lucide-react'

const STATUS_CONFIG: Record<string, { bg: string; color: string; label: string }> = {
  'Active':           { bg: '#D5F5E6', color: '#116B3F', label: 'Active' },
  'Inactive':         { bg: '#F4F2EF', color: '#78716C', label: 'Inactive' },
  'Pending Consent':  { bg: '#FEF3C7', color: '#B45309', label: 'Pending Consent' },
  'Revoked':          { bg: '#FFE4E6', color: '#BE123C', label: 'Revoked' },
}

export default function VoiceProfilesPage() {
  const [profiles, setProfiles] = useState<VoiceProfile[]>(mockVoiceProfiles)
  const [selectedProfile, setSelectedProfile] = useState<VoiceProfile | null>(null)

  const toggleStatus = (id: string) => {
    setProfiles(prev => prev.map(p => {
      if (p.id !== id) return p
      if (p.status === 'Active') return { ...p, status: 'Inactive' as const }
      if (p.status === 'Inactive') return { ...p, status: 'Active' as const }
      return p
    }))
  }

  const deleteProfile = (id: string) => {
    // Basic confirmation
    if (confirm('Are you sure you want to delete this voice profile?')) {
      setProfiles(prev => prev.filter(p => p.id !== id))
    }
  }

  return (
    <DashboardShell role="admin" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)', boxShadow: '0 4px 12px rgba(124,92,252,0.25)' }}>
              <Mic className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917' }}>Voice Profile Management</h1>
              <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Manage consented employee voice profiles for AI calling</p>
            </div>
          </div>
        </motion.div>

        {/* Consent warning */}
        <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="rounded-lg p-4 mb-6 flex items-start gap-3" style={{ background: '#FFFBEB', border: '1.5px solid #FDE68A' }}>
          <AlertTriangle className="w-5 h-5 shrink-0 mt-0.5" style={{ color: '#D97706' }} />
          <div>
            <p className="text-xs font-bold" style={{ color: '#92400E' }}>Voice Consent Required</p>
            <p className="text-[10px] font-semibold" style={{ color: '#B45309' }}>Employee voices can only be replicated with explicit, recorded consent. Azure AI Speech requires a consent statement before creating personal voice profiles.</p>
          </div>
        </motion.div>

        {/* Profile cards */}
        <div className="space-y-4">
          <AnimatePresence>
            {profiles.map((profile, i) => {
              const cfg = STATUS_CONFIG[profile.status]
              return (
                <motion.div key={profile.id} layout initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, scale: 0.95 }} transition={{ delay: 0.1 + i * 0.06 }} className="rounded-xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
                  <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                    {/* Avatar + info */}
                    <div className="flex items-start gap-4 flex-1 min-w-0">
                      <div className="w-12 h-12 rounded-lg flex items-center justify-center shrink-0 text-sm font-black" style={{ background: profile.status === 'Active' ? 'linear-gradient(135deg, #7C5CFC, #9E80FF)' : profile.status === 'Revoked' ? '#FFE4E6' : '#F4F2EF', color: profile.status === 'Active' ? '#FFFFFF' : profile.status === 'Revoked' ? '#BE123C' : '#A8A29E' }}>
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
                            <div className="rounded-md px-3 py-2" style={{ background: '#EDFAF5', border: '1px solid #A8EBCC' }}>
                              <p className="text-[9px] font-bold uppercase" style={{ color: '#178A53' }}>Consent</p>
                              <p className="text-xs font-bold" style={{ color: '#0D512F' }}>{new Date(profile.consentRecordedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</p>
                            </div>
                          )}
                          <div className="rounded-md px-3 py-2" style={{ background: '#F5F2FF', border: '1px solid #D9CCFF' }}>
                            <p className="text-[9px] font-bold uppercase" style={{ color: '#6440E8' }}>Usage</p>
                            <p className="text-xs font-bold" style={{ color: '#4F30C2' }}>{profile.usageCount} games</p>
                          </div>
                          {profile.lastUsedAt && (
                            <div className="rounded-md px-3 py-2" style={{ background: '#F4F2EF', border: '1px solid #E7E5E4' }}>
                              <p className="text-[9px] font-bold uppercase" style={{ color: '#78716C' }}>Last Used</p>
                              <p className="text-xs font-bold" style={{ color: '#57534E' }}>{new Date(profile.lastUsedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</p>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="flex gap-2 shrink-0 self-start sm:self-center">
                      <button onClick={() => setSelectedProfile(profile)} className="flex items-center justify-center w-9 h-9 rounded-md transition-all hover:bg-stone-100 text-stone-500 hover:text-stone-900 border border-stone-200" title="View Info">
                        <Info className="w-4 h-4" />
                      </button>
                      
                      {(profile.status === 'Active' || profile.status === 'Inactive') && (
                        <button onClick={() => toggleStatus(profile.id)} className="flex items-center justify-center gap-1.5 px-3 py-2 rounded-md text-xs font-bold transition-all w-[110px]" style={{ background: profile.status === 'Active' ? '#FEF3C7' : '#D5F5E6', color: profile.status === 'Active' ? '#B45309' : '#116B3F', border: `1.5px solid ${profile.status === 'Active' ? '#FDE68A' : '#A8EBCC'}` }}>
                          {profile.status === 'Active' ? <><Pause className="w-3.5 h-3.5 shrink-0" /> Deactivate</> : <><Play className="w-3.5 h-3.5 shrink-0" /> Activate</>}
                        </button>
                      )}
                      {profile.status === 'Pending Consent' && (
                        <span className="flex items-center justify-center gap-1.5 px-3 py-2 rounded-md text-xs font-bold w-[110px]" style={{ background: '#FEF3C7', color: '#B45309', border: '1.5px solid #FDE68A' }}>
                          <Clock className="w-3.5 h-3.5 shrink-0" /> Awaiting
                        </span>
                      )}
                      {profile.status === 'Revoked' && (
                        <span className="flex items-center justify-center gap-1.5 px-3 py-2 rounded-md text-xs font-bold w-[110px]" style={{ background: '#FFE4E6', color: '#BE123C', border: '1.5px solid #FDA4AF' }}>
                          <XCircle className="w-3.5 h-3.5 shrink-0" /> Revoked
                        </span>
                      )}

                      <button onClick={() => deleteProfile(profile.id)} className="flex items-center justify-center w-9 h-9 rounded-md transition-all hover:bg-red-50 text-red-500 border border-red-100" title="Delete Profile">
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </motion.div>
              )
            })}
          </AnimatePresence>
        </div>
      </div>

      {/* Info Modal */}
      <AnimatePresence>
        {selectedProfile && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} onClick={() => setSelectedProfile(null)} className="absolute inset-0 bg-stone-900/40 backdrop-blur-sm" />
            <motion.div initial={{ opacity: 0, scale: 0.95, y: 10 }} animate={{ opacity: 1, scale: 1, y: 0 }} exit={{ opacity: 0, scale: 0.95, y: 10 }} className="relative w-full max-w-lg bg-white rounded-2xl shadow-2xl border border-stone-200 overflow-hidden z-10 flex flex-col max-h-[90vh]">
              {/* Modal Header */}
              <div className="flex items-center justify-between p-5 border-b border-stone-100" style={{ background: 'linear-gradient(to right, #FAF9F8, #FFFFFF)' }}>
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-black text-white" style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)' }}>
                    {selectedProfile.employeeName.split(' ').map(n => n[0]).join('').slice(0, 2)}
                  </div>
                  <div>
                    <h2 className="text-lg font-black text-stone-900">{selectedProfile.employeeName}</h2>
                    <p className="text-xs font-semibold text-stone-500">{selectedProfile.employeeEmail}</p>
                  </div>
                </div>
                <button onClick={() => setSelectedProfile(null)} className="w-8 h-8 flex items-center justify-center rounded-full hover:bg-stone-100 text-stone-400 hover:text-stone-700 transition-colors">
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Modal Content */}
              <div className="p-5 sm:p-6 overflow-y-auto">
                <div className="space-y-6">
                  
                  {/* Status Section */}
                  <div>
                    <h4 className="text-[10px] font-bold text-stone-400 uppercase tracking-wider mb-2">Current Status</h4>
                    <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md" style={{ background: STATUS_CONFIG[selectedProfile.status].bg, color: STATUS_CONFIG[selectedProfile.status].color }}>
                      <span className="relative flex h-2 w-2">
                        {selectedProfile.status === 'Active' && <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>}
                        <span className={`relative inline-flex rounded-full h-2 w-2 ${selectedProfile.status === 'Active' ? 'bg-emerald-500' : 'bg-current'}`}></span>
                      </span>
                      <span className="text-xs font-bold uppercase tracking-wide">{selectedProfile.status}</span>
                    </div>
                  </div>

                  {/* Authorization Section */}
                  <div>
                    <h4 className="text-[10px] font-bold text-stone-400 uppercase tracking-wider mb-3">Authorization Details</h4>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div className="flex items-center gap-3 p-3 rounded-lg border border-stone-100 bg-stone-50/50">
                        <Calendar className="w-4 h-4 text-emerald-600" />
                        <div>
                          <p className="text-[10px] font-semibold text-stone-500">Consent Recorded</p>
                          <p className="text-xs font-bold text-stone-900">{selectedProfile.consentRecordedAt ? new Date(selectedProfile.consentRecordedAt).toLocaleString() : 'Not recorded'}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-3 p-3 rounded-lg border border-stone-100 bg-stone-50/50">
                        <UserCheck className="w-4 h-4 text-blue-600" />
                        <div>
                          <p className="text-[10px] font-semibold text-stone-500">Approved By</p>
                          <p className="text-xs font-bold text-stone-900">{selectedProfile.approvedBy || 'Pending'}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-3 p-3 rounded-lg border border-stone-100 bg-stone-50/50 sm:col-span-2">
                        <Shield className="w-4 h-4 text-purple-600" />
                        <div>
                          <p className="text-[10px] font-semibold text-stone-500">Approval Date</p>
                          <p className="text-xs font-bold text-stone-900">{selectedProfile.approvedAt ? new Date(selectedProfile.approvedAt).toLocaleString() : 'Not approved yet'}</p>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Voice Samples Section */}
                  <div>
                    <h4 className="text-[10px] font-bold text-stone-400 uppercase tracking-wider mb-3">Voice Samples ({selectedProfile.voiceSamples?.length || 0})</h4>
                    {selectedProfile.voiceSamples && selectedProfile.voiceSamples.length > 0 ? (
                      <div className="space-y-2">
                        {selectedProfile.voiceSamples.map((sample, idx) => (
                          <div key={idx} className="flex items-center justify-between p-3 rounded-lg border border-stone-100 hover:border-stone-200 hover:bg-stone-50 transition-colors group">
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 rounded bg-purple-100 flex items-center justify-center">
                                <FileAudio className="w-4 h-4 text-purple-600" />
                              </div>
                              <p className="text-xs font-bold text-stone-700">{sample}</p>
                            </div>
                            <button className="text-xs font-semibold text-purple-600 opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
                              <Play className="w-3 h-3" /> Play
                            </button>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="p-4 rounded-lg border border-dashed border-stone-200 text-center">
                        <p className="text-xs font-medium text-stone-500">No voice samples recorded yet.</p>
                      </div>
                    )}
                  </div>
                  
                </div>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </DashboardShell>
  )
}
