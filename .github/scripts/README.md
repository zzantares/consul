# Verification Utilities
This path contains some basic script to assist in verification of release artifacts. Details for running them are below although some can be found in the `build.yml` GitHub Action as they are invoked as part of the build process.

## Docker Images
The `verify_docker_post.sh` script verifies the Docker image for a specific release of either Consul OSS or Consul Enterprise for all supported platforms/architectures.

### Examples
If we want to verify Consul Enterprise Docker images in staging for the `v1.12.2+ent` tag, we will need the Git SHA for the build. See [Obtaining Git SHAs](#obtaining-git-shas) below for printing the SHAs. Then to verify all of the images:
```
./verify_docker_post.sh consul-enterprise ${sha} v1.12.2+ent staging
```
This will construct all of the correct URIs for executing a `docker pull` before running a simple smoke test to ensure the version reported matches the expected version.

For Consul OSS, the following could be run to verify production images for `v1.12.2`:
```
./verify_docker_post.sh consul ${sha} v1.12.2 production
```

## Linux Packages
The `verify_linux_packages.sh` script, given a product, distro and version, verifies the existence of the given release version for the relevant package manager. The following distro strings are supported:
* `ubuntu`
* `debian`
* `centos`
* `fedora`
* `amazon`

### Examples
If we want to verify the Ubuntu Linux packages for Consul Enterprise `v1.12.2`:

```
./verify_linux_packages.sh consul-enterprise ubuntu v1.12.2
```

This will verify that the APT source repository for Ubuntu contains a registration of Consul Enterprise for `1.12.2`.

If we wanted to validate all known distros for both Consul and Consul Enterprise for `1.12.2`, we could write a loop:
```
for product in consul consul-enterprise; do
  for distro in ubuntu debian centos fedora amazon; do
    ./verify_linux_packages.sh "${product}" "${distro}" "1.12.2"
  done
done
```

## Obtaining Git SHAs
In order to get the SHA for a build, run the following in the root of a repository:
```
# print the SHA for 1.12.2
git fetch
git rev-parse release/1.12.2
```
