// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"bytes"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupSettings(t *testing.T) {
	web := setupTestWeb(t)

	// Check the default setting values.
	recorder := web.getHttpResponse("/setup/settings")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Untitled Event")
	assert.Contains(t, recorder.Body.String(), "8")
	assert.Contains(t, recorder.Body.String(), "teleopDurationSec")
	assert.Contains(t, recorder.Body.String(), "warningRemainingDurationSec")

	// Change the settings and check the response.
	recorder = web.postHttpResponse(
		"/setup/settings",
		"name=Chezy Champs&code=CC&playoffType=single&numPlayoffAlliances=16&"+
			"eventCode=2014cc&teleopDurationSec=106&warningRemainingDurationSec=25&"+
			"companionEndgameStartPage=1&companionEndgameStartRow=2&companionEndgameStartColumn=3&twoVsTwoMode=on",
	)
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/setup/settings#event", recorder.Header().Get("Location"))
	recorder = web.getHttpResponse("/setup/settings")
	assert.Contains(t, recorder.Body.String(), "Chezy Champs")
	assert.Contains(t, recorder.Body.String(), "16")
	assert.Contains(t, recorder.Body.String(), "2014cc")
	assert.Equal(t, 106, web.arena.EventSettings.TeleopDurationSec)
	assert.Equal(t, 25, web.arena.EventSettings.WarningRemainingDurationSec)
	assert.Equal(t, 106, game.GetTeleopDurationSec())
	assert.Equal(t, 1, web.arena.EventSettings.CompanionEndgameStartPage)
	assert.Equal(t, 2, web.arena.EventSettings.CompanionEndgameStartRow)
	assert.Equal(t, 3, web.arena.EventSettings.CompanionEndgameStartColumn)
	assert.True(t, web.arena.EventSettings.TwoVsTwoMode)

	// Turn 2v2 mode back off and check that it round-trips.
	recorder = web.postHttpResponse(
		"/setup/settings", "name=Chezy Champs&code=CC&playoffType=single&numPlayoffAlliances=16",
	)
	assert.Equal(t, 303, recorder.Code)
	assert.False(t, web.arena.EventSettings.TwoVsTwoMode)

	recorder = web.postHttpResponse("/setup/settings", "name=Field Tab Event&activeSettingsTab=field")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/setup/settings#field", recorder.Header().Get("Location"))
}

func TestSetupSettingsTwoVsTwoModeLocking(t *testing.T) {
	testCases := []struct {
		name                  string
		existingMatchType     model.MatchType
		initialTwoVsTwoMode   bool
		requestTwoVsTwoMode   bool
		wantStatusCode        int
		wantErrorMessage      string
		wantFinalTwoVsTwoMode bool
	}{
		{
			name:                  "no matches, turn on",
			initialTwoVsTwoMode:   false,
			requestTwoVsTwoMode:   true,
			wantStatusCode:        303,
			wantFinalTwoVsTwoMode: true,
		},
		{
			name:                  "no matches, turn off",
			initialTwoVsTwoMode:   true,
			requestTwoVsTwoMode:   false,
			wantStatusCode:        303,
			wantFinalTwoVsTwoMode: false,
		},
		{
			name:                  "practice matches alone do not lock it",
			existingMatchType:     model.Practice,
			initialTwoVsTwoMode:   false,
			requestTwoVsTwoMode:   true,
			wantStatusCode:        303,
			wantFinalTwoVsTwoMode: true,
		},
		{
			name:                "qualification schedule exists, turning on is refused",
			existingMatchType:   model.Qualification,
			initialTwoVsTwoMode: false,
			requestTwoVsTwoMode: true,
			wantStatusCode:      200,
			wantErrorMessage: "The 2v2 setting can't be changed once the qualification schedule exists. Clear the " +
				"qualification schedule first to change it.",
			wantFinalTwoVsTwoMode: false,
		},
		{
			name:                "qualification schedule exists, turning off is refused",
			existingMatchType:   model.Qualification,
			initialTwoVsTwoMode: true,
			requestTwoVsTwoMode: false,
			wantStatusCode:      200,
			wantErrorMessage: "The 2v2 setting can't be changed once the qualification schedule exists. Clear the " +
				"qualification schedule first to change it.",
			wantFinalTwoVsTwoMode: true,
		},
		{
			name:                  "qualification schedule exists, unchanged value is saved",
			existingMatchType:     model.Qualification,
			initialTwoVsTwoMode:   true,
			requestTwoVsTwoMode:   true,
			wantStatusCode:        303,
			wantFinalTwoVsTwoMode: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			web := setupTestWeb(t)
			web.arena.EventSettings.TwoVsTwoMode = testCase.initialTwoVsTwoMode
			assert.Nil(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))
			if testCase.existingMatchType != model.Test {
				assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: testCase.existingMatchType}))
			}

			body := "name=Test Event"
			if testCase.requestTwoVsTwoMode {
				body += "&twoVsTwoMode=on"
			}
			recorder := web.postHttpResponse("/setup/settings", body)

			assert.Equal(t, testCase.wantStatusCode, recorder.Code)
			if testCase.wantErrorMessage != "" {
				assert.Contains(t, recorder.Body.String(), testCase.wantErrorMessage)
			}
			assert.Equal(t, testCase.wantFinalTwoVsTwoMode, web.arena.EventSettings.TwoVsTwoMode)
		})
	}
}

