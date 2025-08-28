package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// version is set at build time
var version = "dev"

// SplunkLikeTimeSpec holds the specification for Splunk-like timestamp operations.
type SplunkLikeTimeSpec struct {
	BaseTime  time.Time
	Relative  string
	Snap      string
	Operation string
}

// ErrInvalidFormat is returned when the input format is invalid.
var ErrInvalidFormat = errors.New("invalid format: must be in the format of [+|-]<quantity><unit>@<unit> or @<unit>[+|-]<quantity><unit>")

// parseInput parses the input string and returns a SplunkLikeTimeSpec.
// If the input string is empty, it returns an empty operation without an error.
func parseInput(input string) (*SplunkLikeTimeSpec, error) {
	// Regular expression to parse relative time and snap operations.
	// Examples: -1d@d, @d-1d, +5h
	re := regexp.MustCompile(`^((?P<rel>[\+\-]\d+[smhdwMy])|(?P<snap>@[smhdwMy]))?((?P<op>[\+\-]\d+[smhdwMy])|(?P<snap2>@[smhdwMy]))?$`) // Added $ to match the whole string
	matches := re.FindStringSubmatch(input)
	if matches == nil || len(matches) < 6 {
		return nil, ErrInvalidFormat
	}

	spec := &SplunkLikeTimeSpec{
		Operation: input,
	}

	// Extract named groups
	relativeIndex := re.SubexpIndex("rel")
	snapIndex := re.SubexpIndex("snap")
	opIndex := re.SubexpIndex("op")
	snap2Index := re.SubexpIndex("snap2")

	if matches[relativeIndex] != "" {
		spec.Relative = matches[relativeIndex]
	}
	if matches[opIndex] != "" {
		spec.Relative = matches[opIndex]
	}

	if matches[snapIndex] != "" {
		spec.Snap = matches[snapIndex]
	}
	if matches[snap2Index] != "" {
		spec.Snap = matches[snap2Index]
	}

	// Return a valid struct even if no operation is specified
	return spec, nil
}

// applyOperation applies the specified operation to the given time.
func applyOperation(t time.Time, spec *SplunkLikeTimeSpec) (time.Time, error) {
	result := t

	// First, apply the snap operation
	if spec.Snap != "" {
		unit := spec.Snap[1:] // Remove '@'
		switch unit {
		case "s":
			result = result.Truncate(time.Second)
		case "m":
			result = result.Truncate(time.Minute)
		case "h":
			result = result.Truncate(time.Hour)
		case "d":
			result = time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, result.Location())
		case "w":
			// Snap to the beginning of the week (Sunday)
			daysToSubtract := int(result.Weekday())
			result = time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, result.Location()).AddDate(0, 0, -daysToSubtract)
		case "M":
			result = time.Date(result.Year(), result.Month(), 1, 0, 0, 0, 0, result.Location())
		case "y":
			result = time.Date(result.Year(), 1, 1, 0, 0, 0, 0, result.Location())
		default:
			return time.Time{}, fmt.Errorf("unknown snap unit: %s", unit)
		}
	}

	// Next, apply the relative time operation
	if spec.Relative != "" {
		sign := string(spec.Relative[0])
		valueStr := spec.Relative[1 : len(spec.Relative)-1]
		unit := string(spec.Relative[len(spec.Relative)-1])

		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid quantity: %s", valueStr)
		}

		if sign == "-" {
			value = -value
		}

		switch unit {
		case "s":
			result = result.Add(time.Duration(value) * time.Second)
		case "m":
			result = result.Add(time.Duration(value) * time.Minute)
		case "h":
			result = result.Add(time.Duration(value) * time.Hour)
		case "d":
			result = result.AddDate(0, 0, value)
		case "w":
			result = result.AddDate(0, 0, value*7)
		case "M":
			result = result.AddDate(0, value, 0)
		case "y":
			result = result.AddDate(value, 0, 0)
		default:
			return time.Time{}, fmt.Errorf("unknown relative unit: %s", unit)
		}
	}

	return result, nil
}

// convertFormat converts a user-friendly format string to a Go time layout string.
func convertFormat(userFormat string) string {
    replacementsList := []struct{ from, to string }{
        {"YYYY", "2006"},
        {"YY", "06"},
        {"MM", "01"},
        {"M", "1"},
        {"DD", "02"},
        {"D", "2"},
        {"hh", "15"},
        {"mm", "04"},
        {"ss", "05"},
		{"SSS", "000"},
		{"UUU", "000000"},
        {"a", "pm"},
		{"TZ", "MST"},
		{"ZZZ", "-0700"},
		{"ZZ", "-07:00"},
    }
    
    // Replace the longest patterns first to avoid partial replacements (e.g., YYYY before YY)
    for _, r := range replacementsList {
        userFormat = strings.ReplaceAll(userFormat, r.from, r.to)
    }
    
    return userFormat
}

