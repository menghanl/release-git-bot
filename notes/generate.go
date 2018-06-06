package notes

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
)

// Filters contains filters to be applied on input PRs.
type Filters struct {
	// If Ignore returns true, the pr will be excluded from the notes.
	Ignore func(pr *github.Issue) bool
	// if SpecialThanks returns true, a special thanks note will be included for
	// the author.
	SpecialThanks func(pr *github.Issue) bool
}

// GenerateNotes generate the release notes from the given prs and maps.
func GenerateNotes(org, repo, version string, prs []*github.Issue, filters Filters) *Notes {
	notes := Notes{
		Org:     org,
		Repo:    repo,
		Version: version,
	}

	sectionsMap := make(map[string]*Section)

	for _, pr := range prs {
		if filters.Ignore != nil && filters.Ignore(pr) {
			continue
		}

		label := pickMostWeightedLabel(pr.Labels)
		_, ok := labelToSectionName[label]
		if !ok {
			continue // If ok==false, ignore this PR in the release note.
		}
		fmt.Printf(" [%v] - ", color.BlueString("%v", pr.GetNumber()))
		fmt.Print(color.GreenString("%-18q", label))
		fmt.Printf(" from: %v\n", labelsToString(pr.Labels))

		section, ok := sectionsMap[label]
		if !ok {
			section = &Section{Name: labelToSectionName[label], LabelName: label}
			sectionsMap[label] = section

			notes.Sections = append(notes.Sections, section)
		}

		user := pr.GetUser()
		milestone := pr.GetMilestone()

		entry := &Entry{
			// head: fmt.Sprintf("%v (#%d)", pr.GetTitle(), pr.GetNumber()),
			IssueNumber: pr.GetNumber(),
			Title:       pr.GetTitle(),
			HTMLURL:     pr.GetHTMLURL(),

			User: &User{
				AvatarURL: user.GetAvatarURL(),
				HTMLURL:   user.GetHTMLURL(),
				Login:     user.GetLogin(),
			},

			MileStone: &MileStone{
				ID:    milestone.GetID(),
				Title: milestone.GetTitle(),
			},
			SpecialThanks: filters.SpecialThanks != nil && filters.SpecialThanks(pr),
		}
		section.Entries = append(section.Entries, entry)
	}
	notes.Sections = sortSections(notes.Sections)
	return &notes
}
