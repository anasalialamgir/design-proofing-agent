package main

import (
	"context"
	"database/sql"
	"fmt"
)

func SendReadyNotifications(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `SELECT id, payload FROM outbox_events WHERE NOT dispatched LIMIT 10`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var payload string
		if err := rows.Scan(&id, &payload); err != nil {
			continue
		}

		// Pretend to send email
		fmt.Printf("Sending notification email for event: %s\n", payload)

		// Mark as sent
		db.ExecContext(ctx, `UPDATE outbox_events SET dispatched = TRUE WHERE id = $1`, id)
	}
	return nil
}
