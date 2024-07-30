#!/bin/sh

# This command outputs a sorted list of tracked files in Git.
git ls-tree -r HEAD --format "./%(path)" | sort > /tmp/authgear-git-ls-tree

# We build an image that prints the build context.
2>/dev/null 1>/dev/null docker build --no-cache -t build-context -f - . <<EOF
FROM busybox
WORKDIR /build-context
COPY . .
CMD find . -type f
EOF
# This command outputs a sorted list of files in the build context.
docker run --rm build-context | sort > /tmp/authgear-dockerignore

# Filter out files listed in .git/info/exclude
grep -v -f .git/info/exclude /tmp/authgear-dockerignore > /tmp/authgear-dockerignore

# This command prints the lines that is unique to the second argument.
# In other words, files that only exist in /tmp/authgear-dockerignore.
# If the output is non-empty, we found some files that are not tracked by Git, but
# accidentally being included in the build context.
comm -13 /tmp/authgear-git-ls-tree /tmp/authgear-dockerignore > /tmp/authgear-dockerignore-comm
if [ -s /tmp/authgear-dockerignore-comm ]; then
	cat /tmp/authgear-dockerignore-comm
	exit 1
fi
