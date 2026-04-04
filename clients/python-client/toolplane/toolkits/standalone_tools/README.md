# Standalone Tools

This folder contains implementations of VS Code-independent tools that can be used for various development and analysis tasks. These tools are designed to work independently of VS Code and can be used from the command line or integrated into other systems.

## Available Tools

### 1. `create_directory.py`
Creates directories recursively with enhanced features.

**Usage:**
```bash
python create_directory.py /path/to/new/directory
python create_directory.py /path/to/new/directory --mode 755 --parents --exist_ok
```

**Parameters:**
- `dirPath` (required): The absolute path to the directory to create
- `--mode`: Permission mode for the directory (default: 755)
- `--parents`: Create parent directories if they don't exist (default: True)
- `--exist_ok`: Don't raise error if directory already exists (default: True)

### 2. `create_file.py`
Creates new files with specified content and automatic directory creation.

**Usage:**
```bash
python create_file.py /path/to/file.txt "File content here"
python create_file.py /path/to/file.txt "Content" --encoding utf-8 --auto_create_dirs --overwrite
```

**Parameters:**
- `filePath` (required): The absolute path to the file to create
- `content` (required): The content to write to the file
- `--encoding`: The encoding to use (default: utf-8)
- `--mode`: File creation mode (default: 'w')
- `--auto_create_dirs`: Create parent directories if needed (default: True)
- `--overwrite`: Overwrite existing files (default: False)

### 3. `file_search.py`
Searches for files using glob patterns with advanced features.

**Usage:**
```bash
python file_search.py "*.py"
python file_search.py "**/*.js" --sort_by size --show_details --max_results 20
```

**Parameters:**
- `query` (required): Glob pattern to search for files
- `--max_results`: Maximum number of results to return
- `--include_hidden`: Include hidden files (default: False)
- `--sort_by`: Sort by name, size, or modified (default: name)
- `--reverse_sort`: Reverse sort order (default: False)
- `--show_details`: Show file details (default: False)
- `--base_path`: Base directory to search from

### 4. `grep_search.py`
Fast text search in files using grep-like functionality.

**Usage:**
```bash
python grep_search.py "search_term"
python grep_search.py "pattern" --is_regexp --context_lines 3 --ignore_case
```

**Parameters:**
- `query` (required): The pattern to search for
- `--is_regexp`: Whether the pattern is a regex (default: False)
- `--include_pattern`: Search files matching this glob pattern
- `--max_results`: Maximum number of results to return
- `--context_lines`: Number of context lines around matches (default: 0)
- `--ignore_case`: Case-insensitive search (default: False)
- `--whole_word`: Match whole words only (default: False)

### 5. `list_dir.py`
Lists directory contents with enhanced features.

**Usage:**
```bash
python list_dir.py /path/to/directory
python list_dir.py /path/to/directory --show_details --recursive --sort_by size
```

**Parameters:**
- `path` (required): The directory path to list
- `--show_hidden`: Show hidden files (default: False)
- `--show_details`: Show detailed information (default: False)
- `--sort_by`: Sort by name, size, modified, or type (default: name)
- `--reverse_sort`: Reverse sort order (default: False)
- `--recursive`: List recursively (default: False)
- `--max_depth`: Maximum depth for recursive listing (default: 3)
- `--file_filter`: Filter files by extension

### 6. `read_file.py`
Reads file contents with line range support and encoding detection.

**Usage:**
```bash
python read_file.py /path/to/file.py 1 50
python read_file.py /path/to/file.py 10 -1 --encoding utf-8 --no_line_numbers
```

**Parameters:**
- `file_path` (required): The absolute path of the file to read
- `start_line` (required): Line number to start reading from (1-based)
- `end_line` (required): Line number to end reading at (1-based, -1 for end)
- `--encoding`: Encoding to use (default: auto-detect)
- `--show_line_numbers`: Show line numbers (default: True)
- `--no_line_numbers`: Hide line numbers
- `--highlight_syntax`: Attempt syntax highlighting (default: False)
- `--max_line_length`: Maximum line length before truncation (default: 1000)

### 7. `replace_string_in_file.py`
Replaces strings in files with safety checks and backup options.

**Usage:**
```bash
python replace_string_in_file.py /path/to/file.py "old_string" "new_string"
python replace_string_in_file.py /path/to/file.py "old" "new" --dry_run --ignore_case
```

**Parameters:**
- `file_path` (required): The absolute path to the file to edit
- `old_string` (required): The string to be replaced
- `new_string` (required): The replacement string
- `--create_backup`: Create backup before editing (default: True)
- `--no_backup`: Don't create backup
- `--dry_run`: Show changes without applying them (default: False)
- `--whole_word`: Only replace whole words (default: False)
- `--ignore_case`: Case-insensitive matching (default: False)
- `--max_replacements`: Maximum number of replacements
- `--check_uniqueness`: Check if old_string is unique

### 8. `semantic_search.py`
Semantic search for relevant code using text similarity.

**Usage:**
```bash
python semantic_search.py "function to handle user authentication"
python semantic_search.py "database connection" --max_results 5 --file_types .py .js
```

**Parameters:**
- `query` (required): The search query in natural language
- `--max_results`: Maximum number of results (default: 10)
- `--file_types`: File types to search in
- `--similarity_threshold`: Minimum similarity score (default: 0.1)
- `--context_size`: Lines of context around matches (default: 3)
- `--search_path`: Directory to search in (default: current)

### 9. `test_failure_analysis.py`
Analyzes test failures and provides debugging suggestions.

**Usage:**
```bash
python test_failure_analysis.py --test_output test_output.txt
python test_failure_analysis.py --test_framework pytest --verbose --suggest_fixes
```

**Parameters:**
- `--test_output`: Path to test output file or direct output
- `--test_framework`: Test framework used (pytest, unittest, jest, etc.)
- `--verbose`: Show detailed analysis (default: False)
- `--suggest_fixes`: Suggest potential fixes (default: True)
- `--group_by_type`: Group failures by error type (default: True)

## Common Features

All tools support:
- JSON output format with `--output_format json`
- Comprehensive error handling and reporting
- Cross-platform compatibility (Windows, Linux, macOS)
- Detailed help with `--help` flag
- UTF-8 encoding support with fallback options

## Installation

These tools are standalone Python scripts that only require Python 3.6+ and standard library modules. No additional dependencies are needed.

## Integration

These tools can be easily integrated into:
- CI/CD pipelines
- Build scripts
- Development workflows
- Other automation systems
- Custom applications via subprocess calls

## Examples

### Search for Python functions containing "auth"
```bash
python grep_search.py "def.*auth" --is_regexp --include_pattern "*.py"
```

### Find all JavaScript files larger than 1KB
```bash
python file_search.py "**/*.js" --show_details --sort_by size
```

### Search for documentation about API endpoints
```bash
python semantic_search.py "API endpoint documentation" --file_types .md .rst .txt
```

### Analyze pytest failures
```bash
python test_failure_analysis.py --test_output pytest_output.txt --test_framework pytest
```

### Replace all occurrences of old API calls
```bash
python replace_string_in_file.py api.py "old_api_call" "new_api_call" --dry_run
```

## Contributing

To add new tools:
1. Follow the existing naming convention: `tool_name.py`
2. Include comprehensive help text and parameter descriptions
3. Support multiple output formats (default, json)
4. Include proper error handling and validation
5. Add documentation to this README

## License

These tools are provided as-is for development and analysis purposes.