// helpMessage returns the detailed help string.
func helpMessage() string {
	return "Usage: sdate [--op <operation>] [--base <time>] [--format <layout>] [--output-tz <timezone>]\n\n" +
		"This tool generates a timestamp based on a Splunk-like relative time and snap operation.\n\n" +
		"Options:\n" +
		"\t--op <operation>\n" +
		"\t\tA string specifying the operation to perform. It can be a relative time, a snap, or a combination.\n" +
		"\t\tThis argument is optional. If not specified, the current time is used without any operation.\n" +
		"\t\tExamples:\n" +
		"\t\t- '-1d@d': 1 day ago, snapped to the beginning of the day.\n" +
		"\t\t- '@h': snapped to the beginning of the hour.\n" +
		"\t\t- '+2h': 2 hours from now.\n\n" +
		"Supported Units:\n" +
		"\ts: seconds\n" +
		"\tm: minutes\n" +
		"\th: hours\n" +
		"\td: days\n" +
		"\tw: weeks (Sunday is the start of the week)\n" +
		"\tM: months\n" +
		"\ty: years\n\n" +
		"\t--base <time>\n" +
		"\t\tThe base time for the calculation. If not specified, the current time is used.\n" +
		"\t\tSupported formats:\n" +
		"\t\t- RFC3339: '2023-10-27T10:00:00Z'\n" +
		"\t\t- Simple Date: '2023-10-27'\n" +
		"\t\t- Unix Time: '1698372000'\n" +
		"\t\t- TZ-aware: 'TZ=Asia/Tokyo 2023-10-27T10:00:00'\n\n" +
		"\t--format <layout>\n" +
		"\t\tThe output format for the final timestamp. The default is RFC3339.\n" +
		"\t\tYou can use Go's time layout string (e.g., '2006-01-02 15:04:05')\n" +
		"\t\tor a more intuitive format like 'YYYY/MM/DD hh:mm:ss'.\n" +
		"\t\tYou can also specify 'unix' or 'epoch' to output as a Unix timestamp.\n\n" +
		"\t--output-tz <timezone>\n" +
		"\t\tSpecifies the timezone for the output timestamp.\n" +
		"\t\tExample: 'Asia/Tokyo', 'America/New_York'.\n\n" +
		"Supported Format Metacharacters:\n" +
		"\tYYYY: 4-digit year\n" +
		"\tYY:   2-digit year\n" +
		"\tMM:   2-digit month\n" +
		"\tM:    1-digit month\n" +
		"\tDD:   2-digit day\n" +
		"\tD:    1-digit day\n" +
		"\thh:   24-hour\n" +
		"\tmm:   minute\n" +
		"\tss:   second\n" +
		"\tSSS:  millisecond\n" +
		"\tUUU:  microsecond\n" +
		"\ta:    AM/PM\n" +
		"\tTZ:   Timezone abbreviation (e.g., JST)\n" +
		"\tZZ:   Timezone offset with colon (e.g., +09:00)\n" +
		"\tZZZ:  Timezone offset without colon (e.g., +0900)\n\n" +
		"Examples:\n" +
		"  # Output current time in a more intuitive format\n" +
		"  ./sdate --format 'YYYY/MM/DD hh:mm:ss'\n" +
		"  # Output current time with milliseconds\n" +
		"  ./sdate --format 'YYYY-MM-DD hh:mm:ss.SSS'\n" +
		"  # Calculate 2 hours after a specified time in a specific timezone, then output in another timezone\n" +
		"  ./sdate --op +2h --base 'TZ=America/New_York 2023-10-27T10:00:00' --output-tz Asia/Tokyo --format 'YYYY-MM-DD hh:mm:ss ZZ'\n" +
		"  # Output the current time with timezone abbreviation\n" +
		"  ./sdate --format 'YYYY-MM-DD hh:mm:ss TZ'\n" +
		"  # Output the current time with timezone offset (with colon)\n" +
		"  ./sdate --format 'YYYY-MM-DD hh:mm:ss ZZ'\n" +
		"  # Output the current time with timezone offset (without colon)\n" +
		"  ./sdate --format 'YYYY-MM-DD hh:mm:ss ZZZ'"
}

