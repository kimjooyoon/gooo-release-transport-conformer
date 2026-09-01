package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-release-transport-conformer/internal/transport"
)

func main() {
	if len(os.Args) < 2 {
		fatal("command is required: generate, conformance, or inspect")
	}
	switch os.Args[1] {
	case "generate":
		generate(os.Args[2:])
	case "conformance":
		generate(os.Args[2:])
	case "inspect":
		inspect(os.Args[2:])
	default:
		fatal(fmt.Sprintf("unknown command %q", os.Args[1]))
	}
}

func generate(args []string) {
	set := flag.NewFlagSet("generate", flag.ContinueOnError)
	set.SetOutput(os.Stderr)
	root := set.String("root", ".", "observed source repository root")
	source := set.String("source", "", "released .gooo source; defaults to ROOT/.gooo/release-transport.gooo")
	output := set.String("output", "", "empty caller-owned output directory")
	operator := set.String("operator-policy-receipt", "", "external immutable operator-policy receipt")
	if err := set.Parse(args); err != nil {
		fatal(err.Error())
	}
	if *output == "" {
		fatal("--output is required")
	}
	receipt, err := transport.Generate(transport.GenerateOptions{Root: *root, Source: *source, Output: *output, OperatorPolicyReceipt: *operator})
	if err != nil {
		fatal(err.Error())
	}
	printSummary(receipt)
}

func inspect(args []string) {
	set := flag.NewFlagSet("inspect", flag.ContinueOnError)
	set.SetOutput(os.Stderr)
	path := set.String("source", "", ".gooo source")
	if err := set.Parse(args); err != nil {
		fatal(err.Error())
	}
	if *path == "" {
		fatal("--source is required")
	}
	contract, _, err := transport.LoadContract(*path)
	if err != nil {
		fatal(err.Error())
	}
	printJSON(struct {
		ContractID  string                   `json:"contract_id"`
		Denominator int                      `json:"denominator"`
		Activities  []string                 `json:"activities"`
		Scenarios   []transport.ScenarioSpec `json:"scenarios"`
	}{ContractID: contract.ContractID, Denominator: contract.Denominator, Activities: transport.RequiredActivities, Scenarios: contract.Scenarios})
}

func printSummary(receipt transport.Receipt) {
	printJSON(struct {
		Decision    transport.Decision `json:"decision"`
		Denominator int                `json:"denominator"`
		Summary     map[string]int     `json:"summary"`
	}{Decision: receipt.Decision, Denominator: receipt.Denominator, Summary: receipt.Summary})
}

func printJSON(value any) {
	raw, err := json.Marshal(value)
	if err != nil {
		fatal(err.Error())
	}
	fmt.Println(string(raw))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
