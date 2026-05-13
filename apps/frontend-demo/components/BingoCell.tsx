import { cn } from "@/lib/utils"
import { motion } from "motion/react"

interface BingoCellProps {
  word: string
  isMarked: boolean
  onClick?: () => void
  disabled?: boolean
  isFreeSpace?: boolean
}

export function BingoCell({ word, isMarked, onClick, disabled, isFreeSpace }: BingoCellProps) {
  return (
    <motion.button
      whileHover={!disabled && !isFreeSpace && !isMarked ? { scale: 1.02 } : {}}
      whileTap={!disabled && !isFreeSpace && !isMarked ? { scale: 0.95 } : {}}
      animate={{
        scale: isMarked && !isFreeSpace ? [1, 1.08, 1] : 1,
      }}
      transition={
        isMarked && !isFreeSpace
          ? { duration: 0.4, times: [0, 0.5, 1], ease: "easeInOut" }
          : { type: "spring", stiffness: 400, damping: 25 }
      }
      onClick={onClick}
      disabled={disabled || isFreeSpace}
      className={cn(
        "group aspect-square rounded-xl sm:rounded-2xl flex items-center justify-center p-2 sm:p-3 text-center outline-none relative transition-all duration-300",
        "focus-visible:ring-4 focus-visible:ring-brand-500 focus-visible:ring-offset-2",
        isFreeSpace
          ? "bg-brand-50 border border-brand-200 shadow-sm"
          : isMarked
            ? "bg-brand-600 shadow-md shadow-brand-500/20 text-white border border-brand-500"
            : "bg-white border border-slate-200 shadow-sm hover:shadow-md hover:border-brand-200 cursor-pointer",
        disabled && !isFreeSpace && !isMarked ? "opacity-60 cursor-not-allowed" : ""
      )}
      aria-pressed={isMarked}
      aria-label={`${word}${isMarked ? ', marked' : ', unmarked'}`}
    >
      <span className={cn(
        "break-words w-full px-0.5 sm:px-1 leading-tight sm:leading-normal z-10 transition-colors duration-300",
        isFreeSpace
          ? "text-[9px] sm:text-[11px] sm:font-black font-bold text-brand-800 uppercase tracking-wider"
          : isMarked
            ? "text-[10px] sm:text-xs font-bold text-white drop-shadow-sm"
            : "text-[10px] sm:text-xs font-semibold text-slate-700 group-hover:text-brand-700",
      )}>
        {word}
      </span>
      {isMarked && !isFreeSpace && (
         <motion.span 
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ type: "spring", stiffness: 400, damping: 20, delay: 0.1 }}
            className="absolute top-1 right-1 sm:top-1.5 sm:right-1.5 text-white bg-white/20 rounded-full w-4 h-4 sm:w-5 sm:h-5 flex items-center justify-center font-bold text-[8px] sm:text-[10px] z-10 backdrop-blur-sm"
         >
            ✓
         </motion.span>
      )}
    </motion.button>
  )
}
