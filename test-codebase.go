// package main

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"syscall"
// 	"time"

// 	"github.com/containerd/containerd"
// 	"github.com/containerd/containerd/cio"
// 	"github.com/containerd/containerd/namespaces"
// 	"github.com/containerd/containerd/oci"
// )

// func main() {
// 	if err := redisExample(); err != nil {
// 		log.Fatal(err)
// 	}
// }

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
