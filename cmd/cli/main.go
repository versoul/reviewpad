// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/reviewpad/v2"
	"github.com/reviewpad/reviewpad/v2/collector"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	dryRun         = flag.Bool("dry-run", false, "Dry run mode")
	reviewpadFile  = flag.String("reviewpad", "", "File path to reviewpad.yml")
	pullRequestUrl = flag.String("pull-request", "", "Pull request GitHub url")
	gitHubToken    = flag.String("github-token", "", "GitHub token")
	mixpanelToken  = flag.String("mixpanel-token", "", "Mixpanel token")
)

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "help" {
		usage()
	}

	if *reviewpadFile == "" {
		log.Printf("Missing argument reviewpad.")
		usage()
	}

	if *pullRequestUrl == "" {
		log.Printf("Missing argument reviewpad.")
		usage()
	}

	if *gitHubToken == "" {
		log.Printf("Missing argument reviewpad.")
		usage()
	}

	pullRequestDetailsRegex := regexp.MustCompile(`github\.com\/(.+)\/(.+)\/pull\/(\d+)`)
	pullRequestDetails := pullRequestDetailsRegex.FindSubmatch([]byte(*pullRequestUrl))

	repositoryOwner := string(pullRequestDetails[1][:])
	repositoryName := string(pullRequestDetails[2][:])
	pullRequestNumber, err := strconv.Atoi(string(pullRequestDetails[3][:]))
	if err != nil {
		log.Fatalf("Error converting pull request number. Details %+q", err.Error())
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *gitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	gitHubClient := github.NewClient(tc)
	gitHubClientGQL := githubv4.NewClient(tc)
	collectorClient := collector.NewCollector(*mixpanelToken, repositoryOwner)

	ghPullRequest, _, err := gitHubClient.PullRequests.Get(ctx, repositoryOwner, repositoryName, pullRequestNumber)
	if err != nil {
		log.Fatal(err)
	}

	headRepoOwner := *ghPullRequest.Head.Repo.Owner.Login
	headRepoName := *ghPullRequest.Head.Repo.Name
	headRef := *ghPullRequest.Head.Ref

	if *ghPullRequest.Merged {
		if ghPullRequest.Head == nil {
			log.Fatal("team-edition: pull request is merged and head branched was automatically deleted\n")
			return
		}

		if ghPullRequest.Head.Repo == nil {
			log.Fatal("team-edition: pull request is merged and head repo branched was automatically deleted\n")
			return
		}

		if ghPullRequest.Head.Repo.DeleteBranchOnMerge == nil || *ghPullRequest.Head.Repo.DeleteBranchOnMerge {
			log.Fatal("team-edition: pull request is merged and head branched was automatically deleted\n")
			return
		}

		_, _, err := gitHubClient.Repositories.GetBranch(ctx, headRepoOwner, headRepoName, headRef, true)
		if err != nil {
			log.Fatal("team-edition: pull request is merged and head branched cannot be retrieved\n")
			return
		}
	}

	data, err := os.ReadFile(*reviewpadFile)
	if err != nil {
		log.Fatalf("Error reading reviewpad file. Details: %v", err.Error())
	}

	buf := bytes.NewBuffer(data)
	file, err := reviewpad.Load(buf)
	if err != nil {
		log.Fatalf("Error running reviewpad team edition. Details %v", err.Error())
	}

	_, err = reviewpad.Run(ctx, gitHubClient, gitHubClientGQL, collectorClient, ghPullRequest, file, *dryRun)
	if err != nil {
		log.Fatalf("Error running reviewpad team edition. Details %v", err.Error())
	}
}
