package git

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type Client struct {
	log     logrus.FieldLogger
	verbose bool
}

func NewClient(log logrus.FieldLogger, verbose bool) *Client {
	return &Client{
		log:     log,
		verbose: verbose,
	}
}

func (c *Client) CloneRepository(source, dest string) error {
	return c.run("", true, "git", "clone", "--quiet", source, dest)
}

func (c *Client) ResetRepository(repo string) error {
	if err := c.run(repo, true, "git", "reset", "--hard", "--quiet"); err != nil {
		return err
	}

	if err := c.run(repo, true, "git", "prune"); err != nil {
		return err
	}

	return nil
}

func (c *Client) CheckoutBranch(repo, branch string) error {
	return c.run(repo, true, "git", "checkout", "--quiet", branch)
}

func (c *Client) CreateBranch(repo, branch string) error {
	return c.run(repo, true, "git", "checkout", "--quiet", "-B", branch)
}

func (c *Client) Commit(repo, message string) error {
	return c.run(repo, true, "git", "commit", "--quiet", "--all", "--message", message)
}

func (c *Client) Push(repo, remote, branch string) error {
	// do not show stderr so we hide github's remote response text
	return c.run(repo, false, "git", "push", "--quiet", remote, branch)
}

func (c *Client) run(directory string, showErr bool, command string, args ...string) error {
	c.log.Debugf("$ %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Dir = directory

	if c.verbose {
		cmd.Stdout = os.Stdout
	}

	if c.verbose || showErr {
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