func TestSetupSettingsOtherFieldsSaveWhileTwoVsTwoModeLocked(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.TwoVsTwoMode = true
	assert.Nil(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))
	assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Qualification}))

	// The checkbox is disabled in the browser, so a real submission still carries the current value via the hidden
	// input; simulate that here.
	recorder := web.postHttpResponse("/setup/settings", "name=Locked Event&twoVsTwoMode=on")

	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "Locked Event", web.arena.EventSettings.Name)
	assert.True(t, web.arena.EventSettings.TwoVsTwoMode)
}

func TestSetupSettingsTwoVsTwoModeLockedRendering(t *testing.T) {
	web := setupTestWeb(t)
	assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Qualification}))

	recorder := web.getHttpResponse("/setup/settings")

	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Locked: the qualification schedule exists.")
	assert.Contains(t, recorder.Body.String(), `name="twoVsTwoMode"  disabled>`)
}

func TestSetupSettingsBlockedDuringMatch(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.Name = "Original Event"
	web.arena.MatchState = field.AutoPeriod

	recorder := web.postHttpResponse("/setup/settings", "name=Changed Event&activeSettingsTab=field")

	assert.Equal(t, 200, recorder.Code)
	assert.Contains(
		t, recorder.Body.String(), "Settings cannot be changed while a match is in progress or is uncommitted.",
	)
	assert.Contains(t, recorder.Body.String(), "hash = \"#field\"")
	assert.Equal(t, "Original Event", web.arena.EventSettings.Name)
}

func TestSetupSettingsAllowedDuringTimeoutStates(t *testing.T) {
	for _, matchState := range []field.MatchState{field.TimeoutActive, field.PostTimeout} {
		web := setupTestWeb(t)
		web.arena.MatchState = matchState

		recorder := web.postHttpResponse("/setup/settings", "name=Changed Event")

		assert.Equal(t, 303, recorder.Code)
		assert.Equal(t, "Changed Event", web.arena.EventSettings.Name)
	}
}

func TestSettingsSaveAllowed(t *testing.T) {
	testCases := []struct {
		matchState field.MatchState
		allowed    bool
	}{
		{field.PreMatch, true},
		{field.StartMatch, false},
		{field.AutoPeriod, false},
		{field.PausePeriod, false},
		{field.TeleopPeriod, false},
		{field.PostMatch, false},
		{field.TimeoutActive, true},
		{field.PostTimeout, true},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.allowed, settingsSaveAllowed(testCase.matchState))
	}
}

func TestSetupSettingsDoubleElimination(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.postHttpResponse("/setup/settings", "playoffType=DoubleEliminationPlayoff&numPlayoffAlliances=4")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, model.DoubleEliminationPlayoff, web.arena.EventSettings.PlayoffType)
	assert.Equal(t, 4, web.arena.EventSettings.NumPlayoffAlliances)

	recorder = web.postHttpResponse("/setup/settings", "playoffType=DoubleEliminationPlayoff&numPlayoffAlliances=3")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Number of alliances for double elimination must be 4 or 8.")
	assert.Equal(t, 4, web.arena.EventSettings.NumPlayoffAlliances)
}

