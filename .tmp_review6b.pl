my $n = 0;
sub edit {
  my ($file, $old, $new) = @_;
  open(my $f, '<', $file) or die "$file: $!"; local $/; my $src = <$f>; close($f);
  die "missing in $file:\n$old\n" unless index($src, $old) >= 0;
  $src =~ s/\Q$old\E/$new/;
  open(my $o, '>', $file) or die $!; print $o $src; close($o);
  $n++;
}

# http session_context: same accumulate fix.
my $hf = 'clients/python-client/toolplane/http_core/http_session_context.py';
edit($hf,
  '        except ToolplaneInvalidArgumentError:
            # The retained window moved past our position; the full result is
            # still fetchable by polling the original request.
            pass
        return self._stream_via_polling(
            tool_name,
            callback,
            params,
            idempotency_key,
            request_id=request_id,
            skip=len(all_chunks),
        )',
  '        except ToolplaneInvalidArgumentError:
            # The retained window moved past our position; the full result is
            # still fetchable by polling the original request.
            self._stream_via_polling(
                tool_name,
                callback,
                params,
                idempotency_key,
                request_id=request_id,
                skip=len(all_chunks),
                accumulate=all_chunks,
            )
        return all_chunks');

edit($hf,
  '    def _stream_via_polling(
        self,
        tool_name: str,
        callback: Callable,
        params: Dict,
        idempotency_key: str = "",
        request_id: Optional[str] = None,
        skip: int = 0,
    ):
        """Stream via polling fallback.

        When request_id is given, polls that request; otherwise invokes the
        tool (under idempotency_key when provided) and polls the result.
        skip suppresses the first skip chunks, which the caller already
        delivered.
        """
        if request_id is None:
            request_id = self.tool_manager.execute_tool(
                self.session_id, tool_name, params, idempotency_key
            )

        all_chunks = []
        last_chunk_count = skip',
  '    def _stream_via_polling(
        self,
        tool_name: str,
        callback: Callable,
        params: Dict,
        idempotency_key: str = "",
        request_id: Optional[str] = None,
        skip: int = 0,
        accumulate: Optional[List] = None,
    ):
        """Stream via polling fallback.

        When request_id is given, polls that request; otherwise invokes the
        tool (under idempotency_key when provided) and polls the result.
        skip suppresses the first skip chunks, which the caller already
        delivered. When accumulate is given, polled chunks append to it so
        callers keep the chunks they already delivered.
        """
        if request_id is None:
            request_id = self.tool_manager.execute_tool(
                self.session_id, tool_name, params, idempotency_key
            )

        all_chunks = accumulate if accumulate is not None else []
        last_chunk_count = skip');

# http resume_stream: parse the real error frame code instead of assuming
# OUT_OF_RANGE, and surface the final marker even with an empty chunk.
my $hr = 'clients/python-client/toolplane/http_core/http_request.py';
edit($hr,
  '                error_text = chunk.get("error", "") or ""
                if error_text:
                    raise api_error_from_http_response(
                        400,
                        json.dumps({"error": {"code": 11, "message": error_text}}),
                        context=f"Failed to resume stream for request {request_id}",
                    )

                value = chunk.get("chunk")
                if value not in (None, ""):
                    yield {
                        "seq": int(chunk.get("seq", 0) or 0),
                        "request_id": chunk.get("requestId", request_id),
                        "chunk": value,
                        "is_final": bool(chunk.get("isFinal", False)),
                        "error": "",
                    }
                if chunk.get("isFinal"):
                    return',
  '                error_frame = chunk.get("error")
                if isinstance(error_frame, dict):
                    # The gateway wraps the gRPC status (with its numeric
                    # code) in the error frame; surface the real code rather
                    # than guessing.
                    raise api_error_from_http_response(
                        400,
                        json.dumps({"error": error_frame}),
                        context=f"Failed to resume stream for request {request_id}",
                    )
                error_text = error_frame if isinstance(error_frame, str) else ""
                if error_text:
                    raise api_error_from_http_response(
                        400,
                        json.dumps({"message": error_text}),
                        context=f"Failed to resume stream for request {request_id}",
                    )

                if chunk.get("isFinal"):
                    # The final marker is part of the stream contract even
                    # when it carries no chunk payload.
                    yield {
                        "seq": int(chunk.get("seq", 0) or 0),
                        "request_id": chunk.get("requestId", request_id),
                        "chunk": chunk.get("chunk") or "",
                        "is_final": True,
                        "error": "",
                    }
                    return

                value = chunk.get("chunk")
                if value not in (None, ""):
                    yield {
                        "seq": int(chunk.get("seq", 0) or 0),
                        "request_id": chunk.get("requestId", request_id),
                        "chunk": value,
                        "is_final": False,
                        "error": "",
                    }');

print "applied $n\n";
