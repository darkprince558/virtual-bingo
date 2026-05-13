import { cn } from "@/lib/utils"

interface CurrentCallDisplayProps {
  word: string | undefined
  aiMessage?: string
  className?: string
}

export function CurrentCallDisplay({ word, aiMessage, className }: CurrentCallDisplayProps) {
  return (
    <section className={cn("bg-white rounded-lg border border-slate-200 p-4 sm:p-6 shadow-sm flex flex-col sm:flex-row items-center justify-between gap-4 sm:gap-6 text-center sm:text-left", className)} aria-live="polite">
      <div className="flex flex-col">
        <span className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400 mb-1">Current Called Word</span>
        <h1 className="text-3xl sm:text-5xl font-black text-brand-600 tracking-tight">
          {word || "WAITING..."}
        </h1>
      </div>
      {aiMessage && (
        <div className="bg-brand-50 border border-brand-100 rounded-xl px-4 sm:px-6 py-3 sm:py-4 max-w-xs w-full sm:w-auto text-center sm:text-left">
          <p className="text-xs font-medium text-brand-800 italic leading-relaxed">
            {aiMessage}
          </p>
          <p className="text-[10px] text-brand-400 font-bold uppercase mt-2">AI Host</p>
        </div>
      )}
    </section>
  )
}
