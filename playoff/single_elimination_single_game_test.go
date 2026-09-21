// Copyright 2026 Team 766. All Rights Reserved.

package playoff

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingleEliminationSingleGameInitialWith4Alliances(t *testing.T) {
	finalMatchup, breakSpecs, err := newSingleEliminationSingleGameBracket(4)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 12, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:6],
			[]expectedMatchSpec{
				{"Semifinal 1-1", "SF1-1", "", 37, "SF1", true, false, "sf", 1, 1},
				{"Semifinal 2-1", "SF2-1", "", 38, "SF2", true, false, "sf", 2, 1},
				{"Semifinal 1-2", "SF1-2", "", 39, "SF1", true, true, "sf", 1, 2},
				{"Semifinal 2-2", "SF2-2", "", 40, "SF2", true, true, "sf", 2, 2},
				{"Semifinal 1-3", "SF1-3", "", 41, "SF1", true, true, "sf", 1, 3},
				{"Semifinal 2-3", "SF2-3", "", 42, "SF2", true, true, "sf", 2, 3},
			},
		)
	}
	// The final is unaffected and remains a best-of-three with all three matches initially visible.
	assertFullFinals(t, matchSpecs, 6)

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "SF1", "SF2", "F")
	assert.Equal(t, 1, matchGroups["SF1"].(*Matchup).NumWinsToAdvance)
	assert.Equal(t, 1, matchGroups["SF2"].(*Matchup).NumWinsToAdvance)
	assert.Equal(t, 2, matchGroups["F"].(*Matchup).NumWinsToAdvance)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:6],
		[]expectedAlliances{
			{1, 4},
			{2, 3},
			{1, 4},
			{2, 3},
			{1, 4},
			{2, 3},
		},
	)
	for i := 6; i < 12; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	if assert.Equal(t, 3, len(breakSpecs)) {
		assert.Equal(t, breakSpec{43, 480, "Field Break"}, breakSpecs[0])
		assert.Equal(t, breakSpec{44, 480, "Field Break"}, breakSpecs[1])
		assert.Equal(t, breakSpec{45, 480, "Field Break"}, breakSpecs[2])
	}
}

func TestSingleEliminationSingleGameInitialWith8Alliances(t *testing.T) {
	finalMatchup, breakSpecs, err := newSingleEliminationSingleGameBracket(8)
	assert.Nil(t, err)
	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	if assert.Equal(t, 24, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs[0:12],
			[]expectedMatchSpec{
				{"Quarterfinal 1-1", "QF1-1", "", 25, "QF1", true, false, "qf", 1, 1},
				{"Quarterfinal 2-1", "QF2-1", "", 26, "QF2", true, false, "qf", 2, 1},
				{"Quarterfinal 3-1", "QF3-1", "", 27, "QF3", true, false, "qf", 3, 1},
				{"Quarterfinal 4-1", "QF4-1", "", 28, "QF4", true, false, "qf", 4, 1},
				{"Quarterfinal 1-2", "QF1-2", "", 29, "QF1", true, true, "qf", 1, 2},
				{"Quarterfinal 2-2", "QF2-2", "", 30, "QF2", true, true, "qf", 2, 2},
				{"Quarterfinal 3-2", "QF3-2", "", 31, "QF3", true, true, "qf", 3, 2},
				{"Quarterfinal 4-2", "QF4-2", "", 32, "QF4", true, true, "qf", 4, 2},
				{"Quarterfinal 1-3", "QF1-3", "", 33, "QF1", true, true, "qf", 1, 3},
				{"Quarterfinal 2-3", "QF2-3", "", 34, "QF2", true, true, "qf", 2, 3},
				{"Quarterfinal 3-3", "QF3-3", "", 35, "QF3", true, true, "qf", 3, 3},
				{"Quarterfinal 4-3", "QF4-3", "", 36, "QF4", true, true, "qf", 4, 3},
			},
		)
		assertMatchSpecs(
			t,
			matchSpecs[12:18],
			[]expectedMatchSpec{
				{"Semifinal 1-1", "SF1-1", "", 37, "SF1", true, false, "sf", 1, 1},
				{"Semifinal 2-1", "SF2-1", "", 38, "SF2", true, false, "sf", 2, 1},
				{"Semifinal 1-2", "SF1-2", "", 39, "SF1", true, true, "sf", 1, 2},
				{"Semifinal 2-2", "SF2-2", "", 40, "SF2", true, true, "sf", 2, 2},
				{"Semifinal 1-3", "SF1-3", "", 41, "SF1", true, true, "sf", 1, 3},
				{"Semifinal 2-3", "SF2-3", "", 42, "SF2", true, true, "sf", 2, 3},
			},
		)
	}
	assertFullFinals(t, matchSpecs, 18)

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "QF1", "QF2", "QF3", "QF4", "SF1", "SF2", "F")
	for _, id := range []string{"QF1", "QF2", "QF3", "QF4", "SF1", "SF2"} {
		assert.Equal(t, 1, matchGroups[id].(*Matchup).NumWinsToAdvance)
	}
	assert.Equal(t, 2, matchGroups["F"].(*Matchup).NumWinsToAdvance)

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:12],
		[]expectedAlliances{
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
		},
	)
	for i := 18; i < 24; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	if assert.Equal(t, 3, len(breakSpecs)) {
		assert.Equal(t, breakSpec{43, 480, "Field Break"}, breakSpecs[0])
		assert.Equal(t, breakSpec{44, 480, "Field Break"}, breakSpecs[1])
		assert.Equal(t, breakSpec{45, 480, "Field Break"}, breakSpecs[2])
	}
}

