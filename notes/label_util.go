package notes

import (
	"sort"
	"strings"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

const (
	labelPrefix  = "Type: "
	defaultLabel = "Bug"
)

var sortWeight = map[string]int{
	"Dependencies":     70,
	"API Change":       60,
	"Behavior Change":  50,
	"Feature":          40,
	"Performance":      30,
	"Bug":              20,
	"Documentation":    10,
	"Testing":          0,
	"Internal Cleanup": 0,
}

func sortLabelName(labels []string) []string {
	sort.Slice(labels, func(i, j int) bool {
		return sortWeight[labels[i]] >= sortWeight[labels[j]]
	})
	return labels
}

func pickMostWeightedLabel(labels []github.Label) string {
	if len(labels) <= 0 {
		log.Info("0 lable was assigned to issue")
		return defaultLabel
	}
	var names []string
	for _, l := range labels {
		names = append(names, strings.TrimPrefix(l.GetName(), labelPrefix))
	}
	sortLabelName(names)
	if _, ok := sortWeight[names[0]]; !ok {
		return defaultLabel
	}
	return names[0]
}

var labelToSectionName = map[string]string{
	"Dependencies":    "Dependencies",
	"API Change":      "API Changes",
	"Behavior Change": "Behavior Changes",
	"Feature":         "New Features",
	"Performance":     "Performance Improvements",
	"Bug":             "Bug Fixes",
	"Documentation":   "Documentation",
}

func sortSections(sections []*Section) []*Section {
	var sss []*Section
	for _, ss := range sections {
		if len(ss.Entries) > 0 {
			sss = append(sss, ss)
		}
	}
	sort.Slice(sss, func(i, j int) bool {
		return sortWeight[sections[i].LabelName] >= sortWeight[sections[j].LabelName]
	})
	return sss
}
