import { cn } from "@/lib/utils"
import { GameState } from "@/types/game"

interface GameStatusBadgeProps {
  status: GameState
  className?: string
}

export function GameStatusBadge({ status, className }: GameStatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border",
        status === 'Live' && "bg-brand-50 text-brand-700 border-brand-200",
        status === 'Waiting' && "bg-gray-100 text-gray-800 border-gray-200",
        status === 'Starting Soon' && "bg-blue-50 text-blue-700 border-blue-200",
        status === 'Paused' && "bg-amber-50 text-amber-700 border-amber-200",
        status === 'Finished' && "bg-green-50 text-green-700 border-green-200",
        className
      )}
    >
      {status === 'Live' && <span className="w-1.5 h-1.5 mr-1.5 bg-brand-600 rounded-full animate-pulse" />}
      {status === 'Paused' && <span className="w-1.5 h-1.5 mr-1.5 bg-amber-500 rounded-full" />}
      {status}
    </span>
  )
}
