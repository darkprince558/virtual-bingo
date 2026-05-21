'use client'

import { useState, useEffect } from 'react'
import { motion } from 'motion/react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { DashboardShell } from '@/components/DashboardShell'
import { apiClient } from '@/lib/apiClient'
import { displayBackendValue } from '@/lib/uiMappers'
import type { AllowedPlayerResponse, GameRunResponse, GameSettingsResponse, WordSetResponse } from '@/types/api'
import { mockGameTemplates, mockWordSet } from '@/lib/mockAdminData'
import {
  CalendarClock,
  ChevronRight,
  Radio,
  Sparkles,
  Users,
  Play,
  Clock,
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

export default function HostDashboardPage() {
  const router = useRouter()
  const [runs, setRuns] = useState<GameRunResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isCreating, setIsCreating] = useState(false)
  const [apiError, setApiError] = useState<string | null>(null)
  const [setupGameForm, setSetupGameForm] = useState({ name: '', code: '', winningPattern: 'single_line', wordSetId: '' })
  const [setupSettingsForm, setSetupSettingsForm] = useState({ markingMode: 'manual', allowPlayerMarkingModeChoice: false, showClaimReadiness: true })
  const [wordSets, setWordSets] = useState<WordSetResponse[]>([])
  const [allowedPlayers, setAllowedPlayers] = useState<AllowedPlayerResponse[]>([])
  const [newPlayersText, setNewPlayersText] = useState('')
  const [setupNotice, setSetupNotice] = useState<string | null>(null)
  const [isSavingSetup, setIsSavingSetup] = useState(false)
  const [themeForm, setThemeForm] = useState({ prompt: '', tone: 'fun' })
  const [isGeneratingTheme, setIsGeneratingTheme] = useState(false)
  const [isSendingInvites, setIsSendingInvites] = useState(false)

  useEffect(() => {
    async function loadGames() {
      try {
        const data = await apiClient<GameRunResponse[]>('/games')
        setRuns(data)
        setApiError(null)
      } catch (err) {
        console.error('Failed to load games:', err)
        setApiError(err instanceof Error ? err.message : 'Failed to load games from the Go backend')
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
      setApiError(err instanceof Error ? err.message : 'Failed to create a quick game')
      setIsCreating(false)
    }
  }

  const activeTemplates = mockGameTemplates.filter(t => t.isActive)
  const upcomingRuns = runs.filter(r => ['pending', 'scheduled'].includes(r.status.toLowerCase()))
  const liveRuns = runs.filter(r => ['live', 'paused'].includes(r.status.toLowerCase()))
  const recentRuns = runs.filter(r => ['finished', 'cancelled'].includes(r.status.toLowerCase())).slice(0, 3)
  const setupRun = upcomingRuns[0] || liveRuns[0] || runs[0]

  useEffect(() => {
    if (!setupRun) return

    let cancelled = false

    async function loadSetup() {
      try {
        const [settings, sets, allowed] = await Promise.all([
          apiClient<GameSettingsResponse>(`/games/${setupRun.id}/settings`, { devUserRole: 'host' }),
          apiClient<WordSetResponse[]>('/word-sets', { devUserRole: 'host' }),
          apiClient<AllowedPlayerResponse[]>(`/games/${setupRun.id}/allowed-players`, { devUserRole: 'host' }),
        ])
        if (!cancelled) {
          setSetupGameForm({
            name: setupRun.name,
            code: setupRun.code,
            winningPattern: setupRun.winningPattern || 'single_line',
            wordSetId: setupRun.wordSetId || '',
          })
          setSetupSettingsForm({
            markingMode: settings.markingMode,
            allowPlayerMarkingModeChoice: settings.allowPlayerMarkingModeChoice,
            showClaimReadiness: settings.showClaimReadiness,
          })
          setWordSets(sets)
          setAllowedPlayers(allowed)
        }
      } catch (err) {
        console.error('Failed to load setup data:', err)
      }
    }

    void loadSetup()

    return () => {
      cancelled = true
    }
  }, [setupRun])

  const refreshGames = async () => {
    const data = await apiClient<GameRunResponse[]>('/games')
    setRuns(data)
  }

  const handleSaveSetup = async () => {
    if (!setupRun) return
    setIsSavingSetup(true)
    setSetupNotice(null)
    try {
      await apiClient<GameRunResponse>(`/games/${setupRun.id}`, {
        method: 'PATCH',
        devUserRole: 'host',
        body: JSON.stringify({
          name: setupGameForm.name,
          code: setupGameForm.code,
          winningPattern: setupGameForm.winningPattern,
          wordSetId: setupGameForm.wordSetId || undefined,
        })
      })
      await apiClient<GameSettingsResponse>(`/games/${setupRun.id}/settings`, {
        method: 'PATCH',
        devUserRole: 'host',
        body: JSON.stringify(setupSettingsForm)
      })
      await refreshGames()
      setSetupNotice('Game setup saved.')
    } catch (err) {
      setApiError(err instanceof Error ? err.message : 'Failed to save game setup')
    } finally {
      setIsSavingSetup(false)
    }
  }

  const handleAddPlayers = async () => {
    if (!setupRun || !newPlayersText.trim()) return
    setIsSavingSetup(true)
    try {
      const players = newPlayersText.split('\n').map(line => {
        const [first, second] = line.split(',').map(part => part.trim())
        const email = second || first
        return {
          email,
          displayName: second ? first : email.split('@')[0].replace(/[._-]+/g, ' '),
        }
      }).filter(player => player.email.includes('@'))
      const allowed = await apiClient<AllowedPlayerResponse[]>(`/games/${setupRun.id}/allowed-players/bulk`, {
        method: 'POST',
        devUserRole: 'host',
        body: JSON.stringify(players),
      })
      setAllowedPlayers(allowed)
      setNewPlayersText('')
      setSetupNotice(`${allowed.length} allowed player${allowed.length === 1 ? '' : 's'} saved.`)
      await refreshGames()
    } catch (err) {
      setApiError(err instanceof Error ? err.message : 'Failed to add allowed players')
    } finally {
      setIsSavingSetup(false)
    }
  }

  const handleGenerateTheme = async () => {
    if (!setupRun || !themeForm.prompt) return
    setIsGeneratingTheme(true)
    setSetupNotice(null)
    try {
      const theme = await apiClient<any>('/themes/generate', {
        method: 'POST',
        devUserRole: 'host',
        body: JSON.stringify({ gameRunId: setupRun.id, prompt: themeForm.prompt, tone: themeForm.tone })
      })
      await apiClient<any>(`/games/${setupRun.id}/theme`, {
        method: 'POST',
        devUserRole: 'host',
        body: JSON.stringify({ themeId: theme.id })
      })
      setSetupNotice(`AI Theme "${theme.name}" applied successfully!`)
      setThemeForm({ prompt: '', tone: 'fun' })
      await refreshGames()
    } catch (err) {
      setApiError(err instanceof Error ? err.message : 'Failed to generate theme')
    } finally {
      setIsGeneratingTheme(false)
    }
  }

  const handleSendInvites = async () => {
    if (!setupRun) return
    setIsSendingInvites(true)
    setSetupNotice(null)
    try {
      await apiClient(`/games/${setupRun.id}/deliveries/player-invites`, {
        method: 'POST',
        devUserRole: 'host'
      })
      setSetupNotice('Invites successfully sent to allowed players!')
    } catch (err) {
      setApiError(err instanceof Error ? err.message : 'Failed to send invites')
    } finally {
      setIsSendingInvites(false)
    }
  }

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

        {apiError && (
          <div className="mb-6 rounded-xl p-5" style={{ background: '#FFF1F2', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
            <p className="text-sm font-extrabold mb-1">Backend connection issue</p>
            <p className="text-sm font-semibold">{apiError}</p>
          </div>
        )}

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
            { icon: Sparkles, label: 'Needs Review', value: setupRun ? 1 : 0, color: '#F59E0B', bg: '#FFFBEB' },
          ].map((stat, i) => (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 + i * 0.05 }}
              className="rounded-xl p-4 sm:p-5 flex flex-col gap-3"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.04)' }}
            >
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center"
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
                  href={`/host/live?gameId=${liveRuns[0].id}`}
                  className="block rounded-xl p-5 sm:p-6 transition-all group"
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
                href={setupRun ? `/host/review?gameId=${setupRun.id}` : '/host'}
                className="block rounded-xl p-5 transition-all group"
                style={{
                  background: 'linear-gradient(135deg, #FFFBEB 0%, #FEF3C7 100%)',
                  border: '1.5px solid #FDE68A',
                }}
              >
                <div className="flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-lg flex items-center justify-center shrink-0"
                    style={{ background: 'linear-gradient(135deg, #FBBF24, #F59E0B)', boxShadow: '0 4px 12px rgba(245,158,11,0.25)' }}
                  >
                    <Sparkles className="w-6 h-6 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-sm font-extrabold" style={{ color: '#92400E' }}>{setupRun ? 'AI Content Needs Review' : 'Create a Game for AI Content'}</h3>
                      <span className="text-[10px] font-black px-2 py-0.5 rounded-full" style={{ background: '#FDE68A', color: '#B45309' }}>
                        1 pending
                      </span>
                    </div>
                    <p className="text-xs font-semibold truncate" style={{ color: '#B45309' }}>
                      {setupRun ? `${mockWordSet.words.length} words can be generated for "${setupRun.name}" - tap to prepare, review, and lock.` : 'Quick start a real backend game before running AI prep.'}
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
              className="rounded-xl p-5 sm:p-6"
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
                  {isLoading ? 'Loading runs...' : 'No upcoming runs. Create a template to get started.'}
                </p>
              ) : (
                <div className="space-y-3">
                  {upcomingRuns.map((run, i) => {
                    const displayStatus = displayBackendValue(run.status)
                    const style = RUN_STATUS_STYLES[displayStatus] || RUN_STATUS_STYLES['Scheduled']
                    return (
                      <motion.div
                        key={run.id}
                        initial={{ opacity: 0, x: -12 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: 0.3 + i * 0.06 }}
                        className="flex items-center gap-4 px-4 py-3.5 rounded-lg"
                        style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                      >
                        <div
                          className="w-10 h-10 rounded-md flex items-center justify-center shrink-0"
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
                          {displayStatus}
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
              className="rounded-xl p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>
                Quick Actions
              </h2>
              <div className="space-y-2.5">
                <button
                  onClick={handleQuickStart}
                  disabled={isCreating}
                  className="w-full flex items-center gap-3 px-4 py-3.5 rounded-lg transition-all group"
                  style={{ background: '#FFF4F0', border: '1.5px solid #FFE4D9', opacity: isCreating ? 0.7 : 1 }}
                >
                  <div className="w-9 h-9 rounded-md flex items-center justify-center shrink-0"
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

                {liveRuns[0] ? (
                  <Link
                    href={`/host/live?gameId=${liveRuns[0].id}`}
                    className="flex items-center gap-3 px-4 py-3.5 rounded-lg transition-all group"
                    style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC' }}
                  >
                    <div className="w-9 h-9 rounded-md flex items-center justify-center"
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
                ) : (
                  <div
                    className="flex items-center gap-3 px-4 py-3.5 rounded-lg"
                    style={{ background: '#F4F2EF', border: '1.5px solid #E7E5E4', opacity: 0.78 }}
                  >
                    <div className="w-9 h-9 rounded-md flex items-center justify-center" style={{ background: '#E7E5E4' }}>
                      <Radio className="w-4 h-4" style={{ color: '#A8A29E' }} />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-bold" style={{ color: '#78716C' }}>No live game selected</p>
                      <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>Quick start a game first</p>
                    </div>
                  </div>
                )}
              </div>
            </motion.div>

            {setupRun && (
              <motion.div
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                className="rounded-xl p-5"
                style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
              >
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>
                    Game Setup
                  </h2>
                  <Link href={`/host/review?gameId=${setupRun.id}`} className="text-xs font-bold flex items-center gap-1" style={{ color: '#FF5A1F' }}>
                    AI Review <ChevronRight className="w-3.5 h-3.5" />
                  </Link>
                </div>

                {setupNotice && (
                  <p className="mb-3 rounded-lg px-3 py-2 text-xs font-bold" style={{ background: '#EDFAF5', color: '#116B3F' }}>{setupNotice}</p>
                )}

                <div className="space-y-3">
                  <input
                    value={setupGameForm.name}
                    onChange={e => setSetupGameForm(prev => ({ ...prev, name: e.target.value }))}
                    className="w-full px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                    style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                    aria-label="Game name"
                  />
                  <div className="grid grid-cols-2 gap-2">
                    <input
                      value={setupGameForm.code}
                      onChange={e => setSetupGameForm(prev => ({ ...prev, code: e.target.value.toUpperCase() }))}
                      className="px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                      aria-label="Game code"
                    />
                    <select
                      value={setupGameForm.winningPattern}
                      onChange={e => setSetupGameForm(prev => ({ ...prev, winningPattern: e.target.value }))}
                      className="px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                      aria-label="Winning pattern"
                    >
                      <option value="single_line">Single Line</option>
                      <option value="four_corners">Four Corners</option>
                      <option value="full_house">Full House</option>
                    </select>
                  </div>
                  <select
                    value={setupGameForm.wordSetId}
                    onChange={e => setSetupGameForm(prev => ({ ...prev, wordSetId: e.target.value }))}
                    className="w-full px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                    style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                    aria-label="Word set"
                  >
                    <option value="">Use generated/assigned word set</option>
                    {wordSets.map(wordSet => (
                      <option key={wordSet.id} value={wordSet.id}>{wordSet.name}</option>
                    ))}
                  </select>
                  <div className="grid grid-cols-2 gap-2">
                    <select
                      value={setupSettingsForm.markingMode}
                      onChange={e => setSetupSettingsForm(prev => ({ ...prev, markingMode: e.target.value }))}
                      className="px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                      aria-label="Marking mode"
                    >
                      <option value="manual">Manual</option>
                      <option value="assist">Assist</option>
                      <option value="auto_mark">Auto Mark</option>
                    </select>
                    <label className="flex items-center gap-2 px-3 py-2.5 rounded-lg text-xs font-bold" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#78716C' }}>
                      <input
                        type="checkbox"
                        checked={setupSettingsForm.showClaimReadiness}
                        onChange={e => setSetupSettingsForm(prev => ({ ...prev, showClaimReadiness: e.target.checked }))}
                      />
                      Readiness
                    </label>
                  </div>
                  <label className="flex items-center gap-2 px-3 py-2.5 rounded-lg text-xs font-bold" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#78716C' }}>
                    <input
                      type="checkbox"
                      checked={setupSettingsForm.allowPlayerMarkingModeChoice}
                      onChange={e => setSetupSettingsForm(prev => ({ ...prev, allowPlayerMarkingModeChoice: e.target.checked }))}
                    />
                    Allow player marking mode choice
                  </label>
                  <button onClick={handleSaveSetup} disabled={isSavingSetup} className="w-full py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FFF4F0', color: '#C23208', opacity: isSavingSetup ? 0.65 : 1 }}>
                    {isSavingSetup ? 'Saving...' : 'Save Game Setup'}
                  </button>
                </div>

                <div className="mt-5 pt-4" style={{ borderTop: '1px solid #F4F2EF' }}>
                  <p className="text-xs font-extrabold mb-2" style={{ color: '#A8A29E' }}>Allowed Players ({allowedPlayers.length})</p>
                  <textarea
                    value={newPlayersText}
                    onChange={e => setNewPlayersText(e.target.value)}
                    rows={3}
                    placeholder="Alex Demo, alex@example.local"
                    className="w-full px-3 py-2.5 rounded-lg text-sm font-bold outline-none resize-none"
                    style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                  />
                  <button onClick={handleAddPlayers} disabled={isSavingSetup || !newPlayersText.trim()} className="mt-2 w-full py-3 rounded-lg text-sm font-extrabold" style={{ background: '#EDFAF5', color: '#116B3F', opacity: isSavingSetup || !newPlayersText.trim() ? 0.55 : 1 }}>
                    Add Allowed Players
                  </button>
                  <button onClick={handleSendInvites} disabled={isSendingInvites || allowedPlayers.length === 0} className="mt-2 w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#F5F2FF', color: '#6440E8', opacity: isSendingInvites || allowedPlayers.length === 0 ? 0.55 : 1 }}>
                    {isSendingInvites ? 'Sending...' : 'Send Invites'}
                  </button>
                </div>
              </motion.div>
            )}

            {setupRun && (
              <motion.div
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.22 }}
                className="rounded-xl p-5"
                style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
              >
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-sm font-extrabold uppercase tracking-widest flex items-center gap-2" style={{ color: '#A8A29E' }}>
                    <Sparkles className="w-4 h-4 text-[#F59E0B]" /> AI Theme Generator
                  </h2>
                </div>
                <div className="space-y-3">
                  <label className="block">
                    <span className="text-xs font-extrabold" style={{ color: '#78716C' }}>Theme Prompt</span>
                    <input
                      value={themeForm.prompt}
                      onChange={e => setThemeForm(prev => ({ ...prev, prompt: e.target.value }))}
                      placeholder="e.g. Space Odyssey, 90s Office Party..."
                      className="mt-1 w-full px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                    />
                  </label>
                  <label className="block">
                    <span className="text-xs font-extrabold" style={{ color: '#78716C' }}>Tone</span>
                    <select
                      value={themeForm.tone}
                      onChange={e => setThemeForm(prev => ({ ...prev, tone: e.target.value }))}
                      className="mt-1 w-full px-3 py-2.5 rounded-lg text-sm font-bold outline-none"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                    >
                      <option value="fun">Fun & Casual</option>
                      <option value="professional">Professional</option>
                      <option value="spooky">Spooky</option>
                      <option value="retro">Retro</option>
                    </select>
                  </label>
                  <button onClick={handleGenerateTheme} disabled={isGeneratingTheme || !themeForm.prompt} className="w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FFF4F0', color: '#E8440A', opacity: isGeneratingTheme || !themeForm.prompt ? 0.65 : 1 }}>
                    <Sparkles className="w-4 h-4" />
                    {isGeneratingTheme ? 'Generating Theme...' : 'Generate & Apply Theme'}
                  </button>
                </div>
              </motion.div>
            )}

            {/* Active templates */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.25 }}
              className="rounded-xl p-5"
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
                    className="flex items-center gap-3 px-4 py-3 rounded-lg"
                    style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                  >
                    <div
                      className="w-9 h-9 rounded-md flex items-center justify-center shrink-0 text-xs font-black"
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
              className="rounded-xl p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-4" style={{ color: '#A8A29E' }}>
                Recent Games
              </h2>
              <div className="space-y-3">
                {recentRuns.map((run, i) => {
                  const displayStatus = displayBackendValue(run.status)
                  const style = RUN_STATUS_STYLES[displayStatus] || RUN_STATUS_STYLES['Complete']
                  return (
                    <Link
                      key={run.id}
                      href={`/summary?gameId=${run.id}`}
                      className="flex items-center gap-3 px-4 py-3 rounded-lg"
                      style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}
                    >
                      <div className="w-9 h-9 rounded-md flex items-center justify-center shrink-0"
                        style={{ background: style.bg }}
                      >
                        {displayStatus === 'Complete' || displayStatus === 'Finished' ? (
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
                    </Link>
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
