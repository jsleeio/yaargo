package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

type config struct {
	duration   time.Duration
	profile    string
	tmux       bool
	docker     bool
	image      string
	entrypoint string
}

func configureFromFlags() config {
	c := config{}
	flag.DurationVar(&c.duration, "duration", time.Hour, "override credential lifetime")
	flag.StringVar(&c.profile, "profile", "", "AWS profile name (from $HOME/.aws/credentials)")
	flag.BoolVar(&c.tmux, "tmux", false, "invoke tmux instead of $SHELL")
	flag.BoolVar(&c.docker, "docker", false, "pass role credentials to a new Docker container")
	flag.StringVar(&c.image, "image", "alpine", "Docker image to use")
	flag.StringVar(&c.entrypoint, "entrypoint", "/bin/sh", "Docker entrypoint to use (set to empty string to use image default)")
	flag.Parse()
	if c.docker && c.tmux {
		log.Fatal("Aborting: specify exactly one of -tmux or -docker, please")
	}
	return c
}

func (c config) tmuxCommand(expiresAt time.Time) *exec.Cmd {
	tmuxsocket := fmt.Sprintf("%s-%d", c.profile, expiresAt.Unix())
	// we start with a dedicated socket, unfortunately, because otherwise the
	// environment variables may cross-pollinate into other tmux sessions via
	// inheritance from the tmux server :-(
	args := []string{"-L", tmuxsocket, "new-session", "-s", tmuxsocket}
	return exec.Command("tmux", args...)
}

func (c config) shellCommand() *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh" // lol windows
	}
	return exec.Command(shell)
}

func (c config) dockerCommand() *exec.Cmd {
	args := []string{
		"run",
		"--interactive",
		"--tty",
		"--rm",
		"--env", "ARG_ROLE_CREDS_EXPIRE_AT",
		"--env", "AWS_ACCESS_KEY_ID",
		"--env", "AWS_PROFILE",
		"--env", "AWS_REGION",
		"--env", "AWS_SECRET_ACCESS_KEY",
		"--env", "AWS_SESSION_TOKEN",
	}
	if c.entrypoint != "" {
		args = append(args, "--entrypoint")
		args = append(args, c.entrypoint)
	}
	args = append(args, c.image)
	return exec.Command("docker", args...)
}

func (c config) command(expiresAt time.Time) *exec.Cmd {
	switch {
	case c.docker:
		return c.dockerCommand()
	case c.tmux:
		return c.tmuxCommand(expiresAt)
	default:
		return c.shellCommand()
	}
}

func main() {
	cfg := configureFromFlags()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		AssumeRoleDuration:      cfg.duration,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
		Profile:                 cfg.profile,
	}))
	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("Credentials.Get: %v", err)
	}
	expiresAt, err := sess.Config.Credentials.ExpiresAt()
	if err != nil {
		log.Fatalf("Credentials.ExpiresAt: %v", err)
	}
	os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
	if sess.Config.Region != nil {
		os.Setenv("AWS_REGION", *sess.Config.Region)
	}
	os.Setenv("ARG_ROLE_CREDS_EXPIRE_AT", fmt.Sprint(expiresAt.Unix()))
	os.Setenv("ARG_PROFILE", cfg.profile)
	cmd := cfg.command(expiresAt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("spawning: %v", err)
	}
}
