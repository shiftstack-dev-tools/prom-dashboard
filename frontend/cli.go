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
	App        *cli.App
}

// NewApp creates and returns a new cli application
func NewApp() CliApp {

	cliApp := CliApp{}
	app := cli.NewApp()
	cliApp.App = app
	app.Description = `This tool pulls down prometheus data from multiple CI test runs,
		aggregates it together, then converts it to a CSV file.`
	app.Version = "1.0.0"
	app.Usage = "A tool to aggregate prometheus data from openshift CI runs"
	app.UsageText = "prom-scrape --config | -c <`FILE`>"

	app.Name = "prom-scrape"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configuration from `FILE`",
			Destination: &cliApp.ConfigPath,
		},
	}

	// Authors
	emilio := cli.Author{
		Name:  "Emilio Garcia",
		Email: "egarcia@redhat.com",
	}
	app.Authors = []cli.Author{emilio}

	return cliApp
}

// ReadInput reads user input data from the cli
func (app *CliApp) ReadInput() (*DataRequest, error) {
	err := app.App.Run(os.Args)
	if err != nil {
		return nil, fmt.Errorf("Failed to run app: %v", err)
	}

	request, err := getRequest(app.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get config: %v", err)
	}

	err = request.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed validation: %v", err)
	}

	return request, nil
}

func getRequest(path string) (*DataRequest, error) {
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
