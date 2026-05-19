interface AIHostPanelProps {
  message: string
}

export function AIHostPanel({ message }: AIHostPanelProps) {
  return (
    <aside className="bg-brand-600 rounded-lg p-4 sm:p-5 text-white shadow-lg mx-4 mb-4 sm:mx-6 sm:mb-6">
      <div className="flex items-center gap-2 mb-2">
        <div className="w-2 h-2 bg-white rounded-full animate-pulse"></div>
        <p className="text-[10px] font-bold uppercase tracking-widest">AI Host Monitoring</p>
      </div>
      <p className="text-sm font-light leading-snug opacity-90 italic">
        {message}
      </p>
    </aside>
  )
}
