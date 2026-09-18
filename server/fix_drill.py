path = '/repo/server/cmd/toolplane/demo.go'
src = open(path).read()

# 1. Remove the old duplicated drill block (before the !drill invoke).
old_block = '''	// 4. Invoke with a wait; on --drill, kill the provider first — the
	// control plane reclaims the request, re-executes, and finishes.
	if drill {
		time.Sleep(300 * time.Millisecond)
		if err := providerCmd.Process.Kill(); err == nil {
			fmt.Fprintln(out, "==> drill: provider killed mid-request; the control plane reclaims and re-executes")
		}
		time.Sleep(2 * time.Second)
		// A second provider takes over the session to finish the work.
		providerCmd2 := exec.Command(providerBin, providerArgs...)
		providerCmd2.Stdout = out
		providerCmd2.Stderr = os.Stderr
		if err := providerCmd2.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return cli.ExitError
		}
		defer func() {
			_ = providerCmd2.Process.Kill()
			_, _ = providerCmd2.Process.Wait()
		}()
		time.Sleep(time.Second)
	}

	if !drill {'''
new_block = '''	if !drill {'''
assert old_block in src, 'old block not found'
src = src.replace(old_block, new_block, 1)

# 2. Drill: create the request, kill the provider, then start a fresh one.
old_create = '''	// --drill: show durable execution. Kill the provider while a request
	// is pending, then bring a fresh provider online — the control plane
	// holds the request, and the new provider claims and finishes it.
	fmt.Fprintln(out, "==> drill: killing the provider with a request pending...")
	_ = providerCmd.Process.Kill()
	_, _ = providerCmd.Process.Wait()'''
new_create = '''	// --drill: show durable execution. Create a request, kill the
	// provider with it pending, then bring a fresh provider online — the
	// control plane holds the request, and the new provider claims and
	// finishes it.
	callCtx, callCancel := context.WithTimeout(ctx, 30*time.Second)
	defer callCancel()
	callCtx = metadata.AppendToOutgoingContext(callCtx, "api_key", "dev-key")
	resp, err := tools.InvokeTool(callCtx, &proto.ExecuteToolRequest{
		SessionId: sessionID,
		ToolName:  "add",
		Input:     `{"a":2,"b":3}`,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "invoke failed: %v\\n", err)
		return cli.ExitError
	}
	pendingRequestID := resp.GetRequestId()
	fmt.Fprintf(out, "==> request %s pending; killing the provider...", pendingRequestID)
	_ = providerCmd.Process.Kill()
	_, _ = providerCmd.Process.Wait()'''
assert old_create in src, 'create block not found'
src = src.replace(old_create, new_create, 1)

open(path, 'w').write(src)
print('DONE')
