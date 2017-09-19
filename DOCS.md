Use this plugin to upload and schedule a test run in AWS Device farm. You can
override the default configuration with the following parameters:

* `access_key` - AWS access key ID - Optional
* `secret_key` - AWS secret access key - Optional
* `region` - AWS availability zone
* `app_name` - The name and location of the app. For example, src/folder/native/app_release.apk
* `tests_name` - The location and filename of the tests. For example, src/e2eTests/features.zip
* `test_project` - The name of the AWS device farm project
* `device_poolname` - the name of the Device Pool
* `upload_app_type` - The app upload type. Refer to `http://docs.aws.amazon.com/devicefarm/latest/APIReference/API_CreateUpload.html#API_CreateUpload_RequestSyntax`
* `tests_type` - The test upload type. Refer to `http://docs.aws.amazon.com/devicefarm/latest/APIReference/API_CreateUpload.html#API_CreateUpload_RequestSyntax`
* `test_type_run` - The test's type. Refer to `http://docs.aws.amazon.com/devicefarm/latest/APIReference/API_ScheduleRunTest.html`

## Example

The following is a sample configuration in your .drone.yml file:

```yaml
deploy:
  behavior-testing:
        image: plugins/drone-aws-device-farm
        access_key: hjwgjhgjhwe
        secret_key: werkjhwekjrhweuyiuwerbyiuweyrui
        region: us-west-2
        app_name: src/folder/native/app_release.apk
        tests_name: src/e2eTests/features.zip
        test_project: androidTestProject
        device_poolname: MostUsedDevicesPool
        upload_app_type: ANDROID_APP
        tests_type: CALABASH_TEST_PACKAGE
        test_type_run: CALABASH
```
