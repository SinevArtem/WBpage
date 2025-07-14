package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/SinevArtem/WBpage.git/internal/cache"
	"github.com/SinevArtem/WBpage.git/internal/model"

	"github.com/segmentio/kafka-go"
)

func Customer(db *sql.DB, cache *cache.Cache, ctx context.Context) error {
	// Проверяем, что БД доступна
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database is not available: %w", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:29092"},
		Topic:   "wb-topic",
		GroupID: "wb-groupID",
	})

	defer reader.Close()
	for {

		select {
		case <-ctx.Done():
			log.Println("stop kafka")
			return nil
		default:
			msg, err := reader.ReadMessage(ctx)

			if err != nil {
				if err == context.Canceled {
					return nil
				}
				log.Printf("Error reading message: %v (retrying)", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if len(msg.Value) == 0 {
				log.Println("Received empty message, skipping")
				continue
			}

			var order model.Order

			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("error parse json %v", err)
				continue

			}

			if err := save(db, &order); err != nil {
				log.Printf("error save order: %v", err)
				continue
			}

			cache.Set(order)
			fmt.Println(order.OrderUid)
		}

	}

}

func save(db *sql.DB, order *model.Order) error {
	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer transaction.Rollback()

	query := `INSERT INTO orders (order_uid, track_number, entry, locale,
		internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created,
		oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = transaction.Exec(query,
		order.OrderUid,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerId,
		order.DeliveryService,
		order.Shardkey,
		order.SmId,
		order.DateCreated,
		order.OofShard,
	)

	if err != nil {
		return fmt.Errorf("error insert order: %v", err)
	}

	query = `INSERT INTO delivery (order_uid, name, phone, zip,
		city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = transaction.Exec(query,
		order.OrderUid,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)

	if err != nil {
		return fmt.Errorf("error insert delivery: %v", err)
	}

	query = `INSERT INTO payment (order_uid, transaction, request_id, currency,
		provider, amount, payment_dt, bank, delivery_cost, goods_total,
		custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = transaction.Exec(query,
		order.OrderUid,
		order.Payment.Transaction,
		order.Payment.RequestId,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)

	if err != nil {
		return fmt.Errorf("error insert payment: %v", err)
	}

	for _, item := range order.Items {
		query = `INSERT INTO items (order_uid, chrt_id, track_number, price,
			rid, name, sale, size, total_price, nm_id,
			brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		_, err = transaction.Exec(query,
			order.OrderUid,
			item.ChrtId,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmId,
			item.Brand,
			item.Status,
		)

		if err != nil {
			return fmt.Errorf("error insert order: %v", err)
		}
	}

	return transaction.Commit()
}
