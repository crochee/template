// Package msg
package msg

import "testing"

func TestName(t *testing.T) {
	t.Log(Exit, ResetTime)
	c := &Status{}
	c.SetStatus(Exit)
	c.AddStatus(ResetTime)
	if !c.HasStatus(ResetTime) {
		t.Error("")
	}
	if c.NotHasStatus(ResetTime) {
		t.Error()
	}
	if c.OnlyHas(ResetTime) {
		t.Error()
	}
	c.DeleteStatus(ResetTime)
	t.Log(c.Flag)
	if c.HasStatus(ResetTime) {
		t.Error()
	}
	if !c.NotHasStatus(ResetTime) {
		t.Error()
	}
	if c.OnlyHas(ResetTime) {
		t.Error()
	}
}
