// Package notes defines the structs for release note and functions to generate
// notes.
package notes

// Notes contains all the note entries for a given release.
type Notes struct {
	Org      string     `json:"org"`
	Repo     string     `json:"repo"`
	Version  string     `json:"version"`
	Sections []*Section `json:"sections"`
}

// Section contains one release note section, for example "Feature".
type Section struct {
	Name      string   `json:"name"`
	LabelName string   `json:"label_name"`
	Entries   []*Entry `json:"entries"`
}

// Entry contains the info for one entry in the release notes.
type Entry struct {
	IssueNumber int    `json:"issue_number"`
	Title       string `json:"title"`
	HTMLURL     string `json:"html_url"`

	User      *User      `json:"user"`
	MileStone *MileStone `json:"milestone"`

	SpecialThanks bool `json:"special_thanks"`
}

// User represents a github user.
type User struct {
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Login     string `json:"login"`
}

// MileStone represents a github milestone.
type MileStone struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// Label represents a github label.
type Label struct {
	Name string `json:"name"`
}
