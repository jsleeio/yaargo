# yaargo

Yet Another Assume-Role in GOlang 

* small, single-file, easily-audited source code
* optionally spawns a `tmux` server
* uses existing AWS SDK config (`$HOME/.aws/credentials` profiles)
* prompts for MFA if required

## usage

```
Usage of ./yaargo:
  -profile string
    	AWS profile name (from $HOME/.aws/credentials)
  -tmux
    	invoke tmux instead of $SHELL
```
