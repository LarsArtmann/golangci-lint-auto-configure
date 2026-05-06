package version

import (
	"fmt"
	"runtime/debug"
	"sync"
)

const (
	shortCommitLength = 7
	tagKeyRevision    = "vcs.revision"
	tagKeyTime        = "vcs.time"
	tagKeyModified    = "vcs.modified"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	treeState = "clean"
)

var (
	infoOnce sync.Once
	cached   Info
)

type Info struct {
	Version   string
	Commit    string
	Date      string
	TreeState string
}

func Get() Info {
	infoOnce.Do(func() {
		cached = buildInfo()
	})

	return cached
}

func buildInfo() Info {
	info := Info{
		Version:   version,
		Commit:    commit,
		Date:      date,
		TreeState: treeState,
	}

	if info.Version != "dev" {
		return info
	}

	return info.withBuildInfoFallback()
}

func (i Info) withBuildInfoFallback() Info {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return i
	}

	result := i

	for _, setting := range buildInfo.Settings {
		result = result.applySetting(setting)
	}

	return result
}

func (i Info) applySetting(setting debug.BuildSetting) Info {
	result := i

	switch setting.Key {
	case tagKeyRevision:
		result = result.applyRevision(setting.Value)
	case tagKeyTime:
		if result.Date == "unknown" {
			result.Date = setting.Value
		}
	case tagKeyModified:
		if result.TreeState == "clean" && setting.Value == "true" {
			result.TreeState = "dirty"
			result.Version += "-dirty"
		}
	}

	return result
}

func (i Info) applyRevision(revision string) Info {
	result := i

	short := revision
	if len(revision) > shortCommitLength {
		short = revision[:shortCommitLength]
	}

	if result.Commit == "none" {
		result.Commit = short
	}

	if result.Version == "dev" {
		result.Version = short
	}

	return result
}

func (i Info) Short() string {
	return i.Version
}

func (i Info) String() string {
	return fmt.Sprintf(
		"%s (commit: %s, built: %s)",
		i.Version,
		i.Commit,
		i.Date,
	)
}

func (i Info) Full() string {
	return fmt.Sprintf(
		"Version:    %s\nCommit:     %s\nBuilt:      %s\nTree:       %s",
		i.Version,
		i.Commit,
		i.Date,
		i.TreeState,
	)
}
