// // Copyright 2018 Drone.IO Inc.
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

package bitbucket

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// Every event registered with Bitbucket at activation time must be handled by
// parseHook, otherwise the delivery arrives and is silently ignored.
func Test_webhookEventsAreParsed(t *testing.T) {
	// a payload carrying just enough of every shape to reach each converter
	const minimalPayload = `{
		"repository": {"full_name": "owner/name", "scm": "git"},
		"push": {"changes": []},
		"pullrequest": {"id": 1, "state": "OPEN"},
		"comment": {"id": 1, "content": {"raw": "/x"}}
	}`

	for _, event := range webhookEvents {
		t.Run(event, func(t *testing.T) {
			buf := bytes.NewBufferString(minimalPayload)
			req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
			req.Header = http.Header{}
			req.Header.Set(hookEvent, event)

			_, _, _, err := parseHook(req)
			// the payload may still be ignored for a stated reason, but it must
			// never be rejected merely for being an unknown event type
			var ignoreErr *types.ErrIgnoreEvent
			if errors.As(err, &ignoreErr) {
				assert.NotEmpty(t, ignoreErr.Reason,
					"event %q is registered at activation but not handled by parseHook", event)
			}
		})
	}
}

func Test_parseHook(t *testing.T) {
	t.Run("unsupported hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, "issue:created")

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("malformed pull-request hook", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		_, _, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("pull-request", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPull)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCreated)

		pr, r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "user_name/repo_name", r.FullName)
		assert.Equal(t, model.EventPull, b.Event)
		assert.Equal(t, "d3022fc0ca3d", b.Commit)
	})

	t.Run("pull-request merged", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestMerged)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullMerged)

		pr, r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "006704dbeab2", b.Commit)
	})

	t.Run("pull-request closed", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestDeclined)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullDeclined)

		pr, r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "anbraten/test-2", r.FullName)
		assert.Equal(t, model.EventPullClosed, b.Event)
		assert.Equal(t, "f90e18fc9d45", b.Commit)
	})

	t.Run("malformed pull-request comment hook", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCommentCreated)

		_, _, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("pull-request comment", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestCommentCreated)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCommentCreated)

		pr, r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
		assert.Equal(t, model.EventPullComment, b.Event)
		assert.Equal(t, "d3022fc0ca3d", b.Commit)
		assert.Equal(t, "/codereview please", b.PullRequestComment)
		// the pipeline links to the comment, not the pull request
		assert.Equal(t, "https://api.bitbucket.org/pullrequest_id#comment-42", b.ForgeURL)
		// the message stays the pull request title, not the comment body
		assert.Equal(t, "Title of pull request", b.Message)
		// the actor of a comment hook is the commenter, which drives approval
		assert.Equal(t, "emmap1", b.Author)
		assert.Equal(t, "emmap1", b.Sender)
	})

	t.Run("pull-request comment on a closed pull request", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestCommentOnClosed)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCommentCreated)

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("pull-request comment with an empty body", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequestCommentEmpty)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullCommentCreated)

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("malformed push", func(t *testing.T) {
		buf := bytes.NewBufferString("[]")
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		_, _, _, err := parseHook(req)
		assert.Error(t, err)
	})

	t.Run("missing commit sha", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPushEmptyHash)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		_, r, b, err := parseHook(req)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("push hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPush)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPush)

		pr, r, b, err := parseHook(req)
		assert.NoError(t, err)
		assert.Nil(t, pr)
		assert.Equal(t, "martinherren1984/publictestrepo", r.FullName)
		assert.Equal(t, "https://bitbucket.org/martinherren1984/publictestrepo", r.Clone)
		assert.Equal(t, "c14c1bb05dfb1fdcdf06b31485fff61b0ea44277", b.Commit)
		assert.Equal(t, "a\n", b.Message)
	})
}