func TestSetupSettingsSingleGameUntilFinals(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.postHttpResponse(
		"/setup/settings", "playoffType=SingleGameUntilFinalsPlayoff&numPlayoffAlliances=6",
	)
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, model.SingleGameUntilFinalsPlayoff, web.arena.EventSettings.PlayoffType)
	assert.Equal(t, 6, web.arena.EventSettings.NumPlayoffAlliances)

	// The setting should round-trip through the settings page and be preserved when the playoff type isn't
	// explicitly resubmitted.
	recorder = web.getHttpResponse("/setup/settings")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "SingleGameUntilFinalsPlayoff")

	recorder = web.postHttpResponse("/setup/settings", "name=Renamed Event")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, model.SingleGameUntilFinalsPlayoff, web.arena.EventSettings.PlayoffType)
	assert.Equal(t, 6, web.arena.EventSettings.NumPlayoffAlliances)

	// Invalid number of alliances for this playoff type.
	recorder = web.postHttpResponse(
		"/setup/settings", "playoffType=SingleGameUntilFinalsPlayoff&numPlayoffAlliances=1",
	)
	assert.Contains(t, recorder.Body.String(), "must be between 2 and 16")
}

func TestSetupSettingsInvalidValues(t *testing.T) {
	web := setupTestWeb(t)
	recorder := web.postHttpResponse("/setup/settings", "playoffType=SingleEliminationPlayoff&numPlayoffAlliances=8")
	assert.Equal(t, 303, recorder.Code)

	// Invalid number of alliances.
	recorder = web.postHttpResponse("/setup/settings", "playoffType=SingleEliminationPlayoff&numAlliances=1")
	assert.Contains(t, recorder.Body.String(), "must be between 2 and 16")

	recorder = web.postHttpResponse("/setup/settings", "playoffType=DoubleEliminationPlayoff&numPlayoffAlliances=3")
	assert.Contains(t, recorder.Body.String(), "must be 4 or 8")

	// Changing the playoff type after alliance selection is finalized.
	assert.Nil(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1}))
	recorder = web.postHttpResponse("/setup/settings", "playoffType=DoubleEliminationPlayoff")
	assert.Contains(t, recorder.Body.String(), "Cannot change playoff type or size after alliance selection")

	// Changing the playoff size after alliance selection is finalized.
	recorder = web.postHttpResponse("/setup/settings", "numPlayoffAlliances=2")
	assert.Contains(t, recorder.Body.String(), "Cannot change playoff type or size after alliance selection")
}

