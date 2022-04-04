package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.xrstf.de/prow-aliases-syncer/pkg/git"
	"go.xrstf.de/prow-aliases-syncer/pkg/github"
	"go.xrstf.de/prow-aliases-syncer/pkg/prow"
	"go.xrstf.de/prow-aliases-syncer/pkg/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type options struct {
	organization   string
	branches       []string
	maxAge         time.Duration
	dryRun         bool
	updateDirectly bool
	strict         bool
	keep           bool
	verbose        bool
}

const prBody = `
This pull request updates the %s file based on the GitHub team associations.

**Release Notes:**
§§§release-note
NONE
§§§
`

func main() {
	opt := options{
		maxAge: 90 * 24 * time.Hour,
	}

	pflag.StringVarP(&opt.organization, "org", "o", opt.organization, "GitHub organization to work with")
	pflag.StringSliceVarP(&opt.branches, "branch", "b", opt.branches, "branch to update (glob expression supported) (can be given multiple times)")
	pflag.BoolVar(&opt.dryRun, "dry-run", opt.dryRun, "do not actually push to GitHub (repositories will still be cloned and locally updated)")
	pflag.BoolVarP(&opt.strict, "strict", "s", opt.strict, "compare owners files byte by byte")
	pflag.BoolVarP(&opt.updateDirectly, "update", "u", opt.updateDirectly, "do not create pull requests, but directly push into the target branches")
	pflag.BoolVarP(&opt.keep, "keep", "k", opt.keep, "keep unknown teams (do not combine with -strict)")
	pflag.BoolVarP(&opt.verbose, "verbose", "v", opt.verbose, "Enable more verbose output")
	pflag.Parse()

	// setup logging
	var log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	})

	if opt.verbose {
		log.SetLevel(logrus.DebugLevel)
	}

	// validate CLI flags
	if opt.organization == "" {
		log.Fatal("No -org given.")
	}

	if len(opt.branches) == 0 {
		log.Fatal("No -branch given.")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		log.Fatal("No GITHUB_TOKEN environment variable defined.")
	}

	logger := log.WithField("org", opt.organization)

	// setup API client
	ctx := context.Background()

	client, err := github.NewClient(ctx, logger, token)
	if err != nil {
		logger.Fatalf("Failed to create API client: %v", err)
	}

	if err := work(ctx, client, logger, opt); err != nil {
		logger.Fatalf("Failure: %v", err)
	}

	logger.Info("Synchronization completed.")
}

func work(ctx context.Context, client *github.Client, log logrus.FieldLogger, opt options) error {
	// list all teams and their members
	log.Info("Listing teams…")

	teams, err := client.GetTeams(opt.organization)
	if err != nil {
		return err
	}

	// list all repos with all branches and the OWNERS_ALIASES file in each of them
	log.Info("Listing repositories and branches…")

	repos, err := client.GetRepositoriesAndBranches(opt.organization)
	if err != nil {
		return err
	}

	log.Infof("Found %d repositories.", len(repos))

	todo, err := createJobs(ctx, log, opt, repos, teams)
	if err != nil {
		return fmt.Errorf("failed to determine tasks: %w", err)
	}

	if len(todo) == 0 {
		return nil
	}

	if err := processTasks(ctx, client, log, opt, todo); err != nil {
		return fmt.Errorf("failed to process: %w", err)
	}

	return nil
}

func createJobs(ctx context.Context, log logrus.FieldLogger, opt options, repos []github.Repository, teams []github.Team) ([]github.Repository, error) {
	todo := []github.Repository{}

	for _, r := range repos {
		rlog := log.WithField("repo", r.Name)
		branchesToUpdate := []github.Branch{}

		for _, b := range r.Branches {
			blog := rlog.WithField("branch", b.Name)

			// apply branch filter
			if !includeBranch(b.Name, opt.branches) {
				blog.Debug("Ignored.")
				continue
			}

			// ignore stale branches
			if time.Since(b.MostRecentCommit) > opt.maxAge {
				blog.Debug("No recent activity, ignored.")
				continue
			}

			// if the branch has no alias file, ignore it
			if b.Aliases == "" {
				blog.Debug("Has no aliases file.")
				continue
			}

			equal, newAliases, err := util.Equal(b.Aliases, teams, opt.strict, opt.keep)
			if err != nil {
				blog.WithError(err).Warn("Invalid aliases file.")
				continue
			}

			if !equal {
				blog.Info("File is not identical.")

				// store the new data so we do not have to generate it again later
				b.Aliases = newAliases

				branchesToUpdate = append(branchesToUpdate, b)
			} else {
				blog.Debug("No changes detected.")
			}
		}

		if len(branchesToUpdate) > 0 {
			todo = append(todo, github.Repository{
				ID:       r.ID,
				Name:     r.Name,
				Branches: branchesToUpdate,
			})
		}
	}

	return todo, nil
}

