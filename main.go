package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var build string

func main() {
	app := cli.NewApp()
	app.Name = "Beanstalk deployment plugin"
	app.Usage = "beanstalk deployment plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.0+%s", build)
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-west-2",
			EnvVar: "PLUGIN_REGION",
		},
		cli.StringFlag{
			Name:   "app-directory",
			Usage:  "Location directory of app",
			EnvVar: "PLUGIN_APP_DIRECTORY",
		},
		cli.StringFlag{
			Name:   "app-name",
			Usage:  "Name of app to upload. With extension",
			EnvVar: "PLUGIN_APP_NAME",
		},
		cli.StringFlag{
			Name:   "tests-directory",
			Usage:  "Location directory of tests",
			EnvVar: "PLUGIN_TEST_DIRECTORY",
		},
		cli.StringFlag{
			Name:   "tests-name",
			Usage:  "Name of the .zip file. With extension",
			EnvVar: "PLUGIN_TESTS_NAME",
		},
		cli.StringFlag{
			Name:   "test-project",
			Usage:  "Name of the AWS Device farm project where you want to upload the app, tests, and schedule the run",
			EnvVar: "PLUGIN_TEST_PROJECT",
		},
		cli.StringFlag{
			Name:   "device-poolname",
			Usage:  "Name of the AWS device farm Device pool name to use when running the tests",
			EnvVar: "PLUGIN_DEVICE_POOLNAME",
		},
		cli.StringFlag{
			Name:   "upload-app-type",
			Usage:  "The type of the app that is going to be tested",
			EnvVar: "PLUGIN_UPLOAD_APP_TYPE",
		},
		cli.StringFlag{
			Name:   "test-type-upload",
			Usage:  "The type tests that you are uploading",
			EnvVar: "PLUGIN_TESTS_TYPE",
		},
		cli.StringFlag{
			Name:   "test-type-run",
			Usage:  "Type of the tests",
			EnvVar: "PLUGIN_TEST_TYPE_RUN",
		},
		cli.BoolTFlag{
			Name:   "yaml-verified",
			Usage:  "Ensure the yaml was signed",
			EnvVar: "DRONE_YAML_VERIFIED",
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func run(c *cli.Context) error {
	plugin := Plugin{
		Key:            c.String("access-key"),
		Secret:         c.String("secret-key"),
		Region:         c.String("region"),
		AppDirectory:   c.String("app-directory"),
		AppName:        c.String("app-name"),
		TestsDirectory: c.String("tests-directory"),
		TestsName:      c.String("tests-name"),
		TestProject:    c.String("test-project"),
		DevicePoolname: c.String("device-poolname"),
		UploadAppType:  c.String("upload-app-type"),
		TestTypeUpload: c.String("test-type-upload"),
		TestTypeRun:    c.String("test-type-run"),
		YamlVerified:   c.BoolT("yaml-verified"),
	}

	return plugin.Exec()
}
