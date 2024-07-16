package pgstore

import (
	"context"
	"fmt"
	"server/internal/api/spec"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (q *Queries) CreateTrip(ctx context.Context, pool *pgxpool.Pool, params spec.CreateTripRequest) (uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to begin tx for CreateTrip: %w", err)
	}
	// Guarantees that connections is closed
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)
	tripID, err := qtx.InsertTrip(ctx, InsertTripParams{
		Destination: params.Destination,
		OwnerEmail:  string(params.OwnerEmail),
		OwnerName:   params.OwnerName,
		StartsAt:    pgtype.Timestamp{Valid: true, Time: params.StartsAt},
		EndsAt:      pgtype.Timestamp{Valid: true, Time: params.EndsAt},
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert Trip for CreateTrip: %w", err)
	}

	participants := make([]InviteParticipantsToTripParams, len(params.EmailsToInvite))
	for i, email := range params.EmailsToInvite {
		participants[i] = InviteParticipantsToTripParams{
			TripID: tripID,
			Email:  string(email),
		}
	}

	_, err = qtx.InviteParticipantsToTrip(ctx, participants)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert Participants for CreateTrip: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to commit tx for CreateTrip: %w", err)
	}

	return tripID, nil
}
