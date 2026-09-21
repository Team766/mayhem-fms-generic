// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/Team254/cheesy-arena/playoff"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestAssignTeam(t *testing.T) {
	arena := setupTestArena(t)

	team := model.Team{Id: 254}
	err := arena.Database.CreateTeam(&team)
	assert.Nil(t, err)
	err = arena.Database.CreateTeam(&model.Team{Id: 1114})
	assert.Nil(t, err)

	err = arena.assignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *arena.AllianceStations["B1"].Team)
	dummyDs := &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["B1"].DsConn = dummyDs

	// Nothing should happen if the same team is assigned to the same station.
	err = arena.assignTeam(254, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team, *arena.AllianceStations["B1"].Team)
	assert.NotNil(t, arena.AllianceStations["B1"])
	assert.Equal(t, dummyDs, arena.AllianceStations["B1"].DsConn) // Pointer equality

	// Test reassignment to another team.
	err = arena.assignTeam(1114, "B1")
	assert.Nil(t, err)
	assert.NotEqual(t, team, *arena.AllianceStations["B1"].Team)
	assert.Nil(t, arena.AllianceStations["B1"].DsConn)

	// Check assigning zero as the team number.
	err = arena.assignTeam(0, "R2")
	assert.Nil(t, err)
	assert.Nil(t, arena.AllianceStations["R2"].Team)
	assert.Nil(t, arena.AllianceStations["R2"].DsConn)

	// Check assigning to a non-existent station.
	err = arena.assignTeam(254, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}
}

func TestArenaCheckCanStartMatch(t *testing.T) {
	arena := setupTestArena(t)

	// Check robot state constraints.
	err := arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: not all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	err = arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: not all robots are connected or bypassed")
	}
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.checkCanStartMatch())

	// Check PLC constraints.
	arena.Plc.SetAddress("1.2.3.4")
	err = arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: PLC is not healthy")
	}
	arena.Plc.SetAddress("")
	assert.Nil(t, arena.checkCanStartMatch())

	var plc FakePlc
	plc.isEnabled = true
	arena.Plc = &plc
	// The M-Ayhem field has no FTA ready switch, so that input never blocks a match start.
	assert.Nil(t, arena.checkCanStartMatch())
	plc.ftaReady = true
	assert.Nil(t, arena.checkCanStartMatch())
}

func TestArenaMatchFlow(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 254})
	assert.Nil(t, arena.assignTeam(254, "B3"))
	dummyDs := &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["B3"].DsConn = dummyDs
	arena.Database.CreateTeam(&model.Team{Id: 1678})
	assert.Nil(t, arena.assignTeam(254, "R2"))
	dummyDs = &DriverStationConnection{TeamId: 1678}
	arena.AllianceStations["R2"].DsConn = dummyDs

	// Check pre-match state and packet timing.
	assert.Equal(t, PreMatch, arena.MatchState)
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-300 * time.Millisecond)
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	lastPacketCount := arena.AllianceStations["B3"].DsConn.packetCount
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-10 * time.Millisecond)
	arena.Update()
	assert.Equal(t, lastPacketCount, arena.AllianceStations["B3"].DsConn.packetCount)
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-550 * time.Millisecond)
	arena.Update()
	assert.Equal(t, lastPacketCount+1, arena.AllianceStations["B3"].DsConn.packetCount)

	// Check match start, autonomous and transition to teleop.
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].DsConn.RobotLinked = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].DsConn.RobotLinked = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(game.MatchTiming.AutoDurationSec) * time.Second,
	)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec,
		) * time.Second,
	)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)

	// Check E-stop and bypass.
	arena.AllianceStations["B3"].EStop = true
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-550 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.AllianceStations["B3"].Bypass = true
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-550 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.AllianceStations["B3"].EStop = false
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-550 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.AllianceStations["B3"].Bypass = false
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-550 * time.Millisecond)
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Enabled)

	// Check match end.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.GetTeleopDurationSec(),
		) * time.Second,
	)
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)

	arena.AllianceStations["R1"].Bypass = true
	arena.ResetMatch()
	arena.lastDsPacketTime = arena.lastDsPacketTime.Add(-550 * time.Millisecond)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["B3"].DsConn.Auto)
	assert.Equal(t, false, arena.AllianceStations["B3"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R1"].Bypass)
}

// Verifies that the Companion "endgame start" event fires exactly once, at the moment the teleop period reaches
// the match warning time -- and not before or after -- including on the default test match.
func TestArenaCompanionEndgameStart(t *testing.T) {
	arena := setupTestArena(t)
	assert.Equal(t, model.Test, arena.CurrentMatch.Type)

	// Stand up a fake Companion listener to capture whatever commands the arena sends it.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(t, err)
	defer listener.Close()
	receivedCommands := make(chan string, 10)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				buf := make([]byte, 256)
				if n, err := conn.Read(buf); err == nil {
					receivedCommands <- string(buf[:n])
				}
			}(conn)
		}
	}()
	addr := listener.Addr().(*net.TCPAddr)
	arena.CompanionClient = partner.NewCompanionClient(
		addr.IP.String(),
		addr.Port,
		map[partner.CompanionEvent]partner.CompanionEventConfig{
			partner.EventEndgameStart: {Page: 1, Row: 2, Column: 3},
		},
	)

	endgameStartTimeSec := game.MatchTiming.AutoDurationSec + game.MatchTiming.PauseDurationSec +
		game.GetTeleopDurationSec() - game.MatchTiming.WarningRemainingDurationSec

	assertNoCommandReceived := func() {
		select {
		case command := <-receivedCommands:
			t.Fatalf("expected no Companion command, but got %q", command)
		case <-time.After(150 * time.Millisecond):
		}
	}

	// Just before the warning time, the event should not have fired yet.
	arena.MatchState = TeleopPeriod
	arena.MatchStartTime = time.Now().Add(-time.Duration(endgameStartTimeSec-1) * time.Second)
	arena.Update()
	assertNoCommandReceived()

	// At the warning time, the event should fire exactly once.
	arena.MatchStartTime = time.Now().Add(-time.Duration(endgameStartTimeSec) * time.Second)
	arena.Update()
	select {
	case command := <-receivedCommands:
		assert.Equal(t, "LOCATION 1/2/3 PRESS\n", command)
	case <-time.After(time.Second):
		t.Fatal("expected a Companion command at the warning time, but got none")
	}

	// Continuing through the rest of teleop should not trigger it again.
	arena.MatchStartTime = time.Now().Add(-time.Duration(endgameStartTimeSec+5) * time.Second)
	arena.Update()
	assertNoCommandReceived()
}

