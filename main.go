package main

import (
	"context"
	"fmt"
)

func main() {
	order := Order{ID: 1, Product: "Laptop", Quantity: 2}
	err := OrderWorkflow(context.Background(), order)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func OrderWorkflow(ctx context.Context, order Order) error {
	// Step 1
	err := receiveOrder(ctx)
	if err != nil {
		return err
	}

	// Step 2
	err = processOrder(ctx, order)
	if err != nil {
		return err
	}

	// Step 3
	err = sendConfirmation(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Workflow completed successfully.")
	return nil
}

type Order struct {
	ID       int
	Product  string
	Quantity int
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
