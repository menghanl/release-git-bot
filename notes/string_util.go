package notes

import (
	"fmt"

	"github.com/google/go-github/github"
)

func labelsToString(ls []github.Label) string {
	var names []string
	for _, l := range ls {
		names = append(names, l.GetName())
	}
	return fmt.Sprintf("%v", names)
}
