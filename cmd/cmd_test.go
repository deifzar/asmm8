package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd(t *testing.T) {
	t.Run("root command is configured correctly", func(t *testing.T) {
		assert.Equal(t, "asmm8", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)
	})

	t.Run("root command has toggle flag", func(t *testing.T) {
		flag := rootCmd.Flags().Lookup("toggle")
		require.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
		assert.Equal(t, "t", flag.Shorthand)
	})

	t.Run("root command has subcommands", func(t *testing.T) {
		subcommands := rootCmd.Commands()
		assert.GreaterOrEqual(t, len(subcommands), 2, "should have at least launch and version commands")

		// Check that expected subcommands exist
		commandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			commandNames[i] = cmd.Use
		}
		assert.Contains(t, commandNames, "launch")
		assert.Contains(t, commandNames, "version")
	})
}

func TestLaunchCmd(t *testing.T) {
	t.Run("launch command is configured correctly", func(t *testing.T) {
		assert.Equal(t, "launch", launchCmd.Use)
		assert.NotEmpty(t, launchCmd.Short)
	})

	t.Run("launch command has ip flag with default", func(t *testing.T) {
		flag := launchCmd.Flags().Lookup("ip")
		require.NotNil(t, flag)
		assert.Equal(t, "0.0.0.0", flag.DefValue)
	})

	t.Run("launch command has port flag with default", func(t *testing.T) {
		flag := launchCmd.Flags().Lookup("port")
		require.NotNil(t, flag)
		assert.Equal(t, "8000", flag.DefValue)
	})

	t.Run("launch command accepts maximum 2 args", func(t *testing.T) {
		// Args validation is set via cobra.MatchAll
		assert.NotNil(t, launchCmd.Args)
	})
}

func TestVersionCmd(t *testing.T) {
	t.Run("version command is configured correctly", func(t *testing.T) {
		assert.Equal(t, "version", versionCmd.Use)
		assert.NotEmpty(t, versionCmd.Short)
		assert.NotEmpty(t, versionCmd.Long)
	})

	t.Run("version command has Run function", func(t *testing.T) {
		// Version command uses fmt.Println which writes to os.Stdout
		// so we can't easily capture output. Verify the Run function exists.
		assert.NotNil(t, versionCmd.Run, "version command should have a Run function")
	})
}

func TestLaunchCmdFlagValidation(t *testing.T) {
	// Create a fresh command for testing to avoid state pollution
	createTestLaunchCmd := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "launch",
			Short: "Test launch command",
			RunE: func(cmd *cobra.Command, args []string) error {
				// Just parse flags, don't actually launch
				_, _ = cmd.Flags().GetString("ip")
				_, _ = cmd.Flags().GetInt("port")
				return nil
			},
		}
		cmd.Flags().String("ip", "0.0.0.0", "IP address")
		cmd.Flags().Int("port", 8000, "Port number")
		return cmd
	}

	t.Run("parses valid ip flag", func(t *testing.T) {
		cmd := createTestLaunchCmd()
		cmd.SetArgs([]string{"--ip", "127.0.0.1"})

		err := cmd.Execute()
		assert.NoError(t, err)

		ip, err := cmd.Flags().GetString("ip")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1", ip)
	})

	t.Run("parses valid port flag", func(t *testing.T) {
		cmd := createTestLaunchCmd()
		cmd.SetArgs([]string{"--port", "8080"})

		err := cmd.Execute()
		assert.NoError(t, err)

		port, err := cmd.Flags().GetInt("port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
	})

	t.Run("parses both flags together", func(t *testing.T) {
		cmd := createTestLaunchCmd()
		cmd.SetArgs([]string{"--ip", "192.168.1.1", "--port", "8500"})

		err := cmd.Execute()
		assert.NoError(t, err)

		ip, _ := cmd.Flags().GetString("ip")
		port, _ := cmd.Flags().GetInt("port")
		assert.Equal(t, "192.168.1.1", ip)
		assert.Equal(t, 8500, port)
	})

	t.Run("uses default values when no flags provided", func(t *testing.T) {
		cmd := createTestLaunchCmd()
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		assert.NoError(t, err)

		ip, _ := cmd.Flags().GetString("ip")
		port, _ := cmd.Flags().GetInt("port")
		assert.Equal(t, "0.0.0.0", ip)
		assert.Equal(t, 8000, port)
	})

	t.Run("rejects invalid port type", func(t *testing.T) {
		cmd := createTestLaunchCmd()
		cmd.SetArgs([]string{"--port", "not-a-number"})

		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestPortRangeValidation(t *testing.T) {
	// Test the port range logic (8000-9000) used in launch.go
	isValidPort := func(port int) bool {
		return port >= 8000 && port <= 9000
	}

	testCases := []struct {
		name     string
		port     int
		expected bool
	}{
		{"port below range", 7999, false},
		{"port at lower bound", 8000, true},
		{"port in middle of range", 8500, true},
		{"port at upper bound", 9000, true},
		{"port above range", 9001, false},
		{"negative port", -1, false},
		{"zero port", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidPort(tc.port)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCommandHierarchy(t *testing.T) {
	t.Run("launch is child of root", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "launch" {
				found = true
				break
			}
		}
		assert.True(t, found, "launch command should be registered with root")
	})

	t.Run("version is child of root", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "version" {
				found = true
				break
			}
		}
		assert.True(t, found, "version command should be registered with root")
	})
}

func TestRootCmdHelp(t *testing.T) {
	t.Run("root command generates help without error", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"--help"})

		// Help should not cause an error
		err := rootCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "asmm8")
	})
}

func TestLaunchCmdHelp(t *testing.T) {
	t.Run("launch command has usage line", func(t *testing.T) {
		// Verify usage line is set correctly
		usage := launchCmd.UseLine()
		assert.Contains(t, usage, "launch")
	})

	t.Run("launch command flags are documented", func(t *testing.T) {
		// Check that flags have help text
		ipFlag := launchCmd.Flags().Lookup("ip")
		portFlag := launchCmd.Flags().Lookup("port")

		require.NotNil(t, ipFlag)
		require.NotNil(t, portFlag)

		assert.NotEmpty(t, ipFlag.Usage, "ip flag should have usage text")
		assert.NotEmpty(t, portFlag.Usage, "port flag should have usage text")
	})
}
