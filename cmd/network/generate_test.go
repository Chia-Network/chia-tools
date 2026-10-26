package network_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/chia-network/chia-tools/cmd"
	"gopkg.in/yaml.v3"
)

func TestNetworkGenerateHelperProcess(t *testing.T) {
	if os.Getenv("CHIA_TOOLS_GENERATE_TEST") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			cmd.RootCmd.SetArgs(os.Args[i+1:])
			cmd.Execute()
			os.Exit(0)
		}
	}
	t.Fatal("missing command arguments")
}

func TestGenerateMaxPlotStrength(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		env     string
		want    *int
		invalid bool
	}{
		{name: "omitted JSON", args: []string{"--as-json"}},
		{name: "omitted YAML"},
		{name: "lower strength JSON", args: []string{"--as-json", "--max-plot-strength", "3"}, want: intPtr(3)},
		{name: "minimum YAML", args: []string{"--max-plot-strength", "2"}, want: intPtr(2)},
		{name: "environment", args: []string{"--as-json"}, env: "5", want: intPtr(5)},
		{name: "maximum", args: []string{"--as-json", "--max-plot-strength", "255"}, want: intPtr(255)},
		{name: "below minimum", args: []string{"--max-plot-strength", "1"}, invalid: true},
		{name: "zero", args: []string{"--max-plot-strength", "0"}, invalid: true},
		{name: "overflow", args: []string{"--max-plot-strength", "256"}, invalid: true},
		{name: "environment overflow", env: "256", invalid: true},
		{name: "negative", args: []string{"--max-plot-strength", "-1"}, invalid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"-test.run=^TestNetworkGenerateHelperProcess$", "--", "network", "generate", "--network", "strength-test"}, tc.args...)
			process := exec.Command(os.Args[0], args...)
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "CHIA_TOOLS_TN_GEN_MAX_PLOT_STRENGTH=") {
					process.Env = append(process.Env, entry)
				}
			}
			process.Env = append(process.Env, "CHIA_TOOLS_GENERATE_TEST=1")
			if tc.env != "" {
				process.Env = append(process.Env, "CHIA_TOOLS_TN_GEN_MAX_PLOT_STRENGTH="+tc.env)
			}
			output, err := process.Output()
			if tc.invalid {
				if err == nil {
					t.Fatal("invalid strength was accepted")
				}
				return
			}
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			var constants map[string]any
			if len(output) > 0 && output[0] == '{' {
				err = json.Unmarshal(output, &constants)
			} else {
				err = yaml.Unmarshal(output, &constants)
			}
			if err != nil {
				t.Fatalf("decode constants: %v", err)
			}
			value, present := constants["MAX_PLOT_STRENGTH"]
			if tc.want == nil {
				if present {
					t.Fatalf("unset strength must be omitted, got %v", value)
				}
			} else if !present || (value != *tc.want && value != float64(*tc.want)) {
				t.Fatalf("MAX_PLOT_STRENGTH = %v, want %d", value, *tc.want)
			}
		})
	}
}

func intPtr(value int) *int { return &value }
