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
  beanstalk:
    image: peloton/drone-elasctic-beanstalk
    access_key: 970d28f4dd477bc184fbd10b376de753
    secret_key: 9c5785d3ece6a9cdefa42eb99b58986f9095ff1c
    region: us-east-1
    version_label: v1
    description: Deployed with DroneCI
    auto_create: true
    bucket_name: my-bucket-name
    bucket_key: 970d28f4dd477bc184fbd10b376de753
```
