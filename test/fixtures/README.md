# Test Fixtures

This directory contains test fixtures and sample data used for testing Fulmen libraries.

## Directory Structure

- `sample-project/`: A typical Go project structure with source files, tests, and configuration
- `empty-dir/`: An empty directory for testing edge cases
- `deep-structure/`: A deeply nested directory structure for testing path traversal

## Sample Project

The `sample-project` contains:

- `src/`: Source code files
- `tests/`: Test files
- `docs/`: Documentation (empty)
- `fulmen.json`: Sample configuration file

## Usage

These fixtures are used by:

- Unit tests for pathfinder functionality
- Integration tests for file discovery
- Examples in documentation
- Cross-language testing consistency

## Adding New Fixtures

When adding new test fixtures:

1. Create a descriptive directory name
2. Include a README.md explaining the fixture's purpose
3. Keep file sizes small
4. Use realistic but minimal content
