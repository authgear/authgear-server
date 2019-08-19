#!/bin/sh -eu

scripts/update-gh-pages.sh
git push https://skygear-bot:$GITHUB_TOKEN@github.com/SkygearIO/skygear-server.git gh-pages > /dev/null 2>/dev/null
