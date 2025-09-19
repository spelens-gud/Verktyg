git describe --tags
read -p "请输入tag名:" TAG
git-chglog --next-tag "${TAG}" --output CHANGELOG.md
echo "package version

var version = \"${TAG}\"" > version/version.go
git add .
git commit -am "doc(CHANGELOG): ${TAG}"
git tag "${TAG}"