func processTasks(ctx context.Context, client *github.Client, log logrus.FieldLogger, opt options, tasks []github.Repository) error {
	tmpDir, err := os.MkdirTemp("", "xrstf*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	gitter := git.NewClient(log, opt.verbose)

	for _, task := range tasks {
		tlog := log.WithField("repo", task.Name)
		tlog.Info("Processing…")

		cloned := false

		repoURL := fmt.Sprintf("git@github.com:%s/%s.git", opt.organization, task.Name)
		repoDir := filepath.Join(tmpDir, task.Name)

		for _, branch := range task.Branches {
			blog := tlog.WithField("branch", branch.Name)
			newBranch := fmt.Sprintf("update-%s-owners", branch.Name)
			newBranch = strings.ReplaceAll(newBranch, "/", "-")

			if !opt.updateDirectly {
				prNumber, err := client.GetPullRequestForBranch(opt.organization, task.Name, branch.Name, newBranch)
				if err != nil {
					blog.WithError(err).Warn("Failed to check for existing pull request.")
					continue
				}

				if prNumber > 0 {
					blog.WithField("pr", prNumber).Info("Pull request already open.")
					continue
				}

			}

			if !cloned {
				tlog.Debug("Cloning…")
				if err := gitter.CloneRepository(repoURL, repoDir); err != nil {
					tlog.WithError(err).Warn("Failed to clone repository.")
					continue
				}

				cloned = true
			}

			// just for safety
			if err := gitter.ResetRepository(repoDir); err != nil {
				blog.WithError(err).Warn("Failed to reset working copy.")
				continue
			}

			if err := gitter.CheckoutBranch(repoDir, branch.Name); err != nil {
				blog.WithError(err).Warn("Failed to checkout branch.")
				continue
			}

			if !opt.updateDirectly {
				if err := gitter.CreateBranch(repoDir, newBranch); err != nil {
					blog.WithError(err).Warn("Failed to create new branch.")
					continue
				}
			}

			filename := filepath.Join(repoDir, prow.OwnersAliasesFilename)
			if err := os.WriteFile(filename, []byte(branch.Aliases), 0644); err != nil {
				blog.WithError(err).Warn("Failed to update file.")
				continue
			}

			commitMsg := fmt.Sprintf("Synchronize %s file with Github teams", prow.OwnersAliasesFilename)
			if branch.Name != "master" && branch.Name != "main" {
				commitMsg = fmt.Sprintf("[%s] %s", branch.Name, commitMsg)
			}

			if err := gitter.Commit(repoDir, commitMsg); err != nil {
				blog.WithError(err).Warn("Failed to commit changes.")
				continue
			}

			if opt.dryRun {
				if !opt.updateDirectly {
					blog = blog.WithField("new-branch", newBranch)
				}

				blog.Info("Dry run, not pushing branch.")
				continue
			}

			if err := gitter.Push(repoDir, "origin", newBranch); err != nil {
				blog.WithError(err).Warn("Failed to push changes.")
				continue
			}

			if opt.updateDirectly {
				blog.Info("Branch updated.")
			} else {
				body := fmt.Sprintf(prBody, prow.OwnersAliasesFilename)
				body = strings.ReplaceAll(body, "§", "`")
				body = strings.TrimSpace(body)

				prNumber, err := client.CreatePullRequest(task.ID, branch.Name, newBranch, commitMsg, body)
				if err != nil {
					blog.WithError(err).Warn("Failed to create pull request.")
					continue
				}

				blog.WithField("pr", prNumber).Info("Pull request created.")
			}
		}
	}

	return nil
}

func includeBranch(branch string, enabled []string) bool {
	for _, b := range enabled {
		if matched, _ := filepath.Match(b, branch); matched {
			return true
		}
	}

	return false
}