func TestArenaStateEnforcement(t *testing.T) {
	arena := setupTestArena(t)

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true

	err := arena.LoadMatch(new(model.Match))
	assert.Nil(t, err)
	err = arena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot abort match when")
	}
	err = arena.StartMatch()
	assert.Nil(t, err)
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: a match is still in progress")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = AutoPeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: a match is still in progress")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = PausePeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: a match is still in progress")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = TeleopPeriod
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: a match is still in progress")
	}
	err = arena.ResetMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot reset match while")
	}
	arena.MatchState = PostMatch
	err = arena.LoadMatch(new(model.Match))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot load match while")
	}
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot start match: a match is still in progress")
	}
	err = arena.AbortMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot abort match when")
	}

	err = arena.ResetMatch()
	assert.Nil(t, err)
	assert.Equal(t, PreMatch, arena.MatchState)
	err = arena.ResetMatch()
	assert.Nil(t, err)
	err = arena.LoadMatch(new(model.Match))
	assert.Nil(t, err)
}

func TestMatchStartRobotLinkEnforcement(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106})
	match := model.Match{Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)

	err := arena.LoadMatch(&match)
	assert.Nil(t, err)
	arena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 101}
	arena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 102}
	arena.AllianceStations["R3"].DsConn = &DriverStationConnection{TeamId: 103}
	arena.AllianceStations["B1"].DsConn = &DriverStationConnection{TeamId: 104}
	arena.AllianceStations["B2"].DsConn = &DriverStationConnection{TeamId: 105}
	arena.AllianceStations["B3"].DsConn = &DriverStationConnection{TeamId: 106}
	for _, station := range arena.AllianceStations {
		station.DsConn.RobotLinked = true
	}
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.MatchState = PreMatch

	// Check with a single team E-stopped, A-stopped, not linked, and bypassed.
	arena.AllianceStations["R1"].EStop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "an emergency stop is active")
	}
	arena.AllianceStations["R1"].EStop = false
	arena.AllianceStations["R1"].aStopReset = false
	arena.AllianceStations["R1"].AStop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "an autonomous stop has not been reset since the previous match")
	}
	arena.AllianceStations["R1"].aStopReset = true
	arena.AllianceStations["R1"].DsConn.RobotLinked = false
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "not all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.AllianceStations["R1"].Bypass = false
	arena.MatchState = PreMatch

	// Check with a team missing.
	err = arena.assignTeam(0, "R1")
	assert.Nil(t, err)
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "not all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.MatchState = PreMatch

	// Check with no teams present.
	arena.LoadMatch(new(model.Match))
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "not all robots are connected or bypassed")
	}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	arena.AllianceStations["B3"].EStop = true
	err = arena.StartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "an emergency stop is active")
	}
	arena.AllianceStations["B3"].EStop = false
	err = arena.StartMatch()
	assert.Nil(t, err)
}

