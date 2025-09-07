// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

type Rule struct {
	Id             int
	RuleNumber     string
	IsMajor        bool
	IsRankingPoint bool
	Description    string
}

// A curated list of generic rules that can apply to many games.
// @formatter:off
var rules = []*Rule{
	// General Conduct
	{1, "G206", false, false, "A team or ALLIANCE may not collude with another team to each purposefully violate a rule."},
	{2, "G210", true, false, "A strategy aimed at forcing an opponent to violate a rule is not allowed."},
	{3, "G301", true, false, "A DRIVE TEAM member may not cause significant delays to the start of their MATCH."},
	{4, "G401", false, false, "In AUTO, each DRIVE TEAM member must remain in their staged areas. A DRIVE TEAM member staged behind a HUMAN STARTING LINE may not contact anything in front of that HUMAN STARTING LINE, unless for personal or equipment safety, to press the E-Stop or A-Stop, or granted permission by a Head REFEREE or FTA."},
	{5, "G402", false, false, "In AUTO, a DRIVE TEAM member may not directly or indirectly interact with a ROBOT or an OPERATOR CONSOLE unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop or A-Stop."},
	{6, "G406", true, false, "A ROBOT may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify the challenge associated with a FIELD element."},
	{7, "G407", false, false, "A ROBOT may not intentionally eject a SCORING ELEMENT from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT)."},
	{8, "G408", true, false, "Neither a ROBOT nor a HUMAN PLAYER may damage a SCORING ELEMENT."},
	{9, "G414", false, false, "BUMPERS must be in the BUMPER ZONE."},
	{10, "G415", false, false, "A ROBOT may not extend more than 1 ft. 6 in. beyond the vertical projection of its ROBOT PERIMETER."},
	{11, "G417", true, false, "A ROBOT is prohibited from the following interactions with FIELD elements: grabbing, grasping, attaching to, becoming entangled with, suspending from."},
	{12, "G422", false, false, "A ROBOT may not use a COMPONENT outside its ROBOT PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of the opponent's ROBOT PERIMETER."},
	{13, "G423", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately. B. regardless of intent, by initiating contact, either directly or transitively via a SCORING ELEMENT CONTROLLED by the ROBOT, inside the vertical projection of an opponent's ROBOT PERIMETER."},
	{14, "G424", true, false, "A ROBOT may not deliberately attach to, tip, or entangle with an opponent ROBOT."},
	{15, "G425", false, false, "A ROBOT may not PIN an opponent's ROBOT for more than 3 seconds."},
	{16, "G429", false, false, "A DRIVE TEAM member must remain in their designated area as follows: A. DRIVERS and COACHES may not contact anything outside their ALLIANCE AREA, B. a DRIVER must use the OPERATOR CONSOLE in the DRIVER STATION to which they are assigned, as indicated on the team sign, C. a HUMAN PLAYER may not contact anything outside their ALLIANCE AREA, and D. a TECHNICIAN may not contact anything outside their designated area."},
	{17, "G430", true, false, "A ROBOT shall be operated only by the DRIVERS and/or HUMAN PLAYERS of that team. A COACH activating their E-Stop or A-Stop is the exception to this rule."},
	{18, "G434", false, false, "COACHES may not touch SCORING ELEMENTS, unless for safety purposes."},
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
