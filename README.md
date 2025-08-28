# sdate

`sdate` is a command-line tool that generates a timestamp using a Splunk-like relative time and snap operation.

## Features

-   Generates a timestamp from the current time or a specified base time.
-   Supports Splunk-like relative time and snap operations, such as `-1d@d`.
-   Supports output in Go language's time layout string format or UNIXTIME (Epoch Time).
-   Provides a detailed help message via the `--help` option.
-   Allows specifying the output timezone.
-   Shows version information via the `--version` option.

## Supported Time Units

-   `s`: seconds
-   `m`: minutes
-   `h`: hours
-   `d`: days
-   `w`: weeks (Monday is assumed to be the start of the week)
-   `M`: months
-   `y`: years

## Installation

With Go and Make installed, build the tool by running the following command.

```sh
make build
```

This will create the `sdate` executable in the current directory.

For cross-compilation or creating a macOS universal binary, use the following commands:

```sh
# Build for Linux, Windows, and macOS (Intel)
make release

# Create a macOS universal binary (Apple Silicon + Intel)
make universal-mac
```

## Usage

The basic usage is as follows:

```sh
./sdate [operation] [--base <time>] [--format <layout>] [--output-tz <timezone>]
```

### Arguments and Options

-   **[operation]** (optional): A string specifying the operation to perform. It can be a relative time, a snap, or a combination.
    -   Examples: `-1d@d` (beginning of the day, one day ago), `@h` (beginning of the current hour), `+2h` (2 hours from now)

-   **--base <time>** (optional): The base time for the calculation. If not specified, the current time is used.
    -   Supported formats:
        -   RFC3339: `2023-10-27T10:00:00Z`
        -   Simple Date: `2023-10-27`
        -   UNIXTIME: `1698372000`
        -   TZ-aware: `TZ=America/New_York 2023-10-27T10:00:00`

-   **--format <layout>** (optional): The output format for the final timestamp. The default is RFC3339.
    -   You can use Go's `time` layout string (e.g., '2006-01-02 15:04:05').
    -   You can also specify a more intuitive format like 'YYYY/MM/DD hh:mm:ss'.
    -   You can also specify 'unix' or 'epoch' to output as a Unix timestamp.

-   **--output-tz <timezone>** (optional): Specifies the timezone for the output timestamp.
    -   Example: 'Asia/Tokyo', 'America/New_York'.

-   **--version**: Show version information.

-   **--help**: Show detailed help message.

### Supported Format Metacharacters
-   `YYYY`: 4-digit year
-   `YY`:   2-digit year
-   `MM`:   2-digit month
-   `M`:    1-digit month
-   `DD`:   2-digit day
-   `D`:    1-digit day
-   `hh`:   24-hour
-   `mm`:   minute
-   `ss`:   second
-   `SSS`:  millisecond
-   `UUU`:  microsecond
-   `a`:    AM/PM
-   `TZ`:   Timezone abbreviation (e.g., JST)
-   `ZZ`:   Timezone offset with colon (e.g., +09:00)
-   `ZZZ`:  Timezone offset without colon (e.g., +0900)

## Examples

```sh
# Show version
./sdate --version

# Output current time in a more intuitive format
./sdate --format 'YYYY/MM/DD hh:mm:ss'

# Output current time with milliseconds
./sdate --format 'YYYY-MM-DD hh:mm:ss.SSS'

# Calculate 2 hours after a specified time in a specific timezone, then output in another timezone
./sdate +2h --base 'TZ=America/New_York 2023-10-27T10:00:00' --output-tz Asia/Tokyo --format 'YYYY-MM-DD hh:mm:ss ZZ'

# Output the current time with timezone abbreviation
./sdate --format 'YYYY-MM-DD hh:mm:ss TZ'

# Output the current time with timezone offset (with colon)
./sdate --format 'YYYY-MM-DD hh:mm:ss ZZ'

# Output the current time with timezone offset (without colon)
./sdate --format 'YYYY-MM-DD hh:mm:ss ZZZ'
```

## Development

### Testing

To run the tests, use the following command:

```sh
make test
```