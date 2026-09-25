package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ProofService struct {
	DB *sql.DB
}

func (s *ProofService) SaveDecision(ctx context.Context, submissionID string, decision string, email string, notes string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the row to prevent conflicting updates
	var currentStatus string
	query := `SELECT status FROM artwork_submissions WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, submissionID).Scan(&currentStatus); err != nil {
		return err
	}

	if currentStatus != "AWAITING_APPROVAL" {
		return fmt.Errorf("this proof cannot be changed because its status is %s", currentStatus)
	}

	update := `UPDATE artwork_submissions SET status = $1, updated_at = $2 WHERE id = $3`
	if _, err := tx.ExecContext(ctx, update, decision, time.Now(), submissionID); err != nil {
		return err
	}

	return tx.Commit()
}
