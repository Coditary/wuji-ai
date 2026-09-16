package cli

import (
	"testing"

	"github.com/coditary/wuji-core/pkg/driver"
)

func TestBuildDataRequestEmbed(t *testing.T) {
	req, err := buildDataRequest([]string{"hello"}, dataRequestFields{
		taskFlags: dataTaskFlags{embed: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Text != "hello" {
		t.Fatalf("text=%q", req.Text)
	}
	if req.TaskOrDefault() != driver.DataTaskEmbed {
		t.Fatalf("task=%s", req.Task)
	}
}

func TestDataTaskFlagsConflict(t *testing.T) {
	_, err := buildDataRequest(nil, dataRequestFields{
		taskFlags: dataTaskFlags{embed: true, detect: true},
	})
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestDataTaskFlagsVectorAlias(t *testing.T) {
	req, err := buildDataRequest([]string{"x"}, dataRequestFields{
		taskFlags: dataTaskFlags{vector: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Task != driver.DataTaskVector {
		t.Fatalf("task=%s", req.Task)
	}
	if req.TaskOrDefault() != driver.DataTaskEmbed {
		t.Fatalf("normalized task=%s", req.TaskOrDefault())
	}
}
