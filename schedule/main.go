package main

import (
	"context"
	"fmt"
	"go.temporal.io/api/enums/v1"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func ScheduledWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	var result string
	err := workflow.ExecuteActivity(ctx, SampleActivity).Get(ctx, &result)
	if err != nil {
		return err
	}

	fmt.Println("Activity executed successfully")
	return nil
}

func SampleActivity(ctx context.Context) (string, error) {
	return "Activity executed successfully", nil
}

func main() {
	c, err := client.NewLazyClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	w := worker.New(c, "schedule-queue", worker.Options{})
	w.RegisterWorkflow(ScheduledWorkflow)
	w.RegisterActivity(SampleActivity)

	_, err = c.ScheduleClient().Create(context.Background(), client.ScheduleOptions{
		ID:      "sample-schedule",
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"30 17 * * 5"},
			TimeZoneName:    "America/Sao_Paulo",
		},
		Action: &client.ScheduleWorkflowAction{
			Workflow:  ScheduledWorkflow,
			TaskQueue: "schedule-queue",
		},
	})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		panic(err)
	}
}
