'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { Shield, Settings, Wifi, WifiOff } from 'lucide-react'
import { GameState } from '@/types/game'
import { SettingsModal } from './SettingsModal'
import { useSettings } from '@/contexts/SettingsContext'

interface TopNavProps {
  gameCode?: string
  playerName?: string
  status?: GameState
  role: 'player' | 'host'
  connectionState?: 'Connected' | 'Reconnecting' | 'Disconnected'
}

const STATUS_STYLES: Record<string, { dot: string; text: string; label: string }> = {
  Live: { dot: '#22AA6A', text: '#116B3F', label: 'Live' },
  Waiting: { dot: '#F59E0B', text: '#B45309', label: 'Waiting' },
  'Starting Soon': { dot: '#F59E0B', text: '#B45309', label: 'Starting Soon' },
  Paused: { dot: '#F59E0B', text: '#D97706', label: 'Paused' },
  'Lobby Open': { dot: '#22AA6A', text: '#116B3F', label: 'Lobby Open' },
  Finished: { dot: '#A8A29E', text: '#78716C', label: 'Finished' },
  Cancelled: { dot: '#F43F5E', text: '#BE123C', label: 'Cancelled' },
  Failed: { dot: '#F43F5E', text: '#BE123C', label: 'Failed' },
}

export function TopNav({ gameCode, playerName, status, role, connectionState = 'Connected' }: TopNavProps) {
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)
  const { avatar } = useSettings()

  const statusStyle = status ? STATUS_STYLES[status] : null

  return (
    <>
      <nav
        className="h-14 sm:h-[68px] px-3 sm:px-8 flex items-center justify-between shrink-0"
        style={{
          background: 'rgba(255,255,255,0.85)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid rgba(231, 229, 228, 0.7)',
        }}
      >
        {/* Left: Logo + Game Code */}
        <div className="flex items-center gap-3 sm:gap-5">
          <Link href="/" className="flex items-center gap-2.5 group">
            <div
              className="w-9 h-9 rounded-lg flex items-center justify-center text-white font-black text-lg shrink-0 transition-transform group-hover:scale-105"
              style={{
                background: 'linear-gradient(135deg, #FF7A42 0%, #FF5A1F 100%)',
                boxShadow: '0 4px 12px rgba(255, 90, 31, 0.35)',
              }}
            >
              B
            </div>
            <span className="text-base font-black tracking-tight hidden sm:inline-block" style={{ color: '#1C1917' }}>
              Virtual Bingo
            </span>
          </Link>

          {gameCode && (
            <>
              <div className="hidden sm:block h-5 w-px bg-stone-200" />
              <div className="flex items-center gap-2">
                <span className="hidden sm:inline text-xs font-700 uppercase tracking-widest" style={{ color: '#A8A29E' }}>Code</span>
                <span
                  className="px-3 py-1 rounded-full text-sm font-black tracking-widest"
                  style={{ background: '#FFF4F0', color: '#E8440A', letterSpacing: '0.12em' }}
                >
                  {gameCode}
                </span>
              </div>
            </>
          )}
        </div>

        {/* Right: Status + Player + Settings */}
        <div className="flex items-center gap-3 sm:gap-5">
          {/* Connection status */}
          {connectionState !== 'Connected' && (
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-bold" style={{ background: '#FEF3C7', color: '#B45309' }}>
              <WifiOff className="w-3.5 h-3.5" />
              <span className="hidden sm:inline">{connectionState}</span>
            </div>
          )}

          {/* Game status pill */}
          {statusStyle && (
            <div className="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-extrabold uppercase tracking-wider"
              style={{
                background: statusStyle.dot === '#22AA6A' ? '#EDFAF5' : statusStyle.dot === '#A8A29E' ? '#F4F2EF' : '#FEF3C7',
                color: statusStyle.text,
              }}
            >
              <span
                className="w-1.5 h-1.5 rounded-full shrink-0"
                style={{
                  background: statusStyle.dot,
                  ...(status === 'Live' ? { animation: 'pulse 2s cubic-bezier(0.4,0,0.6,1) infinite' } : {}),
                }}
              />
              {statusStyle.label}
            </div>
          )}

          {/* Player avatar + name */}
          {playerName && (
            <div className="flex items-center gap-2.5 sm:pl-4 sm:border-l" style={{ borderColor: '#E7E5E4' }}>
              <div className="hidden sm:block text-right">
                <p className="text-sm font-bold leading-tight" style={{ color: '#1C1917' }}>{playerName}</p>
                <p className="text-[10px] font-semibold uppercase tracking-wide" style={{ color: '#A8A29E' }}>
                  {role === 'host' ? 'Host' : 'Player'}
                </p>
              </div>
              <div
                className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 text-lg select-none overflow-hidden"
                style={{
                  background: role === 'host' ? 'linear-gradient(135deg, #7C5CFC, #9E80FF)' : '#FFF4F0',
                  border: role === 'host' ? 'none' : '2px solid #FFC5A8',
                }}
              >
                {role === 'host'
                  ? <Shield className="w-4 h-4 text-white" />
                  : <span style={{ color: '#E8440A' }}>{avatar}</span>
                }
              </div>
            </div>
          )}

          {/* Settings */}
          <button
            onClick={() => setIsSettingsOpen(true)}
            aria-label="Open appearance settings"
            className="w-9 h-9 flex items-center justify-center rounded-lg transition-all"
            style={{ color: '#A8A29E' }}
            onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.background = '#F4F2EF'; (e.currentTarget as HTMLButtonElement).style.color = '#78716C'; }}
            onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; (e.currentTarget as HTMLButtonElement).style.color = '#A8A29E'; }}
          >
            <Settings className="w-5 h-5" />
          </button>
        </div>
      </nav>

      <SettingsModal isOpen={isSettingsOpen} onClose={() => setIsSettingsOpen(false)} />
    </>
  )
}
