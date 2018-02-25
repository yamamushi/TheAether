#!/bin/bash

branch_name=$(git symbolic-ref -q HEAD)
branch_name=${branch_name##refs/heads/}
branch_name=${branch_name:-HEAD}

if [ $branch_name != "master" ]; then
    echo "cannot deploy on non master branch"
    say "cannot deploy on non master branch"
    exit
fi

# Will only work after docker login!
eval $(docker-machine env theaether-hub)

# In case we run this from a different directory
cd $GOPATH/src/github.com/yamamushi/TheAether

docker stop theaether-master-container
yes | docker container prune
yes | docker image prune
docker build -t theaetherbot .
docker tag theaetherbot yamamushi/theaetherbot
docker push yamamushi/theaetherbot
docker run -v /mnt/theaether-hub:/AetherData --name theaether-master-container -d yamamushi/theaetherbot

unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    CYGWIN*)    machine=Cygwin;;
    MINGW*)     machine=MinGw;;
    *)          machine="UNKNOWN:${unameOut}"
esac

echo "deployment completed"

if [ machine != "Mac" ]; then
    say deployment completed
    exit
fi


