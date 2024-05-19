# Sophrosyne documentation

This is the user facing documentation for sophrosyne. For developer facing documentation, please
see the package documentation found at [pkg.go.dev](https://pkg.go.dev/github.com/madsrc/sophrosyne).

## Configuration

Loading binary data into the configuration

To load binary data into the configuration, there is 3 options:
- Load the binary data from a secrets file
- Load the binary data from a YAML file. This requires the binary data to be
    base64 encoded (without newlines and spaces) with a prefix of `!!binary `. For example:
    ```yaml
    binary_data: !!binary Ag4k628yFI2h+SWoypEO7OYzrFqBrEz8az9c6Du7ons=
    ```
- Load the binary data from an environment variable. This requires the binary
    data to be hex encoded with the prefix `0x`. If the string fails decoding
    from hex to binary, the configuration will treat it as a raw string.
