'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { Sparkles, Volume2 } from 'lucide-react'

interface CurrentCallDisplayProps {
  word?: string
  aiMessage?: string
  callNumber?: number
  audioUrl?: string
}

export function CurrentCallDisplay({ word, aiMessage, callNumber, audioUrl }: CurrentCallDisplayProps) {
  const [failedAudioKey, setFailedAudioKey] = useState<string | null>(null)
  const spokenKeyRef = useRef<string | null>(null)
  const speechText = useMemo(() => aiMessage || (word ? `Next word is ${word}.` : ''), [aiMessage, word])
  const audioKey = `${word || 'waiting'}:${callNumber || 'none'}:${audioUrl || 'browser'}`
  const shouldUseBrowserVoice = Boolean(speechText && (!audioUrl || failedAudioKey === audioKey))

  const speakWithBrowserVoice = useCallback(() => {
    if (typeof window === 'undefined' || !('speechSynthesis' in window) || !speechText) return
    window.speechSynthesis.cancel()
    const utterance = new SpeechSynthesisUtterance(speechText)
    utterance.rate = 0.95
    utterance.pitch = 1.03
    utterance.volume = 1
    window.speechSynthesis.speak(utterance)
    spokenKeyRef.current = audioKey
  }, [audioKey, speechText])

  useEffect(() => {
    if (!shouldUseBrowserVoice || spokenKeyRef.current === audioKey) return
    speakWithBrowserVoice()

    return () => {
      if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
        window.speechSynthesis.cancel()
      }
    }
  }, [audioKey, shouldUseBrowserVoice, speakWithBrowserVoice])

  return (
    <div
      className="w-full rounded-xl overflow-hidden relative"
      style={{
        background: 'linear-gradient(135deg, #FFF0F3 0%, #FFFFFF 40%, #F5F2FF 100%)',
        border: '2px solid #FFE4D9',
        boxShadow: '0 4px 24px rgba(232, 0, 45, 0.10)',
      }}
    >
      {/* Decorative shapes */}
      <div
        className="absolute top-0 right-0 w-40 h-40 rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(124,92,252,0.08) 0%, transparent 70%)', transform: 'translate(30%, -40%)' }}
      />
      <div
        className="absolute bottom-0 left-0 w-32 h-32 rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(232,0,45,0.06) 0%, transparent 70%)', transform: 'translate(-30%, 40%)' }}
      />

      {/* Header label */}
      <div className="relative px-5 pt-5 pb-0 flex items-center gap-2">
        <motion.div
          className="w-2.5 h-2.5 rounded-full"
          style={{ background: '#E8002D' }}
          animate={{ scale: [1, 1.3, 1], opacity: [1, 0.7, 1] }}
          transition={{ duration: 1.5, repeat: Infinity, ease: 'easeInOut' }}
        />
        <span
          className="text-xs font-extrabold uppercase tracking-widest"
          style={{ color: '#E8002D', letterSpacing: '0.18em' }}
        >
          Current Word
        </span>
        {callNumber && (
          <span
            className="ml-auto text-[10px] font-black px-2.5 py-1 rounded-full"
            style={{ background: '#FFE4D9', color: '#C40026' }}
          >
            #{callNumber}
          </span>
        )}
        <button
          type="button"
          onClick={speakWithBrowserVoice}
          disabled={!speechText}
          aria-label="Replay caller audio"
          title="Replay caller audio"
          className="ml-auto w-7 h-7 rounded-md flex items-center justify-center transition-all"
          style={{ background: '#FFF0F3' }}
        >
          <Volume2 className="w-3.5 h-3.5" style={{ color: '#FFB0C0' }} />
        </button>
      </div>

      {/* Word display */}
      <div className="relative px-5 py-5 flex items-center justify-center min-h-[110px]">
        <AnimatePresence mode="wait">
          {word ? (
            <motion.div
              key={word}
              initial={{ opacity: 0, scale: 0.5, y: 20, filter: 'blur(8px)' }}
              animate={{ opacity: 1, scale: 1, y: 0, filter: 'blur(0px)' }}
              exit={{ opacity: 0, scale: 0.8, y: -12, filter: 'blur(4px)' }}
              transition={{ type: 'spring', stiffness: 250, damping: 20 }}
              className="text-center relative"
            >
              {/* Glow backdrop */}
              <motion.div
                className="absolute inset-0 rounded-xl -z-10"
                animate={{ opacity: [0.3, 0.6, 0.3] }}
                transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
                style={{
                  background: 'radial-gradient(ellipse, rgba(232, 0, 45, 0.12) 0%, transparent 70%)',
                  transform: 'scale(2)',
                }}
              />
              <h2
                className="font-black leading-tight break-words animate-word-glow"
                style={{
                  color: '#1C1917',
                  fontSize: 'clamp(2rem, 5vw, 3.5rem)',
                }}
              >
                {word}
              </h2>
            </motion.div>
          ) : (
            <motion.div
              key="empty"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center"
            >
              {/* Waiting illustration */}
              <motion.div
                animate={{ y: [0, -5, 0] }}
                transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
                className="mx-auto mb-2 w-12 h-12 rounded-xl flex items-center justify-center"
                style={{ background: '#F4F2EF' }}
              >
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="#D6D3D1" strokeWidth="2" strokeDasharray="4 4" />
                  <path d="M12 8v4l3 3" stroke="#A8A29E" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </motion.div>
              <p className="text-base font-semibold" style={{ color: '#D6D3D1' }}>
                Waiting for first call…
              </p>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* AI message */}
      {aiMessage && (
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="mx-4 mb-4 rounded-lg px-4 py-3 flex items-start gap-3"
          style={{ background: 'linear-gradient(135deg, #F5F2FF, #EDE5FF)', border: '1px solid #D9CCFF' }}
        >
          <div
            className="w-6 h-6 rounded-sm flex items-center justify-center shrink-0 mt-0.5"
            style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)' }}
          >
            <Sparkles className="w-3 h-3 text-white" />
          </div>
          <p className="text-sm font-semibold leading-snug" style={{ color: '#4F30C2' }}>
            {aiMessage}
          </p>
        </motion.div>
      )}

      {/* AI Audio Playback */}
      {audioUrl && (
        <audio
          key={audioKey}
          autoPlay
          src={audioUrl}
          className="hidden"
          onError={() => setFailedAudioKey(audioKey)}
          onPlay={() => {
            spokenKeyRef.current = audioKey
          }}
        />
      )}
    </div>
  )
}