func TestLoadNextMatch(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 1114})
	practiceMatch1 := model.Match{Type: model.Practice, TypeOrder: 1}
	practiceMatch2 := model.Match{Type: model.Practice, TypeOrder: 2, Status: game.RedWonMatch}
	practiceMatch3 := model.Match{Type: model.Practice, TypeOrder: 3}
	arena.Database.CreateMatch(&practiceMatch1)
	arena.Database.CreateMatch(&practiceMatch2)
	arena.Database.CreateMatch(&practiceMatch3)
	qualificationMatch1 := model.Match{Type: model.Qualification, TypeOrder: 1, Status: game.BlueWonMatch}
	qualificationMatch2 := model.Match{Type: model.Qualification, TypeOrder: 2}
	arena.Database.CreateMatch(&qualificationMatch1)
	arena.Database.CreateMatch(&qualificationMatch2)

	// Test match should be followed by another, empty test match.
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	err := arena.SubstituteTeams(1114, 0, 0, 0, 0, 0)
	assert.Nil(t, err)
	arena.CurrentMatch.Status = game.TieMatch
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	assert.Equal(t, 0, arena.CurrentMatch.Red1)
	assert.Equal(t, false, arena.CurrentMatch.IsComplete())

	// Other matches should be loaded by type until they're all complete.
	err = arena.LoadMatch(&practiceMatch2)
	assert.Nil(t, err)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch1.Id, arena.CurrentMatch.Id)
	practiceMatch1.Status = game.RedWonMatch
	arena.Database.UpdateMatch(&practiceMatch1)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, practiceMatch3.Id, arena.CurrentMatch.Id)
	practiceMatch3.Status = game.BlueWonMatch
	arena.Database.UpdateMatch(&practiceMatch3)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, 0, arena.CurrentMatch.Id)
	assert.Equal(t, model.Test, arena.CurrentMatch.Type)

	err = arena.LoadMatch(&qualificationMatch1)
	assert.Nil(t, err)
	err = arena.LoadNextMatch(false)
	assert.Nil(t, err)
	assert.Equal(t, qualificationMatch2.Id, arena.CurrentMatch.Id)
}

func TestSubstituteTeam(t *testing.T) {
	arena := setupTestArena(t)
	tournament.CreateTestAlliances(arena.Database, 2)
	arena.PlayoffTournament, _ = playoff.NewPlayoffTournament(
		arena.EventSettings.PlayoffType, arena.EventSettings.NumPlayoffAlliances,
	)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106})
	arena.Database.CreateTeam(&model.Team{Id: 107})

	// Substitute teams into test match.
	err := arena.SubstituteTeams(0, 0, 0, 101, 0, 0)
	assert.Nil(t, err)
	assert.Equal(t, 101, arena.CurrentMatch.Blue1)
	assert.Equal(t, 101, arena.AllianceStations["B1"].Team.Id)
	err = arena.assignTeam(104, "R4")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Invalid alliance station")
	}

	// Substitute teams into practice match.
	match := model.Match{Type: model.Practice, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	err = arena.SubstituteTeams(107, 102, 103, 104, 105, 106)
	assert.Nil(t, err)
	assert.Equal(t, 107, arena.CurrentMatch.Red1)
	assert.Equal(t, 107, arena.AllianceStations["R1"].Team.Id)
	matchResult := model.NewMatchResult()
	matchResult.MatchId = arena.CurrentMatch.Id

	// Check that substitution is disallowed in qualification matches.
	match = model.Match{Type: model.Qualification, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	err = arena.SubstituteTeams(107, 102, 103, 104, 105, 106)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Can't substitute teams for qualification matches.")
	}
	match = model.Match{Type: model.Playoff, Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	assert.Nil(t, arena.SubstituteTeams(107, 102, 103, 104, 105, 106))

	// Check that loading a nonexistent team fails.
	err = arena.SubstituteTeams(101, 102, 103, 104, 105, 108)
	if assert.NotNil(t, err) {
		assert.Equal(t, err.Error(), "Team 108 is not present at the event.")
	}
}

func TestArenaTimeout(t *testing.T) {
	arena := setupTestArena(t)

	// Test regular ending of timeout.
	timeoutDurationSec := 9
	assert.Nil(t, arena.StartTimeout("Break 1", timeoutDurationSec))
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	assert.Equal(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, "Break 1", arena.breakDescription)
	assert.Equal(t, "Test Match", arena.breakNextMatchName)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, PostTimeout, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec+postTimeoutSec) * time.Second)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)

	// Test ad-hoc timeout display text.
	timeoutDurationSec = 14
	assert.Nil(t, arena.StartAdHocTimeout("Repair Break", "", timeoutDurationSec))
	assert.Equal(t, "Repair Break", arena.breakDescription)
	assert.Equal(t, "", arena.breakNextMatchName)
	arena.SetTimeoutDisplay("Inspection Break", "Practice 1")
	assert.Equal(t, "Inspection Break", arena.breakDescription)
	assert.Equal(t, "Practice 1", arena.breakNextMatchName)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec) * time.Second)
	arena.Update()
	assert.Equal(t, PostTimeout, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec+postTimeoutSec) * time.Second)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)

	// Test early cancellation of timeout.
	timeoutDurationSec = 28
	assert.Nil(t, arena.StartTimeout("Break 2", timeoutDurationSec))
	assert.Equal(t, "Break 2", arena.breakDescription)
	assert.Equal(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	assert.Nil(t, arena.AbortMatch())
	arena.Update()
	assert.Equal(t, PostTimeout, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(-time.Duration(timeoutDurationSec+postTimeoutSec) * time.Second)
	arena.Update()
	assert.Equal(t, PreMatch, arena.MatchState)

	// Test that timeout can't be started during a match.
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.NotNil(t, arena.StartTimeout("Timeout", 1))
	assert.NotEqual(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, timeoutDurationSec, game.MatchTiming.TimeoutDurationSec)
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.GetTeleopDurationSec(),
		) * time.Second,
	)
	for arena.MatchState != PostMatch {
		arena.Update()
		assert.NotNil(t, arena.StartTimeout("Timeout", 1))
	}

	// Test that a match can be loaded during a timeout.
	assert.Nil(t, arena.ResetMatch())
	assert.Nil(t, arena.LoadTestMatch())
	assert.Nil(t, arena.StartTimeout("Break 2", 10))
	assert.Equal(t, TimeoutActive, arena.MatchState)
	match := model.Match{
		Type: model.Playoff, ShortName: "F1", Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6,
	}
	assert.Nil(t, arena.Database.CreateMatch(&match))
	assert.Nil(t, arena.LoadMatch(&match))
	assert.Equal(t, TimeoutActive, arena.MatchState)
	assert.Equal(t, match, *arena.CurrentMatch)
}

