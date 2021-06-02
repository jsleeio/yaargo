# yaargo

Yet Another Assume-Role in GOlang 

* small, single-file, easily-audited source code
* optionally spawns a `tmux` server
* optionally spawns a Docker container (default: `alpine`)
* uses existing AWS SDK config (`$HOME/.aws/credentials` profiles)
* prompts for MFA if required

## usage

```
$ ./yaargo --help
Usage of ./yaargo:
  -docker
    	pass role credentials to a new Docker container
  -duration duration
    	override credential lifetime (default 1h0m0s)
  -entrypoint string
    	Docker entrypoint to use (set to empty string to use image default) (default "/bin/sh")
  -image string
    	Docker image to use (default "alpine")
  -profile string
    	AWS profile name (from $HOME/.aws/credentials)
  -tmux
    	invoke tmux instead of $SHELL
```
