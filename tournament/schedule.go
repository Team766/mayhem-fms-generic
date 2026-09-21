// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for creating practice and qualification match schedules.

package tournament

import (
	"encoding/csv"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	schedulesDir       = "schedules"
	TeamsPerMatch      = 6
	TeamsPerMatch2v2   = 4
	twoVsTwoFilePrefix = "2p_"
)

var schedulePerm = rand.Perm

// Creates a random schedule for the given parameters and returns it as a list of matches. In 2v2 mode, the
// schedule is loaded from a "2p_"-prefixed template with four teams per match; the third slot of each
// alliance is left at team 0 with no surrogate flag.
func BuildRandomSchedule(
	teams []model.Team, scheduleBlocks []model.ScheduleBlock, matchType model.MatchType, twoVsTwo bool,
) ([]model.Match, error) {
	// Load the anonymized, pre-randomized match schedule for the given number of teams and matches per team.
	teamsPerMatch := TeamsPerMatch
	if twoVsTwo {
		teamsPerMatch = TeamsPerMatch2v2
	}
	numTeams := len(teams)
	numMatches := countMatches(scheduleBlocks)
	matchesPerTeam := int(float32(numMatches*teamsPerMatch) / float32(numTeams))

	// Adjust the number of matches to remove any excess from non-perfect block scheduling.
	numMatches = int(math.Ceil(float64(numTeams) * float64(matchesPerTeam) / float64(teamsPerMatch)))

	filePrefix := ""
	if twoVsTwo {
		filePrefix = twoVsTwoFilePrefix
	}
	file, err := os.Open(
		fmt.Sprintf(
			"%s/%s%d_%d.csv", filepath.Join(model.BaseDir, schedulesDir), filePrefix, numTeams, matchesPerTeam,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("No schedule template exists for %d teams and %d matches", numTeams, matchesPerTeam)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	csvLines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(csvLines) != numMatches {
		return nil, fmt.Errorf("Schedule file contains %d matches, expected %d", len(csvLines), numMatches)
	}

	// Convert string fields from schedule to integers.
	anonSchedule := make([][12]int, numMatches)
	for i := 0; i < numMatches; i++ {
		for j := 0; j < 12; j++ {
			anonSchedule[i][j], err = strconv.Atoi(csvLines[i][j])
			if err != nil {
				return nil, err
			}
		}
	}

	// Generate a random permutation of the team ordering to fill into the pre-randomized schedule.
	teamShuffle := schedulePerm(numTeams)
	matches := make([]model.Match, numMatches)
	for i, anonMatch := range anonSchedule {
		matches[i].Type = matchType
		matches[i].TypeOrder = i + 1
		if matchType == model.Practice {
			matches[i].ShortName = fmt.Sprintf("P%d", i+1)
			matches[i].LongName = fmt.Sprintf("Practice %d", i+1)
			matches[i].TbaMatchKey.CompLevel = "p"
		} else if matchType == model.Qualification {
			matches[i].ShortName = fmt.Sprintf("Q%d", i+1)
			matches[i].LongName = fmt.Sprintf("Qualification %d", i+1)
			matches[i].TbaMatchKey.CompLevel = "qm"
		} else {
			return nil, fmt.Errorf("invalid match type %q", matchType)
		}
		matches[i].Red1, matches[i].Red1IsSurrogate = scheduleTeam(teams, teamShuffle, anonMatch[0], anonMatch[1])
		matches[i].Red2, matches[i].Red2IsSurrogate = scheduleTeam(teams, teamShuffle, anonMatch[2], anonMatch[3])
		matches[i].Red3, matches[i].Red3IsSurrogate = scheduleTeam(teams, teamShuffle, anonMatch[4], anonMatch[5])
		matches[i].Blue1, matches[i].Blue1IsSurrogate = scheduleTeam(teams, teamShuffle, anonMatch[6], anonMatch[7])
		matches[i].Blue2, matches[i].Blue2IsSurrogate = scheduleTeam(teams, teamShuffle, anonMatch[8], anonMatch[9])
		matches[i].Blue3, matches[i].Blue3IsSurrogate = scheduleTeam(
			teams, teamShuffle, anonMatch[10], anonMatch[11],
		)
		matches[i].TbaMatchKey.MatchNumber = i + 1
	}

	// Fill in the match times.
	matchIndex := 0
	for _, block := range scheduleBlocks {
		for i := 0; i < block.NumMatches && matchIndex < numMatches; i++ {
			matches[matchIndex].Time = block.StartTime.Add(time.Duration(i*block.MatchSpacingSec) * time.Second)
			matchIndex++
		}
	}

	return matches, nil
}

// Resolves one anonymized (team-index, surrogate-flag) column pair to a real team ID and surrogate flag. A
// team index of 0 (the empty third slot of a 2v2 template) resolves to team 0 with no surrogate flag.
func scheduleTeam(teams []model.Team, teamShuffle []int, teamIndex, surrogateFlag int) (int, bool) {
	if teamIndex == 0 {
		return 0, false
	}
	return teams[teamShuffle[teamIndex-1]].Id, surrogateFlag == 1
}

// Returns the total number of matches that can be run within the given schedule blocks.
func countMatches(scheduleBlocks []model.ScheduleBlock) int {
	numMatches := 0
	for _, block := range scheduleBlocks {
		numMatches += block.NumMatches
	}
	return numMatches
}
