#!/bin/bash

version=$1

if [[ "${version}" = "" ]]; then
  echo "usage: ${0} <version>"
  exit 1
fi

rm -rf dist && mkdir dist

glide install || exit 1
go build main.go || exit 1

defaults write "$(pwd)/info.plist" version "${version}"
plutil -convert xml1  "$(pwd)/info.plist"

git add info.plist
git commit -m "🎉  Release ${version}"
git push

zip -r "dist/calendar-${version}.alfredworkflow" . \
  -x vendor\* .git\* bin\* glide.yaml dist\* README.md glide.lock \*.go docs\*

git tag "${version}" && git push --tags

hub release create \
  -m "🎉  Release ${version}" \
  -a "dist/calendar-$version.alfredworkflow" \
  "${version}"