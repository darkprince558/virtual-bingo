package bingo

import (
	"errors"
	"fmt"
	"strings"
)

const (
	PatternSingleLine   = "single_line"
	PatternFourCorners  = "four_corners"
	PatternFullHouse    = "full_house"
	PatternXPattern     = "x_pattern"
	PatternPlus         = "plus"
	PatternLetterT      = "letter_t"
	PatternPostageStamp = "postage_stamp"
)

var ErrUnsupportedPattern = errors.New("unsupported bingo pattern")

type Cell struct {
	ID          string `json:"id"`
	RowIndex    int    `json:"rowIndex"`
	ColIndex    int    `json:"colIndex"`
	Word        string `json:"word"`
	IsFreeSpace bool   `json:"isFreeSpace"`
}

type ValidationInput struct {
	GameRunID       string
	ClaimGameRunID  string
	PlayerGameRunID string
	CardGameRunID   string
	Pattern         string
	Cells           []Cell
	CalledWords     []string
}

type ValidationResult struct {
	Valid        bool   `json:"valid"`
	MatchedCells []Cell `json:"matchedCells"`
	MissingCells []Cell `json:"missingCells"`
	Reason       string `json:"reason"`
	Pattern      string `json:"pattern"`
}

func NormalizePattern(pattern string) string {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return PatternSingleLine
	}

	return pattern
}

func SupportedPatterns() []string {
	return []string{
		PatternSingleLine,
		PatternFourCorners,
		PatternFullHouse,
		PatternXPattern,
		PatternPlus,
		PatternLetterT,
		PatternPostageStamp,
	}
}

func IsSupportedPattern(pattern string) bool {
	switch NormalizePattern(pattern) {
	case PatternSingleLine,
		PatternFourCorners,
		PatternFullHouse,
		PatternXPattern,
		PatternPlus,
		PatternLetterT,
		PatternPostageStamp:
		return true
	default:
		return false
	}
}

func EnsureSupportedPattern(pattern string) error {
	pattern = NormalizePattern(pattern)
	if !IsSupportedPattern(pattern) {
		return fmt.Errorf("%w: %s", ErrUnsupportedPattern, pattern)
	}

	return nil
}

func Validate(input ValidationInput) ValidationResult {
	pattern := NormalizePattern(input.Pattern)
	result := ValidationResult{
		Valid:        false,
		MatchedCells: []Cell{},
		MissingCells: []Cell{},
		Reason:       "missing_called_words",
		Pattern:      pattern,
	}

	if !IsSupportedPattern(pattern) {
		result.Reason = "unsupported_pattern"
		return result
	}
	if input.GameRunID == "" || input.ClaimGameRunID != input.GameRunID {
		result.Reason = "claim_game_mismatch"
		return result
	}
	if input.PlayerGameRunID != input.GameRunID {
		result.Reason = "player_game_mismatch"
		return result
	}
	if input.CardGameRunID != input.GameRunID {
		result.Reason = "card_game_mismatch"
		return result
	}
	if len(input.Cells) == 0 {
		result.Reason = "card_has_no_cells"
		return result
	}

	called := make(map[string]struct{}, len(input.CalledWords))
	for _, word := range input.CalledWords {
		called[normalizeWord(word)] = struct{}{}
	}

	cellsByPosition := make(map[[2]int]Cell, len(input.Cells))
	for _, cell := range input.Cells {
		cellsByPosition[[2]int{cell.RowIndex, cell.ColIndex}] = cell
	}

	lines := positionsForPattern(pattern)
	best := result
	bestMissingCount := 26
	for _, line := range lines {
		matched := make([]Cell, 0, len(line))
		missing := make([]Cell, 0)

		for _, position := range line {
			cell, ok := cellsByPosition[position]
			if !ok {
				missing = append(missing, Cell{RowIndex: position[0], ColIndex: position[1]})
				continue
			}
			if cell.IsFreeSpace {
				matched = append(matched, cell)
				continue
			}
			if _, ok := called[normalizeWord(cell.Word)]; ok {
				matched = append(matched, cell)
				continue
			}
			missing = append(missing, cell)
		}

		if len(missing) == 0 && len(matched) == len(line) {
			return ValidationResult{
				Valid:        true,
				MatchedCells: matched,
				MissingCells: []Cell{},
				Reason:       pattern + "_complete",
				Pattern:      pattern,
			}
		}
		if len(missing) < bestMissingCount {
			bestMissingCount = len(missing)
			best.MatchedCells = matched
			best.MissingCells = missing
		}
	}

	return best
}

func positionsForPattern(pattern string) [][][2]int {
	switch pattern {
	case PatternFourCorners:
		return [][][2]int{{{0, 0}, {0, 4}, {4, 0}, {4, 4}}}
	case PatternFullHouse:
		all := make([][2]int, 0, 25)
		for row := 0; row < 5; row++ {
			for col := 0; col < 5; col++ {
				all = append(all, [2]int{row, col})
			}
		}
		return [][][2]int{all}
	case PatternXPattern:
		// Both diagonals form an X. The center (2,2) is shared, included once.
		return [][][2]int{{
			{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4},
			{0, 4}, {1, 3}, {3, 1}, {4, 0},
		}}
	case PatternPlus:
		// Middle row plus middle column form a +. Center included once.
		return [][][2]int{{
			{2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4},
			{0, 2}, {1, 2}, {3, 2}, {4, 2},
		}}
	case PatternLetterT:
		// Top row plus middle column form a T.
		return [][][2]int{{
			{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4},
			{1, 2}, {2, 2}, {3, 2}, {4, 2},
		}}
	case PatternPostageStamp:
		// A 2x2 block in any of the four corners qualifies.
		return [][][2]int{
			{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			{{0, 3}, {0, 4}, {1, 3}, {1, 4}},
			{{3, 0}, {3, 1}, {4, 0}, {4, 1}},
			{{3, 3}, {3, 4}, {4, 3}, {4, 4}},
		}
	default:
		return singleLinePositions()
	}
}

func singleLinePositions() [][][2]int {
	lines := make([][][2]int, 0, 12)
	for row := 0; row < 5; row++ {
		line := make([][2]int, 0, 5)
		for col := 0; col < 5; col++ {
			line = append(line, [2]int{row, col})
		}
		lines = append(lines, line)
	}
	for col := 0; col < 5; col++ {
		line := make([][2]int, 0, 5)
		for row := 0; row < 5; row++ {
			line = append(line, [2]int{row, col})
		}
		lines = append(lines, line)
	}

	firstDiagonal := make([][2]int, 0, 5)
	secondDiagonal := make([][2]int, 0, 5)
	for index := 0; index < 5; index++ {
		firstDiagonal = append(firstDiagonal, [2]int{index, index})
		secondDiagonal = append(secondDiagonal, [2]int{index, 4 - index})
	}
	lines = append(lines, firstDiagonal, secondDiagonal)

	return lines
}

func normalizeWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
