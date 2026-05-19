'use client'

import { motion } from 'motion/react'
import Link from 'next/link'
import { DashboardShell } from '@/components/DashboardShell'
import { mockAdminStats, mockGameRuns } from '@/lib/mockAdminData'
import { Users, Radio, CalendarClock, Shield, Mic, Gift, ChevronRight, CheckCircle2, AlertCircle, Clock, XCircle } from 'lucide-react'

const STATUS_STYLES: Record<string, { bg: string; color: string }> = {
  'Scheduled':          { bg: '#EDE5FF', color: '#6440E8' },
  'Content Generating': { bg: '#FEF3C7', color: '#B45309' },
  'Content Review':     { bg: '#FFE4D9', color: '#C23208' },
  'Invites Sent':       { bg: '#D5F5E6', color: '#116B3F' },
  'Lobby Open':         { bg: '#D5F5E6', color: '#116B3F' },
  'Live':               { bg: '#D5F5E6', color: '#116B3F' },
  'Complete':           { bg: '#F4F2EF', color: '#78716C' },
  'Cancelled':          { bg: '#FFE4E6', color: '#BE123C' },
  'Failed':             { bg: '#FFE4E6', color: '#BE123C' },
}

export default function AdminDashboard() {
  const stats = mockAdminStats

  return (
    <DashboardShell role="admin" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1.5" style={{ color: '#A8A29E' }}>Admin Center</p>
          <h1 className="text-3xl sm:text-4xl font-black tracking-tight mb-2" style={{ color: '#1C1917' }}>Operations Overview</h1>
          <p className="text-sm font-semibold" style={{ color: '#78716C' }}>System-wide visibility into games, hosts, and voice profiles.</p>
        </motion.div>

        {/* Stats grid */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-8">
          {[
            { icon: CalendarClock, label: 'Games This Month', value: stats.totalGamesThisMonth, color: '#FF5A1F', bg: '#FFF4F0' },
            { icon: Radio, label: 'Active Now', value: stats.activeGamesNow, color: '#22AA6A', bg: '#EDFAF5' },
            { icon: Users, label: 'Total Players', value: stats.totalPlayers, color: '#7C5CFC', bg: '#F5F2FF' },
            { icon: Shield, label: 'Host Requests', value: stats.pendingHostRequests, color: '#F59E0B', bg: '#FFFBEB' },
            { icon: Mic, label: 'Voice Pending', value: stats.pendingVoiceApprovals, color: '#E8440A', bg: '#FFF4F0' },
            { icon: Gift, label: 'Rewards Sent', value: stats.rewardsFulfilled, color: '#22AA6A', bg: '#EDFAF5' },
          ].map((s, i) => (
            <motion.div key={s.label} initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.08 + i * 0.04 }} className="rounded-2xl p-4 flex flex-col items-center gap-2" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.03)' }}>
              <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={{ background: s.bg }}><s.icon className="w-4 h-4" style={{ color: s.color }} /></div>
              <p className="text-xl font-black" style={{ color: '#1C1917' }}>{s.value}</p>
              <p className="text-[9px] font-bold uppercase tracking-wide text-center" style={{ color: '#A8A29E' }}>{s.label}</p>
            </motion.div>
          ))}
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Game runs table */}
          <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 }} className="lg:col-span-8 rounded-3xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
            <h2 className="text-sm font-extrabold uppercase tracking-widest mb-5" style={{ color: '#A8A29E' }}>All Game Runs</h2>
            <div className="space-y-2.5">
              {mockGameRuns.map((run, i) => {
                const st = STATUS_STYLES[run.status] || STATUS_STYLES['Complete']
                return (
                  <motion.div key={run.id} initial={{ opacity: 0, x: -12 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: 0.2 + i * 0.04 }} className="flex items-center gap-3 px-4 py-3.5 rounded-2xl" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
                    <div className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0" style={{ background: st.bg }}>
                      {run.status === 'Live' ? <Radio className="w-4 h-4" style={{ color: st.color }} /> : run.status === 'Complete' ? <CheckCircle2 className="w-4 h-4" style={{ color: st.color }} /> : run.status === 'Cancelled' || run.status === 'Failed' ? <XCircle className="w-4 h-4" style={{ color: st.color }} /> : <Clock className="w-4 h-4" style={{ color: st.color }} />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-bold truncate" style={{ color: '#1C1917' }}>{run.templateName}</p>
                      <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>Host: {run.hostName} · {new Date(run.scheduledAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} · {run.playerCount} players</p>
                    </div>
                    <span className="text-[9px] font-extrabold px-2.5 py-1 rounded-full uppercase tracking-wide shrink-0" style={{ background: st.bg, color: st.color }}>{run.status}</span>
                  </motion.div>
                )
              })}
            </div>
          </motion.div>

          {/* Quick actions */}
          <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="lg:col-span-4 flex flex-col gap-5">
            <div className="rounded-3xl p-5" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>Actions Required</h2>
              <div className="space-y-2.5">
                <Link href="/admin/requests" className="flex items-center gap-3 px-4 py-3.5 rounded-2xl transition-all group" style={{ background: '#FFFBEB', border: '1.5px solid #FDE68A' }}>
                  <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #FBBF24, #F59E0B)' }}><Shield className="w-4 h-4 text-white" /></div>
                  <div className="flex-1"><p className="text-sm font-bold" style={{ color: '#92400E' }}>Host Requests</p><p className="text-[10px] font-semibold" style={{ color: '#B45309' }}>{stats.pendingHostRequests} pending approvals</p></div>
                  <ChevronRight className="w-4 h-4 transition-transform group-hover:translate-x-1" style={{ color: '#D97706' }} />
                </Link>
                <Link href="/admin/voices" className="flex items-center gap-3 px-4 py-3.5 rounded-2xl transition-all group" style={{ background: '#F5F2FF', border: '1.5px solid #D9CCFF' }}>
                  <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)' }}><Mic className="w-4 h-4 text-white" /></div>
                  <div className="flex-1"><p className="text-sm font-bold" style={{ color: '#4F30C2' }}>Voice Profiles</p><p className="text-[10px] font-semibold" style={{ color: '#7C5CFC' }}>{stats.pendingVoiceApprovals} awaiting consent</p></div>
                  <ChevronRight className="w-4 h-4 transition-transform group-hover:translate-x-1" style={{ color: '#7C5CFC' }} />
                </Link>
              </div>
            </div>

            {/* Security notice */}
            <div className="rounded-3xl p-5" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-3" style={{ color: '#A8A29E' }}>Audit & Security</h2>
              <div className="flex items-start gap-3 px-4 py-3 rounded-2xl" style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC' }}>
                <CheckCircle2 className="w-5 h-5 shrink-0 mt-0.5" style={{ color: '#22AA6A' }} />
                <div><p className="text-xs font-bold" style={{ color: '#0D512F' }}>All systems nominal</p><p className="text-[10px] font-semibold" style={{ color: '#178A53' }}>No failed deliveries, reward issues, or voice consent violations detected.</p></div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </DashboardShell>
  )
}
