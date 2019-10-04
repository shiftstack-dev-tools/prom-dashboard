package frontend

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// CliApp stores a single instance of the cli
// and the inputs expected from the user
type CliApp struct {
	ConfigPath string
	OutFile    string
	App        *cli.App
}

// NewApp creates and returns a new cli application
func NewApp() *CliApp {

	app := CliApp{}
	app.App = cli.NewApp()
	app.App.Description = `This tool pulls down prometheus data from multiple CI test runs,
		aggregates it together, then converts it to a CSV file.`
	app.App.Version = "1.0.0"
	app.App.Usage = "A tool to aggregate prometheus data from openshift CI runs"
	app.App.UsageText = "prom-scrape --config | -c <`FILE`>"

	app.App.Name = "prom-scrape"
	app.App.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configuration from `FILE`",
			Destination: &app.ConfigPath,
		},
		cli.StringFlag{
			Name:        "out, o",
			Usage:       "Write CSV to `FILE`",
			Value:       "prom-data.csv",
			Destination: &app.OutFile,
		},
	}

	app.App.Action = noArgs

	// Authors
	emilio := cli.Author{
		Name:  "Emilio Garcia",
		Email: "egarcia@redhat.com",
	}
	pierre := cli.Author{
		Name:  "Pierre Prinetti",
		Email: "pprinett@redhat.com",
	}
	app.App.Authors = []cli.Author{emilio, pierre}

	return &app
}

func noArgs(c *cli.Context) error {
	return nil
}

// ReadInput reads user input data from the cli
func (app *CliApp) ReadInput() (*DataRequest, error) {
	err := app.App.Run(os.Args)
	if err != nil {
		return nil, fmt.Errorf("Failed to run app: %v", err)
	}

	request, err := readDataRequest(app.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get config: %v", err)
	}

	err = request.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed validation: %v", err)
	}

	return request, nil
}

func readDataRequest(path string) (*DataRequest, error) {
	if path == "" {
		return nil, fmt.Errorf("No config file passed")
	}

	req := NewDataRequest()
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s: %v", path, err)
	}

	err = yaml.Unmarshal(yamlFile, req)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal: %v", err)
	}

	return req, nil
}
