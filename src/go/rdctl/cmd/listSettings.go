/*
Copyright © 2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/reg"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var outputSettingsFlags struct {
	Format              string
	RegistryHive        string
	RegistryProfileType string
}
var InputFile string
var JSONBody string

const jsonFormat = "json"
const regFormat = "reg"
const defaultsRegistrySection = "defaults"
const lockedRegistrySection = "locked"

// listSettingsCmd represents the listSettings command
var listSettingsCmd = &cobra.Command{
	Use:   "list-settings",
	Short: "Lists the current settings.",
	Long: `Lists the current settings in JSON or Windows registry-file format.
The default output format is JSON.

To convert the current settings into a registry file, run the following command:

rdctl list-commands --output reg --reg-hive=X --profile=Y

where X is either "hkcu" or "hklm", depending on whether you want to update HKEY_LOCAL_MACHINE
or HKEY_CURRENT_USER respectively (default: "hklm"),
and Y is either "defaults" or "locked", depending on which deployment profile you want to populate (default: "defaults").

You can convert any arbitrary JSON document to reg file format by specifying either
'--body|-b STRING_CONTAINING_JSON' or
'--input FILE_CONTAINING_JSON' to convert the supplied JSON text, rather than getting the current
settings from the server. Any fields in the JSON that aren't recognized as Rancher Desktop settings
will be ignore.
You can also specify '--input -' to pipe JSON in from standard input.

So something like

	echo '{"kubernetes":{"enabled": false}}' | rdctl list-settings --output=reg --reg-hive=hkcu --input - > k8sOff.reg

can be used to produce a registry file that will set 'kubernetes.enabled' to false in the deployment file.
Then run 'reg import k8sOff.reg' to add that hierarchy to the Windows registry.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		cmd.SilenceUsage = true
		result, err := getListSettings()
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

// the reg file format is directly usable only on Windows,
// but it can be created on any platform for purposes of testing or development

func init() {
	rootCmd.AddCommand(listSettingsCmd)
	listSettingsCmd.Flags().StringVarP(&outputSettingsFlags.Format, "output", "", jsonFormat, fmt.Sprintf("output format: %s|%s", jsonFormat, regFormat))
	listSettingsCmd.Flags().StringVarP(&outputSettingsFlags.RegistryHive, "reg-hive", "", "", fmt.Sprintf(`registry hive: %s|%s (default "%s")`, reg.HklmRegistryHive, reg.HkcuRegistryHive, reg.HklmRegistryHive))
	listSettingsCmd.Flags().StringVarP(&outputSettingsFlags.RegistryProfileType, "section", "", "", fmt.Sprintf(`registry section: %s|%s (default "%s")`, defaultsRegistrySection, lockedRegistrySection, defaultsRegistrySection))
}

func validateOutputFormatFlags() error {
	if outputSettingsFlags.Format != jsonFormat && outputSettingsFlags.Format != regFormat {
		return fmt.Errorf(`invalid output format of "%s"`, outputSettingsFlags.Format)
	}
	if outputSettingsFlags.Format == jsonFormat {
		if outputSettingsFlags.RegistryHive != "" || outputSettingsFlags.RegistryProfileType != "" {
			return fmt.Errorf("registry hive and profile can't be specified with json")
		}
		return nil
	}
	switch strings.ToLower(outputSettingsFlags.RegistryHive) {
	case reg.HklmRegistryHive, reg.HkcuRegistryHive:
		outputSettingsFlags.RegistryHive = strings.ToLower(outputSettingsFlags.RegistryHive)
	case "":
		outputSettingsFlags.RegistryHive = reg.HklmRegistryHive
	default:
		return fmt.Errorf("invalid registry hive of '%s' specified", outputSettingsFlags.RegistryHive)
	}
	switch strings.ToLower(outputSettingsFlags.RegistryProfileType) {
	case defaultsRegistrySection, lockedRegistrySection:
		outputSettingsFlags.RegistryProfileType = strings.ToLower(outputSettingsFlags.RegistryProfileType)
	case "":
		outputSettingsFlags.RegistryProfileType = defaultsRegistrySection
	default:
		return fmt.Errorf("invalid registry section of '%s' specified", outputSettingsFlags.RegistryProfileType)
	}
	return nil
}

func getListSettings() (string, error) {
	err := validateOutputFormatFlags()
	if err != nil {
		return "", err
	}
	var output []byte

	if InputFile != "" && JSONBody != "" {
		return nil, fmt.Errorf("list-settings command: --body|-b and --input options cannot both be specified")
	}
	if InputFile == "" && JSONBody == "" {
		output, err = processRequestForUtility(doRequest("GET", versionCommand("", "settings")))
	} else {
		if outputSettingsFlags.Format != "reg" {
			return nil, fmt.Errorf("--input and --body|-b options are only valid when '--output reg' is also specified")
		}
		if JSONBody != "" {
			output = []byte(JSONBody)
		} else if InputFile == "-" {
			output, err = ioutil.ReadAll(os.Stdin)
		} else {
			output, err = ioutil.ReadFile(InputFile)
		}
	}
	if err != nil {
		return "", err
	} else if outputSettingsFlags.Format == jsonFormat {
		return string(output), nil
	} else if outputSettingsFlags.Format == regFormat {
		lines, err := reg.JsonToReg(outputSettingsFlags.RegistryHive, outputSettingsFlags.RegistryProfileType, string(output))
		if err != nil {
			return "", err
		}
		return strings.Join(lines, "\n"), nil
	} else {
		// This shouldn't happen
		return "", fmt.Errorf("internal error: unexpected output format of %s", outputSettingsFlags.Format)
	}
}
