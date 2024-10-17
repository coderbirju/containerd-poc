package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

func main() {
	if err := redisExample(); err != nil {
		log.Fatal(err)
	}
}

func redisExample() error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "example")

	// Pull the Redis image
	image, err := client.Pull(ctx, "docker.io/library/redis:alpine", containerd.WithPullUnpack)
	if err != nil {
		return err
	}
	log.Printf("Successfully pulled %s image\n", image.Name())

	// Create a new container
	container, err := client.NewContainer(
		ctx,
		"redis-server",
		containerd.WithNewSnapshot("redis-server-snapshot", image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)
	if err != nil {
		return err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	log.Printf("Successfully created container with id %s", container.ID())

	// Subscribe to containerd events
	containerdEventsC, errC := client.Subscribe(ctx, "namespace==example")

	// Loop to rapidly create and delete tasks
	for i := 0; i < 10; i++ {
		task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
		if err != nil {
			return err
		}

		exitStatusC, err := task.Wait(ctx)
		if err != nil {
			return err
		}

		if err := task.Start(ctx); err != nil {
			return err
		}

		log.Printf("Task started for iteration: %d\n", i+1)

		// Simulate processing delay (to overload the event system)
		time.Sleep(5 * time.Second)

		// Forcefully kill the task to simulate container restarts and event drop
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
			return err
		}

		// Check exit status
		status := <-exitStatusC
		code, _, err := status.Result()
		if err != nil {
			return err
		}
		fmt.Printf("Task exited with status: %d\n", code)

		if _, err := task.Delete(ctx); err != nil {
			return err
		}
	}

	// Handle the events
	go func() {
		for {
			select {
			case env := <-containerdEventsC:
				time.Sleep(5 * time.Second) // delay
				fmt.Printf("Received event: topic=%s, namespace=%s, message=%s\n", env.Topic, env.Namespace, env.Event)

			case err := <-errC:
				if err != nil && !errors.Is(err, context.Canceled) {
					fmt.Printf("Error subscribing to events: %v\n", err)
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	time.Sleep(10 * time.Second)

	return nil
}

// func redisExample() error {
// 	client, err := containerd.New("/run/containerd/containerd.sock")
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	ctx := namespaces.WithNamespace(context.Background(), "example")
// 	image, err := client.Pull(ctx, "docker.io/library/redis:alpine", containerd.WithPullUnpack)
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("Successfully pulled %s image\n", image.Name())

// 	container, err := client.NewContainer(
// 		ctx,
// 		"redis-server",
// 		containerd.WithNewSnapshot("redis-server-snapshot", image),
// 		containerd.WithNewSpec(oci.WithImageConfig(image)),
// 	)

// 	if err != nil {
// 		return err
// 	}
// 	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
// 	log.Printf("Successfully rm'd container with id %s and snapshot with id redis-server-snapshot", container.ID())

// 	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
// 	if err != nil {
// 		return err
// 	}
// 	defer task.Delete(ctx)

// 	exitStatusC, err := task.Wait(ctx)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	if err := task.Start(ctx); err != nil {
// 		return err
// 	}

// 	time.Sleep(3 * time.Second)

// 	if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
// 		return err
// 	}

// 	status := <-exitStatusC
// 	code, _, err := status.Result()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("redis-server exited with status: %d\n", code)
// 	containerdEventsC, errC := client.Subscribe(ctx, "namespace==example")

// 	for {
// 		select {
// 		case env := <-containerdEventsC:
// 			fmt.Printf("Received event: topic=%s, namespace=%s, message=%s\n", env.Topic, env.Namespace, env.Event)

// 		case err := <-errC:
// 			if err != nil && !errors.Is(err, context.Canceled) {
// 				fmt.Printf("Error subscribing to events: %v\n", err)
// 			}
// 			return nil
// 		case <-ctx.Done():
// 			return nil

// 		}
// 	}
// }

// --------------------------------------------------------------------------------------------------------------------------------------

// func redisExample() error {
// 	// Connect to the containerd client
// 	client, err := containerd.New("/run/containerd/containerd.sock")
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	// Create a namespace for the example
// 	ctx := namespaces.WithNamespace(context.Background(), "example")

// 	// Pull the Redis image
// 	image, err := client.Pull(ctx, "docker.io/library/redis:alpine", containerd.WithPullUnpack)
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("Successfully pulled %s image\n", image.Name())

// 	// Create a new container, but don't start it immediately
// 	container, err := client.NewContainer(
// 		ctx,
// 		"redis-server",
// 		containerd.WithNewSnapshot("redis-server-snapshot", image),
// 		containerd.WithNewSpec(oci.WithImageConfig(image)),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

// 	// First Task Creation: Create a task and start it
// 	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
// 	if err != nil {
// 		return err
// 	}
// 	defer task.Delete(ctx)

// 	// Wait for task exit status
// 	exitStatusC, err := task.Wait(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// Start the task
// 	if err := task.Start(ctx); err != nil {
// 		return err
// 	}

// 	// Subscribe to containerd events (before restart to catch dropped events)
// 	containerdEventsC, errC := client.Subscribe(ctx, "namespace==example")
// 	log.Println("Subscribed to containerd events")

// 	// Let the task run for a while (simulate workload)
// 	time.Sleep(3 * time.Second)

// 	go func() {
// 		log.Println("Entering event processing loop.")
// 		for {
// 			select {
// 			case env := <-containerdEventsC:
// 				time.Sleep(5 * time.Second)
// 				fmt.Printf("Received event: topic=%s, namespace=%s, message=%s\n", env.Topic, env.Namespace, env.Event)
// 			case err := <-errC:
// 				if err != nil && !errors.Is(err, context.Canceled) {
// 					fmt.Printf("Error subscribing to events: %v\n", err)
// 				}
// 				log.Println("Event subscription error or context canceled.")
// 				return
// 			case <-ctx.Done():
// 				log.Println("Context done, exiting event loop.")
// 				return
// 			}
// 		}
// 	}()

// 	// Simulate Task Restart (this mirrors Fargate's local container restart)
// 	// Kill the running task (simulating restart)
// 	log.Println("Simulating task restart by killing the task.")
// 	if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
// 		return err
// 	}

// 	// Wait for the task to exit after SIGTERM
// 	status := <-exitStatusC
// 	code, _, err := status.Result()
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("Task exited with status: %d\n", code)

// 	// Simulate task recreation (restart the container without deleting it)
// 	task, err = container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
// 	if err != nil {
// 		return err
// 	}
// 	defer task.Delete(ctx)

// 	// Start the new task (restarted task)
// 	exitStatusC, err = task.Wait(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	if err := task.Start(ctx); err != nil {
// 		return err
// 	}

// 	// Let the new task run for some time
// 	time.Sleep(5 * time.Second)

// 	// Clean up: kill the task again to finish
// 	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
// 		return err
// 	}

// 	status = <-exitStatusC
// 	code, _, err = status.Result()
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("Task exited with status: %d\n", code)

// 	return nil
// }

// --------------------------------------------------------------------------------------------------------------------------------------
