import { BingoCellData } from "@/types/player"
import { BingoCell } from "./BingoCell"

interface BingoCardProps {
  cells: BingoCellData[]
  onCellClick?: (id: string) => void
  disabled?: boolean
}

export function BingoCard({ cells, onCellClick, disabled }: BingoCardProps) {
  return (
    <div className="flex flex-col items-center justify-center w-full">
      <div className="grid grid-cols-5 gap-1.5 sm:gap-3 w-full max-w-2xl">
        {['B', 'I', 'N', 'G', 'O'].map((letter, i) => (
          <div key={i} className="text-center text-sm md:text-xl font-black text-slate-300 pb-1 sm:pb-2">{letter}</div>
        ))}
        {cells.map((cell) => (
          <BingoCell
             key={cell.id}
             word={cell.word}
             isMarked={cell.isMarked}
             isFreeSpace={cell.word === 'FREE SPACE'}
             disabled={disabled}
             onClick={() => onCellClick?.(cell.id)}
          />
        ))}
      </div>
    </div>
  )
}