func TestSingleEliminationSingleGameErrors(t *testing.T) {
	_, _, err := newSingleEliminationSingleGameBracket(1)
	if assert.NotNil(t, err) {
		assert.Equal(t, "single-elimination bracket must have at least 2 alliances", err.Error())
	}

	_, _, err = newSingleEliminationSingleGameBracket(17)
	if assert.NotNil(t, err) {
		assert.Equal(t, "single-elimination bracket must have at most 16 alliances", err.Error())
	}
}

// Verifies, for a semifinal matchup, that a single win advances the winning alliance immediately (populating the
// final) while the remaining two matches of that series stay hidden, that a tie reveals exactly one more match of
// the series, and that the series resolves as soon as a decisive match is played.
func TestSingleEliminationSingleGameSemifinalProgression(t *testing.T) {
	playoffTournament, err := NewPlayoffTournament(model.SingleEliminationSingleGamePlayoff, 4)
	assert.Nil(t, err)
	finalMatchup := playoffTournament.FinalMatchup()
	matchSpecs := playoffTournament.matchSpecs
	matchGroups := playoffTournament.MatchGroups()
	sf1 := matchGroups["SF1"].(*Matchup)
	playoffMatchResults := map[int]playoffMatchResult{}

	// SF1-1 (order 37) is the only visible match of the SF1 series to start.
	assert.False(t, matchSpecs[0].isHidden) // SF1-1
	assert.True(t, matchSpecs[2].isHidden)  // SF1-2
	assert.True(t, matchSpecs[4].isHidden)  // SF1-3

	// A decisive win in the first match advances the winner immediately; the rest of the series stays hidden.
	playoffMatchResults[37] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, sf1.IsComplete())
	assert.Equal(t, 1, sf1.WinningAllianceId())
	assert.True(t, matchSpecs[2].isHidden) // SF1-2 stays hidden
	assert.True(t, matchSpecs[4].isHidden) // SF1-3 stays hidden
	assertMatchupOutcome(t, matchGroups["SF1"], "Advances to Final 1", "Eliminated")
	// The final's red alliance is now populated from the semifinal winner.
	assert.Equal(t, 1, matchSpecs[6].redAllianceId) // F1

	// Reverse the outcome to a tie; this should reveal exactly the next match in the series (SF1-2).
	playoffMatchResults[37] = playoffMatchResult{game.TieMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, sf1.IsComplete())
	assert.False(t, matchSpecs[2].isHidden) // SF1-2 is revealed
	assert.True(t, matchSpecs[4].isHidden)  // SF1-3 stays hidden
	assertMatchupOutcome(t, matchGroups["SF1"], "", "")

	// A decisive win in the (now revealed) second match resolves the series without needing a third match.
	playoffMatchResults[39] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, sf1.IsComplete())
	assert.Equal(t, 4, sf1.WinningAllianceId())
	assert.True(t, matchSpecs[4].isHidden) // SF1-3 stays hidden
	assertMatchupOutcome(t, matchGroups["SF1"], "Eliminated", "Advances to Final 1")

	// Tying the second match as well finally reveals the third and decisive match of the series.
	playoffMatchResults[39] = playoffMatchResult{game.TieMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, sf1.IsComplete())
	assert.False(t, matchSpecs[4].isHidden) // SF1-3 is revealed
	assertMatchupOutcome(t, matchGroups["SF1"], "", "")

	playoffMatchResults[41] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, sf1.IsComplete())
	assert.Equal(t, 1, sf1.WinningAllianceId())
	assertMatchupOutcome(t, matchGroups["SF1"], "Advances to Final 1", "Eliminated")
}

// Verifies that the final of the single-game-rounds bracket is unaffected and still requires two wins, including the
// case of a 1-1 series that must be decided by a third match.
func TestSingleEliminationSingleGameFinalStillBestOfThree(t *testing.T) {
	playoffTournament, err := NewPlayoffTournament(model.SingleEliminationSingleGamePlayoff, 4)
	assert.Nil(t, err)
	finalMatchup := playoffTournament.FinalMatchup()
	playoffMatchResults := map[int]playoffMatchResult{
		37: {game.RedWonMatch}, // SF1: alliance 1 advances.
		38: {game.RedWonMatch}, // SF2: alliance 2 advances.
	}
	finalMatchup.update(playoffMatchResults)
	assert.Equal(t, 1, finalMatchup.RedAllianceId)
	assert.Equal(t, 2, finalMatchup.BlueAllianceId)

	playoffMatchResults[43] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())

	playoffMatchResults[44] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete(), "a 1-1 series must not be complete")
	assert.Equal(t, 1, finalMatchup.RedAllianceWins)
	assert.Equal(t, 1, finalMatchup.BlueAllianceWins)

	playoffMatchResults[45] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, finalMatchup.IsComplete())
	assert.Equal(t, 1, finalMatchup.WinningAllianceId())
	assert.Equal(t, 2, finalMatchup.LosingAllianceId())
}
