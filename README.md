# Code Linter

A fast and efficient code linter for C/C++ projects, designed to enforce coding standards and best practices.

## Features

- **License Header Checking**: Ensures all source files have proper license headers
- **Header Guard Validation**: Verifies header files have proper include guards
- **Naming Convention Enforcement**: Checks function and variable naming conventions
- **Code Formatting**: Detects formatting issues like tabs vs spaces, trailing whitespace
- **Line Length Checking**: Warns about lines exceeding configured length
- **Configurable Rules**: Enable/disable specific checks as needed
- **Fast Performance**: Efficient file traversal and parallel processing

## Installation

```bash
go get github.com/nirohfeld/code_linter
```

## Usage

### As a Library

```go
package main

import (
    "fmt"
    "log"
    "github.com/nirohfeld/code_linter"
)

func main() {
    config := codelint.Config{
        RootDir:     ".",
        IncludeDirs: []string{"src", "include"},
        ExcludeDirs: []string{"third_party", "build"},
        FileTypes:   []string{".c", ".cc", ".h"},
        Checks: []string{
            "formatting",
            "naming-conventions",
            "header-guards",
            "license-headers",
        },
    }

    linter := codelint.New(config)
    results, err := linter.Run()
    if err != nil {
        log.Fatal(err)
    }

    codelint.PrintResults(results)
    
    if codelint.HasErrors(results) {
        os.Exit(1)
    }
}
```

### Configuration

The linter is configured through the `Config` struct:

- `RootDir`: Base directory to scan
- `IncludeDirs`: Directories to include (relative to RootDir)
- `ExcludeDirs`: Directories to skip (e.g., "build", ".git")
- `FileTypes`: File extensions to check (e.g., ".c", ".h")
- `Checks`: Which lint rules to enable
- `Verbose`: Enable verbose output
- `MaxErrors`: Stop after this many errors (0 = no limit)

### Available Checks

- `license-headers`: Verify files have proper license headers
- `header-guards`: Check header files for include guards
- `naming-conventions`: Enforce naming standards
- `formatting`: Check code formatting (tabs/spaces, line length, trailing whitespace)

## Lint Rules

### License Headers
Checks that source files contain a license header in the first 10 lines. Looks for common patterns like "Copyright", "SPDX-License-Identifier", etc.

### Header Guards
Ensures header files (.h, .hpp) have proper include guards:
```c
#ifndef MY_HEADER_H
#define MY_HEADER_H
// ... content ...
#endif
```
Also accepts `#pragma once` as an alternative.

### Naming Conventions
- C files: Functions should use snake_case, not camelCase
- Configurable for different project standards

### Formatting
- Checks for consistent use of tabs or spaces
- Warns about trailing whitespace
- Alerts on lines exceeding maximum length (default 100 chars)

## Integration with Build Systems

### CMake Integration

```cmake
add_custom_target(lint
    COMMAND go run github.com/nirohfeld/code_linter/cmd/codelint
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Running code linter..."
)
```

### Make Integration

```makefile
lint:
    @go run github.com/nirohfeld/code_linter/cmd/codelint
```

## Output Format

The linter outputs issues in a standard format:
```
ERROR: src/main.c:45:12: Missing semicolon [syntax]
WARNING: include/utils.h:1:1: Missing license header [license-headers]
INFO: src/helper.c:89:101: Line exceeds 100 characters (105) [line-length]
```

## Exit Codes

- `0`: Success, no errors found
- `1`: Linting errors found
- `2`: Fatal error (couldn't read files, etc.)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - See LICENSE file for details