#!/bin/bash

set -ex

NAME=go_template
REPO=git@172.20.2.149:edge_cloud/cloudcenter/${NAME}.git
COMMIT=${1}

# 建立临时目录并拉取代码
mkdir -p out && cd out && \
git clone ${REPO} && cd ${NAME}
# 切换分支
if [[ -z "${COMMIT}" ]]; then
  git checkout -B master origin/master
else
  git checkout -B  origin/${COMMIT}
fi
# 写入commit信息
# 获取当前 Git 仓库的 commit hash
commit_hash=$(git rev-parse HEAD)
# 将 commit hash 写入指定的文件
echo "Current commit hash: $commit_hash"

# 删除git信息并移动目录
rm -rf .git && cd ../.. && \
rm -rf staging/${NAME} && \
mkdir -p staging/${NAME} && \
mv out/${NAME}/* staging/${NAME}
# 为项目添加依赖信息
go mod edit -require=template@v1.0.0 && go mod edit -replace=template@v1.0.0=./staging/${NAME}
rm -rf out/${NAME}

sed -i "s/^COMMIT=\${1.*}$/COMMIT=\${1:-${commit_hash}}/g" ${0}
echo "finish stage! see your staging dir!"
unset NAME REPO COMMIT commit_hash