func TestSetupSettingsClearDb(t *testing.T) {
	createData := func(web *Web) {
		assert.Nil(t, web.arena.Database.CreateTeam(&model.Team{Id: 254}))
		assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Practice}))
		assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Qualification}))
		assert.Nil(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Playoff}))
		assert.Nil(t, web.arena.Database.CreateMatchResult(&model.MatchResult{MatchId: 1, PlayNumber: 1}))
		assert.Nil(t, web.arena.Database.CreateMatchResult(&model.MatchResult{MatchId: 1, PlayNumber: 2}))
		assert.Nil(t, web.arena.Database.CreateMatchResult(&model.MatchResult{MatchId: 2, PlayNumber: 1}))
		assert.Nil(t, web.arena.Database.CreateMatchResult(&model.MatchResult{MatchId: 3, PlayNumber: 1}))
		assert.Nil(t, web.arena.Database.CreateRanking(&game.Ranking{TeamId: 254}))
		assert.Nil(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1}))
		web.arena.AllianceSelectionAlliances = append(web.arena.AllianceSelectionAlliances, model.Alliance{Id: 1})
	}

	// Test clearing practice data.
	web := setupTestWeb(t)
	createData(web)
	recorder := web.postHttpResponse("/setup/db/clear/practice", "")
	assert.Equal(t, 303, recorder.Code)
	teams, _ := web.arena.Database.GetAllTeams()
	assert.NotEmpty(t, teams)
	matches, _ := web.arena.Database.GetMatchesByType(model.Practice, true)
	assert.Empty(t, matches)
	matchResult, _ := web.arena.Database.GetMatchResultForMatch(1)
	assert.Nil(t, matchResult)
	matches, _ = web.arena.Database.GetMatchesByType(model.Qualification, true)
	assert.NotEmpty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(2)
	assert.NotNil(t, matchResult)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(3)
	assert.NotNil(t, matchResult)
	rankings, _ := web.arena.Database.GetAllRankings()
	assert.NotEmpty(t, rankings)
	tournament.CalculateRankings(web.arena.Database, false)
	assert.NotEmpty(t, rankings)
	alliances, _ := web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	assert.NotEmpty(t, web.arena.AllianceSelectionAlliances)

	// Test clearing qualification data.
	web = setupTestWeb(t)
	createData(web)
	recorder = web.postHttpResponse("/setup/db/clear/qualification", "")
	assert.Equal(t, 303, recorder.Code)
	teams, _ = web.arena.Database.GetAllTeams()
	assert.NotEmpty(t, teams)
	matches, _ = web.arena.Database.GetMatchesByType(model.Practice, true)
	assert.NotEmpty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(1)
	assert.NotNil(t, matchResult)
	matches, _ = web.arena.Database.GetMatchesByType(model.Qualification, true)
	assert.Empty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(2)
	assert.Nil(t, matchResult)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.NotEmpty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(3)
	assert.NotNil(t, matchResult)
	rankings, _ = web.arena.Database.GetAllRankings()
	assert.Empty(t, rankings)
	tournament.CalculateRankings(web.arena.Database, false)
	assert.Empty(t, rankings)
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.NotEmpty(t, alliances)
	assert.NotEmpty(t, web.arena.AllianceSelectionAlliances)

	// Test clearing playoff data.
	web = setupTestWeb(t)
	createData(web)
	recorder = web.postHttpResponse("/setup/db/clear/playoff", "")
	assert.Equal(t, 303, recorder.Code)
	teams, _ = web.arena.Database.GetAllTeams()
	assert.NotEmpty(t, teams)
	matches, _ = web.arena.Database.GetMatchesByType(model.Practice, true)
	assert.NotEmpty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(1)
	assert.NotNil(t, matchResult)
	matches, _ = web.arena.Database.GetMatchesByType(model.Qualification, true)
	assert.NotEmpty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(2)
	assert.NotNil(t, matchResult)
	matches, _ = web.arena.Database.GetMatchesByType(model.Playoff, true)
	assert.Empty(t, matches)
	matchResult, _ = web.arena.Database.GetMatchResultForMatch(3)
	assert.Nil(t, matchResult)
	rankings, _ = web.arena.Database.GetAllRankings()
	assert.NotEmpty(t, rankings)
	tournament.CalculateRankings(web.arena.Database, false)
	assert.NotEmpty(t, rankings)
	alliances, _ = web.arena.Database.GetAllAlliances()
	assert.Empty(t, alliances)
	assert.Empty(t, web.arena.AllianceSelectionAlliances)

	// Test with invalid match types.
	recorder = web.postHttpResponse("/setup/db/clear/all", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid tournament stage to clear")
	recorder = web.postHttpResponse("/setup/db/clear/test", "")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid tournament stage to clear")
}

func TestSetupSettingsBackupRestoreDb(t *testing.T) {
	web := setupTestWeb(t)

	// Modify a parameter so that we know when the database has been restored.
	web.arena.EventSettings.Name = "Chezy Champs"
	assert.Nil(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))

	// Back up the database.
	recorder := web.getHttpResponse("/setup/db/save")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/octet-stream", recorder.HeaderMap["Content-Type"][0])
	backupBody := recorder.Body

	// Wipe the database to reset the defaults.
	web = setupTestWeb(t)
	assert.NotEqual(t, "Chezy Champs", web.arena.EventSettings.Name)

	// Check restoring with a missing file.
	recorder = web.postHttpResponse("/setup/db/restore", "")
	assert.Contains(t, recorder.Body.String(), "No database backup file was specified")
	assert.NotEqual(t, "Chezy Champs", web.arena.EventSettings.Name)

	// Check restoring with a corrupt file.
	recorder = web.postFileHttpResponse("/setup/db/restore", "databaseFile", bytes.NewBufferString("invalid"))
	assert.Contains(t, recorder.Body.String(), "Could not read uploaded database backup file")
	assert.NotEqual(t, "Chezy Champs", web.arena.EventSettings.Name)

	// Check restoring with the backup retrieved before.
	recorder = web.postFileHttpResponse("/setup/db/restore", "databaseFile", backupBody)
	assert.Equal(t, "Chezy Champs", web.arena.EventSettings.Name)
}

func (web *Web) postFileHttpResponse(path string, paramName string, file *bytes.Buffer) *httptest.ResponseRecorder {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(paramName, "file.ext")
	io.Copy(part, file)
	writer.Close()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	web.newHandler().ServeHTTP(recorder, req)
	return recorder
}
