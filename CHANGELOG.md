# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [1.1.0] - 2025-08-28

### Added
- Change the start of the week to Sunday for `@w` snap operation.

### Fixed
- Fix a bug in parsing input strings where invalid inputs were not correctly identified.

## [1.0.0] - 2025-08-28

### Added
- Initial release of `sdate`.
- Command-line tool for Splunk-like time manipulation.
- Support for relative time and snap operations.
- Flexible output formatting, including UNIX timestamp.
- Timezone handling for base time and output.
- Unit tests for core logic.

### Changed
- Improved `Makefile` to provide comprehensive build, test, and packaging targets.
- `make package` now builds and archives for Linux, Windows, and macOS (Universal Binary).