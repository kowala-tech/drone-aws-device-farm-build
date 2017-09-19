package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/devicefarm"
)

// Plugin defines the Device farm plugin parameters.
type Plugin struct {
	Key          string
	Secret       string
	Region       string
	YamlVerified bool

	AppName        string
	TestsName      string
	TestProject    string
	DevicePoolname string
	UploadAppType  string
	TestTypeUpload string
	TestTypeRun    string
}

// Exec runs the plugin
func (p *Plugin) Exec() error {
	fmt.Println("Begin to schedule test run in AWS Device Farm")

	// create the configuration
	conf := &aws.Config{
		Region: aws.String(p.Region),
	}

	// Use key and secret if provided otherwise fall back to ec2 instance profile
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	} else if p.YamlVerified != true {
		return errors.New("Security issue: When using instance role you must have the yaml verified")
	}

	//create Device Farmr service
	svc := devicefarm.New(session.New(), conf)

	//Get AWS Test project
	project := getTestProject(p.TestProject, svc)
	//Get AWS device farm Device pool used for this test
	pool := getDevicePool(p.DevicePoolname, project, svc)

	//Create test upload object
	uploadResponseTests := createUpload(path.Base(p.TestsName), p.TestTypeUpload, project, svc)
	//Upload the tests package to AWS
	uploadFile(p.TestsName, uploadResponseTests, svc)
	//Wait until AWS finishes to process the file
	testsSuccededToUpload := false
	for {
		testsSuccededToUpload = checkToSeeIfFileSucceeded(uploadResponseTests, svc)
		if testsSuccededToUpload {
			break
		}
	}
	//Create app upload object
	uploadResponseApp := createUpload(path.Base(p.AppName), p.UploadAppType, project, svc)
	//Upload the app file to AWS
	uploadFile(p.AppName, uploadResponseApp, svc)
	//Wait until AWS finishes to process the file
	appSuccededToUpload := false
	for {
		appSuccededToUpload = checkToSeeIfFileSucceeded(uploadResponseApp, svc)
		if appSuccededToUpload {
			break
		}
	}
	//Scedule the test run
	scheduleRun("Run", pool, project, uploadResponseApp, uploadResponseTests, p.TestTypeRun, svc)
	fmt.Println("Schedule test run completed")

	return nil
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

	if strings.ToUpper(*result.Upload.Status) == "SUCCEEDED" {
		return true
	} else if strings.ToUpper(*result.Upload.Status) == "FAILED" {
		fmt.Println("The file failed to upload to AWS", result.Upload)
		os.Exit(1)
	}
	return false
}

func uploadFile(file string, uploadResponse *devicefarm.CreateUploadOutput, svc *devicefarm.DeviceFarm) {
	c := exec.Command("curl", "-T", file, *uploadResponse.Upload.Url)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
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
