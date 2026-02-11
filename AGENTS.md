# AGENTS.md - Developer Guide for izapple2

This guide provides essential information for AI coding agents working on the izapple2 Apple II emulator project.

## Project Overview

izapple2 is an Apple II+/IIe emulator written in Go. The project emulates various Apple II models with support for disk drives, expansion cards, and multiple display modes.

- **Language**: Go 1.26.0+
- **Main Package**: `github.com/ivanizag/izapple2`
- **Architecture**: Modular card-based system with memory management, CPU emulation, and video rendering

## Build Commands

### Build the Project
```bash
# Build the main package (library only)
go build .

# Build SDL2 frontend
cd frontend/a2sdl
go build .

# Build WASM frontend
cd frontend/a2wasm
go build .

# Build console frontend
cd frontend/a2console
go build .
```

### Install Dependencies
```bash
# Install Go dependencies
go get -v -t -d ./...

# On Linux, install SDL2 development files
sudo apt-get install libsdl2-dev

# On macOS
brew install SDL2
```

## Testing

### Run All Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Run a Single Test
```bash
# Run a specific test by name
go test -v -run TestName

# Run a specific test in a specific package
go test -v -run TestName ./storage

# Examples:
go test -v -run TestPlusBoots
go test -v -run TestNibBackAndForth ./storage
go test -v -run TestConfigurationModel
```

### Run Tests in a Specific Package
```bash
# Test only the storage package
go test ./storage

# Test with race detector
go test -race ./...
```

### End-to-End Tests
The project includes E2E boot tests (`e2e_boot_test.go`, `e2e_woz_test.go`) that boot various disk images and verify the emulator behavior. These tests use cycle counts and text matching to validate proper operation.

## Code Style Guidelines

### Package Organization
- Main emulator code is in the root package `izapple2`
- Subpackages include: `storage`, `screen`, `fujinet`, `component`
- Frontend implementations are in `frontend/` directory

### Naming Conventions

#### Types and Structs
- Use CamelCase for exported types: `Apple2`, `CardDisk2`, `VideoSource`
- Use camelCase for unexported types: `memoryManager`, `cardBase`, `trackTracer`
- Card implementations follow pattern: `Card<CardName>` (e.g., `CardDisk2`, `CardSmartPort`)
- Card builders follow pattern: `cardBuilder` for the struct

#### Functions
- Exported functions: CamelCase starting with uppercase (e.g., `NewKeyboardChannel`, `LoadResource`)
- Unexported functions: camelCase starting with lowercase (e.g., `configure`, `setupCard`)
- Constructor functions: `New<Type>` (e.g., `NewBlockDiskFile`, `NewSmartPortFujinetNetwork`)
- Factory functions for unexported types: `new<Type>` (e.g., `newVideo`, `newTraceMonitor`)
- Boolean query methods: start with `Is` or `Has` (e.g., `IsPaused`, `isFileWoz`)

#### Variables and Constants
- Exported constants: CamelCase (e.g., `MemoryTypeROM`, `DiskII`)
- Unexported constants: camelCase (e.g., `noCardName`, `wozMaxTrack`)
- Constants for addresses often use hex: `addressLimitZero`, `ioC8Off`

### Imports
- Standard library imports first
- Third-party imports second
- Local package imports last
- Separate groups with blank lines

Example:
```go
import (
    "fmt"
    "strings"

    "github.com/ivanizag/iz6502"
    "golang.org/x/exp/maps"

    "github.com/ivanizag/izapple2/screen"
    "github.com/ivanizag/izapple2/storage"
)
```

### Types and Interfaces
- Use embedded structs for common functionality (e.g., `cardBase` embedded in card implementations)
- Keep interfaces small and focused
- Define interfaces in the package that uses them, not implements them

### Error Handling
- Return errors rather than panicking in most cases
- Use `fmt.Errorf()` for error wrapping with context
- Check errors immediately after function calls
- Test functions should use `t.Fatal(err)` for setup errors, `t.Error(err)` for assertion failures

Example:
```go
if err != nil {
    return fmt.Errorf("failed to load ROM: %v", err)
}
```

### Comments and Documentation
- All exported functions, types, and methods must have doc comments
- Doc comments start with the name being documented
- Use full sentences with proper punctuation
- Include references to specifications or documentation when relevant

Example:
```go
// NewCardDisk2 creates a DiskII controller card
// sectors13 enables support for 13-sector disks (DOS 3.2)
func NewCardDisk2(sectors13 bool) *CardDisk2 {
```

### Testing Conventions
- Test files end with `_test.go`
- Test function names: `Test<FunctionName>` or `Test<Feature>`
- Use table-driven tests when testing multiple scenarios
- Helper functions for tests: `test<Purpose>` (e.g., `testBoots`)
- Use underscores for large numeric literals in tests: `200_000`, `100_000_000`

### Code Formatting
- Use `gofmt` (automatic with most Go tooling)
- Line length: no strict limit, but keep it reasonable (~100-120 chars)
- Use tabs for indentation (Go standard)

### Memory and Hardware Emulation
- Memory addresses use `uint16`
- Use hex notation for hardware addresses: `0xc000`, `0xbfff`
- Bit manipulation is common; comment complex operations
- Cycle counting is important for timing-sensitive operations

### Configuration System
- Configuration uses string-based parameter system
- Card parameters follow format: `cardname,param1=value1,param2=value2`
- Use `configuration` struct for managing settings
- Support both command-line and config file configuration

## Common Patterns

### Card Implementation
1. Define card struct embedding `cardBase`
2. Implement `assign()` method to set up ROM and memory handlers
3. Use builder pattern with `cardBuilder` struct
4. Register in `cardFactory` map

### Resource Loading
- Use `LoadResource()` for files that can be internal, local files, or URLs
- Internal resources use `<internal>/` prefix
- Support gzip and zip compression transparently

### Tracing and Debugging
- Use `executionTracer` interface for CPU tracing
- Implement tracers for different OSes (ProDOS, Pascal, CP/M)
- Use `trace` parameter to enable debug output

## Continuous Integration

The project uses GitHub Actions (`.github/workflows/go.yml`) and CircleCI (`.circleci/config.yml`):
- Builds are tested on Linux with SDL2
- Must pass `go build` and `go test ./...`
- Targets Go 1.26+ (currently using 1.26.0)

## Additional Resources

- See `README.md` for user documentation and features
- See `doc/command_line.md` for command-line options
- See `storage/WozSupportStatus.md` for WOZ format support details
- Reference implementation uses iz6502 CPU emulator: https://github.com/ivanizag/iz6502
