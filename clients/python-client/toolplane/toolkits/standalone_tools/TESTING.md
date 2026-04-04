# Standalone Tools Test Suite

This directory contains a comprehensive test suite for all standalone tools in the toolkit.

## Test Files

### `test_standalone_toolkit.py`
Main test suite containing comprehensive tests for all tools:

- **Basic functionality tests** for each tool
- **Error handling** validation  
- **Edge case** testing
- **Output format** validation (JSON, text)
- **Integration tests** between tools
- **Command-line argument** testing

### `run_tests.py`
Test runner script with additional options:
- Coverage analysis
- Pattern-based test selection  
- Marker-based filtering
- Dependency installation

### `pytest.ini`
Pytest configuration with:
- Test discovery settings
- Output formatting
- Warning filters
- Marker definitions

## Running Tests

### Quick Test Run
```bash
# Run all tests
python -m pytest test_standalone_toolkit.py -v

# Run specific test
python -m pytest test_standalone_toolkit.py::TestStandaloneTools::test_create_file_basic -v

# Run tests for specific tool
python -m pytest test_standalone_toolkit.py -k "create_directory" -v
```

### Using Test Runner
```bash
# Install dependencies and run tests
python run_tests.py --install-deps

# Run with coverage
python run_tests.py --coverage

# Run specific pattern
python run_tests.py --pattern "file_search"

# Verbose output
python run_tests.py --verbose
```

### Coverage Analysis
```bash
python run_tests.py --coverage
# Generates htmlcov/index.html with detailed coverage report
```

## Test Categories

### Unit Tests
- Individual tool functionality
- Parameter validation
- Error handling
- Output format validation

### Integration Tests  
- Tool interactions
- File system operations
- Cross-platform compatibility

### Regression Tests
- Prevent breaking changes
- Validate tool interfaces
- Ensure backward compatibility

## Test Coverage

The test suite covers:

| Tool | Tests | Coverage Areas |
|------|-------|----------------|
| `create_directory.py` | 3 | Basic creation, existing dirs, nested paths |
| `create_file.py` | 3 | File creation, directory auto-creation, overwrite |
| `file_search.py` | 3 | Pattern matching, JSON output, result limiting |
| `grep_search.py` | 4 | Text search, regex, context lines, JSON output |
| `list_dir.py` | 2 | Directory listing, detailed output |
| `read_file.py` | 4 | Line ranges, JSON output, error handling |
| `replace_string_in_file.py` | 3 | String replacement, dry run, backup creation |
| `semantic_search.py` | 2 | Semantic matching, file type filtering |
| `test_failure_analysis.py` | 1 | Test output parsing |
| `launcher.py` | 3 | Tool discovery, search, execution |

### Cross-cutting Tests
- **Help functionality**: All tools provide `--help`
- **JSON output**: Tools with JSON support produce valid JSON
- **Error handling**: Graceful failure modes
- **File system safety**: Proper cleanup and isolation

## Test Environment

### Isolation
- Each test runs in isolated temporary directories
- Automatic cleanup after test completion
- No side effects between tests

### Mocking
- File system operations use real temporary files
- Network operations (if any) can be mocked
- External dependencies are minimized

### Platform Compatibility
- Tests work on Windows, Linux, macOS
- Path handling uses `pathlib` for cross-platform compatibility
- Shell commands handled appropriately per platform

## Adding New Tests

### For New Tools
1. Add tool to `test_all_tools_help()` list
2. Create basic functionality test:
```python
def test_new_tool_basic(self):
    # Test basic functionality
    result = self.run_tool("new_tool.py", ["arg1", "arg2"])
    assert result.returncode == 0
    assert "expected_output" in result.stdout
```

### For New Features
1. Add test method following naming convention `test_tool_feature()`
2. Use existing helper methods (`create_test_file`, `create_temp_dir`)
3. Test both success and failure cases
4. Validate output format if applicable

### Test Guidelines
- **Descriptive names**: Test method names should clearly indicate what's being tested
- **Single responsibility**: Each test should focus on one specific behavior
- **Proper cleanup**: Use fixtures and helper methods for resource management
- **Assertions**: Include meaningful assertions with descriptive messages
- **Edge cases**: Test boundary conditions and error scenarios

## Continuous Integration

### GitHub Actions Integration
```yaml
- name: Run Standalone Tools Tests
  run: |
    cd python/toolplane/toolkits/standalone_tools
    python run_tests.py --install-deps --coverage
```

### Pre-commit Hooks
```bash
# Add to .pre-commit-config.yaml
- repo: local
  hooks:
    - id: standalone-tools-tests
      name: Standalone Tools Tests
      entry: python run_tests.py
      language: system
      pass_filenames: false
```

## Debugging Failed Tests

### Verbose Output
```bash
python -m pytest test_standalone_toolkit.py -v -s
```

### Specific Test Debugging
```bash
python -m pytest test_standalone_toolkit.py::TestStandaloneTools::test_name -v -s --tb=long
```

### Temporary Files Inspection
Failed tests may leave temporary files for inspection:
- Look in `/tmp` or system temp directory
- Enable debugging prints in test methods
- Use `pytest --pdb` for interactive debugging

## Performance Testing

### Benchmark Tests
```python
def test_performance_file_search(self):
    # Create large number of test files
    # Measure search performance
    # Assert reasonable execution time
```

### Memory Usage
- Monitor memory usage for large file operations
- Ensure proper resource cleanup
- Test with large datasets

## Security Testing

### Path Traversal
- Test with malicious file paths
- Validate input sanitization
- Ensure tools don't escape intended directories

### File Permissions
- Test with restricted permissions
- Validate proper error handling
- Ensure security boundaries are respected

## Maintenance

### Regular Updates
- Update tests when tool functionality changes
- Add regression tests for bug fixes
- Keep test dependencies current

### Performance Monitoring
- Track test execution time
- Monitor resource usage
- Optimize slow tests

### Documentation
- Keep this documentation updated
- Document new test patterns
- Explain complex test scenarios
