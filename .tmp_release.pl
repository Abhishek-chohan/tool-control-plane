my $file = 'server/pkg/service/requests.go';
open(my $f, '<', $file) or die $!;
local $/; my $src = <$f>; close($f);
my $n = 0;

# 1. SubmitRequestResult in-memory terminal (resolution/rejection = terminal;
#    streaming returns earlier).
my $old1 = "\trequest.DeadLetter = false\n\n\ts.recordSubmitEvents(request, resultType)\n\ts.notifyRequestUpdate(request.ID)\n\n\treturn nil\n\}";
my $new1 = "\trequest.DeadLetter = false\n\n\ts.recordSubmitEvents(request, resultType)\n\ts.notifyRequestUpdate(request.ID)\n\t// Resolution/rejection are terminal: nothing will broadcast again.\n\ts.releaseRequestSignal(request.ID)\n\n\treturn nil\n\}";
die "r1" unless index($src, $old1) >= 0;
$src =~ s/\Q$old1\E/$new1/; $n++;

# 2. CancelRequest terminal.
my $old2 = "\ts.notifyRequestUpdate(request.ID)\n\n\treturn nil\n}\n\n// SubmitRequestResult";
my $new2 = "\ts.notifyRequestUpdate(request.ID)\n\ts.releaseRequestSignal(request.ID)\n\n\treturn nil\n}\n\n// SubmitRequestResult";
die "r2" unless index($src, $old2) >= 0;
$src =~ s/\Q$old2\E/$new2/; $n++;

# 3. SubmitRequestResult store path (~815): submitted request terminal unless
#    streaming. Read context around line 815.
my $old3 = "\t\ts.mirrorRequestLocked(submitted)\n\t\ts.recordSubmitEvents(submitted, resultType)\n\t\ts.notifyRequestUpdate(submitted.ID)\n\t\treturn nil\n\t\}";
my $new3 = "\t\ts.mirrorRequestLocked(submitted)\n\t\ts.recordSubmitEvents(submitted, resultType)\n\t\ts.notifyRequestUpdate(submitted.ID)\n\t\tif resultType != model.ResultTypeStreaming {\n\t\t\t\/\/ Terminal: nothing will broadcast again.\n\t\t\ts.releaseRequestSignal(submitted.ID)\n\t\t}\n\t\treturn nil\n\t}";
die "r3" unless index($src, $old3) >= 0;
$src =~ s/\Q$old3\E/$new3/; $n++;

open(my $o, '>', $file) or die $!;
print $o $src; close($o);
print "applied $n\n";
