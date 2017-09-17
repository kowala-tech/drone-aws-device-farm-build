package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/devicefarm"
)

func main() {
	fmt.Println("Begin to schedule test run in AWS Device Farm'")
	region := "us-west-2"
	appDirectory := "./"
	appName := "app-release.apk"
	testsDirectory := "./"
	testsName := "features.zip"
	testProject := "kowalaAndroidTest"
	devicePoolname := "KOWALAANDROID"
	uploadAppType := "ANDROID_APP"
	testTypeUpload := "CALABASH_TEST_PACKAGE"
	testTypeRun := "CALABASH"

	fmt.Println(region)
	fmt.Println(appDirectory)
	fmt.Println(appName)
	fmt.Println(testsDirectory)
	fmt.Println(testsName)
	fmt.Println(testProject)
	fmt.Println(devicePoolname)
	fmt.Println(uploadAppType)
	fmt.Println(testTypeUpload)
	fmt.Println(testTypeRun)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	svc := devicefarm.New(sess)

	project := getTestProject(testProject, svc)
	fmt.Println("project", project)
	pool := getDevicePool(devicePoolname, project, svc)
	fmt.Println("pool", pool)

	uploadResponseTests := createUpload(testsName, testTypeUpload, project, svc)
	fmt.Println("uploadResponseTests", uploadResponseTests)
	s := []string{testsDirectory, testsName}
	uploadFile(strings.Join(s, ""), uploadResponseTests, svc)
	testsSuccededToUpload := false
	for {
		testsSuccededToUpload = checkToSeeIfFileSucceeded(uploadResponseTests, svc)
		if testsSuccededToUpload {
			break
		}
	}

	uploadResponseApp := createUpload(appName, uploadAppType, project, svc)
	fmt.Println("uploadResponseApp", uploadResponseApp)
	appLocation := []string{appDirectory, appName}
	uploadFile(strings.Join(appLocation, ""), uploadResponseApp, svc)
	appSuccededToUpload := false
	for {
		appSuccededToUpload = checkToSeeIfFileSucceeded(uploadResponseApp, svc)
		if appSuccededToUpload {
			break
		}
	}
	run := scheduleRun("Run", pool, project, uploadResponseApp, uploadResponseTests, testTypeRun, svc)
	fmt.Println("Run", run)
	fmt.Println("Schedule test run completed")
}

func scheduleRun(runName string, devicePool *devicefarm.DevicePool, project *devicefarm.Project, apkUpload *devicefarm.CreateUploadOutput, uploadTests *devicefarm.CreateUploadOutput, testType string, svc *devicefarm.DeviceFarm) *devicefarm.ScheduleRunOutput {
	result, err := svc.ScheduleRun(&devicefarm.ScheduleRunInput{Name: aws.String(runName),
		DevicePoolArn: aws.String(*devicePool.Arn),
		ProjectArn:    aws.String(*project.Arn),
		AppArn:        aws.String(*apkUpload.Upload.Arn),
		Test: &devicefarm.ScheduleRunTest{
			Type:           aws.String(testType),
			TestPackageArn: aws.String(*uploadTests.Upload.Arn)}})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return result
}

func checkToSeeIfFileSucceeded(uploadResponse *devicefarm.CreateUploadOutput, svc *devicefarm.DeviceFarm) bool {
	result, err := svc.GetUpload(&devicefarm.GetUploadInput{Arn: aws.String(*uploadResponse.Upload.Arn)})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("resultado", result)

	if strings.ToUpper(*result.Upload.Status) == "SUCCEEDED" {
		return true
	} else if strings.ToUpper(*result.Upload.Status) == "FAILED" {
		fmt.Println("The file failed to upload to AWS", result.Upload)
		os.Exit(1)
	}
	return false
}

func uploadFile(file string, uploadResponse *devicefarm.CreateUploadOutput, svc *devicefarm.DeviceFarm) {
	fmt.Println("file", file)
	c := exec.Command("curl", "-T", file, *uploadResponse.Upload.Url)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Listo")
}

func createUpload(filename string, typeUpload string, project *devicefarm.Project, svc *devicefarm.DeviceFarm) *devicefarm.CreateUploadOutput {
	result, err := svc.CreateUpload(&devicefarm.CreateUploadInput{Name: aws.String(filename), Type: aws.String(typeUpload), ProjectArn: aws.String(*project.Arn)})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return result
}

func getDevicePool(devicePoolname string, project *devicefarm.Project, svc *devicefarm.DeviceFarm) *devicefarm.DevicePool {
	result, err := svc.ListDevicePools(&devicefarm.ListDevicePoolsInput{Arn: aws.String(*project.Arn)})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, pool := range result.DevicePools {
		if strings.ToUpper(*pool.Name) == devicePoolname {
			return pool
		}
	}

	fmt.Println("There was no device pool with that name")
	os.Exit(1)
	return nil
}

func getTestProject(testProjectName string, svc *devicefarm.DeviceFarm) *devicefarm.Project {
	result, err := svc.ListProjects(nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, project := range result.Projects {
		if *project.Name == testProjectName {
			return project
		}
	}

	fmt.Println("There was no project with that name")
	os.Exit(1)
	return nil
}
