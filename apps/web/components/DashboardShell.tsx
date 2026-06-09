'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { motion, AnimatePresence } from 'motion/react'
import {
  LayoutDashboard,
  CalendarClock,
  Sparkles,
  Radio,
  Shield,
  Users,
  Mic,
  History,
  Menu,
  X,
  ChevronRight,
  Home,
} from 'lucide-react'
import { useLanguage } from '@/contexts/LanguageContext'
import { LanguageToggle } from './LanguageToggle'

interface NavItem {
  label: string
  href: string
  icon: React.ElementType
  badge?: string | number
}

interface DashboardShellProps {
  children: React.ReactNode
  role: 'host' | 'admin'
  userName: string
}

const HOST_NAV: NavItem[] = [
  { label: 'Overview', href: '/host', icon: LayoutDashboard },
  { label: 'Game Templates', href: '/host/templates', icon: CalendarClock },
  { label: 'Content Review', href: '/host/review', icon: Sparkles, badge: 1 },
  { label: 'Live Game', href: '/host/live', icon: Radio },
]

const ADMIN_NAV: NavItem[] = [
  { label: 'Overview', href: '/admin', icon: LayoutDashboard },
  { label: 'Host Requests', href: '/admin/requests', icon: Users, badge: 3 },
  { label: 'Voice Profiles', href: '/admin/voices', icon: Mic, badge: 1 },
]