func TestSaveTeamHasConnected(t *testing.T) {
	arena := setupTestArena(t)

	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 102})
	arena.Database.CreateTeam(&model.Team{Id: 103})
	arena.Database.CreateTeam(&model.Team{Id: 104})
	arena.Database.CreateTeam(&model.Team{Id: 105})
	arena.Database.CreateTeam(&model.Team{Id: 106, City: "San Jose", HasConnected: true})
	match := model.Match{Red1: 101, Red2: 102, Red3: 103, Blue1: 104, Blue2: 105, Blue3: 106}
	arena.Database.CreateMatch(&match)
	arena.LoadMatch(&match)
	arena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 101}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 102, RobotLinked: true}
	arena.AllianceStations["R3"].DsConn = &DriverStationConnection{TeamId: 103}
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].DsConn = &DriverStationConnection{TeamId: 104}
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].DsConn = &DriverStationConnection{TeamId: 105, RobotLinked: true}
	arena.AllianceStations["B3"].DsConn = &DriverStationConnection{TeamId: 106, RobotLinked: true}
	arena.AllianceStations["B3"].Team.City = "Sand Hosay" // Change some other field to verify that it isn't saved.
	assert.Nil(t, arena.StartMatch())

	// Check that the connection status was saved for the teams that just linked for the first time.
	teams, _ := arena.Database.GetAllTeams()
	if assert.Equal(t, 6, len(teams)) {
		assert.False(t, teams[0].HasConnected)
		assert.True(t, teams[1].HasConnected)
		assert.False(t, teams[2].HasConnected)
		assert.False(t, teams[3].HasConnected)
		assert.True(t, teams[4].HasConnected)
		assert.True(t, teams[5].HasConnected)
		assert.Equal(t, "San Jose", teams[5].City)
	}
}

