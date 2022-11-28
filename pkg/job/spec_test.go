package job

import (
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
