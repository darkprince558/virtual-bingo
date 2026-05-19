import { BingoCellData } from '@/types/player'
import { BingoCell } from './BingoCell'

interface BingoCardProps {
  cells: BingoCellData[]
  onCellClick?: (id: string) => void
  disabled?: boolean
}

const BINGO_LETTERS = ['B', 'I', 'N', 'G', 'O']

export function BingoCard({ cells, onCellClick, disabled }: BingoCardProps) {
  return (
    <div className="flex flex-col items-center justify-center w-full">
      {/* Card wrapper with warm shadow */}
      <div
        className="w-full max-w-2xl rounded-3xl p-3 sm:p-5"
        style={{
          background: '#FFFFFF',
          border: '2px solid #F0EDE8',
          boxShadow: '0 8px 40px rgba(255, 90, 31, 0.08), 0 2px 12px rgba(0,0,0,0.04)',
        }}
      >
        {/* B I N G O header letters */}
        <div className="grid grid-cols-5 gap-1.5 sm:gap-2.5 mb-2 sm:mb-3">
          {BINGO_LETTERS.map((letter) => (
            <div
              key={letter}
              className="text-center text-lg sm:text-2xl font-black pb-1"
              style={{ color: '#FFC5A8', letterSpacing: '0.05em' }}
            >
              {letter}
            </div>
          ))}
        </div>

        {/* Grid */}
        <div className="grid grid-cols-5 gap-1.5 sm:gap-2.5">
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
    </div>
  )
}
