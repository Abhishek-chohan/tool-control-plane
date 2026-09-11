"""Re-exported from toolplane.toolkits.standalone_tools — the single
maintained copy of this tool. The SWE toolkit composes these safe
file tools with its own unsafe ones (execute_bash, file_editor).
"""

from toolplane.toolkits.standalone_tools.read_file import (  # noqa: F401
    detect_encoding,
    format_output,
    get_file_language,
    main,
    read_file,
)
