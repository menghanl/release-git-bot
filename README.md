See page: https://menghanl.github.io/release-git-bot/

---

### Preparation

1. Check the release note and make sure it includes the release PRs, the whole release PRs and nothing but the release PRs.
   - https://github.com/menghanl/release-note-page
1. Sync your fork's master so it's __up-to-date__ with `upstream:master`.
1. Create a [github token](https://github.com/settings/tokens) with `repo`, `read:org` and `user:email` permissions.

### Install or update the tool:

```
go get -u github.com/menghanl/release-git-bot
```

### Nokidding

```
release-git-bot -version <1.14.0> -token <github_token> -nokidding
```

:tada: :tada: :tada: :tada: :tada:
