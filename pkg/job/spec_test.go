package job

import (
	"fmt"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	t1, err := NewParser(SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor).
		Parse("0/3 * * * * ?")
	if err != nil {
		t.Fatal(err)
	}
	p, err := t1.NextFireTime(time.Now().UTC().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(p)
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(p)
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(p)
	if err != nil {
		t.Fatal(err)
	}
	t1, err = NewParser(SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor).
		Parse("0/3 * * * ?")
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(time.Now().UTC().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(p)
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(p)
	if err != nil {
		t.Fatal(err)
	}
	p, err = t1.NextFireTime(p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAt(t *testing.T) {
	now := time.Now()
	t.Log(now)
	t1, err := NewParser(SecondOptional | Minute | Hour | Dom | Month | Dow | Descriptor).
		Parse(fmt.Sprintf("@at %d", now.Add(30*time.Minute).UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	p, err := t1.NextFireTime(time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(time.Unix(p/1e9, p%1e9))
}
