#!/bin/bash

LOCAL=cosmos,iam_client,taskflow,telect,telect_client,sitekeeper,sitekeeper_client,woden,woden_client,woslo,echometer,echometer_client,squill,squill_client,piraty

git add --all
STAGED_GO_FILES=$(git diff --cached --name-only | grep ".go$")

if [[ "$STAGED_GO_FILES" = "" ]]; then
  exit 0
fi

PASS=true

for FILE in $STAGED_GO_FILES
do
  goimports-reviser -file-path $FILE -rm-unused -local $LOCAL --format --output file

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
  printf "COMMIT FAILED\n"
  exit 1
else
  printf "COMMIT SUCCEEDED\n"
  git add --all
fi

exit 0