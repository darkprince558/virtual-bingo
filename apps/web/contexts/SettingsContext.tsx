'use client'

import React, { createContext, useContext, useState, useEffect } from 'react'

export type ThemeType = 'indigo' | 'rose' | 'emerald' | 'amber'

export const AVATARS = [
  '🐶', '🐱', '🐭', '🐹', '🐰', '🦊', '🐻', '🐼', '🐨', '🐯',
  '🦁', '🐮', '🐷', '🐸', '🐵', '🦄', '🐝', '🐙', '🦖', '🐢'
]

interface SettingsContextType {
  theme: ThemeType
  setTheme: (theme: ThemeType) => void
  avatar: string
  setAvatar: (avatar: string) => void
}

const SettingsContext = createContext<SettingsContextType | undefined>(undefined)

function getInitialTheme(): ThemeType {
  return 'indigo'
}

function getInitialAvatar(): string {
  return '🦊'
}

export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<ThemeType>(getInitialTheme)
  const [avatar, setAvatar] = useState<string>(getInitialAvatar)

  useEffect(() => {
    const savedTheme = localStorage.getItem('bingo-theme') as ThemeType | null
    if (savedTheme && ['indigo', 'rose', 'emerald', 'amber'].includes(savedTheme)) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setTheme(savedTheme)
    }
    const savedAvatar = localStorage.getItem('bingo-avatar')
    if (savedAvatar && AVATARS.includes(savedAvatar)) {
      setAvatar(savedAvatar)
    }
  }, [])

  useEffect(() => {
    localStorage.setItem('bingo-theme', theme)
    document.body.classList.remove('theme-indigo', 'theme-rose', 'theme-emerald', 'theme-amber')
    document.body.classList.add(`theme-${theme}`)
  }, [theme])

  useEffect(() => {
    localStorage.setItem('bingo-avatar', avatar)
  }, [avatar])

  return (
    <SettingsContext.Provider value={{ theme, setTheme, avatar, setAvatar }}>
      {children}
    </SettingsContext.Provider>
  )
}

export function useSettings() {
  const context = useContext(SettingsContext)
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  return context
}
