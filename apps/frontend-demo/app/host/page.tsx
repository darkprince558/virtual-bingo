'use client'

import { useState, useEffect } from 'react'
import { motion } from 'motion/react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { DashboardShell } from '@/components/DashboardShell'
import { apiClient } from '@/lib/apiClient'
import type { GameRunResponse } from '@/types/api'
import { mockGameTemplates, mockWordSet } from '@/lib/mockAdminData'
import {
  CalendarClock,
  ChevronRight,
  Radio,
  Sparkles,
  Users,
  Play,
  Clock,
  Plus,
  ArrowRight,
  CheckCircle2,
  AlertCircle,
  Zap,
} from 'lucide-react'

const RUN_STATUS_STYLES: Record<string, { bg: string; color: string; dot: string }> = {
  'Scheduled':           { bg: '#EDE5FF', color: '#6440E8', dot: '#7C5CFC' },
  'Content Generating':  { bg: '#FEF3C7', color: '#B45309', dot: '#F59E0B' },
  'Content Review':      { bg: '#FFE4D9', color: '#C23208', dot: '#FF5A1F' },
  'Invites Sent':        { bg: '#D5F5E6', color: '#116B3F', dot: '#22AA6A' },
  'Lobby Open':          { bg: '#D5F5E6', color: '#116B3F', dot: '#22AA6A' },
  'Live':                { bg: '#D5F5E6', color: '#116B3F', dot: '#22AA6A' },
  'Complete':            { bg: '#F4F2EF', color: '#78716C', dot: '#A8A29E' },
  'Cancelled':           { bg: '#FFE4E6', color: '#BE123C', dot: '#F43F5E' },
  'Failed':              { bg: '#FFE4E6', color: '#BE123C', dot: '#F43F5E' },
}

function formatRelativeDate(dateStr: string) {
  const d = new Date(dateStr)
  const now = new Date()
  const diff = d.getTime() - now.getTime()
  const days = Math.round(diff / 86400000)
  if (days === 0) return 'Today'
  if (days === 1) return 'Tomorrow'
  if (days === -1) return 'Yesterday'
  if (days > 0) return `In ${days} days`
  return `${Math.abs(days)} days ago`
}

