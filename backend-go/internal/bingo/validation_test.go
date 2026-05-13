package bingo

import "testing"

func TestValidateSingleLineRow(t *testing.T) {
	result := Validate(validInput(PatternSingleLine, []string{"Word 0-0", "Word 0-1", "Word 0-2", "Word 0-3", "Word 0-4"}))
	if !result.Valid || result.Reason != "single_line_complete" || len(result.MatchedCells) != 5 {
		t.Fatalf("expected row win, got %+v", result)
	}
}

func TestValidateSingleLineColumn(t *testing.T) {
	result := Validate(validInput(PatternSingleLine, []string{"Word 0-1", "Word 1-1", "Word 2-1", "Word 3-1", "Word 4-1"}))
	if !result.Valid {
		t.Fatalf("expected column win, got %+v", result)
	}
}

func TestValidateSingleLineDiagonal(t *testing.T) {
	result := Validate(validInput(PatternSingleLine, []string{"Word 0-0", "Word 1-1", "Word 3-3", "Word 4-4"}))
	if !result.Valid {
		t.Fatalf("expected diagonal win with free center, got %+v", result)
	}
}

func TestValidateFourCorners(t *testing.T) {
	result := Validate(validInput(PatternFourCorners, []string{"Word 0-0", "Word 0-4", "Word 4-0", "Word 4-4"}))
	if !result.Valid || result.Reason != "four_corners_complete" || len(result.MatchedCells) != 4 {
		t.Fatalf("expected four-corners win, got %+v", result)
	}
}

func TestValidateFullHouse(t *testing.T) {
	called := make([]string, 0, 24)
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			if row == 2 && col == 2 {
				continue
			}
			called = append(called, wordAt(row, col))
		}
	}

	result := Validate(validInput(PatternFullHouse, called))
	if !result.Valid || result.Reason != "full_house_complete" || len(result.MatchedCells) != 25 {
		t.Fatalf("expected full-house win with free center, got %+v", result)
	}
}

func TestValidateFreeSpaceCounts(t *testing.T) {
	result := Validate(validInput(PatternSingleLine, []string{"Word 2-0", "Word 2-1", "Word 2-3", "Word 2-4"}))
	if !result.Valid {
		t.Fatalf("expected row through free space to win, got %+v", result)
	}
}

func TestValidateMissingCalledWord(t *testing.T) {
	result := Validate(validInput(PatternFourCorners, []string{"Word 0-0", "Word 0-4", "Word 4-0"}))
	if result.Valid || len(result.MissingCells) != 1 || result.MissingCells[0].Word != "Word 4-4" {
		t.Fatalf("expected one missing corner, got %+v", result)
	}
}

func TestValidateUnsupportedPattern(t *testing.T) {
	result := Validate(validInput("postage_stamp", []string{"Word 0-0"}))
	if result.Valid || result.Reason != "unsupported_pattern" {
		t.Fatalf("expected unsupported pattern, got %+v", result)
	}
}

func TestValidateGameOwnership(t *testing.T) {
	input := validInput(PatternSingleLine, []string{"Word 0-0", "Word 0-1", "Word 0-2", "Word 0-3", "Word 0-4"})
	input.CardGameRunID = "other-game"

	result := Validate(input)
	if result.Valid || result.Reason != "card_game_mismatch" {
		t.Fatalf("expected card game mismatch, got %+v", result)
	}
}

func validInput(pattern string, calledWords []string) ValidationInput {
	return ValidationInput{
		GameRunID:       "game-1",
		ClaimGameRunID:  "game-1",
		PlayerGameRunID: "game-1",
		CardGameRunID:   "game-1",
		Pattern:         pattern,
		Cells:           testCells(),
		CalledWords:     calledWords,
	}
}

func testCells() []Cell {
	cells := make([]Cell, 0, 25)
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			isFree := row == 2 && col == 2
			word := wordAt(row, col)
			if isFree {
				word = "FREE"
			}
			cells = append(cells, Cell{
				ID:          word,
				RowIndex:    row,
				ColIndex:    col,
				Word:        word,
				IsFreeSpace: isFree,
			})
		}
	}

	return cells
}

func wordAt(row, col int) string {
	return "Word " + string(rune('0'+row)) + "-" + string(rune('0'+col))
}
