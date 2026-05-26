'use client'

import React from 'react'
import { useLanguage } from '@/contexts/LanguageContext'
import { Globe } from 'lucide-react'
import { motion } from 'motion/react'

export function LanguageToggle() {
  const { language, setLanguage } = useLanguage()

  const toggleLanguage = () => {
    setLanguage(language === 'en' ? 'fr' : 'en')
  }

  return (
    <button
      onClick={toggleLanguage}
      className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg transition-colors text-xs font-bold uppercase tracking-widest border border-transparent"
      style={{
        color: '#78716C',
        background: 'transparent',
      }}
      onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.background = '#F4F2EF'; (e.currentTarget as HTMLButtonElement).style.color = '#57534E'; }}
      onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; (e.currentTarget as HTMLButtonElement).style.color = '#78716C'; }}
      aria-label="Toggle language"
    >
      <Globe className="w-4 h-4" />
      <span>{language}</span>
    </button>
  )
}
