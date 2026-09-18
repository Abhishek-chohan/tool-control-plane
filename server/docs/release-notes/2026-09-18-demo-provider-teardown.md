# The demo reaps its provider

Date: 2026-09-18

`toolplane demo` leaked the provider subprocess on every happy-path run:
the teardown killed it only in the `--drill` branch, so a successful demo
exited with the provider still running — holding any inherited stdout
pipe open (a piped `toolplane demo | tail` then never finished) and
accumulating orphaned processes across runs.

- The provider is now killed and reaped on **every** exit path via a
  teardown defer, and the waits use `cmd.Wait()` so the exec machinery is
  fully joined, not just the process reaped.
- The provider subprocess now writes straight to the terminal file
  descriptors instead of the command's output writer: fd passthrough
  removes the parent-side copier goroutine entirely, which also removes
  a data race between that copier and the demo's own prints into the
  same buffer (present on the drill path since the demo landed).

Notes: `--drill` behavior is unchanged — the provider is still killed
mid-request and the replacement provider is torn down on exit. A
regression test drives the failure path with a fake provider and asserts
the process is reaped within a bounded window.
