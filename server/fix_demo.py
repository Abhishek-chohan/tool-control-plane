import re

path = '/repo/server/cmd/toolplane/demo.go'
src = open(path).read()

# 1. Remove the stray duplicated poll block (the second occurrence).
dup = '''		toolCtx, toolCancel := context.WithTimeout(ctx, 2*time.Second)
		_, err := tools.GetTool(toolCtx, &proto.GetToolRequest{
			SessionId: sessionID,
			ToolName:  "add",
		})
		toolCancel()
		if err == nil {
			break
		}
		if time.Now().After(toolDeadline) {
			fmt.Fprintln(os.Stderr, "provider never registered tool add — check the provider output above")
			return cli.ExitError
		}
		time.Sleep(200 * time.Millisecond)
	}
'''
first = src.find(dup)
second = src.find(dup, first + 1)
if second != -1:
    src = src[:second] + src[second + len(dup):]
    print('stray duplicate removed')
else:
    print('no stray duplicate found')

# 2. Create gconn before the readiness poll (if not already there).
anchor = '''	// Wait for the provider to finish registering: poll the tool until it
	// resolves in the session (Python startup + registration take a beat).
	tools := proto.NewToolServiceClient(gconn)'''
new_anchor = '''	gconn, err := conn.Dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer gconn.Close()

	// Wait for the provider to finish registering: poll the tool until it
	// resolves in the session (Python startup + registration take a beat).
	tools := proto.NewToolServiceClient(gconn)'''
if anchor in src and 'gconn, err := conn.Dial()' not in src.split(anchor)[0][-400:]:
    src = src.replace(anchor, new_anchor, 1)
    print('gconn creation added before poll')
else:
    print('gconn already created or anchor missing')

# 3. Remove the later duplicate gconn creation.
late = '''	gconn, err := conn.Dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer gconn.Close()
'''
# Keep only the FIRST occurrence (ours); later ones go.
idx = src.find(new_anchor.split('\t// Wait')[0].rstrip('\n') + '\n\ndefer gconn.Close()')
first_g = src.find(new_anchor.split('\tdefer gconn.Close()')[0])
after = src.find(new_anchor.split('\tdefer gconn.Close()')[0])
# Simpler: count occurrences of the creation and drop later duplicates.
creation = '''	gconn, err := conn.Dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer gconn.Close()
'''
occurrences = src.count(creation)
print('creation count:', occurrences)
while src.count(creation) > 1:
    idx2 = src.find(creation)
    src = src[:idx2] + src[idx2 + len(creation):]

open(path, 'w').write(src)
print('DONE')
