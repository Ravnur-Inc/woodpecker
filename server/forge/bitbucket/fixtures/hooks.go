// Copyright 2018 Drone.IO Inc.
// Copyright 2022 Woodpecker Authors
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

package fixtures

import _ "embed"

//go:embed HookPush.json
var HookPush string

const HookPushEmptyHash = `
{
  "push": {
    "changes": [
      {
        "new": {
          "type": "branch",
          "target": { "hash": "" }
        }
      }
    ]
  }
}
`

//go:embed HookPull.json
var HookPull string

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged string

//go:embed HookPullRequestDeclined.json
var HookPullRequestDeclined string

//go:embed HookPullRequestCommentCreated.json
var HookPullRequestCommentCreated string

// HookPullRequestCommentOnClosed is a comment created on a pull request that
// has already been merged, which must be ignored.
const HookPullRequestCommentOnClosed = `
{
  "actor": { "username": "emmap1" },
  "comment": { "id": 43, "content": { "raw": "/codereview" } },
  "pullrequest": { "id": 1, "state": "MERGED" },
  "repository": { "full_name": "user_name/repo_name", "scm": "git" }
}
`

// HookPullRequestCommentEmpty is a comment carrying no usable body, which must
// be ignored.
const HookPullRequestCommentEmpty = `
{
  "actor": { "username": "emmap1" },
  "comment": { "id": 44, "content": { "raw": "   \n  " } },
  "pullrequest": { "id": 1, "state": "OPEN" },
  "repository": { "full_name": "user_name/repo_name", "scm": "git" }
}
`
