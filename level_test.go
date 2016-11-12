package ligno

import "testing"

func TestBuiltinLevelsRegistered(t *testing.T) {
	for _, buildinLevel := range []Level{
		NOTSET, DEBUG, INFO, WARNING, ERROR, CRITICAL,
	} {
		if _, ok := level2Name[buildinLevel]; !ok {
			t.Errorf("Level %s not registered.\n", buildinLevel)
		}
	}
}

func TestGetBuiltinLevelName(t *testing.T) {
	for level, name := range map[Level]string{
		NOTSET:   "NOTSET",
		DEBUG:    "DEBUG",
		INFO:     "INFO",
		WARNING:  "WARNING",
		ERROR:    "ERROR",
		CRITICAL: "CRITICAL",
		Level(2): "",
	} {
		levelName := getLevelName(level)
		if levelName != name {
			t.Errorf("Wrong level name, expected %s got %s.\n", name, levelName)
		}
	}
}

func TestBuiltinStringer(t *testing.T) {
	for level, name := range map[Level]string{
		NOTSET:   "NOTSET",
		DEBUG:    "DEBUG",
		INFO:     "INFO",
		WARNING:  "WARNING",
		ERROR:    "ERROR",
		CRITICAL: "CRITICAL",
		Level(2): "Level(2)",
	} {
		levelString := level.String()
		if levelString != name {
			t.Errorf("Wrong level string value, expected %s got %s.\n", name, levelString)
		}
	}
}

func TestAddNewLevel(t *testing.T) {
	customLevel := Level(2)
	customName := "CUSTOM"
	AddLevel(customName, customLevel)
	if _, ok := level2Name[customLevel]; !ok {
		t.Error("Custom level not registered to level2Name map.")
	}
	if _, ok := name2Level[customName]; !ok {
		t.Error("Custom level not registered to name2Level map.")
	}
	if getLevelName(customLevel) != customName {
		t.Errorf("Custom level name not returned, got %s, expected %s.\n", getLevelName(customLevel), customName)
	}
}

func TestAddLevelAlreadyExist(t *testing.T) {
	level4 := Level(4)
	level4Name := "Level4"
	AddLevel(level4Name, level4)
	_, err := AddLevel(level4Name, level4)
	if err == nil {
		t.Fatal("Expected error when adding existing level, got nil.")
	}
}

func TestUnregisteredLevelString(t *testing.T) {
	levelRank := 3
	level := Level(levelRank)
	expect := "Level(3)"
	if level.String() != expect {
		t.Fatalf("Unexpected string format for level, expected %s, got %s.\n", expect, level.String())
	}
}
