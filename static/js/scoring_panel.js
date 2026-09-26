// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
let position;
let alliance;
let allianceKey;
let committed = false;

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;

// The counter field on game.Score corresponding to each counter id used by the "treasure" command.
const counterFields = {
  auto_floor: "AutoFloor",
  auto_first: "AutoFirst",
  auto_top: "AutoTop",
  teleop_floor: "TeleopFloor",
  teleop_first: "TeleopFirst",
  teleop_top: "TeleopTop",
  teleop_stacked: "TeleopStacked",
};

// The CrownPlacement enum value corresponding to each counter id that the crown can be placed on.
const crownValueByCounter = {
  auto_floor: 1,
  auto_first: 2,
  auto_top: 3,
  teleop_floor: 4,
  teleop_first: 5,
  teleop_top: 6,
  teleop_stacked: 7,
};

// The currently-known crown placement, used to decide whether tapping a Crown button sets or clears it.
let currentCrown = 0;

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);

  const allianceLetter = alliance === "red" ? "R" : "B";
  for (let i = 1; i <= 3; i++) {
    const team = data.Teams[`${allianceLetter}${i}`];
    const teamId = team ? team.Id : "";
    $(`#team-${i}`).text(teamId);
    $(`#endgameTeam-${i}`).text(teamId);
  }
};

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  switch (matchStates[data.MatchState]) {
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      committed = false;
      break;
    case "POST_MATCH":
      if (!committed) {
        scoringAvailable = true;
        commitAvailable = true;
      }
      break;
    default:
      scoringAvailable = false;
      commitAvailable = false;
      committed = false;
  }
  updateUIMode();
};

// Clear any local ephemeral state that is not maintained by the server
const resetLocalState = function () {
  committed = false;
  updateUIMode();
}

// Refresh which UI controls are enabled/disabled
const updateUIMode = function () {
  $(".scoring-button").prop('disabled', !scoringAvailable);
  $("#commit").prop('disabled', !commitAvailable);
}

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  const score = data[allianceKey].Score;

  $.each(counterFields, function (id, field) {
    $(`#value-${id}`).text(score[field]);
  });

  currentCrown = score.Crown;
  $.each(crownValueByCounter, function (id, value) {
    $(`#crown-${id}`).attr("data-selected", value === currentCrown);
  });

  for (let i = 1; i <= 3; i++) {
    $(`#leave-${i}`).attr("data-selected", score.LeaveStatuses[i - 1]);
    $(`#auto_balance-${i}`).attr("data-selected", score.AutoBalanceStatuses[i - 1]);

    const endgameGroup = $(`#endgame-${i}`);
    endgameGroup.attr("data-value", score.EndgameStatuses[i - 1]);
    endgameGroup.find(".endgame-button").each(function (index) {
      $(this).attr("data-selected", index === score.EndgameStatuses[i - 1]);
    });
  }

  $("#toss").attr("data-selected", score.Toss);
};

// Sends a websocket message to adjust the given counter by the given amount.
const adjustCounter = function (counter, adjustment) {
  websocket.send("treasure", {Counter: counter, Adjustment: adjustment});
};

// Sends a websocket message to set or clear the crown at the given counter's location.
const toggleCrown = function (counter) {
  const value = currentCrown === crownValueByCounter[counter] ? "none" : counter;
  websocket.send("crown", {Value: value});
};

// Sends a websocket message to toggle whether the given robot left its safe house during auto.
const toggleLeave = function (teamPosition) {
  websocket.send("leave", {TeamPosition: teamPosition});
};

// Sends a websocket message to toggle whether the given robot balanced on its mountain top during auto.
const toggleAutoBalance = function (teamPosition) {
  websocket.send("auto_balance", {TeamPosition: teamPosition});
};

// Sends a websocket message to set the given robot's endgame status.
const setEndgame = function (teamPosition, value) {
  websocket.send("endgame", {TeamPosition: teamPosition, Value: value});
};

// Sends a websocket message to toggle whether the alliance's toss scored.
const toggleToss = function () {
  websocket.send("toss", {});
};

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function () {
  websocket.send("commitMatch");

  committed = true;
  scoringAvailable = false;
  commitAvailable = false;
  updateUIMode();
};

$(function () {
  position = window.location.href.split("/").slice(-1)[0];
  alliance = position.split("_")[0];
  allianceKey = alliance === "red" ? "Red" : "Blue";
  $(".container").attr("data-alliance", alliance);
  resetLocalState();

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + position + "/websocket", {
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    resetLocalState: function (event) {
      resetLocalState();
    },
  });
});
