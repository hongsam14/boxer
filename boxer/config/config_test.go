package config_test

import (
	"testing"

	"github.com/hongsam14/boxer/config"
)

func TestCommandChecker(t *testing.T) {
	config := config.VMControlConfig{
		StartCmd:           "echo $machine",
		StopCmd:            "echo $machine",
		RestoreSnapshotCmd: "echo $machine $snapshot",
	}
	if !config.CheckReservedKeyword() {
		t.Errorf("CheckReservedKeyword failed")
	}
}

func TestCommandCheckerFail(t *testing.T) {
	config := config.VMControlConfig{
		StartCmd:           "echo machine",
		StopCmd:            "echo machine",
		RestoreSnapshotCmd: "echo machine $snapshot",
	}
	if config.CheckReservedKeyword() {
		t.Errorf("CheckReservedKeyword failed")
	}
}
