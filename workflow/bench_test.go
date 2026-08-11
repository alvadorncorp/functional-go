package workflow_test

import (
	"context"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow"
	"github.com/alvadorncorp/gopackages/workflow/action"
)

var (
	workflowActionSink action.Action[user]
	workflowUserSink   user
	workflowErrorSink  error
)

func BenchmarkWorkflowInvoke(b *testing.B) {
	subject := loadUserBench()

	b.ReportAllocs()
	for b.Loop() {
		workflowActionSink = subject(firstID)
	}
}

func BenchmarkWorkflowRun(b *testing.B) {
	subject := loadUserBench()

	b.ReportAllocs()
	for b.Loop() {
		workflowUserSink, workflowErrorSink = action.Run(context.Background(), subject(firstID))
	}
}

func BenchmarkWorkflowThen(b *testing.B) {
	subject := workflow.Then(loadUserBench(), workflow.From(action.Pure[user]))

	b.ReportAllocs()
	for b.Loop() {
		workflowUserSink, workflowErrorSink = action.Run(context.Background(), subject(firstID))
	}
}

func BenchmarkWorkflowMap(b *testing.B) {
	subject := workflow.Map(loadUserBench(), func(loaded user) user {
		loaded.Active = true
		return loaded
	})

	b.ReportAllocs()
	for b.Loop() {
		workflowUserSink, workflowErrorSink = action.Run(context.Background(), subject(firstID))
	}
}

func loadUserBench() workflow.Workflow[userID, user] {
	return workflow.From(func(id userID) action.Action[user] {
		return action.Pure(user{ID: id, Name: "ada", Active: true})
	})
}
