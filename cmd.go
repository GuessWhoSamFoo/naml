//
// Copyright © 2021 Kris Nóva <kris@nivenly.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//    ███╗   ██╗ ██████╗ ██╗   ██╗ █████╗
//    ████╗  ██║██╔═████╗██║   ██║██╔══██╗
//    ██╔██╗ ██║██║██╔██║██║   ██║███████║
//    ██║╚██╗██║████╔╝██║╚██╗ ██╔╝██╔══██║
//    ██║ ╚████║╚██████╔╝ ╚████╔╝ ██║  ██║
//    ╚═╝  ╚═══╝ ╚═════╝   ╚═══╝  ╚═╝  ╚═╝

package naml

import (
	"fmt"
	"os"

	"github.com/kris-nova/logger"
	"github.com/urfave/cli/v2"
)

// RunCommandLine is the global NAML command line program.
//
// Use this if you would like to use the built in NAML command line interface.
func RunCommandLine() error {
	// Default options
	RunCommandLineWithOptions()
	return nil
}

// RunCommandLineWithOptions is here so we can default values in RunCommandLine() that
// we would want to pass in here later (tests, etc)
func RunCommandLineWithOptions() error {
	// with is a set of paths that the user has specificed for naml
	// to run with
	var with cli.StringSlice

	// verbose is the logger verbosity
	var verbose bool = false

	// cli assumes "-v" for version.
	// override that here
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "Print the version",
	}

	// ********************************************************
	// [ NAML APPLICATION ]
	// ********************************************************

	app := &cli.App{
		Name:      "naml",
		HelpName:  "naml",
		Usage:     "YAML alternative for managing Kubernetes packages directly with Go.",
		UsageText: " $ naml [options] <arguments>",
		Description: `
NAML Ain't Markup Langauge. Use NAML to encapsulate Kubernetes applications in Go.
`,
		Version: Version,
		Authors: []*cli.Author{
			{
				Name:  "Kris Nóva",
				Email: "kris@nivenly.com",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Value:       false,
				Usage:       "toggle verbose mode for logger.",
				Destination: &verbose,
			},
			&cli.StringSliceFlag{
				Name:        "with",
				Aliases:     []string{"w"},
				Usage:       "include other naml binaries.",
				Destination: &with,
			},
		},
		Commands: []*cli.Command{

			// ********************************************************
			// [ INSTALL ]
			// ********************************************************

			{
				Name:        "install",
				Aliases:     []string{"i"},
				Description: "Will execute the Install method for a specific app.",
				Usage:       "Install a package in Kubernetes.",
				UsageText:   "naml install [app]",
				Action: func(c *cli.Context) error {
					AllInit(verbose, with.Value())

					arguments := c.Args()
					if arguments.Len() != 1 {
						// Feature: We might want to have "naml install" just iterate through every application.
						cli.ShowCommandHelp(c, "install")
						List()
						return nil
					}
					appName := arguments.First()
					app := Find(appName)
					if app == nil {
						return fmt.Errorf("Invalid application name (Application not registered): %s", appName)
					}
					logger.Info("Installing [%s]", appName)
					return Install(app)
				},
			},

			// ********************************************************
			// [ UNINSTALL ]
			// ********************************************************

			{
				Name:        "uninstall",
				Aliases:     []string{"u"},
				Description: "Will execute the Uninstall method for a specific app.",
				Usage:       "Uninstall a package in Kubernetes",
				UsageText:   "naml uninstall [app]",
				Action: func(c *cli.Context) error {
					AllInit(verbose, with.Value())
					arguments := c.Args()
					if arguments.Len() != 1 {
						// Feature: We might want to have "naml install" just iterate through every application.
						cli.ShowCommandHelp(c, "uninstall")
						List()
						return nil
					}
					appName := arguments.First()
					app := Find(appName)
					if app == nil {
						return fmt.Errorf("Invalid application name (Application not registered): %s", appName)
					}
					logger.Info("Uninstalling [%s]", appName)
					return Uninstall(app)
				},
			},

			// ********************************************************
			// [ LIST ]
			// ********************************************************

			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "[local] List applications.",
				Action: func(c *cli.Context) error {
					AllInit(verbose, with.Value())
					List()
					return nil
				},
			},

			// ********************************************************
			// [ CHILD ]
			// ********************************************************

			{
				Name:    "child",
				Aliases: []string{"c"},
				Usage:   "[Local] Run the program in child runtime mode to be used with another NAML.",
				Action: func(c *cli.Context) error {
					err := RuntimeChild()
					if err != nil {
						return fmt.Errorf("unable to run in runtime mode: %v", err)
					}
					return nil
				},
			},
		},
	}

	return app.Run(os.Args)
}

// AllInit is the "constructor" for every command line flag.
// This is how we use naml -w to include sub-namls
func AllInit(verbose bool, with []string) {

	// [ Verbosity System ]
	if verbose {
		fmt.Println("boops")

		logger.BitwiseLevel = logger.LogEverything
		logger.Always("[Verbose Mode]")
	} else {
		logger.BitwiseLevel = logger.LogAlways | logger.LogCritical | logger.LogWarning | logger.LogDeprecated
	}

	// [ Child Runtime System ]
	if len(with) > 0 {
		for _, childPath := range with {
			err := AddRuntimeChild(childPath)
			if err != nil {
				logger.Warning("unable to add child naml %s: %v", childPath, err)
			} else {
				logger.Success("added child naml: %s", childPath)
			}
		}
	}
}

// Install is used to install an application in Kubernetes
func Install(app Deployable) error {
	client, err := Client()
	if err != nil {
		return err
	}
	err = app.Install(client)
	if err != nil {
		return err
	}
	logger.Success("Successfully installed [%s]", app.Meta().Name)
	return nil
}

// List the naml package information in stdout
func List() {
	fmt.Println("")
	for _, app := range Registry() {
		fmt.Printf("[%s]\n", app.Meta().Name)
		fmt.Printf("\tnamespace  : %s\n", app.Meta().Namespace)
		fmt.Printf("\tversion    : %s\n", app.Meta().ResourceVersion)
		if description, ok := app.Meta().Labels["description"]; ok {
			fmt.Printf("\tdescription : %s\n", description)
		}
		fmt.Printf("\n")
	}
	fmt.Println("")
}

// Uninstall is used to uninstall an application in Kubernetes
func Uninstall(app Deployable) error {
	client, err := Client()
	if err != nil {
		return err
	}
	return app.Uninstall(client)
}
