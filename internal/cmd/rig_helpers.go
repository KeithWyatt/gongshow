package cmd

import (
	"fmt"

	"github.com/KeithWyatt/gongshow/internal/config"
	"github.com/KeithWyatt/gongshow/internal/constants"
	"github.com/KeithWyatt/gongshow/internal/git"
	"github.com/KeithWyatt/gongshow/internal/rig"
	"github.com/KeithWyatt/gongshow/internal/workspace"
)

// getRig finds the town root and retrieves the specified rig.
// This is the common boilerplate extracted from get*Manager functions.
// Returns the town root path and rig instance.
func getRig(rigName string) (string, *rig.Rig, error) {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return "", nil, fmt.Errorf("not in a GongShow workspace: %w", err)
	}

	rigsConfigPath := constants.MayorRigsPath(townRoot)
	rigsConfig, err := config.LoadRigsConfig(rigsConfigPath)
	if err != nil {
		rigsConfig = &config.RigsConfig{Rigs: make(map[string]config.RigEntry)}
	}

	g := git.NewGit(townRoot)
	rigMgr := rig.NewManager(townRoot, rigsConfig, g)
	r, err := rigMgr.GetRig(rigName)
	if err != nil {
		return "", nil, fmt.Errorf("rig '%s' not found", rigName)
	}

	return townRoot, r, nil
}
