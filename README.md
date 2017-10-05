# drone-aws-device-farm-build

Drone plugin to upload and schedule a test run in AWS Device farm. For the
usage information and a listing of the available options please take a look at
[the docs](DOCS.md).

## Build

Build the binary with the following commands:

```
drone exec
```

## Docker

Build the docker image with the following commands:

```
drone exec
docker build --rm=true -t kowalatech/drone-aws-device-farm-build .
```
## Usage

Execute from the working directory:

```sh
docker run --rm \
  -e AWS_ACCESS_KEY_ID=<accesskeyid> \
  -e AWS_SECRET_ACCESS_KEY=<accesskey> \
  -e PLUGIN_REGION=<region> \
  -e PLUGIN_APP_NAME=<appname> \
  -e PLUGIN_TESTS_NAME=<testsname> \
  -e PLUGIN_TEST_PROJECT=<testsproject> \
  -e PLUGIN_DEVICE_POOLNAME=<devicepoolname> \
  -e PLUGIN_UPLOAD_APP_TYPE=<uploadapptype> \
  -e PLUGIN_TESTS_TYPE=<testtype> \
  -e PLUGIN_TEST_TYPE_RUN=<testtyperun> \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  kowalatech/drone-aws-device-farm-build
```
