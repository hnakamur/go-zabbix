package zabbix

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type APIVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelType PreRelType
	PreRelVer  int
}

var ErrInvalidZabbixVer = errors.New("invalid Zabbix version")

var versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:(alpha|beta|rc)(\d+))?$`)

func MustParseAPIVersion(ver string) APIVersion {
	v, err := ParseAPIVersion(ver)
	if err != nil {
		panic(err)
	}
	return v
}

func ParseAPIVersion(ver string) (APIVersion, error) {
	var v APIVersion
	m := versionRegex.FindStringSubmatch(ver)
	if len(m) != 6 {
		return v, ErrInvalidZabbixVer
	}

	major, err := strconv.Atoi(m[1])
	if err != nil {
		return v, ErrInvalidZabbixVer
	}
	v.Major = major

	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return v, ErrInvalidZabbixVer
	}
	v.Minor = minor

	patch, err := strconv.Atoi(m[3])
	if err != nil {
		return v, ErrInvalidZabbixVer
	}
	v.Patch = patch

	if m[4] != "" {
		switch m[4] {
		case "alpha":
			v.PreRelType = Alpha
		case "beta":
			v.PreRelType = Beta
		case "rc":
			v.PreRelType = RC
		}

		preReleaseVer, err := strconv.Atoi(m[5])
		if err != nil {
			return v, ErrInvalidZabbixVer
		}
		v.PreRelVer = preReleaseVer
	}
	return v, nil
}

func (v APIVersion) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelType != Release {
		fmt.Fprintf(&b, "%s%d", v.PreRelType, v.PreRelVer)
	}
	return b.String()
}

func (v APIVersion) Compare(w APIVersion) int {
	if v.Major < w.Major {
		return -1
	}
	if v.Major > w.Major {
		return 1
	}

	if v.Minor < w.Minor {
		return -1
	}
	if v.Minor > w.Minor {
		return 1
	}

	if v.Patch < w.Patch {
		return -1
	}
	if v.Patch > w.Patch {
		return 1
	}

	if v.PreRelType < w.PreRelType {
		return -1
	}
	if v.PreRelType > w.PreRelType {
		return 1
	}

	if v.PreRelVer < w.PreRelVer {
		return -1
	}
	if v.PreRelVer > w.PreRelVer {
		return 1
	}

	return 0
}

type PreRelType int

const (
	Alpha PreRelType = iota - 3
	Beta
	RC // Release Candidate
	Release
)

func (t PreRelType) String() string {
	switch t {
	case Alpha:
		return "alpha"
	case Beta:
		return "beta"
	case RC:
		return "rc"
	default:
		return ""
	}
}