export default function HostDashboardPage() {
  const router = useRouter()
  const [runs, setRuns] = useState<GameRunResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isCreating, setIsCreating] = useState(false)

  useEffect(() => {
    async function loadGames() {
      try {
        const data = await apiClient<GameRunResponse[]>('/games')
        setRuns(data)
      } catch (err) {
        console.error('Failed to load games:', err)
      } finally {
        setIsLoading(false)
      }
    }
    loadGames()
  }, [])

  const handleQuickStart = async () => {
    try {
      setIsCreating(true)
      const newGame = await apiClient<GameRunResponse>('/games', {
        method: 'POST',
        body: JSON.stringify({
          name: 'Quick Bingo Game',
          code: Math.random().toString(36).substring(2, 8).toUpperCase(),
        })
      })
      router.push(`/host/live?gameId=${newGame.id}`)
    } catch (err) {
      console.error('Failed to start game:', err)
      setIsCreating(false)
    }
  }

  const activeTemplates = mockGameTemplates.filter(t => t.isActive)
  const upcomingRuns = runs.filter(r => ['pending', 'scheduled'].includes(r.status.toLowerCase()))
  const liveRuns = runs.filter(r => ['live', 'paused'].includes(r.status.toLowerCase()))
  const recentRuns = runs.filter(r => ['finished', 'cancelled'].includes(r.status.toLowerCase())).slice(0, 3)

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">

        {/* Page header */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="mb-8"
        >
          <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1.5" style={{ color: '#A8A29E' }}>
            Host Dashboard
          </p>
          <h1 className="text-3xl sm:text-4xl font-black tracking-tight mb-2" style={{ color: '#1C1917' }}>
            Welcome back 👋
          </h1>
          <p className="text-sm font-semibold" style={{ color: '#78716C' }}>
            Manage your recurring games, review AI content, and launch live sessions.
          </p>
        </motion.div>

        {/* Quick stats row */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.05, duration: 0.4 }}
          className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8"
        >
          {[
            { icon: CalendarClock, label: 'Active Templates', value: activeTemplates.length, color: '#FF5A1F', bg: '#FFF4F0' },
            { icon: Radio, label: 'Live Now', value: liveRuns.length, color: '#22AA6A', bg: '#EDFAF5' },
            { icon: Clock, label: 'Upcoming', value: upcomingRuns.length, color: '#7C5CFC', bg: '#F5F2FF' },
            { icon: Sparkles, label: 'Needs Review', value: 1, color: '#F59E0B', bg: '#FFFBEB' },
          ].map((stat, i) => (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 + i * 0.05 }}
              className="rounded-3xl p-4 sm:p-5 flex flex-col gap-3"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.04)' }}
            >
              <div
                className="w-10 h-10 rounded-2xl flex items-center justify-center"
                style={{ background: stat.bg }}
              >
                <stat.icon className="w-5 h-5" style={{ color: stat.color }} />
              </div>
              <div>
                <p className="text-2xl font-black" style={{ color: '#1C1917' }}>{stat.value}</p>
                <p className="text-[10px] font-bold uppercase tracking-wider" style={{ color: '#A8A29E' }}>{stat.label}</p>
              </div>
            </motion.div>
          ))}
        </motion.div>

        {/* Main grid: 2 columns on desktop */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">

          {/* ─── Left Column ─── */}
          <div className="lg:col-span-7 flex flex-col gap-6">

            {/* Live game banner */}
            {liveRuns.length > 0 && (
              <motion.div
                initial={{ opacity: 0, scale: 0.98 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.15 }}
              >
                <Link
                  href="/host/live"
                  className="block rounded-3xl p-5 sm:p-6 transition-all group"
                  style={{
                    background: 'linear-gradient(135deg, #EDFAF5 0%, #D5F5E6 100%)',
                    border: '1.5px solid #A8EBCC',
                  }}
                >
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2.5">
                      <div className="w-3 h-3 rounded-full animate-pulse" style={{ background: '#22AA6A' }} />
                      <span className="text-xs font-extrabold uppercase tracking-wider" style={{ color: '#116B3F' }}>
                        Live Now
                      </span>
                    </div>
                    <ChevronRight className="w-5 h-5 transition-transform group-hover:translate-x-1" style={{ color: '#22AA6A' }} />
                  </div>
                  <h3 className="text-lg font-black mb-1" style={{ color: '#0D512F' }}>
                    {liveRuns[0].name}
                  </h3>
                  <p className="text-sm font-semibold" style={{ color: '#178A53' }}>
                    {liveRuns[0].allowedPlayerCount} players allowed &middot; Join with code: {liveRuns[0].code}
                  </p>
                  <div className="mt-4 flex items-center gap-2">
                    <Play className="w-4 h-4" style={{ color: '#116B3F' }} />
                    <span className="text-sm font-bold" style={{ color: '#116B3F' }}>
                      Go to Live Control Center
                    </span>
                  </div>
                </Link>
              </motion.div>
            )}

            {/* Content review alert */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <Link
                href="/host/review"
                className="block rounded-3xl p-5 transition-all group"
                style={{
                  background: 'linear-gradient(135deg, #FFFBEB 0%, #FEF3C7 100%)',
                  border: '1.5px solid #FDE68A',
                }}
              >
                <div className="flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0"
                    style={{ background: 'linear-gradient(135deg, #FBBF24, #F59E0B)', boxShadow: '0 4px 12px rgba(245,158,11,0.25)' }}
                  >
                    <Sparkles className="w-6 h-6 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-sm font-extrabold" style={{ color: '#92400E' }}>AI Content Needs Review</h3>
                      <span className="text-[10px] font-black px-2 py-0.5 rounded-full" style={{ background: '#FDE68A', color: '#B45309' }}>
                        1 pending
                      </span>
                    </div>
                    <p className="text-xs font-semibold truncate" style={{ color: '#B45309' }}>
                      {mockWordSet.words.length} words generated for &ldquo;{mockWordSet.templateName}&rdquo; &mdash; tap to review &amp; approve.
                    </p>
                  </div>
                  <ChevronRight className="w-5 h-5 shrink-0 transition-transform group-hover:translate-x-1" style={{ color: '#D97706' }} />
                </div>
              </Link>
            </motion.div>

            {/* Upcoming runs */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.25 }}
              className="rounded-3xl p-5 sm:p-6"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <div className="flex items-center justify-between mb-5">
                <h2 className="text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
                  Upcoming Runs
                </h2>
                <span className="text-xs font-bold px-2.5 py-1 rounded-full" style={{ background: '#F4F2EF', color: '#78716C' }}>
                  {upcomingRuns.length}
                </span>
              </div>
              {upcomingRuns.length === 0 ? (
                <p className="text-sm font-semibold text-center py-8" style={{ color: '#A8A29E' }}>
                  No upcoming runs. Create a template to get started.
                </p>
              ) : (
                <div className="space-y-3">
                  {upcomingRuns.map((run, i) => {
                    const style = RUN_STATUS_STYLES[run.status] || RUN_STATUS_STYLES['Scheduled']
                    return (
                      <motion.div
                        key={run.id}
                        initial={{ opacity: 0, x: -12 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: 0.3 + i * 0.06 }}
                        className="flex items-center gap-4 px-4 py-3.5 rounded-2xl"
                        style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                      >
                        <div
                          className="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
                          style={{ background: style.bg }}
                        >
                          <CalendarClock className="w-5 h-5" style={{ color: style.color }} />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-bold truncate" style={{ color: '#1C1917' }}>{run.name}</p>
                          <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>
                            Code: {run.code}
                          </p>
                        </div>
                        <span
                          className="text-[10px] font-extrabold px-2.5 py-1 rounded-full uppercase tracking-wide shrink-0"
                          style={{ background: style.bg, color: style.color }}
                        >
                          {run.status}
                        </span>
                      </motion.div>
                    )
                  })}
                </div>
              )}
            </motion.div>
          </div>

          {/* ─── Right Column ─── */}
          <div className="lg:col-span-5 flex flex-col gap-6">

            {/* Quick actions */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.15 }}
              className="rounded-3xl p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>
                Quick Actions
              </h2>
              <div className="space-y-2.5">
                <button
                  onClick={handleQuickStart}
                  disabled={isCreating}
                  className="w-full flex items-center gap-3 px-4 py-3.5 rounded-2xl transition-all group"
                  style={{ background: '#FFF4F0', border: '1.5px solid #FFE4D9', opacity: isCreating ? 0.7 : 1 }}
                >
                  <div className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
                    style={{ background: 'linear-gradient(135deg, #FF7A42, #FF5A1F)', boxShadow: '0 4px 10px rgba(255,90,31,0.25)' }}
                  >
                    <Zap className="w-4 h-4 text-white" />
                  </div>
                  <div className="flex-1 text-left">
                    <p className="text-sm font-bold" style={{ color: '#C23208' }}>
                      {isCreating ? 'Starting...' : 'Quick Start Game'}
                    </p>
                    <p className="text-[10px] font-semibold" style={{ color: '#FFA070' }}>Launch an immediate session</p>
                  </div>
                  <ChevronRight className="w-4 h-4 transition-transform group-hover:translate-x-1 shrink-0" style={{ color: '#FF5A1F' }} />
                </button>

                <Link
                  href="/host/live"
                  className="flex items-center gap-3 px-4 py-3.5 rounded-2xl transition-all group"
                  style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC' }}
                >
                  <div className="w-9 h-9 rounded-xl flex items-center justify-center"
                    style={{ background: 'linear-gradient(135deg, #3DC484, #22AA6A)', boxShadow: '0 4px 10px rgba(34,170,106,0.25)' }}
                  >
                    <Radio className="w-4 h-4 text-white" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-bold" style={{ color: '#0D512F' }}>Live Game Control</p>
                    <p className="text-[10px] font-semibold" style={{ color: '#22AA6A' }}>Manage active session</p>
                  </div>
                  <ChevronRight className="w-4 h-4 transition-transform group-hover:translate-x-1" style={{ color: '#22AA6A' }} />
                </Link>
              </div>
            </motion.div>

            {/* Active templates */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.25 }}
              className="rounded-3xl p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
                  My Templates
                </h2>
                <Link
                  href="/host/templates"
                  className="text-xs font-bold flex items-center gap-1 transition-all"
                  style={{ color: '#FF5A1F' }}
                >
                  View All <ChevronRight className="w-3.5 h-3.5" />
                </Link>
              </div>
              <div className="space-y-3">
                {mockGameTemplates.slice(0, 3).map((tmpl, i) => (
                  <motion.div
                    key={tmpl.id}
                    initial={{ opacity: 0, x: -12 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: 0.3 + i * 0.06 }}
                    className="flex items-center gap-3 px-4 py-3 rounded-2xl"
                    style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                  >
                    <div
                      className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 text-xs font-black"
                      style={{
                        background: tmpl.isActive ? 'linear-gradient(135deg, #FF7A42, #FF5A1F)' : '#E7E5E4',
                        color: tmpl.isActive ? '#FFFFFF' : '#A8A29E',
                      }}
                    >
                      {tmpl.name.charAt(0)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-bold truncate" style={{ color: '#1C1917' }}>{tmpl.name}</p>
                      <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>
                        {tmpl.recurrence} · {tmpl.dayOfWeek} · {tmpl.time}
                      </p>
                    </div>
                    <span
                      className="text-[9px] font-extrabold px-2 py-0.5 rounded-full uppercase"
                      style={{
                        background: tmpl.isActive ? '#D5F5E6' : '#F4F2EF',
                        color: tmpl.isActive ? '#116B3F' : '#A8A29E',
                      }}
                    >
                      {tmpl.isActive ? 'Active' : 'Paused'}
                    </span>
                  </motion.div>
                ))}
              </div>
            </motion.div>

            {/* Recent game history */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.35 }}
              className="rounded-3xl p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>
                Recent Games
              </h2>
              <div className="space-y-3">
                {recentRuns.map((run, i) => {
                  const style = RUN_STATUS_STYLES[run.status] || RUN_STATUS_STYLES['Complete']
                  return (
                    <div
                      key={run.id}
                      className="flex items-center gap-3 px-4 py-3 rounded-2xl"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                    >
                      <div className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
                        style={{ background: style.bg }}
                      >
                        {run.status === 'Complete' ? (
                          <CheckCircle2 className="w-4 h-4" style={{ color: style.color }} />
                        ) : (
                          <AlertCircle className="w-4 h-4" style={{ color: style.color }} />
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-bold truncate" style={{ color: '#1C1917' }}>{run.name}</p>
                        <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>
                          Code: {run.code}
                        </p>
                      </div>
                    </div>
                  )
                })}
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </DashboardShell>
  )
}
