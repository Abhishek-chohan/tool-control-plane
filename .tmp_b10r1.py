import io

def edit(path, old, new, count=1):
    src = io.open(path, encoding='utf-8', newline='').read()
    assert old in src, f"missing in {path}:\n{old[:140]}"
    src = src.replace(old, new, count)
    io.open(path, 'w', encoding='utf-8', newline='').write(src)

# 1. HTTP facade astream coroutine (docstring differs from gRPC facade).
edit('clients/python-client/toolplane/toolplane_http_client.py',
'''    def astream(
        self,
        tool_name: str,
        callback: Callable[[Any, bool], None],
        session_id: str,
        **params,
    ) -> List[Any]:
        """Alias for stream method."""
        return self.stream(tool_name, callback, session_id, **params)''',
'''    async def astream(
        self,
        tool_name: str,
        callback: Callable[[Any, bool], None],
        session_id: str,
        **params,
    ) -> List[Any]:
        """Awaitable stream: resolves with the collected chunks."""
        context = self.get_session(session_id)
        if not context:
            raise ToolplaneError(f"Session {session_id} not found")

        return await context.astream(tool_name, callback, **params)''')

# 2. example.py fallback: local ToolException import.
edit('clients/python-client/example.py',
'''        else:
            raise ToolException(f"Tool '{name}' has no run()/arun()")''',
'''        else:
            from langchain_core.tools import ToolException

            raise ToolException(f"Tool '{name}' has no run()/arun()")''')

# 3. Protocol async signatures.
edit('clients/python-client/toolplane/interfaces/session_interface.py',
    '    def ainvoke(self, tool_name: str, **params) -> str:',
    '    async def ainvoke(self, tool_name: str, **params) -> str:')

src = io.open('clients/python-client/toolplane/interfaces/session_interface.py', encoding='utf-8', newline='').read()
needle = '    def stream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):'
if 'async def astream' not in src:
    assert needle in src
    src = src.replace(needle, needle + '''

    async def astream(self, tool_name: str, callback: Callable[[Any, bool], None], **params):''', 1)
    io.open('clients/python-client/toolplane/interfaces/session_interface.py', 'w', encoding='utf-8', newline='').write(src)

# 4. Raise minimum Python to 3.9 (asyncio.to_thread).
for pf in ('clients/python-client/pyproject.toml',):
    src = io.open(pf, encoding='utf-8', newline='').read()
    src = src.replace('requires-python = ">=3.8"', 'requires-python = ">=3.9"')
    src = src.replace('Programming Language :: Python :: 3.8', 'Programming Language :: Python :: 3.9')
    io.open(pf, 'w', encoding='utf-8', newline='').write(src)

print("ok")
