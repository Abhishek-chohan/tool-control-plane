"""Re-exported from toolplane.toolkits.standalone_tools — the single
maintained copy of this tool. The SWE toolkit composes these safe
file tools with its own unsafe ones (execute_bash, file_editor).
"""

from toolplane.toolkits.standalone_tools.grep_search import (  # noqa: F401
    get_all_text_files,
    get_files_by_pattern,
    grep_search,
    is_text_file,
    main,
    search_file,
)
