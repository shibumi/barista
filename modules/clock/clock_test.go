// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clock

import (
	"testing"
	"time"

	"github.com/shibumi/barista/bar"
	"github.com/shibumi/barista/base/watchers/localtz"
	"github.com/shibumi/barista/outputs"
	testBar "github.com/shibumi/barista/testing/bar"
	"github.com/shibumi/barista/timing"

	"github.com/stretchr/testify/require"
)

var fixedTime = time.Date(2017, time.March, 1, 0, 0, 0, 0, time.UTC)

func TestSimpleTicking(t *testing.T) {
	testBar.New(t)
	timing.AdvanceTo(fixedTime)

	testBar.Run(Local())
	testBar.NextOutput().AssertText(
		[]string{"00:00"}, "on start")

	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:01"}, "on next tick")

	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:02"}, "on next tick")
}

func TestAutoGranularities(t *testing.T) {
	testBar.New(t)
	timing.AdvanceTo(fixedTime)
	require := require.New(t)

	local := Local().OutputFormat("15:04:05")
	testBar.Run(local)
	testBar.NextOutput().AssertText(
		[]string{"00:00:00"}, "on start")

	now := timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:00:01"}, "on next tick")
	require.Equal(1, now.Second(), "increases by granularity")
	require.Equal(0, now.Nanosecond(), "triggers at exact granularity")

	timing.AdvanceBy(500 * time.Millisecond)
	testBar.AssertNoOutput("less than granularity")

	now = timing.NextTick()
	require.Equal(2, now.Second(), "increases by granularity")
	require.Equal(0, now.Nanosecond(), "triggers at exact granularity")
	testBar.NextOutput().AssertText(
		[]string{"00:00:02"}, "on next tick")

	local.OutputFormat("15:04")
	testBar.NextOutput().AssertText(
		[]string{"00:00"}, "on output format change")

	now = timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:01"}, "on next tick")
	require.Equal(1, now.Minute(), "triggers on exact granularity")
	require.Equal(0, now.Second(), "triggers on exact granularity")

	local.OutputFormat("15:04:05.0")
	testBar.NextOutput().AssertText(
		[]string{"00:01:00.0"}, "on output format change")
	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:01:00.1"}, "on next tick")

	local.OutputFormat("15:04:05.000")
	testBar.NextOutput().AssertText(
		[]string{"00:01:00.100"}, "on output format change")
	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:01:00.101"}, "on next tick")

	local.OutputFormat("15:04:05.00")
	testBar.NextOutput().AssertText(
		[]string{"00:01:00.10"}, "on output format change")
	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"00:01:00.11"}, "on next tick")

	testBar.AssertNoOutput("when time is frozen")
}

func TestManualGranularities(t *testing.T) {
	testBar.New(t)
	timing.AdvanceTo(fixedTime)

	local := Local().Output(time.Hour, func(now time.Time) bar.Output {
		return outputs.Text(now.Format("15:04:05"))
	})
	testBar.Run(local)
	testBar.NextOutput().AssertText(
		[]string{"00:00:00"}, "on start")

	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"01:00:00"}, "on tick")

	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"02:00:00"}, "on tick")

	local.Output(time.Minute, func(now time.Time) bar.Output {
		return outputs.Text(now.Format("15:04:05.00"))
	})
	testBar.NextOutput().AssertText(
		[]string{"02:00:00.00"}, "on format function + granularity change")

	timing.NextTick()
	testBar.NextOutput().AssertText(
		[]string{"02:01:00.00"}, "on tick")
}

func TestZones(t *testing.T) {
	testBar.New(t)
	timing.AdvanceTo(
		time.Date(2017, time.March, 1, 13, 15, 0, 0, time.UTC))

	la, _ := time.LoadLocation("America/Los_Angeles")
	pst := Zone(la).OutputFormat("15:04:05")

	berlin, err := ZoneByName("Europe/Berlin")
	require.NoError(t, err)
	berlin.OutputFormat("15:04:05")

	tokyo, err := ZoneByName("Asia/Tokyo")
	require.NoError(t, err)
	tokyo.OutputFormat("15:04:05")

	local := Local().OutputFormat("15:04:05")

	testBar.Run(pst, berlin, tokyo, local)

	_, err = ZoneByName("Global/Unknown")
	require.Error(t, err, "when loading unknown zone")

	testBar.LatestOutput().AssertText(
		[]string{"05:15:00", "14:15:00", "22:15:00", "13:15:00"},
		"on start")

	timing.NextTick()
	testBar.LatestOutput().AssertText(
		[]string{"05:15:01", "14:15:01", "22:15:01", "13:15:01"},
		"on tick")

	tok, _ := time.LoadLocation("Asia/Tokyo")
	localtz.SetForTest(tok)
	testBar.NextOutput(3).At(3).AssertText("22:15:01",
		"on local time zone change")

	berlin.Timezone(la)
	testBar.LatestOutput(1).At(1).AssertText(
		"05:15:01", "on timezone change")
}
