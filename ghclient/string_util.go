package ghclient

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
)

func issueToString(ii *github.Issue) string {
	var ret string
	ret += color.CyanString("%v [%v] - %v", ii.GetNumber(), ii.GetState(), ii.GetTitle())
	ret += "\n - "
	ret += color.BlueString("%v", ii.GetHTMLURL())
	ret += "\n - "
	ret += color.BlueString("author: %v", *ii.GetUser().Login)
	return ret
}

func labelsToString(ls []github.Label) string {
	var names []string
	for _, l := range ls {
		names = append(names, l.GetName())
	}
	return fmt.Sprintf("%v", names)
}
