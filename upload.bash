#!/usr/bin/env bash

if [[ -z $1 ]]; then
    echo "Syntax: ${0} <VERSION>"
    exit 1
fi

VERSION=$1

set -Eeuo pipefail
set -x

rm -rf dist/upload
mkdir -p dist/upload
pushd dist/upload/

cat > index.json <<- EOM
{
  "versions": {
    "${VERSION}": {}
  }
}
EOM

cat > ${VERSION}.json <<- EOM
{
  "archives": {
    "darwin_amd64": {
      "url": "terraform-provider-aws_${VERSION}_darwin_amd64.zip"
    },
    "linux_amd64": {
      "url": "terraform-provider-aws_${VERSION}_linux_amd64.zip"
    }
  }
}
EOM

zip -j terraform-provider-aws_${VERSION}_darwin_amd64.zip ../terraform-provider-aws_darwin_amd64/terraform-provider-aws_${VERSION}
zip -j terraform-provider-aws_${VERSION}_linux_amd64.zip ../terraform-provider-aws_linux_amd64/terraform-provider-aws_${VERSION}

aws --profile tgtg-master s3 cp ./ s3://tgtg-public-dist/terraform/providers/terraform-registry.tgtg.ninja/hashicorp/aws/ --acl public-read --recursive
