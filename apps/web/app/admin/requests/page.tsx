'use client'

import { useState } from 'react'
import { motion } from 'motion/react'
import { DashboardShell } from '@/components/DashboardShell'
import { mockHostRequests } from '@/lib/mockAdminData'
import { HostPrivilegeRequest } from '@/types/admin'
import { Check, X, Clock, Shield, CheckCircle2, XCircle, Mail } from 'lucide-react'

export default function HostRequestsPage() {
  const [requests, setRequests] = useState<HostPrivilegeRequest[]>(mockHostRequests)
  const pending = requests.filter(r => r.status === 'Pending')
  const resolved = requests.filter(r => r.status !== 'Pending')

  const approve = (id: string) => {
    setRequests(prev => prev.map(r => r.id === id ? { ...r, status: 'Approved' as const, reviewedAt: new Date().toISOString(), reviewedBy: 'Admin Team' } : r))
  }
  const reject = (id: string) => {
    setRequests(prev => prev.map(r => r.id === id ? { ...r, status: 'Rejected' as const, reviewedAt: new Date().toISOString(), reviewedBy: 'Admin Team' } : r))
  }

  return (
    <DashboardShell role="admin" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #FBBF24, #F59E0B)', boxShadow: '0 4px 12px rgba(245,158,11,0.25)' }}>
              <Shield className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917' }}>Host Privilege Requests</h1>
              <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>Review and manage who can host games</p>
            </div>
          </div>
        </motion.div>

        {/* Pending section */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="rounded-xl p-5 sm:p-6 mb-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>Pending Review</h2>
            <span className="text-xs font-black px-2.5 py-1 rounded-full" style={{ background: '#FFFBEB', color: '#B45309' }}>{pending.length}</span>
          </div>
          {pending.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle2 className="w-10 h-10 mx-auto mb-3" style={{ color: '#22AA6A' }} />
              <p className="text-sm font-bold" style={{ color: '#78716C' }}>All caught up! No pending requests.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {pending.map((req, i) => (
                <motion.div key={req.id} initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 + i * 0.05 }} className="rounded-lg p-4 sm:p-5" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
                  <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                    <div className="flex items-start gap-3 flex-1 min-w-0">
                      <div className="w-10 h-10 rounded-md flex items-center justify-center shrink-0 text-sm font-black" style={{ background: 'linear-gradient(135deg, #FFA070, #FF5A1F)', color: '#FFFFFF' }}>
                        {req.requesterName.split(' ').map(n => n[0]).join('').slice(0, 2)}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-bold" style={{ color: '#1C1917' }}>{req.requesterName}</p>
                        <p className="text-[10px] font-semibold flex items-center gap-1" style={{ color: '#A8A29E' }}><Mail className="w-3 h-3" /> {req.requesterEmail}</p>
                        <p className="text-xs font-semibold mt-2 px-3 py-2 rounded-md" style={{ background: '#F4F2EF', color: '#57534E' }}>
                          {`"${req.reason}"`}
                        </p>
                        <p className="text-[10px] font-semibold mt-1.5 flex items-center gap-1" style={{ color: '#D6D3D1' }}>
                          <Clock className="w-3 h-3" /> {new Date(req.requestedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <motion.button whileTap={{ scale: 0.95 }} onClick={() => approve(req.id)} className="flex items-center gap-1.5 px-4 py-2.5 rounded-md text-xs font-extrabold transition-all" style={{ background: 'linear-gradient(135deg, #3DC484, #22AA6A)', color: '#fff', boxShadow: '0 3px 10px rgba(34,170,106,0.25)' }}>
                        <Check className="w-4 h-4" /> Approve
                      </motion.button>
                      <motion.button whileTap={{ scale: 0.95 }} onClick={() => reject(req.id)} className="flex items-center gap-1.5 px-4 py-2.5 rounded-md text-xs font-extrabold transition-all" style={{ background: '#FFE4E6', color: '#BE123C', border: '1.5px solid #FECDD3' }}>
                        <X className="w-4 h-4" /> Reject
                      </motion.button>
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          )}
        </motion.div>

        {/* Resolved section */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="rounded-xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
          <h2 className="text-sm font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>History</h2>
          <div className="space-y-2.5">
            {resolved.map((req, i) => (
              <div key={req.id} className="flex items-center gap-3 px-4 py-3 rounded-lg" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
                <div className="w-8 h-8 rounded-sm flex items-center justify-center shrink-0" style={{ background: req.status === 'Approved' ? '#D5F5E6' : '#FFE4E6' }}>
                  {req.status === 'Approved' ? <CheckCircle2 className="w-4 h-4" style={{ color: '#116B3F' }} /> : <XCircle className="w-4 h-4" style={{ color: '#BE123C' }} />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-bold truncate" style={{ color: '#1C1917' }}>{req.requesterName}</p>
                  <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>
                    {req.status} · {req.reviewedAt ? new Date(req.reviewedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : ''}
                  </p>
                </div>
                <span className="text-[9px] font-extrabold px-2.5 py-1 rounded-full uppercase" style={{ background: req.status === 'Approved' ? '#D5F5E6' : '#FFE4E6', color: req.status === 'Approved' ? '#116B3F' : '#BE123C' }}>{req.status}</span>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </DashboardShell>
  )
}
