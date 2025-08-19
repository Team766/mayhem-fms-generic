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

// All rules from the 2022 game that carry point penalties.
// @formatter:off
var rules = []*Rule{
	// General Conduct
	{1, "G206", false, true, "A team or ALLIANCE may not collude with another team to each purposefully violate a rule in an attempt to influence Ranking Points."},
	{2, "G210", true, false, "A strategy not consistent with standard gameplay and clearly aimed at forcing the opponent ALLIANCE to violate a rule is not in the spirit of FIRST Robotics Competition and not allowed."},
	{3, "G301", true, false, "A DRIVE TEAM member may not cause significant delays to the start of their MATCH."},
	{4, "G401", false, false, "In AUTO, each DRIVE TEAM member must remain in their staged areas. A DRIVE TEAM member staged behind a HUMAN STARTING LINE may not contact anything in front of that HUMAN STARTING LINE, unless for personal or equipment safety, to press the E-Stop or A-Stop, or granted permission by a Head REFEREE or FTA."},
	{5, "G402", false, false, "In AUTO, a DRIVE TEAM member may not directly or indirectly interact with a ROBOT or an OPERATOR CONSOLE unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop or A-Stop."},
	{6, "G403", false, false, "In AUTO, a HUMAN PLAYER may not enter GAME PIECES onto the field."},
	{7, "G404", true, false, "A ROBOT may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify the challenge associated with a FIELD element."},
	{8, "G405", true, false, "A ROBOT may not intentionally eject a SCORING ELEMENT from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT)."},
	{9, "G406", true, false, "Neither a ROBOT nor a HUMAN PLAYER may damage a SCORING ELEMENT."},
	{10, "G407", false, false, "A ROBOT may not simultaneously CONTROL more than 1 GAME PIECE 1 and 1 GAME PIECE 2 either directly or transitively through other objects."},
	{11, "G408", false, false, "BUMPERS must be in the BUMPER ZONE."}, {27, "G423", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately. B. regardless of intent, by initiating contact, either directly or transitively via a SCORING ELEMENT CONTROLLED by the ROBOT, inside the vertical projection of an opponent's ROBOT PERIMETER."},

	{12, "G409", false, false, "A ROBOT may not extend more than 1 ft. 6 in. beyond the vertical projection of its ROBOT PERIMETER."},
	{13, "G410", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately. B. regardless of intent, by initiating contact, either directly or transitively via a SCORING ELEMENT CONTROLLED by the ROBOT, inside the vertical projection of an opponent's ROBOT PERIMETER."},
	{14, "G411", true, false, "A ROBOT may not deliberately attach to, tip, or entangle with an opponent ROBOT."},
	{15, "G412", false, false, "A ROBOT may not PIN an opponent’s ROBOT for more than 3 seconds."},
	{16, "G413", true, false, "2 or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major element of MATCH play."},
	{17, "G414", false, false, "A DRIVE TEAM member must remain in their designated area as follows: A. DRIVERS and COACHES may not contact anything outside their ALLIANCE AREA, B. a DRIVER must use the OPERATOR CONSOLE in the DRIVER STATION to which they are assigned, as indicated on the team sign, C. a HUMAN PLAYER may not contact anything outside their ALLIANCE AREA or their PROCESSOR AREA, and D. a TECHNICIAN may not contact anything outside their designated area."},
	{18, "G415", false, false, "COACHES may not touch SCORING ELEMENTS, unless for safety purposes."},
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
