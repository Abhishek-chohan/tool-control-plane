export function debugLog(message: string): void {
  if ((process.env.TOOLPLANE_MCP_DEBUG ?? '').trim() !== '1') {
    return;
  }

  console.error(`[toolplane-typescript-mcp-adapter] ${message}`);
}