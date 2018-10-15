package version

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"net/http"

	"github.com/pkg/errors"
)

var AppVersion = "Missing version"
var Githash = "Missing githash"
var Buildstamp = "Missing timestamp"
var AppName = "NanoGrid"

func VersionJson() string {
	body, err := json.MarshalIndent(VersionStringMap(), "", "  ")

	if err != nil {
		panic(errors.Wrap(err, "Failed to marshall version to json"))
	}
	return string(body)
}

func VersionJsonInline() string {
	body, err := json.Marshal(VersionStringMap())

	if err != nil {
		panic(errors.Wrap(err, "Failed to marshall version to json"))
	}
	return string(body)
}

func VersionStringMap() map[string]string {
	return map[string]string{
		"Name":    AppName,
		"Version": AppVersion,
		"Build":   Buildstamp,
		"Githash": Githash}
}

func VersionPartsToString(major, minor, patch uint32, extension string) string {
	base := fmt.Sprintf("%v.%v.%v", major, minor, patch)

	if extension != "" {
		base = base + "-" + extension
	}
	return base
}

type Version struct {
	Major     uint32
	Minor     uint32
	Patch     uint32
	Extension string
}

func (c *Version) Empty() bool {
	return c.Major == 0 && c.Minor == 0 && c.Patch == 0 && c.Extension == ""
}

func (c Version) String() string {
	return VersionPartsToString(c.Major, c.Minor, c.Patch, c.Extension)
}

func (c *Version) Valid() error {
	if c.Major < 0 {
		return errors.New("Version - major is required")
	}
	if c.Minor < 0 {
		return errors.New("Version - minor is required")
	}
	if c.Patch < 0 {
		return errors.New("Version - patch is required")
	}

	if c.Empty() {
		return errors.New("Version - all fields are zero or blank")
	}

	return nil
}

func ParseVersion(version string) (*Version, error) {
	errMsg := errors.Errorf("Invalid version '%v' - expected MAJOR.MINOR.PATCH(-EXTENSION)?", version)

	extensionParts := strings.Split(version, "-")

	var v Version

	switch len(extensionParts) {
	case 0:
		return nil, errMsg
	case 1:
	case 2:
		v.Extension = extensionParts[1]
	default:
		return nil, errors.Wrap(errMsg, "Dash character can be used only between PATCH and EXTENSION")
	}

	parts := strings.Split(extensionParts[0], ".")
	if len(parts) != 3 {
		return nil, errMsg
	}

	if major, err := strconv.Atoi(parts[0]); err != nil {
		return nil, fmt.Errorf("Unparsable MAJOR part '%v', cause %v", parts[0], err)
	} else {
		v.Major = uint32(major)
	}

	if minor, err := strconv.Atoi(parts[1]); err != nil {
		return nil, fmt.Errorf("Unparsable MINOR part '%v', cause %v", parts[1], err)
	} else {
		v.Minor = uint32(minor)
	}

	if patch, err := strconv.Atoi(parts[2]); err != nil {
		return nil, fmt.Errorf("Unparsable MAJOR part '%v', cause %v", parts[2], err)
	} else {
		v.Patch = uint32(patch)
	}

	return &v, nil
}

type Sorter func(p1, p2 *Version) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by Sorter) Sort(vers []*Version) {
	ps := &versionSorter{
		vers: vers,
		by:   by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}

	sort.Sort(ps)

}

func Ascending(p1, p2 *Version) bool {
	if p1.Major < p2.Major {
		return true
	} else if p1.Major == p2.Major {
		if p1.Minor < p2.Minor {
			return true
		} else if p1.Minor == p2.Minor {
			if p1.Patch < p2.Patch {
				return true
			} else if p1.Patch < p2.Patch {
				return p1.Extension < p2.Extension
			} else {
				return false
			}

		} else {
			return false
		}
	} else {
		return false
	}
}

func Descending(p1, p2 *Version) bool {
	return !Ascending(p1, p2)
}

type versionSorter struct {
	vers []*Version
	by   func(p1, p2 *Version) bool // Closure used in the Less method.
}

func (s *versionSorter) Len() int {
	return len(s.vers)
}

func (s *versionSorter) Swap(i, j int) {
	s.vers[i], s.vers[j] = s.vers[j], s.vers[i]
}

func (s *versionSorter) Less(i, j int) bool {
	return s.by(s.vers[i], s.vers[j])
}

func HandlerFunc(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJson(w, VersionStringMap(), http.StatusOK)
}
