// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a game-specific rule.

package game

type Rule struct {
	Id             int
	RuleNumber     string
	IsMajor        bool
	IsRankingPoint bool
	Description    string
}

// All rules from the Medieval Mayhem manual that carry point penalties. Rules that the referee may assess at either
// severity (G423, MA2615) appear twice, since the rule dropdown is filtered by the foul's type. Card-only rules and the
// no-score rule MA2622 are not listed, as they carry no foul points.
// @formatter:off
var rules = []*Rule{
	{1, "G210", true, false, "Do not force an opponent robot to commit a foul."},
	{2, "G401", true, false, "Drive team members must stay behind the lines during auto."},
	{3, "G402", true, false, "Drive team members must not control their robot during auto, except to press A-stop or E-stop."},
	{4, "G403", false, false, "A robot may not cross the midfield line into the opponent's half during auto."},
	{5, "G408", true, false, "No human or human player may damage a treasure."},
	{6, "G415", true, false, "Contacting another robot inside its perimeter. Can stack with MA2618."},
	{7, "G423", false, false, "Damaging or impairing an opponent robot (referee's choice of severity)."},
	{8, "G423", true, false, "Damaging or impairing an opponent robot (referee's choice of severity)."},
	{9, "G424", true, false, "Intentionally attaching to, entangling or tipping another robot."},
	{10, "G425", false, false, "Pinning a robot for longer than 3 seconds."},
	{11, "G429", true, false, "Drive team members must stay in their designated areas."},
	{12, "G430", true, false, "A robot may be operated only by its drivers."},
	{13, "G434", true, false, "Coaches may not touch treasures, unless for safety."},
	{14, "MA2601", true, true, "Contacting the opposing alliance's balance beam during endgame. Opponent gets the Endgame RP. Can stack with MA2602."},
	{15, "MA2602", true, true, "Contacting an opposing robot at all while it is on its own balance beam during endgame. Opponent gets the Endgame RP. Can stack with MA2601."},
	{16, "MA2603", true, true, "Entering the opposing alliance's safe house or safe zone during auto. Opponent gets the Auton RP."},
	{17, "MA2604", true, false, "Hoarding: more than 6 non-scoring treasures in your safe house or safe zone. Repeats every 10 seconds."},
	{18, "MA2606", false, false, "Entering the opposing human loading zone, safe house or safe zone, or touching their balance beam before endgame. Repeats every 10 seconds. The apron facing midfield is allowed."},
	{19, "MA2607", false, false, "Putting a treasure into the field other than through the human loading holes (not during the Toss)."},
	{20, "MA2608", false, false, "Loaded treasure does not first touch an own-alliance robot or the human loading zone floor."},
	{21, "MA2609", false, false, "Contacting own shelf repeatedly or in an unsafe manner."},
	{22, "MA2610", false, false, "Starting auto with more than 1 treasure. That robot cannot earn auto placement points."},
	{23, "MA2611", false, false, "Descoring the other alliance's treasures. Can stack with MA2620."},
	{24, "MA2612", false, false, "Robot deliberately destroying a treasure."},
	{25, "MA2613", false, false, "Adding treasures to the field during endgame. The treasure does not count."},
	{26, "MA2614", false, false, "Placing or shooting treasures into the opposing alliance's zones."},
	{27, "MA2615", false, false, "Failure to obey referee-area signs (referee's choice of severity)."},
	{28, "MA2615", true, false, "Failure to obey referee-area signs (referee's choice of severity)."},
	{29, "MA2616", false, false, "Controlling more than 2 treasures for more than a moment. Extra treasures do not count."},
	{30, "MA2617", false, false, "Throwing a treasure onto the shelf from outside own safe house. The treasure does not count."},
}

// @formatter:on
var ruleMap map[int]*Rule

// Returns the rule having the given ID, or nil if no such rule exists.
func GetRuleById(id int) *Rule {
	return GetAllRules()[id]
}

// Returns a slice of all defined rules that carry point penalties.
func GetAllRules() map[int]*Rule {
	if ruleMap == nil {
		ruleMap = make(map[int]*Rule, len(rules))
		for _, rule := range rules {
			ruleMap[rule.Id] = rule
		}
	}
	return ruleMap
}