func TestPlcEStopAStop(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	plc.ftaReady = true
	arena.Plc = &plc

	arena.Database.CreateTeam(&model.Team{Id: 254})
	err := arena.assignTeam(254, "R1")
	assert.Nil(t, err)
	dummyDs := &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["R1"].DsConn = dummyDs
	arena.Database.CreateTeam(&model.Team{Id: 148})
	err = arena.assignTeam(148, "R2")
	assert.Nil(t, err)
	dummyDs = &DriverStationConnection{TeamId: 148}
	arena.AllianceStations["R2"].DsConn = dummyDs

	arena.AllianceStations["R1"].DsConn.RobotLinked = true
	arena.AllianceStations["R1"].aStopReset = true
	arena.AllianceStations["R2"].DsConn.RobotLinked = true
	arena.AllianceStations["R2"].aStopReset = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["R3"].aStopReset = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B1"].aStopReset = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B2"].aStopReset = true
	arena.AllianceStations["B3"].Bypass = true
	arena.AllianceStations["B3"].aStopReset = true
	err = arena.StartMatch()
	assert.Nil(t, err)
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)

	// Press the R1 A-stop.
	plc.redAStops[0] = true
	plc.redEStops[0] = false
	plc.redAStops[1] = false
	plc.redEStops[1] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.EStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.Enabled)

	// Unpress the R1 A-stop and press the R2 E-stop.
	plc.redAStops[0] = false
	plc.redEStops[0] = false
	plc.redAStops[1] = false
	plc.redEStops[1] = true
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.EStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.AStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.AStop)

	// Unpress the R2 E-stop.
	plc.redAStops[0] = false
	plc.redEStops[0] = false
	plc.redAStops[1] = false
	plc.redEStops[1] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	// Transition into the teleop period without any stops.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(game.MatchTiming.AutoDurationSec) * time.Second,
	)
	arena.Update()
	assert.Equal(t, PausePeriod, arena.MatchState)
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec,
		) * time.Second,
	)
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	// Press the R1 E-stop and the R2 A-stop.
	plc.redAStops[0] = false
	plc.redEStops[0] = true
	plc.redAStops[1] = true
	plc.redEStops[1] = false
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].EStop)
	arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].DsConn.Enabled)

	// Ensure the other stations A-stops are working as well.
	plc.redAStops[2] = true
	plc.redEStops[2] = false
	plc.blueAStops[0] = true
	plc.blueEStops[0] = false
	plc.blueAStops[1] = true
	plc.blueEStops[1] = false
	plc.blueAStops[2] = true
	plc.blueEStops[2] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R3"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R3"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["B1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B2"].AStop)
	assert.Equal(t, false, arena.AllianceStations["B2"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B3"].AStop)
	assert.Equal(t, false, arena.AllianceStations["B3"].EStop)

	// Ensure the other stations E-stops are working as well.
	plc.redAStops[2] = false
	plc.redEStops[2] = true
	plc.blueAStops[0] = false
	plc.blueEStops[0] = true
	plc.blueAStops[1] = false
	plc.blueEStops[1] = true
	plc.blueAStops[2] = false
	plc.blueEStops[2] = true
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R3"].AStop)
	assert.Equal(t, true, arena.AllianceStations["R3"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B1"].AStop)
	assert.Equal(t, true, arena.AllianceStations["B1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B2"].AStop)
	assert.Equal(t, true, arena.AllianceStations["B2"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B3"].AStop)
	assert.Equal(t, true, arena.AllianceStations["B3"].EStop)

	// Ensure unpressed E-stops are cleared at the end of the match.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.GetTeleopDurationSec(),
		) * time.Second,
	)
	arena.Update()
	plc.blueEStops[2] = false
	arena.Update()
	assert.Equal(t, true, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R3"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["B2"].EStop)
	assert.Equal(t, false, arena.AllianceStations["B3"].EStop)
}

func TestPlcEStopAStopWithPlcDisabled(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = false
	arena.Plc = &plc

	arena.Database.CreateTeam(&model.Team{Id: 254})
	err := arena.assignTeam(254, "R1")
	assert.Nil(t, err)
	arena.AllianceStations["R1"].DsConn = &DriverStationConnection{TeamId: 254}
	arena.AllianceStations["R2"].DsConn = &DriverStationConnection{TeamId: 1323}

	arena.AllianceStations["R1"].DsConn.RobotLinked = true
	arena.AllianceStations["R2"].DsConn.RobotLinked = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)

	plc.redEStops[0] = true
	plc.redAStops[1] = true
	arena.Update()
	assert.Equal(t, false, arena.AllianceStations["R1"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R1"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R1"].DsConn.Enabled)
	assert.Equal(t, false, arena.AllianceStations["R2"].AStop)
	assert.Equal(t, false, arena.AllianceStations["R2"].EStop)
	assert.Equal(t, true, arena.AllianceStations["R2"].DsConn.Enabled)
}

func TestPlcFieldEStop(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	plc.ftaReady = true
	arena.Plc = &plc

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)

	plc.fieldEStop = true
	arena.Update()
	assert.True(t, arena.matchAborted)
	assert.Equal(t, PostMatch, arena.MatchState)
}

func TestPlcFieldEStopWithPlcDisabled(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = false
	arena.Plc = &plc

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)

	plc.fieldEStop = true
	arena.Update()
	assert.False(t, arena.matchAborted)
	assert.Equal(t, AutoPeriod, arena.MatchState)
}

func TestPlcMatchCycleEvergreen(t *testing.T) {
	arena := setupTestArena(t)
	var plc FakePlc
	plc.isEnabled = true
	plc.ftaReady = true
	arena.Plc = &plc

	arena.Update()
	assert.Equal(t, [4]bool{true, true, false, false}, plc.stackLights)

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.Update()
	assert.Equal(t, [4]bool{true, true, false, false}, plc.stackLights)

	arena.AllianceStations["R3"].Bypass = true
	arena.Update()
	assert.Equal(t, [4]bool{false, true, false, false}, plc.stackLights)
	assert.Equal(t, false, plc.stackLightBuzzer)

	// All teams are ready.
	arena.AllianceStations["B3"].Bypass = true
	plc.cycleState = true
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, true}, plc.stackLights)
	assert.Equal(t, true, plc.stackLightBuzzer)

	// Green light when blink cycle is off.
	plc.cycleState = false
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, false}, plc.stackLights)

	// Start the match.
	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)
	assert.Equal(t, [4]bool{false, false, false, true}, plc.stackLights)
	assert.Equal(t, false, plc.stackLightBuzzer)

	// End the match.
	arena.MatchStartTime = time.Now().Add(
		-time.Duration(
			game.MatchTiming.AutoDurationSec+game.MatchTiming.PauseDurationSec+game.GetTeleopDurationSec(),
		) * time.Second,
	)
	arena.Update()
	arena.Update()
	arena.Update()
	assert.Equal(t, PostMatch, arena.MatchState)
	assert.Equal(t, [4]bool{false, false, true, false}, plc.stackLights)
	assert.Equal(t, false, plc.fieldResetLight)

	// Ready the score.
	arena.RedRealtimeScore.FoulsCommitted = true
	arena.BlueRealtimeScore.FoulsCommitted = true
	redWs := &websocket.Websocket{}
	arena.ScoringPanelRegistry.RegisterPanel("red", redWs)
	arena.ScoringPanelRegistry.SetScoreCommitted("red", redWs)
	arena.Update()
	assert.Equal(t, [4]bool{false, false, true, false}, plc.stackLights)
	blueWs := &websocket.Websocket{}
	arena.ScoringPanelRegistry.RegisterPanel("blue", blueWs)
	arena.ScoringPanelRegistry.SetScoreCommitted("blue", blueWs)
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, false}, plc.stackLights)

	arena.FieldReset = true
	arena.Update()
	assert.Equal(t, true, plc.fieldResetLight)

	assert.Equal(t, false, plc.awardsModeLight)

	arena.SetAllianceStationDisplayMode("logo")
	arena.Update()
	assert.Equal(t, true, plc.awardsModeLight)

	arena.SetAllianceStationDisplayMode("match")
	arena.Update()
	assert.Equal(t, false, plc.awardsModeLight)
}

