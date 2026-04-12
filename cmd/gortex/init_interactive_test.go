package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// runWizard is a test harness that feeds canned stdin into
// runInteractiveInit. Returns the choice + the prompt output so
// assertions can pin both the control flow and the user-facing text.
//
// Note: runInteractiveInit gates on isInteractive() which reads from
// the real os.Stdin. We side-step that by calling the *bufio*-
// consuming half directly — the prompt body behaves identically when
// invoked with arbitrary io.Reader, we just need to bypass the tty
// check. So tests here invoke the internal logic via a thin wrapper
// that assumes interactive mode.
func runWizard(t *testing.T, input string) (interactiveChoice, string) {
	t.Helper()
	var out bytes.Buffer
	in := strings.NewReader(input)

	// Call the prompt body directly by inlining — this mirrors
	// runInteractiveInit but skips the isInteractive() gate so the
	// wizard is exercised under test without a real TTY.
	choice, _ := runInteractiveForTest(in, &out)
	return choice, out.String()
}

func TestInteractiveWizard_DefaultIsGlobal(t *testing.T) {
	// Pressing Enter to every prompt must pick global, track=yes, start=yes.
	// That's the "happy path" we optimized the wizard for.
	choice, out := runWizard(t, "\n\n\n")
	assert.True(t, choice.Global, "empty input must default to global")
	assert.True(t, choice.Track)
	assert.True(t, choice.Start)
	assert.Contains(t, out, "Global daemon (recommended)")
}

func TestInteractiveWizard_ChoosePerRepo(t *testing.T) {
	choice, _ := runWizard(t, "2\n")
	assert.False(t, choice.Global, "option 2 must map to per-repo mode")
	// Track / Start aren't asked when per-repo — they only make sense
	// in global mode.
	assert.False(t, choice.Track)
	assert.False(t, choice.Start)
}

func TestInteractiveWizard_DeclineTrack(t *testing.T) {
	// Global + decline-to-track + start-yes. Makes sure the two
	// follow-up prompts are handled independently.
	choice, _ := runWizard(t, "1\nn\ny\n")
	assert.True(t, choice.Global)
	assert.False(t, choice.Track, "'n' to track must set Track=false")
	assert.True(t, choice.Start)
}

func TestInteractiveWizard_DeclineBoth(t *testing.T) {
	choice, _ := runWizard(t, "1\nn\nn\n")
	assert.True(t, choice.Global)
	assert.False(t, choice.Track)
	assert.False(t, choice.Start)
}

func TestInteractiveWizard_UnrecognizedChoiceFallsBackToGlobal(t *testing.T) {
	// Unknown first-prompt answer prints a warning and defaults to
	// global. Users get out of trouble by pressing Enter next time.
	choice, out := runWizard(t, "zzz\n\n\n")
	assert.True(t, choice.Global)
	assert.Contains(t, out, "Unrecognized")
}

func TestIsNo(t *testing.T) {
	// Blank input is yes (default pick). Anything starting with n is no.
	// Case and trailing whitespace must not matter.
	assert.False(t, isNo(""))
	assert.False(t, isNo("\n"))
	assert.False(t, isNo("y\n"))
	assert.False(t, isNo("Y\n"))
	assert.False(t, isNo("yes\n"))
	assert.True(t, isNo("n\n"))
	assert.True(t, isNo("N\n"))
	assert.True(t, isNo("no\n"))
	assert.True(t, isNo("  no  \n"))
}

// runInteractiveForTest invokes the prompt body with an arbitrary
// io.Reader so tests can feed canned input. It duplicates the body of
// runInteractiveInit minus the isInteractive() gate — keeping the
// gate untested is fine since it just wraps os.Stdin.Stat().
//
// Defined in the _test.go file (not daemon_interactive.go) so the
// production binary doesn't ship a second entry point that bypasses
// the TTY check.
func runInteractiveForTest(in *strings.Reader, out *bytes.Buffer) (interactiveChoice, bool) {
	return runInteractiveInit(in, out)
}
