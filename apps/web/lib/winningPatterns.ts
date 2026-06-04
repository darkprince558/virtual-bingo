// Single source of truth for winning-pattern shapes used by the host-screen
// preview. Each pattern maps its backend key to a label, a short description,
// and the set of [row, col] cells (on the 5x5 board) that light up to
// illustrate the pattern. Keep the keys in sync with the backend
// `bingo` package (see backend-go/internal/bingo/validation.go).

export type WinningPatternKey =
  | 'single_line'
  | 'four_corners'
  | 'full_house'
  | 'x_pattern'
  | 'plus'
  | 'letter_t'
  | 'postage_stamp'

export interface WinningPatternShape {
  key: WinningPatternKey
  label: string
  description: string
  /** Highlighted board cells as [rowIndex, colIndex] pairs (0-4). */
  cells: Array<[number, number]>
  /**
   * Set when the highlighted cells are only one representative example and the
   * win can be achieved in other equivalent placements (e.g. any row, or any
   * corner block). Surfaced as a hint under the preview.
   */
  exampleNote?: string
}

export const BOARD_SIZE = 5

function row(r: number): Array<[number, number]> {
  return Array.from({ length: BOARD_SIZE }, (_, c) => [r, c] as [number, number])
}

function col(c: number): Array<[number, number]> {
  return Array.from({ length: BOARD_SIZE }, (_, r) => [r, c] as [number, number])
}

export const WINNING_PATTERNS: Record<WinningPatternKey, WinningPatternShape> = {
  single_line: {
    key: 'single_line',
    label: 'Single Line',
    description: 'Any full row, column, or diagonal.',
    // Representative example: the middle row.
    cells: row(2),
    exampleNote: 'Any row, column, or diagonal wins — the middle row is shown as an example.',
  },
  four_corners: {
    key: 'four_corners',
    label: 'Four Corners',
    description: 'The four corner squares.',
    cells: [
      [0, 0],
      [0, 4],
      [4, 0],
      [4, 4],
    ],
  },
  full_house: {
    key: 'full_house',
    label: 'Full House',
    description: 'Every square on the card.',
    cells: Array.from({ length: BOARD_SIZE }, (_, r) => row(r)).flat(),
  },
  x_pattern: {
    key: 'x_pattern',
    label: 'X Pattern',
    description: 'Both diagonals crossing in the center.',
    cells: [
      [0, 0],
      [1, 1],
      [2, 2],
      [3, 3],
      [4, 4],
      [0, 4],
      [1, 3],
      [3, 1],
      [4, 0],
    ],
  },
  plus: {
    key: 'plus',
    label: 'Plus / Cross',
    description: 'The middle row and middle column.',
    cells: [...row(2), ...col(2).filter(([r]) => r !== 2)],
  },
  letter_t: {
    key: 'letter_t',
    label: 'Letter T',
    description: 'The top row and the middle column.',
    cells: [...row(0), ...col(2).filter(([r]) => r !== 0)],
  },
  postage_stamp: {
    key: 'postage_stamp',
    label: 'Postage Stamp',
    description: 'A 2x2 block in any corner.',
    // Representative example: the top-right corner block.
    cells: [
      [0, 3],
      [0, 4],
      [1, 3],
      [1, 4],
    ],
    exampleNote: 'Any corner 2x2 block wins — the top-right corner is shown as an example.',
  },
}

export const WINNING_PATTERN_LIST: WinningPatternShape[] = [
  WINNING_PATTERNS.single_line,
  WINNING_PATTERNS.four_corners,
  WINNING_PATTERNS.full_house,
  WINNING_PATTERNS.x_pattern,
  WINNING_PATTERNS.plus,
  WINNING_PATTERNS.letter_t,
  WINNING_PATTERNS.postage_stamp,
]

/** Resolve any backend pattern string (snake_case, spaced, etc.) to a shape. */
export function resolveWinningPattern(pattern?: string | null): WinningPatternShape {
  const normalized = (pattern || '').toLowerCase().replace(/[\s-]+/g, '_')
  switch (normalized) {
    case 'four_corners':
      return WINNING_PATTERNS.four_corners
    case 'full_house':
      return WINNING_PATTERNS.full_house
    case 'x_pattern':
    case 'x':
      return WINNING_PATTERNS.x_pattern
    case 'plus':
    case 'cross':
      return WINNING_PATTERNS.plus
    case 'letter_t':
    case 't':
      return WINNING_PATTERNS.letter_t
    case 'postage_stamp':
      return WINNING_PATTERNS.postage_stamp
    case 'single_line':
    case 'line':
    default:
      return WINNING_PATTERNS.single_line
  }
}

/** Membership test for a [row, col] cell within a pattern's highlighted set. */
export function isCellInPattern(shape: WinningPatternShape, rowIndex: number, colIndex: number): boolean {
  return shape.cells.some(([r, c]) => r === rowIndex && c === colIndex)
}
