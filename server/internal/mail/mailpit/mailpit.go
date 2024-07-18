package mailpit

import (
	"context"
	"fmt"
	"server/internal/pgstore"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type store interface {
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
}

type Mailpit struct {
	store store
}

func NewMailpit(pool *pgxpool.Pool) Mailpit {
	return Mailpit{pgstore.New(pool)}
}

func (mp Mailpit) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	const timeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("Mailpit: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	msg := mail.NewMsg()
	err = msg.From("mailpit@travelplanner.com")
	if err != nil {
		return fmt.Errorf("Mailpit: failed to set From in email for SendConfirmTripEmailToTripOwner: %w", err)
	}
	err = msg.To(trip.OwnerEmail)
	if err != nil {
		return fmt.Errorf("Mailpit: failed to set To in email for SendConfirmTripEmailToTripOwner: %w", err)
	}
	msg.Subject(fmt.Sprintf("Confirm your trip to %s", trip.Destination))
	body := fmt.Sprintf(`
		Hello, %s!
		
		Your trip to %s needs confirmation!
		
		Trip Details:
		ID: %s
		Destination: %s
		Starts At: %s
		Ends At: %s
		
		Best regards,
		Travel Planner"`,
		trip.OwnerName,
		trip.Destination,
		trip.ID,
		trip.Destination,
		trip.StartsAt.Time.Format(time.DateOnly),
		trip.EndsAt.Time.Format(time.DateOnly),
	)
	msg.SetBodyString(mail.TypeTextPlain, body)

	client, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return fmt.Errorf("Mailpit: failed to create mail client for SendConfirmTripEmailToTripOwner: %w", err)
	}

	err = client.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("Mailpit: failed to send e-mail message for SendCon	firmTripEmailToTripOwner: %w", err)
	}

	return nil
}
