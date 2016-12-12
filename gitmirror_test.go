package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"io"
	"testing"
)

func TestHMACCompare(t *testing.T) {
	tests := []struct {
		data string
		hash string
		eq   bool
	}{
		{"", "sha1=8a3b873f8dcaebf748c60464fb16878b2953d6df", true},
		{"xxx", "sha1=950408a7db2d17330d8a288417c9d38fd8c6bfef", true},
		{"xxx", "sha1=950408a7db2d17330d8a288417c9d38fd8c6bfee", false},
		{"xxx", "", false},
	}

	for _, test := range tests {
		h := hmac.New(sha1.New, []byte{'h', 'i'})
		io.WriteString(h, test.data)
		if checkHMAC(h, test.hash) != test.eq {
			t.Errorf("On %q, expected %v, got %x", test.data, test.eq, h.Sum(nil))
		}
	}
}

const testOrgPushHook = `{
  "zen": "Encourage flow.",
  "hook_id": 5564070,
  "hook": {
    "url": "https://api.github.com/repos/rotorbench/data/hooks/5564070",
    "test_url": "https://api.github.com/repos/rotorbench/data/hooks/5564070/test",
    "ping_url": "https://api.github.com/repos/rotorbench/data/hooks/5564070/pings",
    "id": 5564070,
    "name": "web",
    "active": true,
    "events": [
      "push"
    ],
    "config": {
      "url": "http://wwcp540.appspot.com/q/push/aglzfnd3Y3A1NDByEQsSBEZlZWQYgICAgPjChAoM",
      "content_type": "json",
      "insecure_ssl": "0",
      "secret": ""
    },
    "last_response": {
      "code": null,
      "status": "unused",
      "message": null
    },
    "updated_at": "2015-08-12T16:10:55Z",
    "created_at": "2015-08-12T16:10:55Z"
  },
  "repository": {
    "id": 40608334,
    "name": "data",
    "full_name": "rotorbench/data",
    "owner": {
      "login": "rotorbench",
      "id": 13767952,
      "avatar_url": "https://avatars.githubusercontent.com/u/13767952?v=3",
      "gravatar_id": "",
      "url": "https://api.github.com/users/rotorbench",
      "html_url": "https://github.com/rotorbench",
      "followers_url": "https://api.github.com/users/rotorbench/followers",
      "following_url": "https://api.github.com/users/rotorbench/following{/other_user}",
      "gists_url": "https://api.github.com/users/rotorbench/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/rotorbench/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/rotorbench/subscriptions",
      "organizations_url": "https://api.github.com/users/rotorbench/orgs",
      "repos_url": "https://api.github.com/users/rotorbench/repos",
      "events_url": "https://api.github.com/users/rotorbench/events{/privacy}",
      "received_events_url": "https://api.github.com/users/rotorbench/received_events",
      "type": "Organization",
      "site_admin": false
    },
    "private": false,
    "html_url": "https://github.com/rotorbench/data",
    "description": "",
    "fork": false,
    "url": "https://api.github.com/repos/rotorbench/data",
    "forks_url": "https://api.github.com/repos/rotorbench/data/forks",
    "keys_url": "https://api.github.com/repos/rotorbench/data/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/rotorbench/data/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/rotorbench/data/teams",
    "hooks_url": "https://api.github.com/repos/rotorbench/data/hooks",
    "issue_events_url": "https://api.github.com/repos/rotorbench/data/issues/events{/number}",
    "events_url": "https://api.github.com/repos/rotorbench/data/events",
    "assignees_url": "https://api.github.com/repos/rotorbench/data/assignees{/user}",
    "branches_url": "https://api.github.com/repos/rotorbench/data/branches{/branch}",
    "tags_url": "https://api.github.com/repos/rotorbench/data/tags",
    "blobs_url": "https://api.github.com/repos/rotorbench/data/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/rotorbench/data/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/rotorbench/data/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/rotorbench/data/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/rotorbench/data/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/rotorbench/data/languages",
    "stargazers_url": "https://api.github.com/repos/rotorbench/data/stargazers",
    "contributors_url": "https://api.github.com/repos/rotorbench/data/contributors",
    "subscribers_url": "https://api.github.com/repos/rotorbench/data/subscribers",
    "subscription_url": "https://api.github.com/repos/rotorbench/data/subscription",
    "commits_url": "https://api.github.com/repos/rotorbench/data/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/rotorbench/data/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/rotorbench/data/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/rotorbench/data/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/rotorbench/data/contents/{+path}",
    "compare_url": "https://api.github.com/repos/rotorbench/data/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/rotorbench/data/merges",
    "archive_url": "https://api.github.com/repos/rotorbench/data/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/rotorbench/data/downloads",
    "issues_url": "https://api.github.com/repos/rotorbench/data/issues{/number}",
    "pulls_url": "https://api.github.com/repos/rotorbench/data/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/rotorbench/data/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/rotorbench/data/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/rotorbench/data/labels{/name}",
    "releases_url": "https://api.github.com/repos/rotorbench/data/releases{/id}",
    "created_at": "2015-08-12T15:26:37Z",
    "updated_at": "2015-08-12T15:26:37Z",
    "pushed_at": "2015-08-12T15:40:27Z",
    "git_url": "git://github.com/rotorbench/data.git",
    "ssh_url": "git@github.com:rotorbench/data.git",
    "clone_url": "https://github.com/rotorbench/data.git",
    "svn_url": "https://github.com/rotorbench/data",
    "homepage": null,
    "size": 0,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": null,
    "has_issues": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": false,
    "forks_count": 0,
    "mirror_url": null,
    "open_issues_count": 0,
    "forks": 0,
    "open_issues": 0,
    "watchers": 0,
    "default_branch": "master"
  },
  "sender": {
    "login": "dustin",
    "id": 1779,
    "avatar_url": "https://avatars.githubusercontent.com/u/1779?v=3",
    "gravatar_id": "",
    "url": "https://api.github.com/users/dustin",
    "html_url": "https://github.com/dustin",
    "followers_url": "https://api.github.com/users/dustin/followers",
    "following_url": "https://api.github.com/users/dustin/following{/other_user}",
    "gists_url": "https://api.github.com/users/dustin/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/dustin/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/dustin/subscriptions",
    "organizations_url": "https://api.github.com/users/dustin/orgs",
    "repos_url": "https://api.github.com/users/dustin/repos",
    "events_url": "https://api.github.com/users/dustin/events{/privacy}",
    "received_events_url": "https://api.github.com/users/dustin/received_events",
    "type": "User",
    "site_admin": false
  }
}`

func TestExists(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"gitmirror.go", true},
		{"svnmirror.go", false},
	}

	for _, test := range tests {
		got := exists(test.path)
		if got != test.want {
			t.Errorf("exists(%q) = %v; want %v", test.path, got, test.want)
		}
	}
}
