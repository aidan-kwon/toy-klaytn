package utils

import (
	"gopkg.in/urfave/cli.v1"
	"sort"
)

const uncategorized = "MISC" // Uncategorized flags will belong to this group

// FlagGroup is a collection of flags belonging to a single topic.
type FlagGroup struct {
	Name  string
	Flags []cli.Flag
}

var FlagGroups = []FlagGroup{
	{
		Name: "KLAY",
		Flags: []cli.Flag{
			DataDirFlag,
		},
	},
}


// CategorizeFlags classifies each flag into pre-defined flagGroups.
func CategorizeFlags(flags []cli.Flag) []FlagGroup {
	flagGroupsMap := make(map[string][]cli.Flag)
	isFlagAdded := make(map[string]bool) // Check duplicated flags

	// Find its group for each flag
	for _, flag := range flags {
		if isFlagAdded[flag.GetName()] {
			logger.Debug("a flag is added in the help description more than one time", "flag", flag.GetName())
			continue
		}

		// Find a group of the flag. If a flag doesn't belong to any groups, categorize it as a MISC flag
		group := flagCategory(flag, FlagGroups)
		flagGroupsMap[group] = append(flagGroupsMap[group], flag)
		isFlagAdded[flag.GetName()] = true
	}

	// Convert flagGroupsMap to a slice of FlagGroup
	flagGroups := []FlagGroup{}
	for group, flags := range flagGroupsMap {
		flagGroups = append(flagGroups, FlagGroup{Name: group, Flags: flags})
	}

	// Sort flagGroups in ascending order of name
	sortFlagGroup(flagGroups, uncategorized)

	return flagGroups
}

// sortFlagGroup sorts a slice of FlagGroup in ascending order of name,
// but an uncategorized group is exceptionally placed at the end.
func sortFlagGroup(flagGroups []FlagGroup, uncategorized string) []FlagGroup {
	sort.Slice(flagGroups, func(i, j int) bool {
		if flagGroups[i].Name == uncategorized {
			return false
		}
		if flagGroups[j].Name == uncategorized {
			return true
		}
		return flagGroups[i].Name < flagGroups[j].Name
	})

	// Sort flags in each group i ascending order of flag name.
	for _, group := range flagGroups {
		sort.Slice(group.Flags, func(i, j int) bool {
			return group.Flags[i].GetName() < group.Flags[j].GetName()
		})
	}

	return flagGroups
}

// flagCategory returns belonged group of the given flag.
func flagCategory(flag cli.Flag, fg []FlagGroup) string {
	for _, category := range fg {
		for _, flg := range category.Flags {
			if flg.GetName() == flag.GetName() {
				return category.Name
			}
		}
	}
	return uncategorized
}

