package kpackdeploy

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	kpackdeployv1alpha1 "github.com/cpage-pivotal/kpack-deploy-operator/pkg/apis/kpackdeploy/v1alpha1"
)

// maybe add to configuration file??
const branchName = "master"

func writeToGitTarget(latestImage string, git kpackdeployv1alpha1.Git) error {
	if git.WriteMethod == kpackdeployv1alpha1.GIT_PULL_REQUEST {
		return errors.New("write method 'pullrequest' is not supported yet")
	} else if git.WriteMethod != kpackdeployv1alpha1.GIT_COMMIT {
		return fmt.Errorf("write method '%s' is not supported", git.WriteMethod)
	}

	// setup GitHub API client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: git.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner, repo := getOwnerAndRepo(git.Url)
	branch, _, err := client.Repositories.GetBranch(ctx, owner, repo, branchName)
	if err != nil {
		return err
	}

	repoCommit := branch.GetCommit()
	baseCommit := repoCommit.GetCommit()
	baseCommit.SHA = repoCommit.SHA // this is stupid but you have to do it
	baseTreeSha := baseCommit.GetTree().SHA

	baseTree, _, err := client.Git.GetTree(ctx, owner, repo, *baseTreeSha, false)
	if err != nil {
		return err
	}

	newTreeEntries := make([]*github.TreeEntry, 0, len(git.Paths))
pathLoop:
	for _, entry := range baseTree.Entries {
		// this logic will not work if a git.Path is not a direct subfolder of the repository root
		if stringSliceContains(git.Paths, *entry.Path) &&
			// this logic will not work if it is a submodule
			*entry.Type == "tree" {
			subTree, _, err := client.Git.GetTree(ctx, owner, repo, *entry.SHA, false)
			if err != nil {
				return err
			}
			for _, entry2 := range subTree.Entries {
				if *entry2.Path == git.DeploymentFile &&
					*entry2.Type == "blob" {
					// maybe don't return on errors here? maybe do? maybe goroutines?
					blob, _, err := client.Git.GetBlobRaw(ctx, owner, repo, *entry2.SHA)
					if err != nil {
						return err
					}

					newContent, err := updateDeployment(latestImage, string(blob))
					if err != nil {
						return err
					}

					newBlob, _, err := client.Git.CreateBlob(ctx, owner, repo, &github.Blob{
						Content: github.String(newContent),
					})

					newTreeEntries = append(newTreeEntries, &github.TreeEntry{
						Path: github.String(*entry.Path + "/" + *entry2.Path),
						Mode: github.String("100644"),
						Type: github.String("blob"),
						SHA:  newBlob.SHA,
					})

					continue pathLoop
				}
			}
			return errors.New("missing deployment file(s)")
		}
	}

	if len(newTreeEntries) != len(git.Paths) {
		return errors.New("missing deployment path(s)")
	}

	newTree, _, err := client.Git.CreateTree(ctx, owner, repo, *baseTreeSha, newTreeEntries)
	if err != nil {
		return err
	}

	// maybe allow customization of this commit?
	newCommitInfo := &github.Commit{
		Message: github.String("Update from kpack-deploy-operator"),
		Tree:    newTree,
		Parents: []*github.Commit{
			baseCommit,
		},
	}

	newCommit, _, err := client.Git.CreateCommit(ctx, owner, repo, newCommitInfo)
	if err != nil {
		return err
	}

	ref, _, err := client.Git.GetRef(ctx, owner, repo, "heads/"+branchName)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA

	_, _, err = client.Git.UpdateRef(ctx, owner, repo, ref, false)
	if err != nil {
		return err
	}

	return nil
}

var ownerRepoRegex = regexp.MustCompile("^https?://github.com/(.+)/(.+)/?$")

func getOwnerAndRepo(url string) (string, string) {
	matches := ownerRepoRegex.FindStringSubmatch(url)
	return matches[1], matches[2]
}

func stringSliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
