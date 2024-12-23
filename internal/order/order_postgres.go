package order

import (
	"context"
	"errors"
	"fmt"

	"l0/pkg/client/postgres"
)

func insertOrder(pool postgres.Pool, order Order) error {
	if exist, err := CheckRecordExists(pool, order.OrderUID); err != nil {
		return errors.New("Error check record exist")
	} else if exist {
		return errors.New(fmt.Sprintf("Record with uid = %s already exist", order.OrderUID))
	}

	// Вставка доставки
	var deliveryID int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO deliveries (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	).Scan(&deliveryID)
	if err != nil {
		return err
	}

	// Вставка платежа
	var paymentID int
	err = pool.QueryRow(context.Background(),
		`INSERT INTO payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	).Scan(&paymentID)
	if err != nil {
		return err
	}

	// Вставка заказа
	_, err = pool.Exec(context.Background(),
		`INSERT INTO orders (order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		deliveryID,
		paymentID,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SMID,
		order.DateCreated,
		order.OOFShard,
	)
	if err != nil {
		return err
	}

	// Вставка элементов заказа
	for _, item := range order.Items {
		_, err = pool.Exec(context.Background(),
			`INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.RID,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NMID,
			item.Brand,
			item.Status,
			order.OrderUID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetOrderFromDB(pool postgres.Pool, uid string) (order Order, err error) {

	paymentIdQuery := "SELECT payment_id FROM orders where order_uid = $1"
	var paymentID int
	var deliveryID int
	err = pool.QueryRow(context.Background(), paymentIdQuery, uid).Scan(&paymentID)
	if err != nil {
		return order, err
	}

	paymentQuery := "SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE id = $1"
	err = pool.QueryRow(context.Background(), paymentQuery, paymentID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)

	if err != nil {
		return order, err
	}

	deliveryIdQuery := "SELECT delivery_id FROM orders where order_uid = $1"

	err = pool.QueryRow(context.Background(), deliveryIdQuery, uid).Scan(&deliveryID)
	if err != nil {
		return order, err

	}

	deliveryQuery := "SELECT  name, phone, zip, city, address, region, email FROM deliveries WHERE id = $1"
	err = pool.QueryRow(context.Background(), deliveryQuery, deliveryID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return order, err

	}

	orderQuery := "SELECT order_uid, track_number, entry, locale, internal_signature,customer_id,delivery_service,shardkey,sm_id,date_created,oof_shard FROM orders where order_uid = $1"

	err = pool.QueryRow(context.Background(), orderQuery, uid).Scan(&order.OrderUID, &order.TrackNumber, &order.Entry,
		&order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SMID, &order.DateCreated, &order.OOFShard)

	if err != nil {
		return order, err

	}

	itemsQuery := "SELECT  chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1"
	rows, err := pool.Query(context.Background(), itemsQuery, uid)
	if err != nil {
		return order, err

	}
	defer rows.Close()

	for rows.Next() {
		var item Items
		if err := rows.Scan(
			&item.ChrtID, &item.TrackNumber,
			&item.Price, &item.RID, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
		); err != nil {
			return order, err

		}
		order.Items = append(order.Items, item)
	}
	return order, nil

}

func GetAllOrdersFromDB(conn postgres.Pool) ([]Order, error) {
	var orders []Order

	// Запрос для получения всех заказов
	orderQuery := `
	SELECT order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature,
	customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	FROM orders`

	rows, err := conn.Query(context.Background(), orderQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		var deliveryID, paymentID int

		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &deliveryID, &paymentID,
			&order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SMID, &order.DateCreated, &order.OOFShard,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Получаем данные о доставке
		deliveryQuery := "SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE id = $1"
		err = conn.QueryRow(context.Background(), deliveryQuery, deliveryID).Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get delivery for order %s: %w", order.OrderUID, err)
		}

		// Получаем данные о платеже
		paymentQuery := "SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE id = $1"
		err = conn.QueryRow(context.Background(), paymentQuery, paymentID).Scan(
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
			&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment for order %s: %w", order.OrderUID, err)
		}

		// Получаем элементы заказа
		itemsQuery := "SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1"
		itemRows, err := conn.Query(context.Background(), itemsQuery, order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get items for order %s: %w", order.OrderUID, err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item Items
			if err := itemRows.Scan(
				&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
				&item.Sale, &item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
			); err != nil {
				return nil, fmt.Errorf("failed to scan item for order %s: %w", order.OrderUID, err)
			}
			order.Items = append(order.Items, item)
		}

		// Добавляем заказ в слайс
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", err)
	}

	return orders, nil
}
func CheckRecordExists(pool postgres.Pool, fieldValue string) (bool, error) {
	var exists bool
	// SQL-запрос для проверки наличия записи
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)"

	// Выполнение запроса
	err := pool.QueryRow(context.Background(), query, fieldValue).Scan(&exists)
	if err != nil {
		return false, err // Обработка ошибки
	}
	return exists, nil // Возвращаем результат
}
