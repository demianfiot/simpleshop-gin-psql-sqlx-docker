package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"prac/pkg/repository"
	"prac/todo"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializating configs: %s", err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: viper.GetStringSlice("kafka.bootstrap_servers"),
		Topic:   viper.GetString("kafka.topic_orders"),
		GroupID: viper.GetString("kafka.group_id"), // dl9 grupi - mashtabuvann9
	})
	defer reader.Close()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{viper.GetString("clickhouse.host") + ":" + viper.GetString("clickhouse.port")},
		Auth: clickhouse.Auth{
			Database: viper.GetString("clickhouse.database"),
			Username: viper.GetString("clickhouse.user"),
			Password: viper.GetString("clickhouse.password"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewOrderAnalyticsRepository(conn)

	// graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down consumer...")
		cancel()

	}()

	for {
		msg, err := reader.FetchMessage(ctx) // readmsg має автоматичний коміт
		if err != nil {
			log.Println("reader stopped:", err)
			return
		}
		var event todo.OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("invalid message:", err)
			continue
		}
		if err := repo.InsertOrder(ctx, event); err != nil {
			log.Println("failed to insert:", err)
			continue // offset НЕ комітимо
		}
		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Println("commit failed:", err)
		}
		log.Printf(
			"Order %d created by user %d, total %.2f",
			event.OrderID,
			event.UserID,
			event.Total,
		)
	}
}
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
