package license

import (
	"embed"
	"fmt"
	"os"
	"text/template"
	"time"
)

type License struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Url       string `json:"url"`
}

type Result struct {
	Name    string
	Year    string
	License string
}

var licenseList = GetLicenseList()

func GetLicenseList() []string {
	var names []string
	for _, l := range licenseDefinitions {
		names = append(names, l.ShortName)
	}
	return names
}

var licenseDefinitions = []License{
	{Key: "agpl-3.0", Name: "GNU Affero General Public License v3.0", ShortName: "AGPL-3.0", Url: "https://api.github.com/licenses/agpl-3.0"},
	{Key: "apache-2.0", Name: "Apache License 2.0", ShortName: "Apache-2.0", Url: "https://api.github.com/licenses/apache-2.0"},
	{Key: "cc0-1.0", Name: "Creative Commons Zero v1.0 Universal", ShortName: "CC0-1.0", Url: "https://api.github.com/licenses/cc0-1.0"},
	{Key: "gpl-3.0", Name: "GNU General Public License v3.0", ShortName: "GPL-3.0", Url: "https://api.github.com/licenses/gpl-3.0"},
	{Key: "mit", Name: "MIT License", ShortName: "MIT", Url: "https://api.github.com/licenses/mit"},
	{Key: "mpl-2.0", Name: "Mozilla Public License 2.0", ShortName: "MPL-2.0", Url: "https://api.github.com/licenses/mpl-2.0"},
	{Key: "unlicense", Name: "The Unlicense", ShortName: "Unlicense", Url: "https://api.github.com/licenses/unlicense"},
}

//go:embed templates/*
var embeddedTemplates embed.FS

func GenerateLicense(name, year, licenseType, outputPath string) error {
	if name == "" {
		cwd, err := os.Getwd()
		if err != nil {
			name = "Unknown"
		} else {
			name = cwd
		}
	}

	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	if outputPath == "" {
		outputPath = "./LICENSE"
	}

	validLicense := false
	for _, l := range licenseDefinitions {
		if l.ShortName == licenseType {
			validLicense = true
			break
		}
	}
	if !validLicense {
		return fmt.Errorf("invalid license type: %s. Available: %v", licenseType, GetLicenseList())
	}

	tmplFile := fmt.Sprintf("templates/%s.txt", licenseType)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	tmpl, err := template.ParseFS(embeddedTemplates, tmplFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	result := Result{
		Name:    name,
		Year:    year,
		License: licenseType,
	}

	err = tmpl.Execute(file, result)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func MustGenerateLicense(name, year, licenseType, outputPath string) {
	if err := GenerateLicense(name, year, licenseType, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