export function DashboardShell({ children, role, userName }: DashboardShellProps) {
  const pathname = usePathname()
  const [mobileOpen, setMobileOpen] = useState(false)
  const { t } = useLanguage()

  const navItems = role === 'admin' ? ADMIN_NAV : HOST_NAV
  const roleLabel = role === 'admin' ? t('dashboard.admin_center', 'Admin Center') : t('dashboard.title', 'Host Dashboard')
  const roleBg = role === 'admin' ? 'linear-gradient(135deg, #7C5CFC, #9E80FF)' : 'linear-gradient(135deg, #C0003D, #E8002D)'
  const roleGlow = role === 'admin' ? 'rgba(124,92,252,0.30)' : 'rgba(232,0,45,0.30)'

  function isActive(href: string) {
    if (href === '/host' || href === '/admin') return pathname === href
    return pathname.startsWith(href)
  }

  const NavLink = ({ item }: { item: NavItem }) => {
    const active = isActive(item.href)
    return (
      <Link
        href={item.href}
        onClick={() => setMobileOpen(false)}
        className="flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-bold transition-all relative group"
        style={{
          background: active ? (role === 'admin' ? '#F5F2FF' : '#FFF0F3') : 'transparent',
          color: active ? (role === 'admin' ? '#6440E8' : '#C40026') : '#78716C',
        }}
      >
        <item.icon className="w-[18px] h-[18px] shrink-0" />
        <span className="flex-1">{item.label}</span>
        {item.badge && (
          <span
            className="text-[10px] font-black px-2 py-0.5 rounded-full"
            style={{
              background: role === 'admin' ? '#EDE5FF' : '#FFE4D9',
              color: role === 'admin' ? '#6440E8' : '#C40026',
            }}
          >
            {item.badge}
          </span>
        )}
        {active && (
          <motion.div
            layoutId="activeNav"
            className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 rounded-full"
            style={{ background: role === 'admin' ? '#7C5CFC' : '#E8002D' }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          />
        )}
      </Link>
    )
  }

  return (
    <div
      className="min-h-screen flex relative overflow-x-hidden"
      style={{
        backgroundColor: '#FAF8F5',
        color: '#1C1917',
        fontFamily: "'Nunito', ui-rounded, system-ui, sans-serif",
      }}
    >
      {/* Decorative blobs */}
      <div
        aria-hidden="true"
        className="fixed top-[-200px] right-[-200px] w-[600px] h-[600px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(255,164,112,0.12) 0%, transparent 70%)' }}
      />
      <div
        aria-hidden="true"
        className="fixed bottom-[-150px] left-[-150px] w-[500px] h-[500px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(124,92,252,0.08) 0%, transparent 70%)' }}
      />

      {/* ─── Desktop Sidebar ─── */}
      <aside
        className="hidden lg:flex flex-col w-[260px] shrink-0 border-r sticky top-0 h-screen"
        style={{ borderColor: '#F0EDE8', background: 'rgba(255,255,255,0.75)', backdropFilter: 'blur(16px)' }}
      >
        {/* Logo */}
        <div className="p-5 pb-4">
          <Link href="/" className="flex items-center gap-2.5 group mb-5">
            <div
              className="w-9 h-9 rounded-lg flex items-center justify-center text-white font-black text-lg shrink-0 transition-transform group-hover:scale-105"
              style={{ background: 'linear-gradient(135deg, #C0003D, #E8002D)', boxShadow: '0 4px 12px rgba(232,0,45,0.30)' }}
            >
              B
            </div>
            <span className="font-black text-base tracking-tight" style={{ color: '#1C1917' }}>
              Virtual Bingo
            </span>
          </Link>

          {/* Role badge */}
          <div
            className="flex items-center gap-2.5 px-4 py-3 rounded-lg"
            style={{ background: role === 'admin' ? '#F5F2FF' : '#FFF0F3' }}
          >
            <div
              className="w-8 h-8 rounded-md flex items-center justify-center"
              style={{ background: roleBg, boxShadow: `0 4px 12px ${roleGlow}` }}
            >
              <Shield className="w-4 h-4 text-white" />
            </div>
            <div>
              <p className="text-xs font-extrabold" style={{ color: role === 'admin' ? '#4F30C2' : '#C23208' }}>
                {roleLabel}
              </p>
              <p className="text-[10px] font-semibold" style={{ color: '#A8A29E' }}>{userName}</p>
            </div>
          </div>
        </div>

        {/* Nav links */}
        <nav className="flex-1 px-3 space-y-1 overflow-y-auto">
          {navItems.map(item => (
            <NavLink key={item.href + item.label} item={{...item, label: t(`dashboard.nav_${item.label.toLowerCase().replace(/ /g, '_')}`, item.label)}} />
          ))}
        </nav>

        {/* Bottom: language and back to home */}
        <div className="p-3 border-t flex flex-col gap-2" style={{ borderColor: '#F0EDE8' }}>
          <div className="px-4 py-1">
            <LanguageToggle />
          </div>
          <Link
            href="/"
            className="flex items-center gap-2.5 px-4 py-2 rounded-lg text-sm font-bold transition-all"
            style={{ color: '#A8A29E' }}
          >
            <Home className="w-4 h-4" />
            {t('nav.back_to_home', 'Back to Home')}
          </Link>
        </div>
      </aside>

      {/* ─── Mobile Top Bar ─── */}
      <div className="lg:hidden fixed top-0 left-0 right-0 z-50 h-[60px] flex items-center justify-between px-4"
        style={{ background: 'rgba(255,255,255,0.9)', backdropFilter: 'blur(12px)', borderBottom: '1px solid #F0EDE8' }}
      >
        <Link href="/" className="flex items-center gap-2">
          <div
            className="w-8 h-8 rounded-md flex items-center justify-center text-white font-black text-sm"
            style={{ background: 'linear-gradient(135deg, #C0003D, #E8002D)' }}
          >
            B
          </div>
          <span className="font-black text-sm tracking-tight" style={{ color: '#1C1917' }}>
            {roleLabel}
          </span>
        </Link>
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          aria-label="Toggle menu"
          className="w-9 h-9 flex items-center justify-center rounded-md"
          style={{ color: '#78716C' }}
        >
          {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
        </button>
      </div>

      {/* ─── Mobile Drawer ─── */}
      <AnimatePresence>
        {mobileOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="lg:hidden fixed inset-0 z-40"
              style={{ background: 'rgba(0,0,0,0.2)' }}
              onClick={() => setMobileOpen(false)}
            />
            <motion.aside
              initial={{ x: -280 }}
              animate={{ x: 0 }}
              exit={{ x: -280 }}
              transition={{ type: 'spring', stiffness: 300, damping: 30 }}
              className="lg:hidden fixed left-0 top-0 bottom-0 z-50 w-[260px] flex flex-col"
              style={{ background: '#FFFFFF', boxShadow: '4px 0 24px rgba(0,0,0,0.08)' }}
            >
              <div className="h-[60px] flex items-center justify-between px-4 border-b" style={{ borderColor: '#F0EDE8' }}>
                <span className="font-black text-sm" style={{ color: '#1C1917' }}>{roleLabel}</span>
                <button onClick={() => setMobileOpen(false)} className="w-8 h-8 flex items-center justify-center" style={{ color: '#A8A29E' }}>
                  <X className="w-5 h-5" />
                </button>
              </div>
              <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
                {navItems.map(item => (
                  <NavLink key={item.href + item.label} item={{...item, label: t(`dashboard.nav_${item.label.toLowerCase().replace(/ /g, '_')}`, item.label)}} />
                ))}
              </nav>
              <div className="p-3 border-t flex flex-col gap-2" style={{ borderColor: '#F0EDE8' }}>
                <div className="px-4 py-1">
                  <LanguageToggle />
                </div>
                <Link
                  href="/"
                  onClick={() => setMobileOpen(false)}
                  className="flex items-center gap-2.5 px-4 py-2 rounded-lg text-sm font-bold"
                  style={{ color: '#A8A29E' }}
                >
                  <Home className="w-4 h-4" />
                  {t('nav.back_to_home', 'Back to Home')}
                </Link>
              </div>
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* ─── Main Content Area ─── */}
      <main className="flex-1 min-w-0 lg:pt-0 pt-[60px]">
        <div className="relative" style={{ isolation: 'isolate' }}>
          {children}
        </div>
      </main>
    </div>
  )
}
