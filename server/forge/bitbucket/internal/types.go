// Copyright 2018 Drone.IO Inc.
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

package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// cspell:words pagelen

type Account struct {
	UUID  string `json:"uuid"`
	Login string `json:"username"`
	Name  string `json:"display_name"`
	Type  string `json:"type"`
	Links Links  `json:"links"`
}

type Workspace struct {
	UUID  string `json:"uuid"`
	Slug  string `json:"slug"`
	Type  string `json:"type"`
	Links Links  `json:"links"`
}

type WorkspaceAccess struct {
	Type          string     `json:"type"`
	Administrator bool       `json:"administrator"`
	Workspace     *Workspace `json:"workspace"`
}

type WorkspacesResp struct {
	Page   int                `json:"page"`
	Pages  int                `json:"pagelen"`
	Size   int                `json:"size"`
	Next   string             `json:"next"`
	Values []*WorkspaceAccess `json:"values"`
}

type PipelineStatus struct {
	State   string `json:"state"`
	Key     string `json:"key"`
	Name    string `json:"name,omitempty"`
	URL     string `json:"url"`
	Desc    string `json:"description,omitempty"`
	Refname string `json:"refname,omitempty"`
}

type Email struct {
	Email       string `json:"email"`
	IsConfirmed bool   `json:"is_confirmed"`
	IsPrimary   bool   `json:"is_primary"`
}

type EmailResp struct {
	Page   int      `json:"page"`
	Pages  int      `json:"pagelen"`
	Size   int      `json:"size"`
	Next   string   `json:"next"`
	Values []*Email `json:"values"`
}

type Hook struct {
	UUID   string   `json:"uuid,omitempty"`
	Desc   string   `json:"description"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

type HookResp struct {
	Page   int     `json:"page"`
	Pages  int     `json:"pagelen"`
	Size   int     `json:"size"`
	Next   string  `json:"next"`
	Values []*Hook `json:"values"`
}

type Links struct {
	Self   Link   `json:"self"`
	Avatar Link   `json:"avatar"`
	HTML   Link   `json:"html"`
	Clone  []Link `json:"clone"`
}

type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type LinkClone struct {
	Link
}

type Repo struct {
	UUID       string  `json:"uuid"`
	Owner      Account `json:"owner"`
	Name       string  `json:"name"`
	FullName   string  `json:"full_name"`
	Language   string  `json:"language"`
	IsPrivate  bool    `json:"is_private"`
	Scm        string  `json:"scm"`
	Desc       string  `json:"desc"`
	Links      Links   `json:"links"`
	MainBranch struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"mainbranch"` // cspell:ignore mainbranch
}

type RepoResp struct {
	Page   int     `json:"page"`
	Pages  int     `json:"pagelen"`
	Size   int     `json:"size"`
	Next   string  `json:"next"`
	Values []*Repo `json:"values"`
}

type Change struct {
	New struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Target struct {
			Type    string    `json:"type"`
			Hash    string    `json:"hash"`
			Message string    `json:"message"`
			Date    time.Time `json:"date"`
			Links   Links     `json:"links"`
			Author  struct {
				Raw  string  `json:"raw"`
				User Account `json:"user"`
			} `json:"author"`
		} `json:"target"`
	} `json:"new"`
}

type PushHook struct {
	Actor Account `json:"actor"`
	Repo  Repo    `json:"repository"`
	Push  struct {
		Changes []Change `json:"changes"`
	} `json:"push"`
}

