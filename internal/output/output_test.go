package output_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/siggimoo/task/v3/internal/templater"
	"github.com/siggimoo/task/v3/taskfile"
	"github.com/stretchr/testify/assert"

	"github.com/siggimoo/task/v3/internal/output"
)

func TestInterleaved(t *testing.T) {
	var b bytes.Buffer
	var o output.Output = output.Interleaved{}
	var w = o.WrapWriter(&b, "", nil)

	fmt.Fprintln(w, "foo\nbar")
	assert.Equal(t, "foo\nbar\n", b.String())
	fmt.Fprintln(w, "baz")
	assert.Equal(t, "foo\nbar\nbaz\n", b.String())
}

func TestGroup(t *testing.T) {
	var b bytes.Buffer
	var o output.Output = output.Group{}
	var w = o.WrapWriter(&b, "", nil).(io.WriteCloser)

	fmt.Fprintln(w, "foo\nbar")
	assert.Equal(t, "", b.String())
	fmt.Fprintln(w, "baz")
	assert.Equal(t, "", b.String())
	assert.NoError(t, w.Close())
	assert.Equal(t, "foo\nbar\nbaz\n", b.String())
}

func TestGroupWithBeginEnd(t *testing.T) {
	tmpl := templater.Templater{
		Vars: &taskfile.Vars{
			Keys: []string{"VAR1"},
			Mapping: map[string]taskfile.Var{
				"VAR1": {Static: "example-value"},
			},
		},
	}

	var o output.Output = output.Group{
		Begin: "::group::{{ .VAR1 }}",
		End:   "::endgroup::",
	}
	t.Run("simple", func(t *testing.T) {
		var b bytes.Buffer
		var w = o.WrapWriter(&b, "", &tmpl).(io.WriteCloser)

		fmt.Fprintln(w, "foo\nbar")
		assert.Equal(t, "", b.String())
		fmt.Fprintln(w, "baz")
		assert.Equal(t, "", b.String())
		assert.NoError(t, w.Close())
		assert.Equal(t, "::group::example-value\nfoo\nbar\nbaz\n::endgroup::\n", b.String())
	})
	t.Run("no output", func(t *testing.T) {
		var b bytes.Buffer
		var w = o.WrapWriter(&b, "", &tmpl).(io.WriteCloser)
		assert.NoError(t, w.Close())
		assert.Equal(t, "", b.String())
	})
}

func TestPrefixed(t *testing.T) {
	var b bytes.Buffer
	var o output.Output = output.Prefixed{}
	var w = o.WrapWriter(&b, "prefix", nil).(io.WriteCloser)

	t.Run("simple use cases", func(t *testing.T) {
		b.Reset()

		fmt.Fprintln(w, "foo\nbar")
		assert.Equal(t, "[prefix] foo\n[prefix] bar\n", b.String())
		fmt.Fprintln(w, "baz")
		assert.Equal(t, "[prefix] foo\n[prefix] bar\n[prefix] baz\n", b.String())
	})

	t.Run("multiple writes for a single line", func(t *testing.T) {
		b.Reset()

		for _, char := range []string{"T", "e", "s", "t", "!"} {
			fmt.Fprint(w, char)
			assert.Equal(t, "", b.String())
		}

		assert.NoError(t, w.Close())
		assert.Equal(t, "[prefix] Test!\n", b.String())
	})
}
