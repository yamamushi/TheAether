#!/bin/bash

cp pre-push.sh .git/hooks/pre-push
cp pre-commit.sh .git/hooks/pre-commit

chmod +x .git/hooks/pre-push
chmod +x .git/hooks/pre-commit