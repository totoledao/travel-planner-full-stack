package email

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

type Email struct {
	store  store
	client *mail.Client
}

func NewEmail(pool *pgxpool.Pool, client *mail.Client) Email {
	return Email{pgstore.New(pool), client}
}

func (m Email) getTripDetails(tripID uuid.UUID) (pgstore.Trip, error) {
	const timeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	trip, err := m.store.GetTrip(ctx, tripID)
	if err != nil {
		return pgstore.Trip{}, err
	}

	return trip, nil
}

func (m Email) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	trip, err := m.getTripDetails(tripID)
	if err != nil {
		return fmt.Errorf("Email: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	msg := mail.NewMsg()
	err = msg.From("no-reply@travelplanner.com")
	if err != nil {
		return fmt.Errorf("Email: failed to set From in email for SendConfirmTripEmailToTripOwner: %w", err)
	}
	err = msg.To(trip.OwnerEmail)
	if err != nil {
		return fmt.Errorf("Email: failed to set To in email for SendConfirmTripEmailToTripOwner: %w", err)
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
		Travel Planner`,
		trip.OwnerName,
		trip.Destination,
		trip.ID,
		trip.Destination,
		trip.StartsAt.Time.Format(time.DateOnly),
		trip.EndsAt.Time.Format(time.DateOnly),
	)
	msg.SetBodyString(mail.TypeTextPlain, body)

	err = m.client.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("Email: failed to send e-mail message for SendConfirmTripEmailToTripOwner: %w", err)
	}

	return nil
}

func (m Email) SendInviteToTripEmail(tripID uuid.UUID, email string) error {
	trip, err := m.getTripDetails(tripID)
	if err != nil {
		return fmt.Errorf("Email: failed to get trip for SendInviteToTripEmail: %w", err)
	}

	msg := mail.NewMsg()
	err = msg.From("no-reply@travelplanner.com")
	if err != nil {
		return fmt.Errorf("Email: failed to set From in email for SendInviteToTripEmail: %w", err)
	}
	err = msg.To(email)
	if err != nil {
		return fmt.Errorf("Email: failed to set To in email for SendInviteToTripEmail: %w", err)
	}
	msg.Subject(fmt.Sprintf("You are invited on a trip to %s!", trip.Destination))
	body := fmt.Sprintf(`
		Hey!
		%s is inviting you for a trip to %s and is waiting for your confirmation!

		Trip Details:
		ID: %s
		Destination: %s
		Starts At: %s
		Ends At: %s
		
		Best regards,
		Travel Planner`,
		trip.OwnerName,
		trip.Destination,
		trip.ID,
		trip.Destination,
		trip.StartsAt.Time.Format(time.DateOnly),
		trip.EndsAt.Time.Format(time.DateOnly),
	)
	msg.SetBodyString(mail.TypeTextPlain, body)

	err = m.client.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("Email: failed to send e-mail message for SendInviteToTripEmail: %w", err)
	}

	return nil
}
