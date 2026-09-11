// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package constraint

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// matchCommentCommand reports whether the comment body invokes one of the given
// commands. A command matches when it appears at the start of a line and is
// followed by whitespace or the end of that line, so "/deploy" matches
// "/deploy production" but neither "/deployment" nor "please run /deploy".
// Matching is case-sensitive, like every other constraint.
func matchCommentCommand(commands []string, body string) bool {
	for _, rawLine := range strings.Split(body, "\n") {
		// TrimSpace also removes the trailing \r of CRLF line endings
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		for _, command := range commands {
			command = strings.TrimSpace(command)
			if command == "" {
				continue
			}

			rest, found := strings.CutPrefix(line, command)
			if !found {
				continue
			}
			if rest == "" {
				return true
			}
			if r, _ := utf8.DecodeRuneInString(rest); unicode.IsSpace(r) {
				return true
			}
		}
	}

	return false
}
