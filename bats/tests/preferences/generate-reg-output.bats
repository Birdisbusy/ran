load '../helpers/load'

# No rancher-desktop needed for this test
local_setup() {
    if is_windows; then
        # We need to use a directory that exists on the Win32 filesystem
        # so the ctrctl clients can correctly map the bind mounts.
        TEMP="$(win32env TEMP)"
    else
        TEMP=/tmp
    fi
}

verify_registry_output() {
    assert_success
    assert_output --partial '[HKEY_CURRENT_USER\SOFTWARE\Policies\Rancher Desktop\defaults\containerEngine]'
    assert_output --partial '[HKEY_CURRENT_USER\SOFTWARE\Policies\Rancher Desktop\defaults\containerEngine\allowedImages]'
    assert_output --partial '[HKEY_CURRENT_USER\SOFTWARE\Policies\Rancher Desktop\defaults\containerEngine\allowedImages]'
    assert_output --partial '"patterns"=hex(7):61,00,62,00,63,00,00,00,67,00,68,00,69,00,00,00,64,00,65,00,66,00,00,00,00,00'
    assert_output --partial '[HKEY_CURRENT_USER\SOFTWARE\Policies\Rancher Desktop\defaults\WSL\integrations]'
    assert_output --partial '"first"=dword:1'
    assert_output --partial '"second"=dword:0'
}

@test 'generates registry output from stdin' {
    run bash -c $'echo \'{"kubernetes": {"enabled": false}, "containerEngine": { "allowedImages": {"patterns": ["abc", "ghi", "def"] } }, "WSL": { "integrations": { "first": true, "second": false } } }\' | rdctl list-settings --output reg --reg-hive=hkcu --input - | cat -n'
    verify_registry_output
}

@test 'generates registry output from inline string' {
    run rdctl list-settings --output reg --reg-hive hkcu -b '{"kubernetes": {"enabled": false}, "containerEngine": { "allowedImages": {"patterns": ["abc", "ghi", "def"] } }, "WSL": { "integrations": { "first": true, "second": false } } }'
    verify_registry_output
}

@test 'generates registry output from a file' {
    local JSONFILE="$TEMP"/rdctl-reg-output.txt
    echo '{"kubernetes": {"enabled": false}, "containerEngine": { "allowedImages": {"patterns": ["abc", "ghi", "def"] } }, "WSL": { "integrations": { "first": true, "second": false } } }' >"$JSONFILE"
    run rdctl list-settings --output reg --reg-hive=hkcu --input "$JSONFILE"
    verify_registry_output
    rm -f "$JSONFILE"
}

@test 'complains about both --input and --body' {
    run rdctl list-settings --output reg --reg-hive=hkcu --input - --body blip
    assert_failure
    assert_output "Error: list-settings command: --body|-b and --input options cannot both be specified"
}

@test 'complains about both --input and -b' {
    run rdctl list-settings --output reg --reg-hive=hkcu --input - -b blip
    assert_failure
    assert_output "Error: list-settings command: --body|-b and --input options cannot both be specified"
}

@test 'complains about non-json body' {
    run rdctl list-settings --output reg,hkcu --body "ceci n'est pas quelque json"
    assert_failure
    assert_output "Error: error in json: invalid character 'c' looking for beginning of value"
}

@test 'complains about --body without reg' {
    run rdctl list-settings --output json --body "ceci n'est pas quelque json"
    assert_failure
    assert_output "Error: --input and --body|-b options are only valid when '--output reg' is also specified"
}

@test 'complains about --input without reg' {
    run rdctl list-settings --output json --input "ceci n'est pas quelque json"
    assert_failure
    assert_output "Error: --input and --body|-b options are only valid when '--output reg' is also specified"
}

@test 'complains about --input stdin without reg' {
    run rdctl list-settings --input -
    assert_failure
    assert_output "Error: --input and --body|-b options are only valid when '--output reg' is also specified"
}