func TestSignalVolunteers(t *testing.T) {
	arena := setupTestArena(t)

	// Test that SignalVolunteers only works in PreMatch and PostMatch states.
	for _, state := range []MatchState{StartMatch, AutoPeriod, PausePeriod, TeleopPeriod, PostTimeout} {
		arena.MatchState = state
		arena.FieldVolunteers = false
		arena.SignalVolunteers()
		assert.False(t, arena.FieldVolunteers)
		assert.NotEqual(t, "signalCount", arena.AllianceStationDisplayMode)
	}

	// Test SignalVolunteers in PreMatch state.
	arena.MatchState = PreMatch
	arena.FieldReset = true
	arena.AllianceStationDisplayMode = "match"
	arena.SignalVolunteers()
	assert.True(t, arena.FieldVolunteers)
	assert.False(t, arena.FieldReset)
	assert.Equal(t, "signalCount", arena.AllianceStationDisplayMode)

	// Test SignalVolunteers in PostMatch state.
	arena.MatchState = PostMatch
	arena.FieldVolunteers = false
	arena.FieldReset = false
	arena.AllianceStationDisplayMode = "match"
	arena.SignalVolunteers()
	assert.True(t, arena.FieldVolunteers)
	assert.False(t, arena.FieldReset)
	assert.Equal(t, "signalCount", arena.AllianceStationDisplayMode)
}

func TestSignalReset(t *testing.T) {
	arena := setupTestArena(t)

	// Test that SignalReset only works in PreMatch and PostMatch states.
	for _, state := range []MatchState{StartMatch, AutoPeriod, PausePeriod, TeleopPeriod, PostTimeout} {
		arena.MatchState = state
		arena.FieldReset = false
		arena.FieldVolunteers = false
		arena.AllianceStationDisplayMode = "match"
		arena.SignalReset()
		assert.False(t, arena.FieldReset)
		assert.False(t, arena.FieldVolunteers)
		assert.NotEqual(t, "fieldReset", arena.AllianceStationDisplayMode)
	}

	// Test SignalReset in PreMatch state.
	arena.MatchState = PreMatch
	arena.FieldReset = false
	arena.FieldVolunteers = true
	arena.AllianceStationDisplayMode = "match"
	arena.SignalReset()
	assert.False(t, arena.FieldVolunteers)
	assert.True(t, arena.FieldReset)
	assert.Equal(t, "fieldReset", arena.AllianceStationDisplayMode)

	// Test SignalReset in PostMatch state.
	arena.MatchState = PostMatch
	arena.FieldReset = false
	arena.FieldVolunteers = true
	arena.AllianceStationDisplayMode = "match"
	arena.SignalReset()
	assert.False(t, arena.FieldVolunteers)
	assert.True(t, arena.FieldReset)
	assert.Equal(t, "fieldReset", arena.AllianceStationDisplayMode)
}

