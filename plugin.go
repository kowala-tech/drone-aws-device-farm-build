package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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

	AppDirectory   string
	AppName        string
	TestsDirectory string
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

	// create the client
	conf := &aws.Config{
		Region: aws.String(p.Region),
	}

	fmt.Println("*conf.Region", *conf.Region)
	fmt.Println("p.Key", p.Key)
	fmt.Println("p.Secret", p.Secret)

	// Use key and secret if provided otherwise fall back to ec2 instance profile
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	} else if p.YamlVerified != true {
		return errors.New("Security issue: When using instance role you must have the yaml verified")
	}

	// log.WithFields(log.Fields{
	// 	"Key":            p.Key,
	// 	"Secret-name":    p.Secret,
	// 	"Region":         p.Region,
	// 	"AppDirectory":   p.AppDirectory,
	// 	"AppName":        p.AppName,
	// 	"TestsDirectory": p.TestsDirectory,
	// 	"TestsName":      p.TestsName,
	// 	"TestProject":    p.TestProject,
	// 	"DevicePoolname": p.DevicePoolname,
	// 	"UploadAppType":  p.UploadAppType,
	// 	"TestTypeUpload": p.TestTypeUpload,
	// 	"TestTypeRun":    p.TestTypeRun,
	// }).Info("Attempting to create and update")
	fmt.Println("Credentials", conf.Credentials)
	cred, _ := conf.Credentials.Get()

	fmt.Println("AccessKeyID", cred.AccessKeyID)
	fmt.Println("SecretAccessKey", cred.SecretAccessKey)
	fmt.Println("ProviderName", cred.ProviderName)
	// fmt.Println("ProviderName", conf.Credentials.Get().ProviderName)
	// fmt.Println("SecretAccessKey", conf.Credentials.Get().SecretAccessKey)

	// sess := session.Must(session.NewSession(conf))
	// sess := s3.New(session.New(), conf)

	fmt.Println("2")

	svc := devicefarm.New(session.New(), conf)
	fmt.Println("3")

	project := getTestProject(p.TestProject, svc)
	fmt.Println("project", project)
	pool := getDevicePool(p.DevicePoolname, project, svc)
	fmt.Println("pool", pool)

	uploadResponseTests := createUpload(p.TestsName, p.TestTypeUpload, project, svc)
	fmt.Println("uploadResponseTests", uploadResponseTests)
	s := []string{p.TestsDirectory, p.TestsName}
	uploadFile(strings.Join(s, ""), uploadResponseTests, svc)
	testsSuccededToUpload := false
	for {
		testsSuccededToUpload = checkToSeeIfFileSucceeded(uploadResponseTests, svc)
		if testsSuccededToUpload {
			break
		}
	}

	uploadResponseApp := createUpload(p.AppName, p.UploadAppType, project, svc)
	fmt.Println("uploadResponseApp", uploadResponseApp)
	appLocation := []string{p.AppDirectory, p.AppName}
	uploadFile(strings.Join(appLocation, ""), uploadResponseApp, svc)
	appSuccededToUpload := false
	for {
		appSuccededToUpload = checkToSeeIfFileSucceeded(uploadResponseApp, svc)
		if appSuccededToUpload {
			break
		}
	}
	run := scheduleRun("Run", pool, project, uploadResponseApp, uploadResponseTests, p.TestTypeRun, svc)
	fmt.Println("Run", run)
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
