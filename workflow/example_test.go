package workflow_test

import (
	"context"
	"errors"
	"fmt"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
)

// Example shows the three moments the module separates: dependencies enter at
// construction, the input enters at application, and the context enters at Run.
func Example() {
	dependency := users{
		Find: func(id userID) action.Action[user] {
			return action.Pure(user{ID: id, Name: "ada", Active: true})
		},
	}

	find := workflow.From(dependency.Find)

	program := find(1)
	// No effect has executed yet.

	loaded, err := action.Run(context.Background(), program)
	fmt.Println(loaded.Name, err)
	// Output: ada <nil>
}

// ExampleThen shows a composed workflow becoming a reusable vocabulary word:
// LoadUser + RequireActive = LoadActiveUser, which larger workflows consume
// exactly like a primitive one.
func ExampleThen() {
	inactive := errors.New("user is not active")

	loadUser := workflow.From(func(id userID) action.Action[user] {
		return action.Pure(user{ID: id, Name: "ada", Active: false})
	})
	requireActive := workflow.MapFallible(
		workflow.From(action.Pure[user]),
		func(loaded user) (user, error) {
			if !loaded.Active {
				return user{}, inactive
			}
			return loaded, nil
		},
	)

	loadActiveUser := workflow.Then(loadUser, requireActive)

	_, err := action.Run(context.Background(), loadActiveUser(1))
	fmt.Println(err)
	// Output: user is not active
}

// ExampleWorkflow_dependencyInjection shows that swapping the adapter changes
// only the composition root, never the workflow.
func ExampleWorkflow_dependencyInjection() {
	inMemory := users{
		Find: func(id userID) action.Action[user] {
			return action.Pure(user{ID: id, Name: "in-memory", Active: true})
		},
	}
	remote := users{
		Find: func(id userID) action.Action[user] {
			return action.Pure(user{ID: id, Name: "remote", Active: true})
		},
	}

	for _, dependency := range []users{inMemory, remote} {
		loaded, _ := action.Run(context.Background(), workflow.From(dependency.Find)(1))
		fmt.Println(loaded.Name)
	}
	// Output:
	// in-memory
	// remote
}

func ExampleMap() {
	load := workflow.From(func(id userID) action.Action[user] {
		return action.Pure(user{ID: id, Name: "ada", Active: true})
	})
	loadName := workflow.Map(load, func(loaded user) string { return loaded.Name })

	name, err := action.Run(context.Background(), workflow.Provide(loadName, userID(7)))
	fmt.Println(name, err)
	// Output: ada <nil>
}

func ExampleProvide() {
	load := workflow.From(func(id userID) action.Action[user] {
		return action.Pure(user{ID: id, Name: "provided", Active: true})
	})

	loaded, err := action.Run(context.Background(), workflow.Provide(load, userID(7)))
	fmt.Println(loaded.Name, err)
	// Output: provided <nil>
}

func ExampleContramap() {
	type request struct{ ID userID }
	load := workflow.From(func(id userID) action.Action[user] {
		return action.Pure(user{ID: id, Name: "adapted", Active: true})
	})
	fromRequest := workflow.Contramap(load, func(input request) userID { return input.ID })

	loaded, err := action.Run(context.Background(), fromRequest(request{ID: 9}))
	fmt.Println(loaded.ID, loaded.Name, err)
	// Output: 9 adapted <nil>
}
