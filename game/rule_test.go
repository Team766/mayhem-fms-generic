// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Rule IDs used by the tests in this package, from the rules list in rule.go.
const (
	ruleIdG415   = 6  // Major, no ranking point.
	ruleIdMa2601 = 14 // Major, grants the opponent the Endgame RP.
	ruleIdMa2602 = 15 // Major, grants the opponent the Endgame RP.
	ruleIdMa2603 = 16 // Major, grants the opponent the Auton RP.
	ruleIdMa2604 = 17 // Major, no ranking point.
	ruleIdMa2606 = 18 // Minor, no ranking point.
	ruleIdMa2607 = 19 // Minor, no ranking point.
	ruleIdMa2616 = 29 // Minor, no ranking point.
)

func TestGetRuleById(t *testing.T) {
	assert.Nil(t, GetRuleById(0))
	assert.Equal(t, rules[0], GetRuleById(1))
	assert.Equal(t, rules[20], GetRuleById(21))
	assert.Nil(t, GetRuleById(1000))
}

func TestGetAllRules(t *testing.T) {
	allRules := GetAllRules()
	assert.Equal(t, len(rules), len(allRules))
	for _, rule := range rules {
		assert.Equal(t, rule, allRules[rule.Id])
	}
}

func TestRuleIdsAreSequentialAndUnique(t *testing.T) {
	for i, rule := range rules {
		assert.Equal(t, i+1, rule.Id, "rule %s is out of order", rule.RuleNumber)
		assert.NotEmpty(t, rule.RuleNumber)
		assert.NotEmpty(t, rule.Description)
	}

	// A duplicate ID would silently hide a rule, since the lookup is a map keyed by ID.
	assert.Equal(t, len(rules), len(GetAllRules()))
}

func TestRuleCounts(t *testing.T) {
	numMajor, numMinor, numRankingPoint := 0, 0, 0
	for _, rule := range rules {
		if rule.IsMajor {
			numMajor++
		} else {
			numMinor++
		}
		if rule.IsRankingPoint {
			numRankingPoint++
			// Only a major foul can grant the opposing alliance a ranking point.
			assert.True(t, rule.IsMajor, "%s grants a ranking point but is not major", rule.RuleNumber)
		}
	}

	assert.Equal(t, 30, len(rules))
	assert.Equal(t, 15, numMajor)
	assert.Equal(t, 15, numMinor)
	assert.Equal(t, 3, numRankingPoint)
}

func TestRankingPointRules(t *testing.T) {
	// The ranking-point conditions look these up by rule number, so both the numbers and the IDs matter.
	for ruleId, ruleNumber := range map[int]string{
		ruleIdMa2601: "MA2601",
		ruleIdMa2602: "MA2602",
		ruleIdMa2603: "MA2603",
	} {
		rule := GetRuleById(ruleId)
		if assert.NotNil(t, rule) {
			assert.Equal(t, ruleNumber, rule.RuleNumber)
			assert.True(t, rule.IsMajor)
			assert.True(t, rule.IsRankingPoint)
		}
	}

	// Rules that may be assessed at either severity appear twice, once per severity.
	for _, ruleNumber := range []string{"G423", "MA2615"} {
		severities := map[bool]int{}
		for _, rule := range rules {
			if rule.RuleNumber == ruleNumber {
				severities[rule.IsMajor]++
			}
		}
		assert.Equal(t, map[bool]int{false: 1, true: 1}, severities, "%s", ruleNumber)
	}
}
