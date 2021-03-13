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
	profile string
	tmux    bool
}

func configureFromFlags() config {
	c := config{}
	flag.StringVar(&c.profile, "profile", "", "AWS profile name (from $HOME/.aws/credentials)")
	flag.BoolVar(&c.tmux, "tmux", false, "invoke tmux instead of $SHELL")
	flag.Parse()
	return c
}

func main() {
	cfg := configureFromFlags()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
		Profile:                 cfg.profile,
	}))
	expiry := time.Now().Add(time.Hour).Unix()
	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("Credentials.Get: %v", err)
	}
	os.Setenv("AWS_ACCESS_KEY_ID", creds.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", creds.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", creds.SessionToken)
	if sess.Config.Region != nil {
		os.Setenv("AWS_REGION", *sess.Config.Region)
	}
	os.Setenv("ARG_ROLE_CREDS_EXPIRE_AT", fmt.Sprint(expiry))
	os.Setenv("ARG_PROFILE", cfg.profile)
	var cmd *exec.Cmd
	if cfg.tmux {
		tmuxsocket := fmt.Sprintf("%s-%d", cfg.profile, expiry)
		// we start with a dedicated socket, unfortunately, because otherwise the
		// environment variables may cross-pollinate into other tmux sessions via
		// inheritance from the tmux server :-(
		args := []string{"-L", tmuxsocket, "new-session", "-s", tmuxsocket}
		cmd = exec.Command("tmux", args...)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh" // lol windows
		}
		cmd = exec.Command(shell)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("spawning: %v", err)
	}
}
