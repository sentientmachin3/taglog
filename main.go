package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	internal "taglog/internal"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ObjectMessageDate struct {
	date    time.Time
	content string
}

func sortObjects(objects []ObjectMessageDate) []ObjectMessageDate {
	slices.SortFunc(objects, func(a, b ObjectMessageDate) int {
		if a.date.Before(b.date) {
			return 1
		} else if a.date.After(b.date) {
			return -1
		} else {
			return 0
		}
	})
	return objects
}

func getTags(r *git.Repository) []ObjectMessageDate {
	tagRefIter, _ := r.Tags()
	defer tagRefIter.Close()
	var tags []ObjectMessageDate
	tagRefIter.ForEach(func(ref *plumbing.Reference) error {
		commit, _ := r.CommitObject(ref.Hash())
		date := commit.Author.When
		msgPathParts := strings.Split(ref.Name().String(), "/")
		msg := msgPathParts[len(msgPathParts)-1]
		tags = append(tags, ObjectMessageDate{date: date, content: msg})
		return nil
	})
	return sortObjects(tags)
}

func isConventional(commitMsg string, prefixes []string) bool {
	for _, pref := range prefixes {
		if strings.HasPrefix(strings.ToLower(commitMsg), strings.ToLower(pref)) {
			return true
		}
	}
	return false
}

func getConventionalCommits(r *git.Repository, prefixes []string) []ObjectMessageDate {
	commitIter, _ := r.CommitObjects()
	var commits []ObjectMessageDate
	commitIter.ForEach(func(comm *object.Commit) error {
		date := comm.Author.When
		if isConventional(comm.Message, prefixes) {
			commits = append(commits, ObjectMessageDate{date: date, content: comm.Message})
		}
		return nil
	})
	return sortObjects(commits)
}

func buildClusters(tags []ObjectMessageDate, commits []ObjectMessageDate) ([]string, map[string][]string) {
	order := []string{"untagged"}
	clusters := make(map[string]([]string))
	clusters["untagged"] = []string{}

	for tagIndex, tag := range tags {
		order = append(order, tag.content)
		for _, commit := range commits {
			if tagIndex == len(tags)-1 {
				clusters[tag.content] = append(clusters[tag.content], commit.content)
				continue
			}
			predecessorTag := tags[tagIndex+1]
			if (commit.date.Before(tag.date) || commit.date.Equal(tag.date)) && commit.date.After(predecessorTag.date) {
				clusters[tag.content] = append(clusters[tag.content], commit.content)
				continue
			}
		}
	}
	return order, clusters
}

func pageOutput(sortedTagNames []string, clusters map[string][]string) {
	maxTagNameLength := internal.MaxStringLength(sortedTagNames)
	padding := strings.Repeat(" ", maxTagNameLength+1)
	// tagFuncColored := color.New(color.FgYellow).SprintFunc()
	cmd := exec.Command("less")
	var lines []string
	for _, tagName := range sortedTagNames {
		for commitIndex, commitMessage := range clusters[tagName] {
			// 	coloredTagName := tagFuncColored(tagName)
			if commitIndex == 0 {
				lines = append(lines, strings.Join([]string{tagName + strings.Repeat(" ", len(padding)-len(tagName)), commitMessage}, " "))
			} else {
				lines = append(lines, strings.Join([]string{padding, commitMessage}, " "))
			}
		}
	}
	cmd.Stdin = strings.NewReader(strings.Join(lines, ""))
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal("Error running pager", err)
	}
}

func main() {
	configPath := flag.String("config", "./taglog.json", "Path to the config list with prefixes")
	repoPath := flag.String("repo", ".", "Path to the repo")
	flag.Parse()
	prefixes := internal.LoadConfig(configPath)

	repo, err := git.PlainOpen(*repoPath)
	if err != nil {
		log.Fatal("Unable to open repository")
	}
	// Tags and commits are sorted in decreasing order
	tags := getTags(repo)
	commits := getConventionalCommits(repo, prefixes)

	order, clusters := buildClusters(tags, commits)
	pageOutput(order, clusters)
}