func main() {
	// Add --op, --format, --base, and --output-tz options
	operation := flag.String("op", "", "The operation to perform (e.g., '-1d@d').")
	outputFormat := flag.String("format", time.RFC3339, "Output timestamp format (Go time layout string)")
	baseTimeStr := flag.String("base", "", "Base time for calculation (e.g., '2023-10-27T10:00:00Z')")
	outputTZ := flag.String("output-tz", "", "Timezone for the output (e.g., 'Asia/Tokyo')")
	showHelp := flag.Bool("help", false, "Show detailed help message")
	showVersion := flag.Bool("version", false, "Show version information")

	// Parse flags
	flag.Parse()

	// If --version flag is specified, show the version and exit
	if *showVersion {
		fmt.Printf("sdate version %s\n", version)
		os.Exit(0)
	}

	// If --help flag is specified, show the help message and exit
	if *showHelp {
		fmt.Println(helpMessage())
		os.Exit(0)
	}

	// Get the operation argument (a positional argument)
	args := flag.Args()

	// Make the operation optional
	var op string
	if len(args) > 0 {
		op = args[0]
	}

	// If the --op flag is set, it takes precedence over the positional argument.
	if *operation != "" {
		op = *operation
	}

	// Use --base option for the base time, otherwise use the current time.
	baseTime := time.Now()

	if *baseTimeStr != "" {
		// Handle TZ=... format
		tzRegex := regexp.MustCompile(`^TZ=(?P<tz>[\w\/]+)\s+(?P<time>.+)$`)
		matches := tzRegex.FindStringSubmatch(*baseTimeStr)

		if matches != nil {
			tzName := matches[tzRegex.SubexpIndex("tz")]
			timeStr := matches[tzRegex.SubexpIndex("time")]

			loc, err := time.LoadLocation(tzName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid timezone name: %s\n", tzName)
				os.Exit(1)
			}

			baseTime, err = time.ParseInLocation("2006-01-02T15:04:05", timeStr, loc)
			if err != nil {
				baseTime, err = time.ParseInLocation("2006-01-02", timeStr, loc)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Invalid base time format. Use RFC3339 (e.g., 2023-10-27T10:00:00), YYYY-MM-DD, or Unix time (e.g., 1698372000)\n")
					os.Exit(1)
				}
			}
		} else {
			// Try to parse as a Unix timestamp first
			if i, err := strconv.ParseInt(*baseTimeStr, 10, 64); err == nil {
				baseTime = time.Unix(i, 0)
			} else {
				// If it's not a Unix timestamp, try RFC3339 or a simple date format
				var err error
				baseTime, err = time.Parse(time.RFC3339, *baseTimeStr)
				if err != nil {
					// Also try simple date format
					baseTime, err = time.Parse("2006-01-02", *baseTimeStr)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: Invalid base time format. Use RFC3339 (e.g., 2023-10-27T10:00:00Z), YYYY-MM-DD, or Unix time (e.g., 1698372000)\n")
						os.Exit(1)
					}
				}
			}
		}
	}

	// If no operation is specified, do not perform any calculation.
	if op == "" {
		// Output the base time directly with the specified format
		calculatedTime := baseTime

		if *outputTZ != "" {
			loc, err := time.LoadLocation(*outputTZ)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid output timezone name: %s\n", *outputTZ)
				os.Exit(1)
			}
			calculatedTime = calculatedTime.In(loc)
		}

		finalFormat := *outputFormat
		if strings.ToLower(finalFormat) != "unix" && strings.ToLower(finalFormat) != "epoch" {
			finalFormat = convertFormat(finalFormat)
		}

		if strings.ToLower(*outputFormat) == "unix" || strings.ToLower(*outputFormat) == "epoch" {
			fmt.Println(calculatedTime.Unix())
		} else {
			fmt.Println(calculatedTime.Format(finalFormat))
		}
		return
	}

	// Parse the input
	spec, err := parseInput(op)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Apply the operation and calculate the final time
	calculatedTime, err := applyOperation(baseTime, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error applying operation: %v\n", err)
		os.Exit(1)
	}

	// Convert to output timezone if specified
	if *outputTZ != "" {
		loc, err := time.LoadLocation(*outputTZ)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid output timezone name: %s\n", *outputTZ)
			os.Exit(1)
		}
		calculatedTime = calculatedTime.In(loc)
	}

	// Convert user-friendly format to Go time layout
	finalFormat := *outputFormat
	if strings.ToLower(finalFormat) != "unix" && strings.ToLower(finalFormat) != "epoch" {
		finalFormat = convertFormat(finalFormat)
	}

	// Print the result based on the format
	if strings.ToLower(*outputFormat) == "unix" || strings.ToLower(*outputFormat) == "epoch" {
		fmt.Println(calculatedTime.Unix())
	} else {
		fmt.Println(calculatedTime.Format(finalFormat))
	}
}