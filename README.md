# Rebase from upstream

git remote add upstream git@github.com:hashicorp/terraform-provider-aws.git

git fetch upstream

git rebase -i upstream/master

# Build

install go and goreleaser

goreleaser check

git tag -f v3.24.1

rm -rf ./dist

goreleaser build

./upload.bash 3.24.1

# Use

Setup the following in your `~/.terraformrc`:
```
provider_installation {
  network_mirror {
    url     = "https://tgtg-public-dist.s3-eu-west-1.amazonaws.com/terraform/providers/"
    include = ["terraform-registry.tgtg.ninja/*/*"]
  }
  direct {
    exclude = ["terraform-registry.tgtg.ninja/*/*"]
  }
}
```
