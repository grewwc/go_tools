package internal

import (
	"testing"

	"github.com/grewwc/go_tools/src/strw"
)

func TestBuildRemoteSQLiteReplaceCommand(t *testing.T) {
	got := buildRemoteSQLiteReplaceCommand()
	want := `rm -f -- $HOME/.go_tools_memo.sqlite3-wal $HOME/.go_tools_memo.sqlite3-shm && mv -- $HOME/.go_tools_memo.sqlite3.incoming $HOME/.go_tools_memo.sqlite3`
	if got != want {
		t.Fatalf("unexpected replace command:\n got: %s\nwant: %s", got, want)
	}
}

func TestRunSSHCommandUsesSingleRemoteCommandArg(t *testing.T) {
	remoteCommand := buildRemoteSQLiteReplaceCommand()
	cmd := buildCommandLine("ssh", "-o", "ConnectTimeout=8", "user@host", remoteCommand)
	parts := strw.SplitByStrKeepQuotes(cmd, " ", `"'`, false)
	if len(parts) != 5 {
		t.Fatalf("unexpected ssh command parts: %+v", parts)
	}
	if parts[4] != remoteCommand {
		t.Fatalf("remote command should remain a single arg: %+v", parts)
	}
	if strw.Contains(parts[4], `"`) {
		t.Fatalf("remote command should not contain literal escaped quotes: %s", parts[4])
	}
}