func TestTwoVsTwoActiveStations(t *testing.T) {
	testCases := []struct {
		name         string
		twoVsTwoMode bool
		teams        map[string]int
		expected     []string
	}{
		{"3v3, all empty", false, nil, []string{"R1", "R2", "R3", "B1", "B2", "B3"}},
		{
			"3v3, all full",
			false,
			map[string]int{"R1": 1, "R2": 2, "R3": 3, "B1": 4, "B2": 5, "B3": 6},
			[]string{"R1", "R2", "R3", "B1", "B2", "B3"},
		},
		{"2v2, all empty", true, nil, []string{"R1", "R2", "B1", "B2"}},
		{"2v2, R3 holds a team", true, map[string]int{"R3": 3}, []string{"R1", "R2", "R3", "B1", "B2"}},
		{"2v2, B3 holds a team", true, map[string]int{"B3": 6}, []string{"R1", "R2", "B1", "B2", "B3"}},
		{
			"2v2, 3v3 match still loaded",
			true,
			map[string]int{"R1": 1, "R2": 2, "R3": 3, "B1": 4, "B2": 5, "B3": 6},
			[]string{"R1", "R2", "R3", "B1", "B2", "B3"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			arena := setupTestArena(t)
			for station, teamId := range testCase.teams {
				assert.Nil(t, arena.Database.CreateTeam(&model.Team{Id: teamId}))
				assert.Nil(t, arena.assignTeam(teamId, station))
			}
			arena.EventSettings.TwoVsTwoMode = testCase.twoVsTwoMode
			assert.Equal(t, testCase.expected, arena.activeStations())
		})
	}

	// A driver station connection alone also keeps R3 in play.
	arena := setupTestArena(t)
	arena.EventSettings.TwoVsTwoMode = true
	arena.AllianceStations["R3"].DsConn = &DriverStationConnection{TeamId: 3}
	assert.Equal(t, []string{"R1", "R2", "R3", "B1", "B2"}, arena.activeStations())
}

// setupLinkedTestArenaWithPlc returns an arena with a PLC in which each of the given stations holds a team with a
// linked robot and a reset A-stop.
func setupLinkedTestArenaWithPlc(t *testing.T, stations ...string) (*Arena, *FakePlc) {
	arena := setupTestArena(t)
	plc := &FakePlc{isEnabled: true, ftaReady: true}
	arena.Plc = plc
	for i, station := range stations {
		teamId := 100 + i
		assert.Nil(t, arena.Database.CreateTeam(&model.Team{Id: teamId}))
		assert.Nil(t, arena.assignTeam(teamId, station))
		// lastPacketTime is set to now so that Update()'s driver station packet send doesn't immediately time out
		// and clear RobotLinked back out from under the test.
		arena.AllianceStations[station].DsConn = &DriverStationConnection{
			TeamId: teamId, RobotLinked: true, lastPacketTime: time.Now(),
		}
		arena.AllianceStations[station].aStopReset = true
	}
	return arena, plc
}

// Verifies that in 2v2, a PLC E-stop on any of the four stations in play is honoured and blocks the match start.
func TestTwoVsTwoPlcEStopOnActiveStation(t *testing.T) {
	for i, station := range []string{"R1", "R2", "B1", "B2"} {
		t.Run(station, func(t *testing.T) {
			arena, plc := setupLinkedTestArenaWithPlc(t, "R1", "R2", "B1", "B2")
			arena.EventSettings.TwoVsTwoMode = true
			plc.cycleState = true
			arena.Update()
			assert.Nil(t, arena.checkCanStartMatch())

			if i < 2 {
				plc.redEStops[i] = true
			} else {
				plc.blueEStops[i-2] = true
			}
			arena.Update()
			assert.True(t, arena.AllianceStations[station].EStop)
			err := arena.StartMatch()
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), fmt.Sprintf("an emergency stop is active (%s)", station))
			}
			assert.Equal(t, PreMatch, arena.MatchState)
			assert.False(t, plc.stackLights[3])
		})
	}
}

// Verifies that turning 2v2 on while a 3v3 match is loaded doesn't take R3/B3 out of play: their E-stops are still
// honoured (before and during the match) and they still block the start when not ready.
func TestTwoVsTwoSettingTurnedOnWithThreeVsThreeMatchLoaded(t *testing.T) {
	testCases := []struct {
		station  string
		eStopIdx int
		red      bool
	}{
		{"R3", 2, true},
		{"B3", 2, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.station, func(t *testing.T) {
			arena, plc := setupLinkedTestArenaWithPlc(t, "R1", "R2", "R3", "B1", "B2", "B3")
			match := model.Match{
				Type: model.Practice, Red1: 100, Red2: 101, Red3: 102, Blue1: 103, Blue2: 104, Blue3: 105,
			}
			assert.Nil(t, arena.LoadMatch(&match))
			arena.EventSettings.TwoVsTwoMode = true
			plc.cycleState = true
			pressEStop := func(pressed bool) {
				if testCase.red {
					plc.redEStops[testCase.eStopIdx] = pressed
				} else {
					plc.blueEStops[testCase.eStopIdx] = pressed
				}
			}

			// A pressed E-stop blocks the start.
			pressEStop(true)
			arena.Update()
			assert.True(t, arena.AllianceStations[testCase.station].EStop)
			err := arena.checkCanStartMatch()
			if assert.NotNil(t, err) {
				assert.Contains(t, err.Error(), fmt.Sprintf("an emergency stop is active (%s)", testCase.station))
			}
			assert.False(t, plc.stackLights[3])

			// An unlinked, unbypassed robot blocks the start and keeps the stack light from going green.
			pressEStop(false)
			arena.AllianceStations[testCase.station].DsConn.RobotLinked = false
			arena.Update()
			assert.False(t, arena.AllianceStations[testCase.station].EStop)
			err = arena.checkCanStartMatch()
			if assert.NotNil(t, err) {
				assert.Contains(
					t, err.Error(), fmt.Sprintf("not all robots are connected or bypassed (%s)", testCase.station),
				)
			}
			assert.False(t, plc.stackLights[3])

			// Once ready, the match starts, and an E-stop pressed mid-match disables that robot.
			arena.AllianceStations[testCase.station].DsConn.RobotLinked = true
			arena.Update()
			assert.Equal(t, [4]bool{false, false, false, true}, plc.stackLights)
			assert.Nil(t, arena.StartMatch())
			arena.Update()
			assert.Equal(t, AutoPeriod, arena.MatchState)
			assert.True(t, arena.AllianceStations[testCase.station].DsConn.Enabled)
			pressEStop(true)
			arena.Update()
			arena.lastDsPacketTime = time.Unix(0, 0) // Force a DS packet.
			arena.Update()
			assert.True(t, arena.AllianceStations[testCase.station].EStop)
			assert.True(t, arena.AllianceStations[testCase.station].DsConn.EStop)
			assert.False(t, arena.AllianceStations[testCase.station].DsConn.Enabled)
		})
	}
}

