package k8s

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePrompter scripts the interactive answers for tests.
type fakePrompter struct {
	confirm      bool
	confirmErr   error
	chooseIdx    int
	chooseErr    error
	confirmAsked bool
	chooseAsked  bool
	lastOptions  []string
}

func (f *fakePrompter) Confirm(string, bool) (bool, error) {
	f.confirmAsked = true
	return f.confirm, f.confirmErr
}

func (f *fakePrompter) Choose(_ string, options []string) (int, error) {
	f.chooseAsked = true
	f.lastOptions = options
	return f.chooseIdx, f.chooseErr
}

var threeContexts = []ContextInfo{
	{Name: "ctx-a", Cluster: "cluster-a"},
	{Name: "ctx-b", Cluster: "cluster-b", Current: true},
	{Name: "ctx-c"},
}

func TestSelectContext_AcceptCurrent(t *testing.T) {
	p := &fakePrompter{confirm: true}
	got, err := SelectContext(threeContexts, "ctx-b", p)
	require.NoError(t, err)
	assert.Equal(t, "ctx-b", got)
	assert.True(t, p.confirmAsked)
	assert.False(t, p.chooseAsked, "accepting current must not show the list")
}

func TestSelectContext_DeclineCurrentThenChoose(t *testing.T) {
	p := &fakePrompter{confirm: false, chooseIdx: 0}
	got, err := SelectContext(threeContexts, "ctx-b", p)
	require.NoError(t, err)
	assert.Equal(t, "ctx-a", got, "index 0 → first context")
	assert.True(t, p.chooseAsked)
	// labels include the cluster hint for clarity
	assert.Contains(t, p.lastOptions[0], "ctx-a")
	assert.Contains(t, p.lastOptions[0], "cluster-a")
}

func TestSelectContext_NoCurrentGoesStraightToList(t *testing.T) {
	p := &fakePrompter{chooseIdx: 2}
	got, err := SelectContext(threeContexts, "", p)
	require.NoError(t, err)
	assert.Equal(t, "ctx-c", got)
	assert.False(t, p.confirmAsked, "no current context → no confirm step")
	assert.True(t, p.chooseAsked)
}

func TestSelectContext_Empty(t *testing.T) {
	_, err := SelectContext(nil, "", &fakePrompter{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no kubeconfig contexts")
}

func TestSelectContext_PropagatesErrors(t *testing.T) {
	_, err := SelectContext(threeContexts, "ctx-b", &fakePrompter{confirmErr: errors.New("boom")})
	require.Error(t, err)

	_, err = SelectContext(threeContexts, "", &fakePrompter{chooseErr: errors.New("boom")})
	require.Error(t, err)
}
