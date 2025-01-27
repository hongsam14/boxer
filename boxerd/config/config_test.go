package config_test

import (
	"boxerd/config"
	"testing"
)

func TestCommandChecker(t *testing.T) {
	config := config.VMControlConfig{
		StartCmd:   "echo $machine",
		StopCmd:    "echo $machine",
		RestoreCmd: "echo $machine $snapshot",
	}
	if !config.CheckReservedKeyword() {
		t.Errorf("CheckReservedKeyword failed")
	}
}

func TestCommandCheckerFail(t *testing.T) {
	config := config.VMControlConfig{
		StartCmd:   "echo machine",
		StopCmd:    "echo machine",
		RestoreCmd: "echo machine $snapshot",
	}
	if config.CheckReservedKeyword() {
		t.Errorf("CheckReservedKeyword failed")
	}
}
