package main

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"strconv"
	"time"
)

type Order struct {
	ID       int
	Product  string
	Quantity int
}

// Activities
func validateOrder(ctx context.Context, order Order) (string, error) {
	fmt.Printf("Validating order %d\n", order.ID)
	return "Order validated successfully", nil
}

func processPayment(ctx context.Context, order Order) (string, error) {
	fmt.Printf("Processing payment for order %d\n", order.ID)
	return "Payment processed successfully", nil
}

func prepareShipping(ctx context.Context, order Order) (string, error) {
	fmt.Printf("Preparing shipping for order %d\n", order.ID)
	return "Shipping prepared successfully", nil
}

func notifyDelivery(ctx context.Context, order Order) error {
	fmt.Printf("Notifying delivery for order %d\n", order.ID)
	return nil
}

func OrderWorkflow(ctx workflow.Context, order Order) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	var result string

	// Order validation
	err := workflow.ExecuteActivity(ctx, validateOrder, order).Get(ctx, &result)
	if err != nil {
		return err
	}

	// Payment processing
	err = workflow.ExecuteActivity(ctx, processPayment, order).Get(ctx, &result)
	if err != nil {
		return err
	}

	// Shipping preparation
	err = workflow.ExecuteActivity(ctx, prepareShipping, order).Get(ctx, &result)
	if err != nil {
		return err
	}

	// Delivery notification
	return workflow.ExecuteActivity(ctx, notifyDelivery, order).Get(ctx, nil)
}

func main() {
	// Temporal client configuration
	c, err := client.NewLazyClient(client.Options{})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Worker registration
	w := worker.New(c, "order-processing", worker.Options{})
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterActivity(validateOrder)
	w.RegisterActivity(processPayment)
	w.RegisterActivity(prepareShipping)
	w.RegisterActivity(notifyDelivery)

	order := Order{ID: 1, Product: "T-shirt", Quantity: 2}
	we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "order_" + strconv.Itoa(order.ID), // "order-1
		TaskQueue: "order-processing",
	}, OrderWorkflow, order)
	if err != nil {
		fmt.Println("Unable to start workflow:", err)
		return
	}

	fmt.Println("Workflow started:", we.GetID())

	// Start the worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Unable to start worker:", err)
		return
	}
}
