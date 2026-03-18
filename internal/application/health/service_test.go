package health

import (
	"context"
	"testing"
)

func TestServiceCheck(t *testing.T) {
	service := NewService("procureflow-api", "test", "1.0.0")

	status := service.Check(context.Background())

	if status.Name != "procureflow-api" {
		t.Fatalf("expected name procureflow-api, got %q", status.Name)
	}

	if status.Environment != "test" {
		t.Fatalf("expected environment test, got %q", status.Environment)
	}

	if status.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %q", status.Version)
	}

	if status.Status != "ok" {
		t.Fatalf("expected status ok, got %q", status.Status)
	}
}
