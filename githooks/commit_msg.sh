#!/bin/bash
LC_ALL=C

commit_msg=`awk '{printf("%s",$0)}' $1`
msg_re="^\[(Refactor|Bugfix|Perf|Config|Test|Other)\](\s)*[\s\S]*"
if [[ $commit_msg =~ $msg_re ]]
then
  echo "不合法的 commit 消息提交格式，请使用正确的格式"
  exit 1
fi