func TestTwoVsTwoCheckCanStartMatch(t *testing.T) {
	arena := setupTestArena(t)
	arena.EventSettings.TwoVsTwoMode = true

	// R3/B3 must never block the match from starting, whether or not they are bypassed, and without a PLC.
	err := arena.checkCanStartMatch()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "not all robots are connected or bypassed (R1, R2, B1, B2)")
		assert.NotContains(t, err.Error(), "R3")
		assert.NotContains(t, err.Error(), "B3")
	}
	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	assert.Nil(t, arena.checkCanStartMatch())
}

func TestTwoVsTwoLoadMatchRejectsThirdTeam(t *testing.T) {
	arena := setupTestArena(t)
	arena.EventSettings.TwoVsTwoMode = true
	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 104})

	match := model.Match{Type: model.Practice, Red1: 101, Red3: 103, Blue1: 104}
	err := arena.LoadMatch(&match)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "a third robot is not allowed in 2v2 mode")
	}

	match = model.Match{Type: model.Practice, Red1: 101, Blue1: 104, Blue3: 106}
	err = arena.LoadMatch(&match)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "a third robot is not allowed in 2v2 mode")
	}

	match = model.Match{Type: model.Practice, Red1: 101, Blue1: 104}
	assert.Nil(t, arena.LoadMatch(&match))
	assert.Nil(t, arena.AllianceStations["R3"].Team)
	assert.Nil(t, arena.AllianceStations["B3"].Team)
}

func TestTwoVsTwoSubstituteTeamsRejectsThirdTeam(t *testing.T) {
	arena := setupTestArena(t)
	arena.EventSettings.TwoVsTwoMode = true
	arena.Database.CreateTeam(&model.Team{Id: 101})
	arena.Database.CreateTeam(&model.Team{Id: 104})

	match := model.Match{Type: model.Practice, Red1: 101, Blue1: 104}
	arena.Database.CreateMatch(&match)
	assert.Nil(t, arena.LoadMatch(&match))

	err := arena.SubstituteTeams(101, 0, 103, 104, 0, 0)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "a third robot is not allowed in 2v2 mode")
	}
	err = arena.SubstituteTeams(101, 0, 0, 104, 0, 106)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "a third robot is not allowed in 2v2 mode")
	}
	assert.Nil(t, arena.SubstituteTeams(101, 0, 0, 104, 0, 0))
}

// Verifies that in 2v2, empty R3/B3 -- no team, never bypassed -- do not stop the readiness stack light from
// going green or the match from starting, and that their E-stops are ignored; and that the same E-stop press
// is NOT ignored once back in 3v3.
func TestTwoVsTwoPlcReadyAndEStop(t *testing.T) {
	arena, plc := setupLinkedTestArenaWithPlc(t, "R1", "R2", "B1", "B2")
	arena.EventSettings.TwoVsTwoMode = true

	// R3/B3 are empty and never bypassed; readiness must still reach green and the match must be startable.
	plc.cycleState = true
	arena.Update()
	assert.Equal(t, [4]bool{false, false, false, true}, plc.stackLights)
	assert.Nil(t, arena.checkCanStartMatch())
	assert.False(t, arena.AllianceStations["R3"].Bypass)
	assert.False(t, arena.AllianceStations["B3"].Bypass)

	// Empty R3/B3 E-stops are ignored in 2v2.
	plc.redEStops[2] = true
	plc.blueEStops[2] = true
	arena.Update()
	assert.False(t, arena.AllianceStations["R3"].EStop)
	assert.False(t, arena.AllianceStations["B3"].EStop)
	assert.Nil(t, arena.checkCanStartMatch())

	// The same E-stop press is NOT ignored in 3v3.
	arena.EventSettings.TwoVsTwoMode = false
	arena.Update()
	assert.True(t, arena.AllianceStations["R3"].EStop)
	assert.True(t, arena.AllianceStations["B3"].EStop)
}
