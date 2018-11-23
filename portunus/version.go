// Copyright © 2018 Carl P. Corliss <carl@corliss.name>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package portunus

import (
	"fmt"
	"strconv"
	"strings"
)

type version struct {
	Major    int
	Minor    int
	Patch    int
	Revision string
}

func (self *version) String() string {
	if Revision == "" {
		return fmt.Sprintf("%d.%d.%d (unstable)",
			self.Major, self.Minor, self.Patch)
	} else {
		return fmt.Sprintf("%d.%d.%d (%s)",
			self.Major, self.Minor, self.Patch, self.Revision)
	}
}

func NewVersionFromString(value string, revision string) *version {
	var parts = []int{0, 0, 0}

	for idx, ver := range strings.Split(value, `.`) {
		if val, err := strconv.Atoi(ver); err == nil {
			parts[idx] = val
		} else {
			panic(err)
		}
	}

	return &version{parts[0], parts[1], parts[2], revision}
}

var (
	Revision   = "unstable"
	VersionStr = "0.1.0"
	Version    = NewVersionFromString(VersionStr, Revision)
)
