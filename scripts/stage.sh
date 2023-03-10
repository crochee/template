#!/bin/bash

set -ex

NAME=go_template
REPO=git@172.20.2.149:edge_cloud/cloudcenter/${NAME}.git
# 建立临时目录并拉取代码
mkdir -p out && cd out && \
git clone ${REPO} && cd ${NAME}
# 切换分支
if [[ -z "${1}" ]]; then
  git checkout -B master origin/master
else
  git checkout -B ${1} origin/${1}
fi
# 删除git信息并移动目录
rm -rf .git && cd ../.. && \
rm -rf staging/${NAME} && \
mkdir -p staging/${NAME} && \
mv out/${NAME}/* staging/${NAME}
# 为项目添加依赖信息
go mod edit -require=template@v1.0.0 && go mod edit -replace=template@v1.0.0=./staging/${NAME}
rm -rf out/${NAME}



