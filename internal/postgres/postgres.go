package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/SinevArtem/WBpage.git/internal/model"
	"github.com/SinevArtem/WBpage.git/internal/postgres/migrations"
	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

func New() (*Storage, error) {
	connStr := "user=user password=1111 dbname=WBdatabase host=localhost port=5428 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// Ждём готовности БД с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	migrator := MustGetNewMigrator(migrations.Files, ".")

	if err = migrator.ApplyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) Insert(response string, args ...any) error {
	_, err := s.DB.Exec(response, args...)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	return nil
}

func (s *Storage) LoadOrders() ([]model.Order, error) {
	query := `SELECT order_uid, track_number, entry, locale, internal_signature,
	                customer_id, delivery_service, shardkey, sm_id, date_created,
					 oof_shard FROM orders`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer rows.Close()

	var orders []model.Order

	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.OrderUid, &o.TrackNumber, &o.Entry, &o.Locale,
			&o.InternalSignature, &o.CustomerId, &o.DeliveryService,
			&o.Shardkey, &o.SmId, &o.DateCreated, &o.OofShard,
		); err != nil {
			return nil, err
		}

		// Delivery
		err = s.DB.QueryRow(`SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`, o.OrderUid).
			Scan(&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email)
		if err != nil {
			return nil, fmt.Errorf("load delivery: %w", err)
		}

		// Payment
		err = s.DB.QueryRow(`SELECT transaction, request_id, currency, provider, amount, payment_dt, 
		bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1`, o.OrderUid).
			Scan(&o.Payment.Transaction, &o.Payment.RequestId, &o.Payment.Currency, &o.Payment.Provider, &o.Payment.Amount,
				&o.Payment.PaymentDt, &o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee)
		if err != nil {
			return nil, fmt.Errorf("load payment: %w", err)
		}

		// Items
		itemRows, err := s.DB.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status 
		FROM items WHERE order_uid = $1`, o.OrderUid)
		if err != nil {
			return nil, fmt.Errorf("query items: %w", err)
		}

		for itemRows.Next() {
			var item model.Items
			if err := itemRows.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmId,
				&item.Brand, &item.Status); err != nil {

				itemRows.Close()
				return nil, err

			}
			o.Items = append(o.Items, item)
		}
		itemRows.Close()

		orders = append(orders, o)
	}
	return orders, nil
}
