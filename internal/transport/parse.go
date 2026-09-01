package transport

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func DigestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func LoadContract(path string) (Contract, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, nil, err
	}
	contract, err := ParseContract(raw)
	if err != nil {
		return Contract{}, nil, fmt.Errorf("parse contract: %w", err)
	}
	return contract, raw, nil
}

func ParseContract(raw []byte) (Contract, error) {
	var contract Contract
	contract.Precedence = []Decision{}
	contract.UnknownFields = []string{}
	contract.Activities = []Activity{}
	contract.Scenarios = []ScenarioSpec{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(stripComment(scanner.Text()))
		if line == "" || line == "{" || line == "}" {
			continue
		}
		tokens, err := lex(line)
		if err != nil {
			return Contract{}, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "contract":
			if len(tokens) != 2 || contract.ContractID != "" {
				return Contract{}, fmt.Errorf("line %d: invalid contract declaration", lineNo)
			}
			contract.ContractID = tokens[1]
		case "authority":
			if len(tokens) != 2 {
				return Contract{}, fmt.Errorf("line %d: invalid authority declaration", lineNo)
			}
			contract.Authority = tokens[1]
		case "precedence":
			if len(tokens) != 4 {
				return Contract{}, fmt.Errorf("line %d: precedence must have three values", lineNo)
			}
			for _, value := range tokens[1:] {
				contract.Precedence = append(contract.Precedence, Decision(value))
			}
		case "unknown_fields":
			contract.UnknownFields = append(contract.UnknownFields, tokens[1:]...)
		case "denominator":
			values, err := pairs(tokens[1:])
			if err != nil {
				return Contract{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			if values["count"] == "" || values["id"] == "" {
				return Contract{}, fmt.Errorf("line %d: denominator needs id and count", lineNo)
			}
			count, err := strconv.Atoi(values["count"])
			if err != nil {
				return Contract{}, fmt.Errorf("line %d: invalid denominator count", lineNo)
			}
			contract.Denominator = count
		case "activity":
			values, err := pairs(tokens[1:])
			if err != nil {
				return Contract{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			activity := Activity{ID: values["id"], Name: values["name"], Proof: values["proof"], Artifact: values["artifact"], Authority: values["authority"]}
			if activity.ID == "" || activity.Name == "" || activity.Proof == "" || activity.Artifact == "" || activity.Authority == "" {
				return Contract{}, fmt.Errorf("line %d: incomplete activity", lineNo)
			}
			contract.Activities = append(contract.Activities, activity)
		case "scenario":
			values, err := pairs(tokens[1:])
			if err != nil {
				return Contract{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			ordinal, err := strconv.Atoi(values["ordinal"])
			if err != nil {
				return Contract{}, fmt.Errorf("line %d: invalid scenario ordinal", lineNo)
			}
			contract.Scenarios = append(contract.Scenarios, ScenarioSpec{Ordinal: ordinal, ID: values["id"], Expected: Decision(values["expected"]), UnknownClass: values["unknown_class"]})
		default:
			return Contract{}, fmt.Errorf("line %d: unsupported declaration %s", lineNo, tokens[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return Contract{}, err
	}
	if err := ValidateContract(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func ValidateContract(contract Contract) error {
	if contract.ContractID == "" || contract.Authority != "metacode" || contract.Denominator != 20 || len(contract.Scenarios) != 20 {
		return errors.New("contract must declare the fixed 20-scenario denominator")
	}
	if len(contract.Precedence) != 3 || contract.Precedence[0] != Refuted || contract.Precedence[1] != Unknown || contract.Precedence[2] != Closed {
		return errors.New("contract precedence must be REFUTED > UNKNOWN > CLOSED")
	}
	wantUnknownFields := []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}
	if !sameStrings(contract.UnknownFields, wantUnknownFields) {
		return errors.New("contract unknown fields are not the six required fields")
	}
	if len(contract.Activities) != len(RequiredActivities) {
		return fmt.Errorf("expected exactly %d activities, got %d", len(RequiredActivities), len(contract.Activities))
	}
	for i, activity := range contract.Activities {
		if activity.Name != RequiredActivities[i] || activity.Authority != "READ_ONLY" {
			return fmt.Errorf("activity %d must be %s with READ_ONLY authority", i+1, RequiredActivities[i])
		}
		for j := 0; j < i; j++ {
			if contract.Activities[j].Name == activity.Name {
				return fmt.Errorf("activity %s is bound more than once", activity.Name)
			}
		}
	}
	want := []struct {
		id       string
		decision Decision
		unknown  string
	}{
		{"draft-assets-publish-immutable", Closed, ""},
		{"deterministic-replay", Closed, ""},
		{"all-asset-digests-match", Closed, ""},
		{"exact-annotated-tag-target", Closed, ""},
		{"missing-operator-immutable-policy-receipt", Unknown, "DIRECT_MISSING"},
		{"stale-source-run", Unknown, "STALE"},
		{"missing-git-identity", Unknown, "DIRECT_MISSING"},
		{"tag-collision", Refuted, ""},
		{"publish-before-assets", Refuted, ""},
		{"published-immutable-false", Refuted, ""},
		{"checksum-path-mismatch", Refuted, ""},
		{"user-token-secret-or-admin-endpoint-in-actions", Refuted, ""},
		{"resume-existing-exact-draft-by-list-id", Closed, ""},
		{"existing-draft-target-or-assets-mismatch", Refuted, ""},
		{"upload-assets-via-release-upload-url", Closed, ""},
		{"upload-assets-via-api-endpoint", Refuted, ""},
		{"reconcile-symbolic-target-with-peeled-tag-target", Closed, ""},
		{"treat-symbolic-target-commitish-as-exact-commit", Refuted, ""},
		{"continue-with-create-response-draft-id", Closed, ""},
		{"require-immediate-draft-list-visibility-after-create", Refuted, ""},
	}
	for i, scenario := range contract.Scenarios {
		wantScenario := want[i]
		if scenario.Ordinal != i+1 || scenario.ID != wantScenario.id || scenario.Expected != wantScenario.decision || scenario.UnknownClass != wantScenario.unknown {
			return fmt.Errorf("scenario %d does not match the fixed denominator", i+1)
		}
	}
	return nil
}

func LoadJSONContract(path string) (Contract, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, nil, err
	}
	var contract Contract
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, nil, err
	}
	if contract.Schema != ContractSchema {
		return Contract{}, nil, errors.New("JSON contract has wrong schema")
	}
	if err := ValidateContract(contract); err != nil {
		return Contract{}, nil, err
	}
	return contract, raw, nil
}

func stripComment(line string) string {
	inQuote := false
	for i, r := range line {
		switch r {
		case '"':
			inQuote = !inQuote
		case '#':
			if !inQuote {
				return line[:i]
			}
		}
	}
	return line
}

func lex(line string) ([]string, error) {
	var tokens []string
	for i := 0; i < len(line); {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == '{' || line[i] == '}') {
			i++
		}
		if i == len(line) {
			break
		}
		if line[i] == '"' {
			start := i
			i++
			for i < len(line) {
				if line[i] == '\\' {
					i += 2
					continue
				}
				if line[i] == '"' {
					i++
					break
				}
				i++
			}
			if i > len(line) || line[i-1] != '"' {
				return nil, errors.New("unterminated quoted value")
			}
			value, err := strconv.Unquote(line[start:i])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, value)
			continue
		}
		start := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' && line[i] != '{' && line[i] != '}' {
			i++
		}
		tokens = append(tokens, line[start:i])
	}
	return tokens, nil
}

func pairs(tokens []string) (map[string]string, error) {
	if len(tokens)%2 != 0 {
		return nil, errors.New("key/value declaration has an odd number of tokens")
	}
	values := make(map[string]string, len(tokens)/2)
	for i := 0; i < len(tokens); i += 2 {
		if _, exists := values[tokens[i]]; exists {
			return nil, fmt.Errorf("duplicate key %s", tokens[i])
		}
		values[tokens[i]] = tokens[i+1]
	}
	return values, nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
