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

const (
	deliveryNotificationSignal = "delivery-notification-signal"
	paymentNotificationSignal  = "payment-notification-signal"
	orderProcessingQueue       = "order-processing"
)

type Order struct {
	ID       int
	Product  string
	Quantity int
}

type User struct {
	ID int
}

func queuePayment(flow interface{}, queue string, order Order, user User) error {
	time.Sleep(2 * time.Second) // Simulating payment processing delay
	c, err := client.NewLazyClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		return err
	}

	we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		TaskQueue: queue,
	}, flow, user, order)
	if err != nil {
		return err
	}

	fmt.Println("Payment Workflow started:", we.GetID())
	return nil
}

func queueDelivery(flow interface{}, queue string, order Order) error {
	time.Sleep(2 * time.Second) // Simulating payment processing delay
	c, err := client.NewLazyClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		return err
	}

	we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		TaskQueue: queue,
	}, flow, order)
	if err != nil {
		return err
	}

	fmt.Println("Delivery Workflow started:", we.GetID())
	return nil
}

// Activities
func validateOrder(ctx context.Context, order Order) (string, error) {
	fmt.Printf("Validating order %d\n", order.ID)
	return "Order validated successfully", nil
}

func processPayment(ctx context.Context, order Order) (string, error) {
	fmt.Printf("Processing payment for order %d\n", order.ID)

	// Payment processing logic here...
	user := User{ID: 1}
	err := queuePayment(PaymentWorkflow, orderProcessingQueue, order, user)
	if err != nil {
		return "", err
	}
	return "Payment processed successfully", nil
}

func prepareShipping(ctx context.Context, order Order) (string, error) {
	fmt.Printf("Preparing shipping for order %d\n", order.ID)
	err := queueDelivery(DeliveryWorkflow, orderProcessingQueue, order)
	if err != nil {
		return "", err
	}
	return "Shipping prepared successfully", nil
}

func notifyDelivery(ctx context.Context, order Order) error {
	fmt.Printf("Notifying delivery for order %d\n", order.ID)
	return nil
}

func validatePayment(ctx context.Context, user User, order Order) (float64, error) {
	fmt.Printf("Validating payment for order %d\n", order.ID)
	// Your validation logic here...
	return 100.00, nil // Placeholder value
}

func chargeCreditCard(ctx context.Context, user User, payment float64) (string, error) {
	fmt.Printf("Charging credit card with payment amount: %.2f\n", payment)
	// Your credit card charging logic here...
	return "charged", nil // Placeholder value
}

func verifyPayment(ctx context.Context, orderId int) (string, error) {
	fmt.Printf("Verifying payment transaction with ID: %d\n", orderId)
	// Your payment verification logic here...
	return "verified", nil // Placeholder value
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

	// Wait for delivery notification
	workflow.GetSignalChannel(ctx, paymentNotificationSignal).Receive(ctx, nil)

	// Shipping preparation
	err = workflow.ExecuteActivity(ctx, prepareShipping, order).Get(ctx, &result)
	if err != nil {
		return err
	}

	// Wait for delivery notification
	workflow.GetSignalChannel(ctx, deliveryNotificationSignal).Receive(ctx, nil)

	// Delivery notification
	return workflow.ExecuteActivity(ctx, notifyDelivery, order).Get(ctx, nil)
}

func PaymentWorkflow(ctx workflow.Context, user User, order Order) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	// Validating payment
	var payment float64
	err := workflow.ExecuteActivity(ctx, validatePayment, user, order).Get(ctx, &payment)
	if err != nil {
		return err
	}

	// Executing charge credit card
	var paymentResult string
	err = workflow.ExecuteActivity(ctx, chargeCreditCard, user, payment).Get(ctx, &paymentResult)
	if err != nil {
		return err
	}

	// Verifying payment
	err = workflow.ExecuteActivity(ctx, verifyPayment, order.ID).Get(ctx, &paymentResult)
	if err != nil {
		return err
	}

	// Signal the order workflow
	err = workflow.SignalExternalWorkflow(ctx, orderWorkflowId(order), "", paymentNotificationSignal, nil).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func DeliveryWorkflow(ctx workflow.Context, order Order) error {
	// Signal the order workflow
	err := workflow.SignalExternalWorkflow(ctx, orderWorkflowId(order), "", deliveryNotificationSignal, nil).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func orderWorkflowId(order Order) string {
	return "order_" + strconv.Itoa(order.ID)
}

func main() {
	// Temporal client configuration
	c, err := client.NewLazyClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Worker registration
	w := worker.New(c, orderProcessingQueue, worker.Options{})
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterWorkflow(PaymentWorkflow)
	w.RegisterWorkflow(DeliveryWorkflow)

	w.RegisterActivity(validateOrder)
	w.RegisterActivity(processPayment)
	w.RegisterActivity(prepareShipping)
	w.RegisterActivity(notifyDelivery)

	w.RegisterActivity(validatePayment)
	w.RegisterActivity(chargeCreditCard)
	w.RegisterActivity(verifyPayment)

	order := Order{ID: 2, Product: "T-shirt", Quantity: 2}
	we, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "order_" + strconv.Itoa(order.ID), // "order-1
		TaskQueue: "order-processing",
	}, OrderWorkflow, order)
	if err != nil {
		fmt.Println("Unable to start workflow:", err)
		return
	}

	fmt.Println("Workflow started:", we.GetID())

	//Start the worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Unable to start worker:", err)
		return
	}
}
