package main

import (
	"fmt"
	"log"
	"os"
	"gopkg.in/yaml.v2"
)

type CoverageTargets struct {
	Files map[string]struct {
		Criticality    string `yaml:"criticality"`
		CurrentCoverage int    `yaml:"current_coverage"`
		TargetCoverage  int    `yaml:"target_coverage"`
	} `yaml:"files"`
	Gates struct {
		GlobalMinimum int `yaml:"global_minimum"`
	} `yaml:"gates"`
}

func main() {
	targetsFile := "../coverage_targets.yaml"
	data, err := os.ReadFile(targetsFile)
	if err != nil {
		log.Fatalf("Error reading %s: %v", targetsFile, err)
	}

	var targets CoverageTargets
	err = yaml.Unmarshal(data, &targets)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	fail := false
	for file, target := range targets.Files {
		if target.CurrentCoverage < target.TargetCoverage {
			fmt.Printf("%s is below target: %d%% < %d%%\n", file, target.CurrentCoverage, target.TargetCoverage)
			fail = true
		} else if target.CurrentCoverage < targets.Gates.GlobalMinimum {
			fmt.Printf("%s is below global minimum: %d%% < %d%%\n", file, target.CurrentCoverage, targets.Gates.GlobalMinimum)
			fail = true
		}
	}

	if fail {
		os.Exit(1)
	}

	fmt.Println("All files meet coverage targets.")
}

