#!/bin/bash
LC_ALL=C

local_branch="$(git rev-parse --abbrev-ref HEAD)"

valid_branch_regex="^(master|develop)$|(feature|release|hotfix|wec)\/[a-z0-9._-]+$|^HEAD$"

message="There is something wrong with your branch name. Branch names in this project must adhere to this contract: $valid_branch_regex.
Your commit will be rejected. You should rename your branch to a valid name and try again."

if [[ ! $local_branch =~ $valid_branch_regex ]]
then
    echo "$message"
    exit 1
fi

LOCAL=cosmos,iam_client,taskflow,telect,telect_client,sitekeeper,sitekeeper_client,woden,woden_client,woslo

git add --all
STAGED_GO_FILES=$(git diff --cached --name-only | grep ".go$")

if [[ "$STAGED_GO_FILES" = "" ]]; then
  exit 0
fi

PASS=true

for FILE in $STAGED_GO_FILES
do
  goimports-reviser -file-path $FILE -rm-unused -local $LOCAL --format --output file  --project-name  devt

  # golangci-lint run --disable=typecheck $FILE
  # if [[ $? == 1 ]]; then
  #   PASS=false
  # fi

  # go vet $FILE
  # if [[ $? != 0 ]]; then
  #   PASS=false
  # fi
done

if ! $PASS; then
  echo "COMMIT FAILED\n"
  exit 1
else
  echo "COMMIT SUCCEEDED\n"
  git add --all
fi
