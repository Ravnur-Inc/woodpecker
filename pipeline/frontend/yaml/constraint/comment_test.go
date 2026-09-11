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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchCommentCommand(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		body     string
		want     bool
	}{
		{name: "exact match", commands: []string{"/codereview"}, body: "/codereview", want: true},
		{name: "with arguments", commands: []string{"/codereview"}, body: "/codereview --deep", want: true},
		{name: "surrounding whitespace", commands: []string{"/codereview"}, body: "  /codereview  ", want: true},
		{name: "on a later line", commands: []string{"/codereview"}, body: "LGTM, but:\n/codereview", want: true},
		{name: "crlf line endings", commands: []string{"/codereview"}, body: "a\r\n/codereview\r\n", want: true},
		{name: "second command matches", commands: []string{"/deploy", "/codereview"}, body: "/codereview", want: true},
		{name: "tab separated argument", commands: []string{"/codereview"}, body: "/codereview\tnow", want: true},

		{name: "longer command is not a match", commands: []string{"/codereview"}, body: "/codereview2", want: false},
		{name: "command as a prefix of a word", commands: []string{"/codereview"}, body: "/codereviewer", want: false},
		{name: "not at the start of a line", commands: []string{"/codereview"}, body: "please run /codereview", want: false},
		{name: "case sensitive", commands: []string{"/codereview"}, body: "/CODEREVIEW", want: false},
		{name: "unrelated body", commands: []string{"/codereview"}, body: "looks good to me", want: false},
		{name: "empty body", commands: []string{"/codereview"}, body: "", want: false},
		{name: "no commands", commands: []string{}, body: "/codereview", want: false},
		{name: "blank command is ignored", commands: []string{"  "}, body: "anything", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matchCommentCommand(tt.commands, tt.body))
		})
	}
}
