# Authgear Custom Build

## Setup

The following packages are private packages.

1. `github.com/authgear/iamsmart`

You need to doing the following so that you can access those packages

1. Ask for permission to private repositories
1. Update your git config
	```
	[url "ssh://git@github.com/authgear/the-private-repository"]
		insteadOf = https://github.com/authgear/the-private-repository
	```
1. Include the private packages in `GOPRIVATE`