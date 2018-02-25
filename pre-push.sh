#!/bin/bash

branch_name=$(git symbolic-ref -q HEAD)
branch_name=${branch_name##refs/heads/}
branch_name=${branch_name:-HEAD}

if [ $branch_name == "master" ]; then
    echo "cannot push on master branch"
    exit
fi