type PullRequestHook struct {
	Actor       Account `json:"actor"`
	Repo        Repo    `json:"repository"`
	PullRequest struct {
		ID      int       `json:"id"`
		Type    string    `json:"type"`
		Reason  string    `json:"reason"`
		Desc    string    `json:"description"`
		Title   string    `json:"title"`
		State   string    `json:"state"`
		Links   Links     `json:"links"`
		Created time.Time `json:"created_on"`
		Updated time.Time `json:"updated_on"`

		MergeCommit struct {
			Hash string `json:"hash"`
		} `json:"merge_commit"`

		Source struct {
			Repo   Repo `json:"repository"`
			Commit struct {
				Hash  string `json:"hash"`
				Links Links  `json:"links"`
			} `json:"commit"`
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`

		Dest struct {
			Repo   Repo `json:"repository"`
			Commit struct {
				Hash  string `json:"hash"`
				Links Links  `json:"links"`
			} `json:"commit"`
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"destination"`
	} `json:"pullrequest"`
}

// PullRequestCommentHook is the payload of a "pullrequest:comment_created"
// webhook. It carries the same actor/repository/pullrequest objects as a plain
// pull request hook, plus the comment itself.
type PullRequestCommentHook struct {
	PullRequestHook
	Comment Comment `json:"comment"`
}

type Comment struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created_on"`
	Updated time.Time `json:"updated_on"`
	Deleted bool      `json:"deleted"`
	Content struct {
		Raw    string `json:"raw"`
		Markup string `json:"markup"`
		HTML   string `json:"html"`
	} `json:"content"`
	User  Account `json:"user"`
	Links Links   `json:"links"`
}

type WorkspaceMembershipResp struct {
	Page   int    `json:"page"`
	Pages  int    `json:"pagelen"`
	Size   int    `json:"size"`
	Next   string `json:"next"`
	Values []struct {
		Permission string `json:"permission"`
		User       struct {
			Nickname string `json:"nickname"`
		}
	} `json:"values"`
}

type ListOpts struct {
	Page    int
	PageLen int
}

func (o *ListOpts) Encode() string {
	params := url.Values{}
	if o.Page != 0 {
		params.Set("page", strconv.Itoa(o.Page))
	}
	if o.PageLen != 0 {
		params.Set("pagelen", strconv.Itoa(o.PageLen))
	}
	return params.Encode()
}

type Error struct {
	Status int    `json:"-"`
	Method string `json:"-"`
	URL    string `json:"-"`
	// Raw holds the unparsed response body, so failures that do not follow
	// Bitbucket's error schema (or are not JSON at all) can still be diagnosed.
	Raw  string `json:"-"`
	Body struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
		// Fields holds per-field validation errors, e.g. an invalid webhook URL.
		Fields map[string][]string `json:"fields"`
	} `json:"error"`
}

func (e Error) Error() string {
	msg := e.Body.Message
	if e.Body.Detail != "" {
		msg = fmt.Sprintf("%s: %s", msg, e.Body.Detail)
	}
	for field, errs := range e.Body.Fields {
		msg = fmt.Sprintf("%s [%s: %s]", msg, field, strings.Join(errs, ", "))
	}
	if strings.TrimSpace(msg) == "" {
		msg = strings.TrimSpace(e.Raw)
	}
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	if e.Method != "" && e.URL != "" {
		return fmt.Sprintf("%s %s: %d: %s", e.Method, e.URL, e.Status, msg)
	}
	return msg
}

type RepoPermResp struct {
	Page   int         `json:"page"`
	Pages  int         `json:"pagelen"`
	Values []*RepoPerm `json:"values"`
}

type RepoPerm struct {
	Permission string `json:"permission"`
	Repo       Repo   `json:"repository"`
}

type BranchResp struct {
	Values []*Branch `json:"values"`
}

type Branch struct {
	Name string `json:"name"`
}

type PullRequestResp struct {
	Page    uint           `json:"page"`
	PageLen uint           `json:"pagelen"`
	Size    uint           `json:"size"`
	Values  []*PullRequest `json:"values"`
}

type PullRequest struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

type CommitsResp struct {
	Values []*Commit `json:"values"`
}

type Commit struct {
	Hash  string `json:"hash"`
	Links struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

type DirResp struct {
	Page    uint    `json:"page"`
	PageLen uint    `json:"pagelen"`
	Next    *string `json:"next"`
	Values  []*Dir  `json:"values"`
}

type Dir struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size uint   `json:"size"`
}

type DiffStatResp struct {
	Next   *string `json:"next"`
	Values []*Diff `json:"values"`
}

type Diff struct {
	Old *DiffFile `json:"old"`
	New *DiffFile `json:"new"`
}

type DiffFile struct {
	Path string `json:"path"`
}
