'use client'

import { useState, useEffect, type ReactNode } from 'react'
import { motion } from 'motion/react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { DashboardShell } from '@/components/DashboardShell'
import { apiClient } from '@/lib/apiClient'
import { displayBackendValue } from '@/lib/uiMappers'
import type { AllowedPlayerResponse, GameRunResponse, GameSettingsResponse, WordSetResponse } from '@/types/api'
import { mockGameTemplates } from '@/lib/mockAdminData'
import {
  CalendarClock,
  ChevronRight,
  Radio,
  Sparkles,
  Play,
  Clock,
  CheckCircle2,
  AlertCircle,
  Zap,
  Settings2,
  Send,
  Lock,
  Search,
  UserPlus,
  X,
  Check,
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

const MARKING_MODE_HELP: Record<string, string> = {
  manual: 'Players mark their own card. Best for a classic bingo feel.',
  assist: 'Players still mark manually, but the app can show claim-readiness hints.',
  auto_mark: 'The backend marks matching called words automatically for eligible players.',
}

const WINNING_PATTERN_HELP: Record<string, string> = {
  single_line: 'A player wins with any full row, column, or diagonal.',
  four_corners: 'A player wins by marking the four corner squares.',
  full_house: 'A player wins when the entire card is marked.',
}

type AutomationForm = {
  gameName: string
  themePrompt: string
  players: string
  reviewBeforeLobby: boolean
}

const DEFAULT_AUTOMATION_FORM: AutomationForm = {
  gameName: 'Team Bingo',
  themePrompt: 'team wins, workplace shout-outs, and Friday energy',
  players: 'Alex Local, alex@example.local\nJamie Local, jamie@example.local',
  reviewBeforeLobby: true,
}

const LOCKED_SETUP_STATUSES = new Set(['live', 'paused', 'finished', 'cancelled', 'failed'])

function isEditableSetupRun(run: GameRunResponse) {
  return !LOCKED_SETUP_STATUSES.has(run.status.toLowerCase())
}

function formatHostError(err: unknown, fallback: string) {
  const message = err instanceof Error ? err.message : fallback

  if (message.includes('game can only be updated before it is live')) {
    return 'That game is already running, so setup edits are locked. Use Live Monitor for the current game, or create the next game from this panel.'
  }

  return message
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
  const [isSendingInvites, setIsSendingInvites] = useState(false)
  const [automationForm, setAutomationForm] = useState<AutomationForm>(DEFAULT_AUTOMATION_FORM)
  const [isRunningAutomation, setIsRunningAutomation] = useState(false)
  const [automationProgress, setAutomationProgress] = useState<string | null>(null)
  const [manualEditsOpen, setManualEditsOpen] = useState(false)
  const [rosterMode, setRosterMode] = useState<'individual' | 'mass'>('individual')
  const [rosterSearch, setRosterSearch] = useState('')
  const [rosterInput, setRosterInput] = useState('')
  const [showAutocomplete, setShowAutocomplete] = useState(false)
  const [activeStep, setActiveStep] = useState(0)

  const mockDirectory = [
    { name: 'Alex Local', email: 'alex@example.local' },
    { name: 'Jamie Local', email: 'jamie@example.local' },
    { name: 'Taylor Swift', email: 'taylor@example.local' },
    { name: 'Morgan Freeman', email: 'morgan@example.local' },
    { name: 'Jordan Carter', email: 'jordan@example.local' },
    { name: 'Casey Smith', email: 'casey@example.local' },
    { name: 'Riley Jones', email: 'riley@example.local' }
  ]

  const addedPlayers = automationForm.players.split('\n').filter(p => p.trim() !== '')
  const filteredDirectory = mockDirectory.filter(p => 
    (p.name.toLowerCase().includes(rosterInput.toLowerCase()) || p.email.toLowerCase().includes(rosterInput.toLowerCase())) &&
    !addedPlayers.some(ap => ap.toLowerCase().includes(p.email.toLowerCase()))
  )
  const displayPlayers = addedPlayers.filter(p => p.toLowerCase().includes(rosterSearch.toLowerCase()))

  const handleAddPlayer = (name: string, email: string) => {
    const entry = `${name}, ${email}`
    if (!addedPlayers.includes(entry)) {
      updateAutomationForm('players', automationForm.players ? automationForm.players + '\n' + entry : entry)
    }
    setRosterInput('')
    setShowAutocomplete(false)
  }

  const handleRemovePlayer = (entryToRemove: string) => {
    updateAutomationForm('players', addedPlayers.filter(p => p !== entryToRemove).join('\n'))
  }

  const getActiveStepIndex = () => {
    if (!isRunningAutomation) return -1
    if (!automationProgress) return 0
    if (['Creating game shell', 'Saving smart defaults'].includes(automationProgress)) return 0
    if (['Adding players', 'Generating theme'].includes(automationProgress)) return 1
    if (['Preparing words', 'Locking deck'].includes(automationProgress)) return 2
    if (['Sending invites', 'Opening lobby'].includes(automationProgress)) return 3
    return 0
  }

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
      setApiError(formatHostError(err, 'Failed to create a quick game'))
      setIsCreating(false)
    }
  }

  const activeTemplates = mockGameTemplates.filter(t => t.isActive)
  const upcomingRuns = runs.filter(r => ['pending', 'scheduled'].includes(r.status.toLowerCase()))
  const liveRuns = runs.filter(r => ['live', 'paused'].includes(r.status.toLowerCase()))
  const recentRuns = runs.filter(r => ['finished', 'cancelled'].includes(r.status.toLowerCase())).slice(0, 3)
  const editableRuns = runs.filter(isEditableSetupRun)
  const setupRun = upcomingRuns[0] || editableRuns[0]
  const liveRun = liveRuns[0]

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
          setAutomationForm(prev => ({
            ...prev,
            gameName: prev.gameName === DEFAULT_AUTOMATION_FORM.gameName ? setupRun.name : prev.gameName,
          }))
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
      setApiError(formatHostError(err, 'Failed to save game setup'))
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
      setApiError(formatHostError(err, 'Failed to add allowed players'))
    } finally {
      setIsSavingSetup(false)
    }
  }

  const updateAutomationForm = <K extends keyof AutomationForm>(key: K, value: AutomationForm[K]) => {
    setAutomationForm(prev => ({ ...prev, [key]: value }))
  }

  const handleAutomatedSetup = async () => {
    setIsRunningAutomation(true)
    setSetupNotice(null)
    setApiError(null)
    try {
      setAutomationProgress('Creating game shell')
      let run = setupRun
      if (!run) {
        run = await apiClient<GameRunResponse>('/games', {
          method: 'POST',
          devUserRole: 'host',
          body: JSON.stringify({
            name: automationForm.gameName.trim() || 'Team Bingo',
            code: generateJoinCode(),
            winningPattern: 'single_line',
          }),
        })
      } else {
        await apiClient<GameRunResponse>(`/games/${run.id}`, {
          method: 'PATCH',
          devUserRole: 'host',
          body: JSON.stringify({
            name: automationForm.gameName.trim() || setupGameForm.name,
            code: setupGameForm.code,
            winningPattern: setupGameForm.winningPattern,
            wordSetId: setupGameForm.wordSetId || undefined,
          }),
        })
      }

      setAutomationProgress('Saving smart defaults')
      await apiClient<GameSettingsResponse>(`/games/${run.id}/settings`, {
        method: 'PATCH',
        devUserRole: 'host',
        body: JSON.stringify({
          markingMode: 'auto_mark',
          allowPlayerMarkingModeChoice: false,
          showClaimReadiness: true,
          callerMode: 'tts',
          themeMode: automationForm.themePrompt.trim() ? 'ai_generated' : 'default',
        }),
      })

      const players = parseAllowedPlayers(automationForm.players)
      if (players.length > 0) {
        setAutomationProgress('Adding players')
        let allowed: AllowedPlayerResponse[]
        try {
          allowed = await apiClient<AllowedPlayerResponse[]>(`/games/${run.id}/allowed-players/bulk`, {
            method: 'POST',
            devUserRole: 'host',
            body: JSON.stringify(players),
          })
        } catch (err) {
          if (err instanceof Error && err.message.includes('already exists')) {
            allowed = await apiClient<AllowedPlayerResponse[]>(`/games/${run.id}/allowed-players`, {
              devUserRole: 'host',
            })
          } else {
            throw err
          }
        }
        setAllowedPlayers(allowed)
      }

      if (automationForm.themePrompt.trim()) {
        setAutomationProgress('Generating theme')
        const theme = await apiClient<{ id: string; name?: string }>('/themes/generate', {
          method: 'POST',
          devUserRole: 'host',
          body: JSON.stringify({ gameRunId: run.id, prompt: automationForm.themePrompt.trim(), tone: 'fun' }),
        })
        await apiClient(`/themes/${theme.id}/approve`, {
          method: 'POST',
          devUserRole: 'host',
        })
        await apiClient<GameSettingsResponse>(`/games/${run.id}/theme`, {
          method: 'POST',
          devUserRole: 'host',
          body: JSON.stringify({ themeId: theme.id }),
        })
      }

      setAutomationProgress('Preparing words')
      await apiClient(`/games/${run.id}/content/prepare`, {
        method: 'POST',
        devUserRole: 'host',
      })

      if (automationForm.reviewBeforeLobby) {
        await refreshGames()
        setSetupNotice('Game prepared. Review and edit the generated content next.')
        router.push(`/host/review?gameId=${run.id}`)
        return
      }

      setAutomationProgress('Locking deck')
      await apiClient(`/games/${run.id}/content/lock`, {
        method: 'POST',
        devUserRole: 'host',
      })
      await apiClient(`/games/${run.id}/caller-assets/generate`, {
        method: 'POST',
        devUserRole: 'host',
      })
      if (players.length > 0) {
        setAutomationProgress('Sending invites')
        await apiClient(`/games/${run.id}/deliveries/player-invites`, {
          method: 'POST',
          devUserRole: 'host',
        })
      }
      setAutomationProgress('Opening lobby')
      await apiClient<GameRunResponse>(`/games/${run.id}/lobby/open`, {
        method: 'POST',
        devUserRole: 'host',
      })
      await refreshGames()
      setSetupNotice('Game is ready in the lobby.')
      router.push(`/host/live?gameId=${run.id}`)
    } catch (err) {
      console.error('Failed to automate game setup:', err)
      setApiError(formatHostError(err, 'Failed to prepare the game automatically'))
    } finally {
      setIsRunningAutomation(false)
      setAutomationProgress(null)
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
      setApiError(formatHostError(err, 'Failed to send invites'))
    } finally {
      setIsSendingInvites(false)
    }
  }

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-5 lg:p-6 max-w-7xl mx-auto">

        {/* Page header */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"
        >
          <div>
            <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] mb-1" style={{ color: '#A8A29E' }}>
              Host Dashboard
            </p>
            <h1 className="text-2xl sm:text-3xl font-black tracking-tight mb-1" style={{ color: '#1C1917' }}>
              Game Control
            </h1>
            <p className="text-sm font-semibold" style={{ color: '#78716C' }}>
              Run setup, AI prep, invites, and live control from one place.
            </p>
          </div>
          <button
            onClick={handleQuickStart}
            disabled={isCreating}
            className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg text-sm font-extrabold"
            style={{ background: '#FF5A1F', color: '#FFFFFF', opacity: isCreating ? 0.7 : 1 }}
          >
            <Zap className="w-4 h-4" /> {isCreating ? 'Starting...' : 'Quick Start'}
          </button>
        </motion.div>

        {apiError && (
          <div className="mb-6 rounded-xl p-5" style={{ background: '#FFF1F2', border: '1.5px solid #FECDD3', color: '#BE123C' }}>
            <p className="text-sm font-extrabold mb-1">Game setup needs attention</p>
            <p className="text-sm font-semibold">{apiError}</p>
          </div>
        )}

        {/* Quick stats row */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.05, duration: 0.4 }}
          className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-5"
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
              className="rounded-lg p-3 flex items-center gap-3"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 12px rgba(0,0,0,0.04)' }}
            >
              <div
                className="w-9 h-9 rounded-md flex items-center justify-center shrink-0"
                style={{ background: stat.bg }}
              >
                <stat.icon className="w-5 h-5" style={{ color: stat.color }} />
              </div>
              <div>
                <p className="text-xl font-black leading-none" style={{ color: '#1C1917' }}>{stat.value}</p>
                <p className="text-[10px] font-bold uppercase tracking-wider" style={{ color: '#A8A29E' }}>{stat.label}</p>
              </div>
            </motion.div>
          ))}
        </motion.div>

        {/* Main grid: 2 columns on desktop */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">

          {/* ─── Left Column ─── */}
          <div className="lg:col-span-8 flex flex-col gap-4">

            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.12 }}
              className="rounded-xl p-4 sm:p-5"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="mb-2 inline-flex items-center gap-2 rounded-full px-2.5 py-1 text-[10px] font-extrabold uppercase tracking-wider" style={{ background: '#FFF4F0', color: '#C23208' }}>
                    <Sparkles className="h-3.5 w-3.5" /> Automated prep
                  </div>
                  <h2 className="text-xl font-black tracking-tight" style={{ color: '#1C1917' }}>
                    {setupRun ? `Prepare "${setupRun.name}"` : 'Create the next game'}
                  </h2>
                  <p className="mt-1 text-sm font-semibold" style={{ color: '#78716C' }}>
                    AI handles the defaults. Open manual edits only when something needs changing.
                  </p>
                </div>
                <div className="flex gap-2">
                  {setupRun && (
                    <Link href={`/host/review?gameId=${setupRun.id}`} className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-xs font-extrabold" style={{ background: '#FFFBEB', color: '#B45309' }}>
                      Review <ChevronRight className="h-3.5 w-3.5" />
                    </Link>
                  )}
                  {liveRun && (
                    <Link href={`/host/live?gameId=${liveRun.id}`} className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-xs font-extrabold" style={{ background: '#EDFAF5', color: '#116B3F' }}>
                      Live Monitor <Radio className="h-3.5 w-3.5" />
                    </Link>
                  )}
                </div>
              </div>

              {setupNotice && (
                <p className="mb-4 rounded-lg px-3 py-2 text-xs font-bold" style={{ background: '#EDFAF5', color: '#116B3F' }}>{setupNotice}</p>
              )}

              {/* Process Flow Roadmap */}
              <div className="mb-6 relative flex flex-col md:flex-row justify-between items-center gap-4 p-4 rounded-xl z-0" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8' }}>
                <div className="absolute top-1/2 left-8 right-8 h-0.5 bg-[#E7E5E4] -translate-y-1/2 hidden md:block -z-10" />
                {[
                  { number: '1', title: 'Setup Run', desc: 'Name & players' },
                  { number: '2', title: 'Theme Gen', desc: 'AI topics' },
                  { number: '3', title: 'Review Deck', desc: 'Launch strategy' },
                  { number: '4', title: 'Launch Lobby', desc: 'Finalize & start' },
                ].map((step, idx) => {
                  const currentDisplayIndex = isRunningAutomation ? getActiveStepIndex() : activeStep
                  const isCurrent = currentDisplayIndex === idx
                  const isCompleted = currentDisplayIndex > idx
                  const isUpcoming = currentDisplayIndex < idx

                  let circleBg = '#E7E5E4'
                  let circleColor = '#78716C'
                  let borderColor = '#E7E5E4'

                  if (isCompleted) {
                    circleBg = '#EDFAF5'
                    circleColor = '#116B3F'
                    borderColor = '#22AA6A'
                  } else if (isCurrent) {
                    circleBg = '#FFF4F0'
                    circleColor = '#FF5A1F'
                    borderColor = '#FF5A1F'
                  }

                  return (
                    <button
                      key={step.title}
                      onClick={() => !isRunningAutomation && setActiveStep(idx)}
                      disabled={isRunningAutomation}
                      className="relative z-10 flex flex-row md:flex-col items-center gap-3 md:gap-1.5 text-left md:text-center flex-1 w-full focus:outline-none group"
                    >
                      <div 
                        className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs shrink-0 border-2 transition-all ${isCurrent && isRunningAutomation ? 'animate-pulse' : ''} ${!isRunningAutomation ? 'group-hover:scale-105' : ''}`}
                        style={{ background: circleBg, color: circleColor, borderColor: borderColor }}
                      >
                        {isCompleted ? '✓' : step.number}
                      </div>
                      <div className="min-w-0">
                        <p className="text-xs font-black transition-colors" style={{ color: isUpcoming ? '#78716C' : '#1C1917' }}>{step.title}</p>
                        <p className="text-[10px] font-semibold truncate transition-colors" style={{ color: isUpcoming ? '#A8A29E' : '#78716C' }}>{step.desc}</p>
                      </div>
                    </button>
                  )
                })}
              </div>

              {/* Step-by-Step Interactive Form */}
              <div className="space-y-4">
                {/* Section 01: Core Details & Theme */}
                <div className="p-4 rounded-xl" style={{ border: '1.5px solid #F0EDE8', background: '#FFFFFF' }}>
                  <div className="flex items-center gap-2 mb-3">
                    <span className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-black text-white" style={{ background: '#FF5A1F' }}>1</span>
                    <h3 className="text-xs font-black uppercase tracking-wider" style={{ color: '#1C1917' }}>Game Identity & AI Theme</h3>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-1">
                      <label className="text-[10px] font-extrabold uppercase tracking-wider" style={{ color: '#A8A29E' }}>Game Name</label>
                      <input
                        value={automationForm.gameName}
                        onChange={e => updateAutomationForm('gameName', e.target.value)}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20"
                        placeholder="e.g. Monthly All-Hands Bingo"
                        style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                      />
                    </div>
                    <div className="space-y-1">
                      <label className="text-[10px] font-extrabold uppercase tracking-wider flex items-center gap-1.5" style={{ color: '#A8A29E' }}>
                        <Sparkles className="h-3 w-3 text-amber-500" /> AI Theme Topic Prompt
                      </label>
                      <input
                        value={automationForm.themePrompt}
                        onChange={e => updateAutomationForm('themePrompt', e.target.value)}
                        className="w-full rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20"
                        placeholder="e.g. office jokes, team wins, retro comments"
                        style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                      />
                    </div>
                  </div>
                </div>

                {/* Section 02: Invite Player Roster */}
                <div className="p-4 rounded-xl" style={{ border: '1.5px solid #F0EDE8', background: '#FFFFFF' }}>
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                      <span className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-black text-white" style={{ background: '#FF5A1F' }}>2</span>
                      <h3 className="text-xs font-black uppercase tracking-wider" style={{ color: '#1C1917' }}>Participant Roster</h3>
                    </div>
                    <div className="flex bg-[#F0EDE8] rounded-lg p-0.5">
                      <button
                        onClick={() => setRosterMode('individual')}
                        className={`px-3 py-1 rounded-md text-[10px] font-bold transition-all ${rosterMode === 'individual' ? 'bg-white text-[#1C1917] shadow-sm' : 'text-[#78716C] hover:text-[#1C1917]'}`}
                      >
                        Add & Search
                      </button>
                      <button
                        onClick={() => setRosterMode('mass')}
                        className={`px-3 py-1 rounded-md text-[10px] font-bold transition-all ${rosterMode === 'mass' ? 'bg-white text-[#1C1917] shadow-sm' : 'text-[#78716C] hover:text-[#1C1917]'}`}
                      >
                        Mass Paste
                      </button>
                    </div>
                  </div>

                  {rosterMode === 'mass' ? (
                    <div>
                      <p className="text-[10px] font-semibold mb-2" style={{ color: '#78716C' }}>
                        Type or paste one participant per line using the format: <code className="bg-orange-50 px-1 py-0.5 rounded text-orange-600 font-mono text-[10px]">Name, Email</code>
                      </p>
                      <textarea
                        value={automationForm.players}
                        onChange={e => updateAutomationForm('players', e.target.value)}
                        rows={5}
                        placeholder="Alex Local, alex@example.local&#10;Jamie Local, jamie@example.local"
                        className="w-full resize-none rounded-lg px-3.5 py-2.5 text-sm font-bold outline-none transition-all focus:ring-2 focus:ring-orange-500/20"
                        style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                      />
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {/* Add New Player Input */}
                      <div className="relative">
                        <div className="flex items-center bg-[#FAFAF9] rounded-lg px-3.5 py-2" style={{ border: '1.5px solid #F0EDE8' }}>
                          <UserPlus className="h-4 w-4 text-[#A8A29E] mr-2" />
                          <input
                            type="text"
                            value={rosterInput}
                            onChange={e => {
                              setRosterInput(e.target.value)
                              setShowAutocomplete(true)
                            }}
                            onFocus={() => setShowAutocomplete(true)}
                            placeholder="Add player by name or email..."
                            className="flex-1 bg-transparent text-sm font-bold outline-none text-[#1C1917] placeholder:text-[#A8A29E]"
                          />
                        </div>
                        
                        {/* Autocomplete Dropdown */}
                        {showAutocomplete && rosterInput && (
                          <div className="absolute top-full left-0 right-0 mt-1 bg-white rounded-lg shadow-xl z-20 max-h-48 overflow-y-auto" style={{ border: '1.5px solid #F0EDE8' }}>
                            {filteredDirectory.length > 0 ? (
                              filteredDirectory.map((player) => (
                                <div 
                                  key={player.email}
                                  onClick={() => handleAddPlayer(player.name, player.email)}
                                  className="flex flex-col px-3.5 py-2 hover:bg-[#FAFAF9] cursor-pointer transition-colors border-b border-[#F0EDE8] last:border-b-0"
                                >
                                  <span className="text-sm font-bold text-[#1C1917]">{player.name}</span>
                                  <span className="text-xs text-[#78716C]">{player.email}</span>
                                </div>
                              ))
                            ) : (
                              <div className="px-3.5 py-3 text-xs text-[#78716C] font-semibold flex items-center justify-between">
                                <span>No matches in directory.</span>
                                {rosterInput.includes('@') && (
                                  <button 
                                    onClick={() => handleAddPlayer(rosterInput.split('@')[0], rosterInput)}
                                    className="px-2 py-1 bg-orange-100 text-orange-700 rounded-md hover:bg-orange-200 transition-colors"
                                  >
                                    Add as New
                                  </button>
                                )}
                              </div>
                            )}
                          </div>
                        )}
                      </div>

                      {/* Added Players List & Search */}
                      <div className="pt-2 border-t border-[#F0EDE8]">
                        <div className="flex items-center justify-between mb-3">
                          <h4 className="text-[11px] font-extrabold uppercase tracking-wider text-[#A8A29E]">
                            Current Roster ({addedPlayers.length})
                          </h4>
                          <div className="relative">
                            <Search className="h-3 w-3 absolute left-2 top-1/2 -translate-y-1/2 text-[#A8A29E]" />
                            <input
                              type="text"
                              value={rosterSearch}
                              onChange={e => setRosterSearch(e.target.value)}
                              placeholder="Search roster..."
                              className="pl-6 pr-2 py-1 text-xs bg-[#FAFAF9] rounded-md outline-none text-[#1C1917] font-semibold w-32 focus:w-40 transition-all"
                              style={{ border: '1px solid #E7E5E4' }}
                            />
                          </div>
                        </div>

                        <div className="flex flex-wrap gap-2 max-h-40 overflow-y-auto">
                          {displayPlayers.length > 0 ? (
                            displayPlayers.map((player, idx) => (
                              <div key={idx} className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-[#EDFAF5] border border-[#22AA6A]/20">
                                <span className="text-xs font-bold text-[#116B3F]">{player.split(',')[0]}</span>
                                <button 
                                  onClick={() => handleRemovePlayer(player)}
                                  className="p-0.5 rounded-full hover:bg-[#22AA6A]/20 text-[#116B3F] transition-colors"
                                >
                                  <X className="h-3 w-3" />
                                </button>
                              </div>
                            ))
                          ) : (
                            <p className="text-xs font-semibold text-[#A8A29E] w-full text-center py-4">
                              {rosterSearch ? 'No players found.' : 'No players added yet.'}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {/* Section 03: Launch Strategy */}
                <div className="p-4 rounded-xl" style={{ border: '1.5px solid #F0EDE8', background: '#FFFFFF' }}>
                  <div className="flex items-center gap-2 mb-3">
                    <span className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-black text-white" style={{ background: '#FF5A1F' }}>3</span>
                    <h3 className="text-xs font-black uppercase tracking-wider" style={{ color: '#1C1917' }}>Lobby Launch Strategy</h3>
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
                    {/* Strategy A: Interactive Review */}
                    <div 
                      onClick={() => updateAutomationForm('reviewBeforeLobby', true)}
                      className={`cursor-pointer rounded-xl p-4 border-2 transition-all flex flex-col justify-between hover:border-orange-500/50 ${automationForm.reviewBeforeLobby ? 'border-orange-500 bg-orange-50/10' : 'border-gray-200'}`}
                    >
                      <div>
                        <div className="flex items-center justify-between mb-1.5">
                          <span className="text-xs font-black" style={{ color: '#1C1917' }}>Review Words First</span>
                          <div className={`w-3.5 h-3.5 rounded-full border-2 flex items-center justify-center ${automationForm.reviewBeforeLobby ? 'border-orange-500 bg-orange-500' : 'border-gray-300 bg-white'}`}>
                            {automationForm.reviewBeforeLobby && <div className="w-1.5 h-1.5 rounded-full bg-white" />}
                          </div>
                        </div>
                        <p className="text-[11px] font-semibold leading-relaxed" style={{ color: '#57534E' }}>
                          Let AI generate a customized list of 30 vocabulary words. You will be taken to a review board to inspect and edit them before starting the game.
                        </p>
                      </div>
                      <span className="mt-3 text-[9px] font-extrabold uppercase tracking-wide px-2 py-0.5 rounded-full self-start" style={{ background: '#FFF4F0', color: '#C23208' }}>
                        Recommended
                      </span>
                    </div>

                    {/* Strategy B: Instant Launch */}
                    <div 
                      onClick={() => updateAutomationForm('reviewBeforeLobby', false)}
                      className={`cursor-pointer rounded-xl p-4 border-2 transition-all flex flex-col justify-between hover:border-indigo-500/50 ${!automationForm.reviewBeforeLobby ? 'border-indigo-500 bg-indigo-50/10' : 'border-gray-200'}`}
                    >
                      <div>
                        <div className="flex items-center justify-between mb-1.5">
                          <span className="text-xs font-black" style={{ color: '#1C1917' }}>Instant Lobby Launch</span>
                          <div className={`w-3.5 h-3.5 rounded-full border-2 flex items-center justify-center ${!automationForm.reviewBeforeLobby ? 'border-indigo-500 bg-indigo-500' : 'border-gray-300 bg-white'}`}>
                            {!automationForm.reviewBeforeLobby && <div className="w-1.5 h-1.5 rounded-full bg-white" />}
                          </div>
                        </div>
                        <p className="text-[11px] font-semibold leading-relaxed" style={{ color: '#57534E' }}>
                          Generates custom vocabulary words and locks them, issues player invitations via email, and launches straight into the live player lobby immediately.
                        </p>
                      </div>
                      <span className="mt-3 text-[9px] font-extrabold uppercase tracking-wide px-2 py-0.5 rounded-full self-start" style={{ background: '#F5F2FF', color: '#6440E8' }}>
                        Fast Track
                      </span>
                    </div>
                  </div>

                  {/* Actions & Status Feed */}
                  <div className="pt-2">
                    <button
                      onClick={handleAutomatedSetup}
                      disabled={isRunningAutomation}
                      className="inline-flex w-full items-center justify-center gap-2 rounded-xl py-3.5 text-sm font-extrabold shadow-md hover:shadow-lg transition-all"
                      style={{
                        background: isRunningAutomation ? '#E7E5E4' : 'linear-gradient(135deg, #FF7A42, #FF5A1F)',
                        color: '#FFFFFF',
                        opacity: isRunningAutomation ? 0.72 : 1,
                      }}
                    >
                      <Zap className={`h-4 w-4 ${isRunningAutomation ? 'animate-spin' : ''}`} />
                      {isRunningAutomation ? (automationProgress || 'Preparing game...') : 'Generate & Prepare Live Game'}
                    </button>

                    {/* Progress details indicator */}
                    {isRunningAutomation && automationProgress && (
                      <div className="mt-3 p-3 rounded-lg flex items-center gap-2.5 justify-center animate-pulse" style={{ background: '#FFF4F0', border: '1px solid #FFE4D9' }}>
                        <div className="w-2 h-2 rounded-full bg-orange-500" />
                        <span className="text-xs font-bold" style={{ color: '#C23208' }}>
                          AI Engine: {automationProgress}...
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              </div>

              <div className="mt-4 rounded-lg" style={{ border: '1.5px solid #F0EDE8' }}>
                <button
                  onClick={() => setManualEditsOpen(open => !open)}
                  className="flex w-full items-center justify-between px-3 py-2.5 text-left text-xs font-extrabold uppercase tracking-wider"
                  style={{ color: '#78716C' }}
                >
                  <span className="inline-flex items-center gap-2"><Settings2 className="h-4 w-4" /> Manual edits</span>
                  <ChevronRight className={`h-4 w-4 transition-transform ${manualEditsOpen ? 'rotate-90' : ''}`} />
                </button>

                {manualEditsOpen && (
                  <div className="border-t p-3" style={{ borderColor: '#F0EDE8' }}>
                    {!setupRun ? (
                      <p className="text-sm font-semibold" style={{ color: '#A8A29E' }}>Create a game first to edit advanced settings.</p>
                    ) : (
                      <div className="space-y-3">
                        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                          <FieldHelp label="Join code" help="The short code players use to find this game from the lobby.">
                            <input
                              value={setupGameForm.code}
                              onChange={e => setSetupGameForm(prev => ({ ...prev, code: e.target.value.toUpperCase() }))}
                              className="w-full rounded-md px-3 py-2 text-sm font-bold outline-none"
                              style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                            />
                          </FieldHelp>
                          <FieldHelp label="Winning pattern" help={WINNING_PATTERN_HELP[setupGameForm.winningPattern] || 'Controls which marked card pattern counts as a win.'}>
                            <select
                              value={setupGameForm.winningPattern}
                              onChange={e => setSetupGameForm(prev => ({ ...prev, winningPattern: e.target.value }))}
                              className="w-full rounded-md px-3 py-2 text-sm font-bold outline-none"
                              style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                            >
                              <option value="single_line">Single Line</option>
                              <option value="four_corners">Four Corners</option>
                              <option value="full_house">Full House</option>
                            </select>
                          </FieldHelp>
                          <FieldHelp label="Word set" help="Choose the words that fill player cards and the caller deck.">
                            <select
                              value={setupGameForm.wordSetId}
                              onChange={e => setSetupGameForm(prev => ({ ...prev, wordSetId: e.target.value }))}
                              className="w-full rounded-md px-3 py-2 text-sm font-bold outline-none"
                              style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                            >
                              <option value="">Use generated/assigned word set</option>
                              {wordSets.map(wordSet => (
                                <option key={wordSet.id} value={wordSet.id}>{wordSet.name}</option>
                              ))}
                            </select>
                          </FieldHelp>
                          <FieldHelp label="Marking mode" help={MARKING_MODE_HELP[setupSettingsForm.markingMode] || 'Controls how player card cells get marked during the game.'}>
                            <select
                              value={setupSettingsForm.markingMode}
                              onChange={e => setSetupSettingsForm(prev => ({ ...prev, markingMode: e.target.value }))}
                              className="w-full rounded-md px-3 py-2 text-sm font-bold outline-none"
                              style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                            >
                              <option value="manual">Manual Marking</option>
                              <option value="assist">Assist Hints</option>
                              <option value="auto_mark">Auto Mark</option>
                            </select>
                          </FieldHelp>
                        </div>
                        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                          <label className="flex min-h-[40px] items-center gap-2 rounded-md px-3 py-2 text-xs font-bold" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#78716C' }}>
                            <input
                              type="checkbox"
                              checked={setupSettingsForm.showClaimReadiness}
                              onChange={e => setSetupSettingsForm(prev => ({ ...prev, showClaimReadiness: e.target.checked }))}
                            />
                            Claim readiness hints
                          </label>
                          <label className="flex min-h-[40px] items-center gap-2 rounded-md px-3 py-2 text-xs font-bold" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#78716C' }}>
                            <input
                              type="checkbox"
                              checked={setupSettingsForm.allowPlayerMarkingModeChoice}
                              onChange={e => setSetupSettingsForm(prev => ({ ...prev, allowPlayerMarkingModeChoice: e.target.checked }))}
                            />
                            Player marking choice
                          </label>
                        </div>
                        <FieldHelp label="Add players" help="Paste one player per line, using Name, email.">
                          <textarea
                            value={newPlayersText}
                            onChange={e => setNewPlayersText(e.target.value)}
                            rows={2}
                            placeholder="Alex Local, alex@example.local"
                            className="w-full resize-none rounded-md px-3 py-2 text-sm font-bold outline-none"
                            style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }}
                          />
                        </FieldHelp>
                        <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                          <button onClick={handleSaveSetup} disabled={isSavingSetup} className="rounded-lg py-2.5 text-xs font-extrabold sm:col-span-1" style={{ background: '#FFF4F0', color: '#C23208', opacity: isSavingSetup ? 0.65 : 1 }}>
                            {isSavingSetup ? 'Saving...' : 'Save settings'}
                          </button>
                          <button onClick={handleAddPlayers} disabled={isSavingSetup || !newPlayersText.trim()} className="rounded-lg py-2.5 text-xs font-extrabold" style={{ background: '#EDFAF5', color: '#116B3F', opacity: isSavingSetup || !newPlayersText.trim() ? 0.55 : 1 }}>
                            Add pasted players
                          </button>
                          <button onClick={handleSendInvites} disabled={isSendingInvites || allowedPlayers.length === 0} className="inline-flex items-center justify-center gap-2 rounded-lg py-2.5 text-xs font-extrabold" style={{ background: '#F5F2FF', color: '#6440E8', opacity: isSendingInvites || allowedPlayers.length === 0 ? 0.55 : 1 }}>
                            {isSendingInvites ? 'Sending...' : 'Send invites'}
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </motion.div>

            {/* Live game banner */}
            {liveRuns.length > 0 && (
              <motion.div
                initial={{ opacity: 0, scale: 0.98 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.15 }}
              >
                <Link
                  href={`/host/live?gameId=${liveRuns[0].id}`}
                  className="block rounded-xl p-4 transition-all group"
                  style={{
                    background: 'linear-gradient(135deg, #EDFAF5 0%, #D5F5E6 100%)',
                    border: '1.5px solid #A8EBCC',
                  }}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2.5">
                      <div className="w-3 h-3 rounded-full animate-pulse" style={{ background: '#22AA6A' }} />
                      <span className="text-xs font-extrabold uppercase tracking-wider" style={{ color: '#116B3F' }}>
                        Live Now
                      </span>
                    </div>
                    <ChevronRight className="w-5 h-5 transition-transform group-hover:translate-x-1" style={{ color: '#22AA6A' }} />
                  </div>
                  <h3 className="text-base font-black mb-1" style={{ color: '#0D512F' }}>
                    {liveRuns[0].name}
                  </h3>
                  <p className="text-sm font-semibold" style={{ color: '#178A53' }}>
                    {liveRuns[0].allowedPlayerCount} players allowed &middot; Join with code: {liveRuns[0].code}
                  </p>
                  <div className="mt-3 flex items-center gap-2">
                    <Play className="w-4 h-4" style={{ color: '#116B3F' }} />
                    <span className="text-sm font-bold" style={{ color: '#116B3F' }}>
                      Go to Live Control Center
                    </span>
                  </div>
                </Link>
              </motion.div>
            )}

            {/* Upcoming runs */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.25 }}
              className="rounded-xl p-4"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <div className="flex items-center justify-between mb-3">
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
                <div className="space-y-2">
                  {upcomingRuns.map((run, i) => {
                    const displayStatus = displayBackendValue(run.status)
                    const style = RUN_STATUS_STYLES[displayStatus] || RUN_STATUS_STYLES['Scheduled']
                    return (
                      <motion.div
                        key={run.id}
                        initial={{ opacity: 0, x: -12 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: 0.3 + i * 0.06 }}
                        className="flex items-center gap-3 px-3 py-2.5 rounded-lg"
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
          <div className="lg:col-span-4 flex flex-col gap-4">

            {/* Active templates */}
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.25 }}
              className="rounded-xl p-4"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <div className="flex items-center justify-between mb-3">
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
              <div className="space-y-2">
                {mockGameTemplates.slice(0, 3).map((tmpl, i) => (
                  <motion.div
                    key={tmpl.id}
                    initial={{ opacity: 0, x: -12 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: 0.3 + i * 0.06 }}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-lg"
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
              className="rounded-xl p-4"
              style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}
            >
              <h2 className="text-sm font-extrabold uppercase tracking-widest mb-3" style={{ color: '#A8A29E' }}>
                Recent Games
              </h2>
              <div className="space-y-2">
                {recentRuns.map((run, i) => {
                  const displayStatus = displayBackendValue(run.status)
                  const style = RUN_STATUS_STYLES[displayStatus] || RUN_STATUS_STYLES['Complete']
                  return (
                    <Link
                      key={run.id}
                      href={`/summary?gameId=${run.id}`}
                      className="flex items-center gap-3 px-3 py-2.5 rounded-lg"
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

function generateJoinCode() {
  return Math.random().toString(36).slice(2, 8).toUpperCase()
}

function parseAllowedPlayers(raw: string) {
  return raw.split('\n').map(line => {
    const [first, second] = line.split(',').map(part => part.trim())
    const email = second || first
    return {
      email,
      displayName: second ? first : email.split('@')[0].replace(/[._-]+/g, ' '),
    }
  }).filter(player => player.email.includes('@'))
}

function FieldHelp({ label, help, children }: { label: string; help: string; children: ReactNode }) {
  return (
    <div className="block" title={help}>
      <span className="mb-1 block text-[10px] font-extrabold uppercase tracking-wider" style={{ color: '#A8A29E' }}>
        {label}
      </span>
      {children}
    </div>
  )
}
