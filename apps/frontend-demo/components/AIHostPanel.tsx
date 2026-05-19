import { Sparkles } from 'lucide-react'

interface AIHostPanelProps {
  message?: string
}

export function AIHostPanel({ message }: AIHostPanelProps) {
  if (!message) return null

  return (
    <div
      className="mx-4 mb-4 rounded-3xl p-4"
      style={{
        background: 'linear-gradient(135deg, #F5F2FF 0%, #EDE5FF 100%)',
        border: '1.5px solid #D9CCFF',
      }}
    >
      <div className="flex items-center gap-2 mb-2.5">
        <div
          className="w-7 h-7 rounded-xl flex items-center justify-center"
          style={{ background: 'linear-gradient(135deg, #7C5CFC, #9E80FF)', boxShadow: '0 4px 12px rgba(124, 92, 252, 0.30)' }}
        >
          <Sparkles className="w-3.5 h-3.5 text-white" />
        </div>
        <span className="text-xs font-extrabold uppercase tracking-widest" style={{ color: '#7C5CFC' }}>
          AI Host
        </span>
      </div>
      <p className="text-sm font-semibold leading-snug" style={{ color: '#4F30C2' }}>
        {message}
      </p>
    </div>
  )
}
