package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"google.golang.org/api/iterator"
	"time"
)

var (
	project       string
	subscriptions []string
	confirm       bool
)

func getSubscriptions(project string) []string {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	it := client.Subscriptions(ctx)
	subs := make([]string, 0)

	for {
		sub, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		subs = append(subs, sub.ID())
	}

	return subs
}

func purgeEvents(project string, subscriptions []string) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	seekTo := time.Now().Add(time.Minute * 10)

	for _, s := range subscriptions {
		sub := client.Subscription(s)
		fmt.Printf("Purging events in subscription: [%s]\n", sub)
		sub.SeekToTime(ctx, seekTo)
	}
}

func main() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Which Project").
				Value(&project),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which Subscriptions?").
				OptionsFunc(func() []huh.Option[string] {
					subs := getSubscriptions(project)
					return huh.NewOptions(subs...)
				}, &project).
				Value(&subscriptions),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirm Purge?").
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		),
	)
	form.WithTheme(huh.ThemeBase())
	form.Run()

	if confirm {
		purgeEvents(project, subscriptions)
	}
}
