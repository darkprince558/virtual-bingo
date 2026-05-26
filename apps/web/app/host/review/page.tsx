'use client'

import { useCallback, useEffect, useMemo, useState, Suspense } from 'react'
import { motion } from 'motion/react'
import { useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { DashboardShell } from '@/components/DashboardShell'
import { apiClient } from '@/lib/apiClient'
import { displayBackendValue, mapContentStatus } from '@/lib/uiMappers'
import type { CallerAssetResponse, GameContentResponse, GameRunResponse } from '@/types/api'
import { ChevronLeft, Sparkles, Check, RefreshCw, AlertTriangle, Edit3, CheckCircle2, Lock, Radio } from 'lucide-react'

export default function ContentReviewPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading...</div>}>
      <ContentReviewContent />
    </Suspense>
  )
}

function ContentReviewContent() {
  const searchParams = useSearchParams()
  const gameId = searchParams.get('gameId')

  const [game, setGame] = useState<GameRunResponse | null>(null)
  const [content, setContent] = useState<GameContentResponse | null>(null)
  const [wordsText, setWordsText] = useState('')
  const [topic, setTopic] = useState('')
  const [summary, setSummary] = useState('')
  const [callerStyle, setCallerStyle] = useState('')
  const [assets, setAssets] = useState<CallerAssetResponse[]>([])
  const [isLoading, setIsLoading] = useState(Boolean(gameId))
  const [isWorking, setIsWorking] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const words = useMemo(() => wordsText.split('\n').map(word => word.trim()).filter(Boolean), [wordsText])
  const isLocked = Boolean(content?.lockedAt)

  const syncContent = useCallback((next: GameContentResponse) => {
    setContent(next)
    setTopic(next.topic)
    setSummary(next.summary)
    setCallerStyle(next.callerStyle || '')
    setWordsText(next.words.join('\n'))
  }, [])

  const load = useCallback(async () => {
    if (!gameId) return
    const run = await apiClient<GameRunResponse>(`/games/${gameId}`)
    setGame(run)
    try {
      const nextContent = await apiClient<GameContentResponse>(`/games/${gameId}/content`, { devUserRole: 'host' })
      syncContent(nextContent)
      setError(null)
    } catch (err) {
      setContent(null)
      setWordsText('')
      setTopic('')
      setSummary('')
      setCallerStyle('')
      setError(err instanceof Error ? err.message : 'No generated content yet')
    }
  }, [gameId, syncContent])

  useEffect(() => {
    if (!gameId) return
    let cancelled = false

    async function run() {
      try {
        await load()
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load AI content review')
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    void run()

    return () => {
      cancelled = true
    }
  }, [gameId, load])

  const runAction = async (label: string, action: () => Promise<void>) => {
    setIsWorking(true)
    setNotice(null)
    try {
      await action()
      setNotice(label)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed')
    } finally {
      setIsWorking(false)
    }
  }

  const prepareContent = () => runAction('AI content generated with the local backend pipeline.', async () => {
    if (!gameId) return
    const next = await apiClient<GameContentResponse>(`/games/${gameId}/content/prepare`, { method: 'POST', devUserRole: 'host' })
    syncContent(next)
    setError(null)
  })

  const saveEdits = () => runAction('Content edits saved.', async () => {
    if (!gameId) return
    const next = await apiClient<GameContentResponse>(`/games/${gameId}/content`, {
      method: 'PATCH',
      devUserRole: 'host',
      body: JSON.stringify({ topic, summary, words, callerStyle: callerStyle || undefined })
    })
    syncContent(next)
    setError(null)
  })

  const lockContent = () => runAction('Content locked and ready for live play.', async () => {
    if (!gameId) return
    const next = await apiClient<GameContentResponse>(`/games/${gameId}/content/lock`, { method: 'POST', devUserRole: 'host' })
    syncContent(next)
    setError(null)
  })

  const generateAssets = () => runAction('Caller assets generated.', async () => {
    if (!gameId) return
    const nextAssets = await apiClient<CallerAssetResponse[]>(`/games/${gameId}/caller-assets/generate`, { method: 'POST', devUserRole: 'host' })
    setAssets(nextAssets)
  })

  const openLobby = () => runAction('Lobby opened.', async () => {
    if (!gameId) return
    const run = await apiClient<GameRunResponse>(`/games/${gameId}/lobby/open`, { method: 'POST', devUserRole: 'host' })
    setGame(run)
  })

  if (!gameId) {
    return (
      <DashboardShell role="host" userName="Admin Team">
        <div className="p-6 max-w-xl mx-auto">
          <Link href="/host" className="flex items-center gap-1.5 text-sm font-bold mb-6" style={{ color: '#A8A29E' }}>
            <ChevronLeft className="w-4 h-4" /> Back to Dashboard
          </Link>
          <div className="rounded-xl p-6 text-center" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
            <Sparkles className="w-8 h-8 mx-auto mb-3" style={{ color: '#F59E0B' }} />
            <h1 className="text-2xl font-black mb-2" style={{ color: '#1C1917' }}>No game selected</h1>
            <p className="text-sm font-semibold mb-5" style={{ color: '#78716C' }}>Choose a real game from the host dashboard to prepare AI words.</p>
            <Link href="/host" className="inline-flex px-5 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FFF4F0', color: '#E8440A' }}>Open Host Dashboard</Link>
          </div>
        </div>
      </DashboardShell>
    )
  }

  return (
    <DashboardShell role="host" userName="Admin Team">
      <div className="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
        <Link href="/host" className="flex items-center gap-1.5 text-sm font-bold mb-6" style={{ color: '#A8A29E' }}>
          <ChevronLeft className="w-4 h-4" /> Back to Dashboard
        </Link>

        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #FBBF24, #F59E0B)', boxShadow: '0 4px 12px rgba(245,158,11,0.25)' }}>
              <Sparkles className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-black tracking-tight" style={{ color: '#1C1917' }}>AI Content Review</h1>
              <p className="text-xs font-semibold" style={{ color: '#A8A29E' }}>{game?.name || 'Loading game'} · {content ? mapContentStatus(content.status) : 'Not generated'}</p>
            </div>
          </div>
        </motion.div>

        {notice && <div className="mb-5 rounded-lg p-4 text-sm font-bold" style={{ background: '#EDFAF5', border: '1.5px solid #A8EBCC', color: '#116B3F' }}>{notice}</div>}
        {error && <div className="mb-5 rounded-lg p-4 text-sm font-bold" style={{ background: '#FFF1F2', border: '1.5px solid #FECDD3', color: '#BE123C' }}>{error}</div>}

        {isLoading ? (
          <div className="rounded-xl p-8" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>Loading AI review...</div>
        ) : !content ? (
          <div className="rounded-xl p-6 text-center" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
            <AlertTriangle className="w-8 h-8 mx-auto mb-3" style={{ color: '#F59E0B' }} />
            <h2 className="text-xl font-black mb-2" style={{ color: '#1C1917' }}>No generated content yet</h2>
            <p className="text-sm font-semibold mb-5" style={{ color: '#78716C' }}>Use the Go backend pipeline to generate a reviewable topic, summary, and word list.</p>
            <button onClick={prepareContent} disabled={isWorking} className="inline-flex items-center gap-2 px-5 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FF5A1F', color: '#FFFFFF', opacity: isWorking ? 0.7 : 1 }}>
              <Sparkles className="w-4 h-4" /> {isWorking ? 'Generating...' : 'Generate AI Content'}
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
            <div className="lg:col-span-8 rounded-xl p-5 sm:p-6" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8', boxShadow: '0 2px 16px rgba(0,0,0,0.04)' }}>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-extrabold uppercase tracking-widest" style={{ color: '#A8A29E' }}>Generated Package</h2>
                <span className="text-[10px] font-black px-2.5 py-1 rounded-full uppercase" style={{ background: isLocked ? '#EDFAF5' : '#FEF3C7', color: isLocked ? '#116B3F' : '#B45309' }}>
                  {isLocked ? 'Locked' : mapContentStatus(content.status)}
                </span>
              </div>
              <div className="space-y-4">
                <label className="block">
                  <span className="text-xs font-extrabold" style={{ color: '#78716C' }}>Topic</span>
                  <input value={topic} onChange={e => setTopic(e.target.value)} disabled={isLocked} className="mt-1 w-full px-4 py-3 rounded-lg text-sm font-bold outline-none" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }} />
                </label>
                <label className="block">
                  <span className="text-xs font-extrabold" style={{ color: '#78716C' }}>Summary</span>
                  <textarea value={summary} onChange={e => setSummary(e.target.value)} disabled={isLocked} rows={3} className="mt-1 w-full px-4 py-3 rounded-lg text-sm font-bold outline-none resize-none" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }} />
                </label>
                <label className="block">
                  <span className="text-xs font-extrabold" style={{ color: '#78716C' }}>Caller Style</span>
                  <input value={callerStyle} onChange={e => setCallerStyle(e.target.value)} disabled={isLocked} placeholder="Energetic, concise, work-friendly" className="mt-1 w-full px-4 py-3 rounded-lg text-sm font-bold outline-none" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }} />
                </label>
                <label className="block">
                  <span className="text-xs font-extrabold" style={{ color: '#78716C' }}>Words ({words.length})</span>
                  <textarea value={wordsText} onChange={e => setWordsText(e.target.value)} disabled={isLocked} rows={14} className="mt-1 w-full px-4 py-3 rounded-lg text-sm font-bold outline-none resize-none" style={{ background: '#FAFAF9', border: '1.5px solid #F0EDE8', color: '#1C1917' }} />
                </label>
              </div>
            </div>

            <div className="lg:col-span-4 flex flex-col gap-4">
              <div className="rounded-xl p-5" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
                <h3 className="text-sm font-extrabold mb-4" style={{ color: '#1C1917' }}>Pipeline Actions</h3>
                <div className="space-y-2.5">
                  <button onClick={prepareContent} disabled={isWorking || isLocked} className="w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#F5F2FF', color: '#6440E8', opacity: isWorking || isLocked ? 0.55 : 1 }}>
                    <RefreshCw className="w-4 h-4" /> Regenerate Draft
                  </button>
                  <button onClick={saveEdits} disabled={isWorking || isLocked || words.length < 24} className="w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FFF4F0', color: '#C23208', opacity: isWorking || isLocked || words.length < 24 ? 0.55 : 1 }}>
                    <Edit3 className="w-4 h-4" /> Save Edits
                  </button>
                  <button onClick={lockContent} disabled={isWorking || isLocked || words.length < 24} className="w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#EDFAF5', color: '#116B3F', opacity: isWorking || isLocked || words.length < 24 ? 0.55 : 1 }}>
                    <Lock className="w-4 h-4" /> Lock Word Set
                  </button>
                  <button onClick={generateAssets} disabled={isWorking || !isLocked} className="w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FEF3C7', color: '#B45309', opacity: isWorking || !isLocked ? 0.55 : 1 }}>
                    <Radio className="w-4 h-4" /> Generate Caller Assets
                  </button>
                  <button onClick={openLobby} disabled={isWorking || !isLocked} className="w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#FF5A1F', color: '#FFFFFF', opacity: isWorking || !isLocked ? 0.55 : 1 }}>
                    <CheckCircle2 className="w-4 h-4" /> Open Lobby
                  </button>
                </div>
              </div>

              <div className="rounded-xl p-5" style={{ background: '#FFFFFF', border: '1.5px solid #F0EDE8' }}>
                <h3 className="text-sm font-extrabold mb-3" style={{ color: '#1C1917' }}>Backend State</h3>
                <div className="space-y-2 text-xs font-bold" style={{ color: '#78716C' }}>
                  <p>Status: {mapContentStatus(content.status)}</p>
                  <p>Provider: {displayBackendValue(content.generationProvider)}</p>
                  <p>Locked word set: {content.lockedWordSetId || 'Not locked'}</p>
                  <p>Caller assets: {assets.length > 0 ? assets.length : 'Not generated in this view'}</p>
                  <p>Game status: {displayBackendValue(game?.status)}</p>
                </div>
                {isLocked && (
                  <Link href={`/host/live?gameId=${gameId}`} className="mt-4 w-full flex items-center justify-center gap-2 py-3 rounded-lg text-sm font-extrabold" style={{ background: '#F4F2EF', color: '#44403C' }}>
                    Open Live Control
                  </Link>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </DashboardShell>
  )
}
