package main

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"time"

	"go.temporal.io/sdk/workflow"
)

func main() {
	c, err := client.NewLazyClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		fmt.Println("Unable to create client:", err)
		return
	}
	defer c.Close()

	w := worker.New(c, "order-processing", worker.Options{})

	w.RegisterWorkflow(OrderWorkflow)

	w.RegisterActivity(receiveOrder)

	w.RegisterActivity(processOrder)

	w.RegisterActivity(sendConfirmation)

	options := client.StartWorkflowOptions{
		TaskQueue: "order-processing",
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, OrderWorkflow, Order{ID: 1, Product: "T-shirt", Quantity: 2})
	if err != nil {
		fmt.Println("Unable to start workflow:", err)
		return
	}
	fmt.Println("Workflow started:", we.GetID())

	err = w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Unable to start worker:", err)
		return
	}
}

type Order struct {
	ID       int
	Product  string
	Quantity int
}

func OrderWorkflow(ctx workflow.Context, order Order) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	// Step 1
	err := workflow.ExecuteActivity(ctx, receiveOrder).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Step 2
	err = workflow.ExecuteActivity(ctx, processOrder, order).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Step 3
	err = workflow.ExecuteActivity(ctx, sendConfirmation).Get(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Println("Workflow completed successfully.")
	return nil
}

// Step 1: Receive the order
func receiveOrder(ctx context.Context) error {
	// Receiving order...
	fmt.Println("Order received.")
	return nil
}

// Step 2: Process the order
func processOrder(ctx context.Context, order Order) error {
	// Processing order...
	fmt.Println("Order processed:", order)
	return nil
}

// Step 3: Send confirmation email
func sendConfirmation(ctx context.Context) error {
	// Sending confirmation email...
	fmt.Println("Order confirmation sent via email.")
	return nil
}
