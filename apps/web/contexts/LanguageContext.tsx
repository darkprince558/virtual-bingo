'use client'

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'

export type Language = 'en' | 'fr'

type Translations = Record<string, string>

const enTranslations: Translations = {
  // Common
  'common.loading': 'Loading...',
  'common.save': 'Save',
  'common.cancel': 'Cancel',
  'common.join': 'Join',
  'common.host': 'Host',
  'common.player': 'Player',

  // Home / Login Page
  'home.title': 'Virtual Bingo',
  'home.subtitle': 'Centralized cards, live word calls, and real-time leaderboards for your next team event.',
  'home.player_login': 'Join as Player',
  'home.host_login': 'Host a Game',
  'home.enter_code': 'Enter 6-digit code',
  'home.name_placeholder': 'Your Name',
  'home.join_game': 'Join Game',
  'home.welcome': 'Welcome to CGI Virtual Bingo',
  'home.entra_login': 'Sign in with Microsoft Entra',
  'home.entra_disclaimer': 'For CGI partners only — account will be verified',
  'home.join_with_code': 'Join with a Code',
  'home.game_code': 'GAME CODE',
  'home.join': 'Join',
  'home.feature_cards': 'Live Cards',
  'home.feature_cards_desc': 'Auto-generated',
  'home.feature_leaderboard': 'Leaderboard',
  'home.feature_leaderboard_desc': 'Real-time',
  'home.feature_ai': 'AI Host',
  'home.feature_ai_desc': 'Commentary',
  'home.host_link': 'I\'m a host, take me to the dashboard',
  'home.admin_link': 'Admin Operations Center',
  'home.footer': 'Internal corporate tool · Authentication coming soon',
  
  // Dashboard
  'dashboard.title': 'Host Dashboard',
  'dashboard.welcome': 'Welcome back, Host',
  'dashboard.active_games': 'Active Games',
  'dashboard.create_game': 'Create New Game',
  'dashboard.templates': 'Templates',
  
  // Nav
  'nav.home': 'Home',
  'nav.dashboard': 'Dashboard',
  'nav.settings': 'Settings',
  
  // Host Controls
  'host.call_next': 'Call Next Word',
  'host.end_game': 'End Game',
  'host.review_claims': 'Review Claims',
}

const frTranslations: Translations = {
  // Common
  'common.loading': 'Chargement...',
  'common.save': 'Enregistrer',
  'common.cancel': 'Annuler',
  'common.join': 'Rejoindre',
  'common.host': 'Héberger',
  'common.player': 'Joueur',

  // Home / Login Page
  'home.title': 'Bingo Virtuel',
  'home.subtitle': 'Cartes centralisées, tirages en direct et classements en temps réel pour votre prochain événement d\'équipe.',
  'home.player_login': 'Rejoindre en tant que joueur',
  'home.host_login': 'Héberger une partie',
  'home.enter_code': 'Entrez le code à 6 chiffres',
  'home.name_placeholder': 'Votre Nom',
  'home.join_game': 'Rejoindre la partie',
  'home.welcome': 'Bienvenue au CGI Bingo Virtuel',
  'home.entra_login': 'Se connecter avec Microsoft Entra',
  'home.entra_disclaimer': 'Pour les partenaires CGI uniquement — le compte sera vérifié',
  'home.join_with_code': 'Rejoindre avec un code',
  'home.game_code': 'CODE DU JEU',
  'home.join': 'Rejoindre',
  'home.feature_cards': 'Cartes en Direct',
  'home.feature_cards_desc': 'Générées auto.',
  'home.feature_leaderboard': 'Classement',
  'home.feature_leaderboard_desc': 'En temps réel',
  'home.feature_ai': 'Animateur IA',
  'home.feature_ai_desc': 'Commentaires',
  'home.host_link': 'Je suis un hôte, emmenez-moi au tableau de bord',
  'home.admin_link': 'Centre des Opérations Admin',
  'home.footer': 'Outil d\'entreprise interne · Authentification à venir',
  
  // Dashboard
  'dashboard.title': 'Tableau de bord de l\'hôte',
  'dashboard.welcome': 'Bon retour, Hôte',
  'dashboard.active_games': 'Parties Actives',
  'dashboard.create_game': 'Créer une Nouvelle Partie',
  'dashboard.templates': 'Modèles',
  
  // Nav
  'nav.home': 'Accueil',
  'nav.dashboard': 'Tableau de bord',
  'nav.settings': 'Paramètres',
  
  // Host Controls
  'host.call_next': 'Appeler le mot suivant',
  'host.end_game': 'Terminer la partie',
  'host.review_claims': 'Examiner les réclamations',
}

const dictionaries: Record<Language, Translations> = {
  en: enTranslations,
  fr: frTranslations,
}

interface LanguageContextType {
  language: Language
  setLanguage: (lang: Language) => void
  t: (key: string, fallback?: string) => string
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined)

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>('en')

  useEffect(() => {
    const savedLang = localStorage.getItem('bingo-language') as Language | null
    if (savedLang && ['en', 'fr'].includes(savedLang)) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setLanguage(savedLang)
    }
  }, [])

  useEffect(() => {
    localStorage.setItem('bingo-language', language)
    document.documentElement.lang = language
  }, [language])

  const t = (key: string, fallback?: string): string => {
    const dict = dictionaries[language]
    return dict[key] || fallback || key
  }

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  )
}

export function useLanguage() {
  const context = useContext(LanguageContext)
  if (context === undefined) {
    throw new Error('useLanguage must be used within a LanguageProvider')
  }
  return context
}
