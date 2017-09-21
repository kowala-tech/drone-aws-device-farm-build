package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

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

	}
	//create Device Farm service
	svc := devicefarm.New(session.New(), conf)

	//Get AWS device farm Test project
	project, err := getTestProject(p.TestProject, svc)
	if err != nil {
		return err
	}
	//Get AWS device farm Device pool used for this test
	pool, err := getDevicePool(p.DevicePoolname, project, svc)
	if err != nil {
		return err
	}
	//upload tests to AWS device farm
	uploadResponseTests, err := handleUpload(p.TestsName, p.TestTypeUpload, project, svc)
	if err != nil {
		return err
	}
	//upload app to AWS device farm
	uploadResponseApp, err := handleUpload(p.AppName, p.UploadAppType, project, svc)
	if err != nil {
		return err
	}
	//Schedule the test run
	_, err = scheduleRun("Run", pool, project, uploadResponseApp, uploadResponseTests, p.TestTypeRun, svc)
	if err != nil {
		return err
	}
	fmt.Println("Schedule test run completed")

	return nil
}

func scheduleRun(runName string, devicePool *devicefarm.DevicePool, project *devicefarm.Project, apkUpload *devicefarm.CreateUploadOutput, uploadTests *devicefarm.CreateUploadOutput, testType string, svc *devicefarm.DeviceFarm) (*devicefarm.ScheduleRunOutput, error) {
	return svc.ScheduleRun(&devicefarm.ScheduleRunInput{Name: aws.String(runName),
		DevicePoolArn: aws.String(*devicePool.Arn),
		ProjectArn:    aws.String(*project.Arn),
		AppArn:        aws.String(*apkUpload.Upload.Arn),
		Test: &devicefarm.ScheduleRunTest{
			Type:           aws.String(testType),
			TestPackageArn: aws.String(*uploadTests.Upload.Arn)}})
}

func handleUpload(filename string, uploadType string, project *devicefarm.Project, svc *devicefarm.DeviceFarm) (*devicefarm.CreateUploadOutput, error) {
	//Create app upload object
	uploadResponse, err := createUpload(path.Base(filename), uploadType, project, svc)
	if err != nil {
		return nil, err
	}
	//Upload the app file to AWS
	err = uploadFile(filename, uploadResponse, svc)
	if err != nil {
		return nil, err
	}

	//Wait until AWS finishes to process the file
	//Poll AWS every 2 seconds if the file was processed and succeeded
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		filesuccededToUpload, err := checkToSeeIfFileSucceeded(uploadResponse, svc)
		if err != nil {
			return nil, err
		}

		if filesuccededToUpload {
			break
		}
	}

	return uploadResponse, nil
}

func checkToSeeIfFileSucceeded(uploadResponse *devicefarm.CreateUploadOutput, svc *devicefarm.DeviceFarm) (bool, error) {
	result, err := svc.GetUpload(&devicefarm.GetUploadInput{Arn: aws.String(*uploadResponse.Upload.Arn)})
	if err != nil {
		return false, err
	}

	if strings.ToUpper(*result.Upload.Status) == "SUCCEEDED" {
		return true, nil
	} else if strings.ToUpper(*result.Upload.Status) == "FAILED" {
		return false, errors.New(*result.Upload.Metadata)
	}
	return false, nil
}

func uploadFile(file string, uploadResponse *devicefarm.CreateUploadOutput, svc *devicefarm.DeviceFarm) error {
	c := exec.Command("curl", "-T", file, *uploadResponse.Upload.Url)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func createUpload(filename string, typeUpload string, project *devicefarm.Project, svc *devicefarm.DeviceFarm) (*devicefarm.CreateUploadOutput, error) {
	return svc.CreateUpload(&devicefarm.CreateUploadInput{Name: aws.String(filename), Type: aws.String(typeUpload), ProjectArn: aws.String(*project.Arn)})
}

func getDevicePool(devicePoolname string, project *devicefarm.Project, svc *devicefarm.DeviceFarm) (*devicefarm.DevicePool, error) {
	result, err := svc.ListDevicePools(&devicefarm.ListDevicePoolsInput{Arn: aws.String(*project.Arn)})
	if err != nil {
		return nil, err
	}

	for _, pool := range result.DevicePools {
		if strings.ToUpper(*pool.Name) == strings.ToUpper(devicePoolname) {
			return pool, nil
		}
	}
	return nil, fmt.Errorf("There was no device pool with the name %s", devicePoolname)
}

func getTestProject(testProjectName string, svc *devicefarm.DeviceFarm) (*devicefarm.Project, error) {
	result, err := svc.ListProjects(nil)
	if err != nil {
		return nil, err
	}

	for _, project := range result.Projects {
		if *project.Name == testProjectName {
			return project, nil
		}
	}

	return nil, fmt.Errorf("There was no project with the name %s", testProjectName)
}
