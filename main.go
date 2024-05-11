package main

import (
	"fmt"
	"log"
	"slices"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func getTags(r *git.Repository) []*plumbing.Reference {
	tagRefIter, _ := r.Tags()
	defer tagRefIter.Close()
	var tags []*plumbing.Reference
	tagRefIter.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref)
		return nil
	})
	slices.SortFunc(tags, func(a, b *plumbing.Reference) int {
		aCommit, _ := r.CommitObject(a.Hash())
		aDate := aCommit.Author.When
		bCommit, _ := r.CommitObject(b.Hash())
		bDate := bCommit.Author.When
		if aDate.Before(bDate) {
			return -1
		} else if aDate.After(bDate) {
			return 1
		} else {
			return 0
		}
	})
	return tags
}

func isConventional(commitMsg string) bool {
	prefixes := [6]string{"Add", "Fix", "Change", "Ref", "Perf", "Doc"}
	for _, pref := range prefixes {
		if strings.HasPrefix(commitMsg, pref) || strings.HasPrefix(commitMsg, strings.ToLower(pref)) {
			return true
		}
	}
	return false
}

func getConventionalCommits(r *git.Repository) []*object.Commit {
	commitIter, _ := r.CommitObjects()
	var commits []*object.Commit
	commitIter.ForEach(func(comm *object.Commit) error {
		if isConventional(comm.Message) {
			commits = append(commits, comm)
		}
		return nil
	})
	slices.SortFunc(commits, func(a, b *object.Commit) int {
		aDate := a.Author.When
		bDate := b.Author.When
		if aDate.Before(bDate) {
			return -1
		} else if aDate.After(bDate) {
			return 1
		} else {
			return 0
		}
	})
	return commits
}

func main() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatal("Unable to open repository")
	}
	tags := getTags(repo)
	commits := getConventionalCommits(repo)
	fmt.Println(commits, tags)
}
