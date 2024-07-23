package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"server/internal/api/spec"
	"server/internal/pgstore"

	"github.com/discord-gophers/goapi-gen/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type store interface {
	ConfirmParticipant(ctx context.Context, id uuid.UUID) error
	CreateActivity(ctx context.Context, arg pgstore.CreateActivityParams) (uuid.UUID, error)
	CreateTripLink(ctx context.Context, arg pgstore.CreateTripLinkParams) (uuid.UUID, error)
	GetParticipant(ctx context.Context, id uuid.UUID) (pgstore.Participant, error)
	GetParticipants(ctx context.Context, tripID uuid.UUID) ([]pgstore.Participant, error)
	GetTrip(ctx context.Context, id uuid.UUID) (pgstore.Trip, error)
	GetTripActivities(ctx context.Context, tripID uuid.UUID) ([]pgstore.Activity, error)
	GetTripLinks(ctx context.Context, tripID uuid.UUID) ([]pgstore.Link, error)
	InsertTrip(ctx context.Context, arg pgstore.InsertTripParams) (uuid.UUID, error)
	InviteParticipantToTrip(ctx context.Context, arg pgstore.InviteParticipantToTripParams) (uuid.UUID, error)
	UpdateTrip(ctx context.Context, arg pgstore.UpdateTripParams) error
	CreateTrip(ctx context.Context, pool *pgxpool.Pool, params spec.CreateTripRequest) (uuid.UUID, error)
}

type mailer interface {
	SendConfirmTripEmailToTripOwner(uuid.UUID) error
	SendInviteToTripEmail(uuid.UUID, string) error
}

type API struct {
	store     store
	logger    *zap.Logger
	validator *validator.Validate
	pool      *pgxpool.Pool
	mailer    mailer
}

func NewAPI(pool *pgxpool.Pool, logger *zap.Logger, mailer mailer) API {
	validator := validator.New(validator.WithRequiredStructEnabled())
	return API{pgstore.New(pool), logger, validator, pool, mailer}
}

// Confirms a participant on a trip.
// (PATCH /participants/{participantId}/confirm)
func (api *API) PatchParticipantsParticipantIDConfirm(w http.ResponseWriter, r *http.Request, participantID string) *spec.Response {
	id, err := uuid.Parse(participantID)
	if err != nil {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	participant, err := api.store.GetParticipant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Participant not found"})
		}

		api.logger.Error("Failed to get participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong finding participant, try again"})
	}

	if participant.IsConfirmed {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Participant already confirmed"})
	}

	err = api.store.ConfirmParticipant(r.Context(), id)
	if err != nil {
		api.logger.Error("Failed to confirm participant", zap.Error(err), zap.String("participant_ id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong confirming participant, try again"})
	}

	return spec.PatchParticipantsParticipantIDConfirmJSON204Response(nil)
}

// Create a new trip
// (POST /trips)
func (api *API) PostTrips(w http.ResponseWriter, r *http.Request) *spec.Response {
	var body spec.CreateTripRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid JSON"})
	}

	err = api.validator.Struct(body)
	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}

	tripID, err := api.store.CreateTrip(r.Context(), api.pool, body)
	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Failed to create trip, try again"})
	}

	go func() {
		err := api.mailer.SendConfirmTripEmailToTripOwner(tripID)
		if err != nil {
			api.logger.Error(
				"failed to send email on PostTrips",
				zap.Error(err),
				zap.String("trip_id", tripID.String()),
			)
		}
	}()

	return spec.PostTripsJSON201Response(spec.CreateTripResponse{TripID: tripID.String()})
}

// Get a trip details.
// (GET /trips/{tripId})
func (api *API) GetTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Trip not found"})
		}

		api.logger.Error("Failed to get trip", zap.Error(err), zap.String("trip_id", trip.ID.String()))
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Something went wrong finding trip, try again"})
	}

	return spec.GetTripsTripIDJSON200Response(spec.GetTripDetailsResponse{
		Trip: spec.GetTripDetailsResponseTripObj{
			ID:          trip.ID.String(),
			Destination: trip.Destination,
			StartsAt:    trip.StartsAt.Time,
			EndsAt:      trip.EndsAt.Time,
			IsConfirmed: trip.IsConfirmed,
		},
	})
}

// Update a trip.
// (PUT /trips/{tripId})
func (api *API) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	var body spec.PutTripsTripIDJSONBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid JSON"})
	}

	err = api.validator.Struct(body)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}

	err = api.store.UpdateTrip(r.Context(),
		pgstore.UpdateTripParams{
			Destination: body.Destination,
			EndsAt:      pgtype.Timestamp{Valid: true, Time: body.EndsAt},
			StartsAt:    pgtype.Timestamp{Valid: true, Time: body.StartsAt},
			IsConfirmed: false,
			ID:          id,
		},
	)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Failed to update trip, try again"})
	}

	return spec.PutTripsTripIDJSON204Response(nil)
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (api *API) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	activities, err := api.store.GetTripActivities(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Activities not found"})
		}

		api.logger.Error("Failed to get activities from trip", zap.Error(err), zap.String("trip_id", id.String()))
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Something went wrong finding activities from trip, try again"})
	}

	var response spec.GetTripActivitiesResponse
	for _, act := range activities {
		response.Activities = append(response.Activities, spec.GetTripActivitiesResponseOuterArray{
			Date: act.OccursAt.Time,
			Activities: []spec.GetTripActivitiesResponseInnerArray{
				{
					ID:       act.ID.String(),
					OccursAt: act.OccursAt.Time,
					Title:    act.Title,
				},
			},
		})
	}

	return spec.GetTripsTripIDActivitiesJSON200Response(response)
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (api *API) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	var body spec.PostTripsTripIDActivitiesJSONBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid JSON"})
	}

	err = api.validator.Struct(body)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}

	actID, err := api.store.CreateActivity(r.Context(),
		pgstore.CreateActivityParams{
			TripID:   id,
			Title:    body.Title,
			OccursAt: pgtype.Timestamp{Valid: true, Time: body.OccursAt},
		},
	)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Failed to create activity for trip, try again"})
	}

	return spec.PostTripsTripIDActivitiesJSON201Response(spec.CreateActivityResponse{ActivityID: actID.String()})
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (api *API) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Trip not found"})
		}

		api.logger.Error("Failed to get trip", zap.Error(err), zap.String("trip_id", trip.ID.String()))
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Something went wrong finding trip, try again"})
	}

	if trip.IsConfirmed {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Trip already confirmed"})
	}

	err = api.store.UpdateTrip(r.Context(),
		pgstore.UpdateTripParams{
			Destination: trip.Destination,
			EndsAt:      trip.EndsAt,
			StartsAt:    trip.StartsAt,
			IsConfirmed: true,
			ID:          id,
		},
	)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Failed to update trip for confirmation, try again"})
	}

	participants, err := api.store.GetParticipants(r.Context(), id)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Failed to get participants for trip, try again"})
	}

	go func() {
		sem := make(chan struct{}, 1)

		for _, v := range participants {
			sem <- struct{}{} // Acquire a slot
			go func(email string) {
				defer func() { <-sem }() // Release the slot
				err := api.mailer.SendInviteToTripEmail(id, email)
				if err != nil {
					api.logger.Error(
						"failed to send email on GetTripsTripIDConfirm",
						zap.Error(err),
						zap.String("trip_id", tripID),
					)
				}
			}(string(v.Email))
		}

		// Wait for all goroutines to finish
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
	}()

	return spec.GetTripsTripIDConfirmJSON204Response(nil)
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (api *API) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	var body spec.PostTripsTripIDInvitesJSONBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid JSON"})
	}

	err = api.validator.Struct(body)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}

	_, err = api.store.InviteParticipantToTrip(r.Context(),
		pgstore.InviteParticipantToTripParams{
			TripID: id,
			Email:  string(body.Email),
		},
	)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Failed to invite to trip, try again"})
	}

	go func() {
		err := api.mailer.SendInviteToTripEmail(id, string(body.Email))
		if err != nil {
			api.logger.Error(
				"failed to send email on PostTripsTripIDInvites",
				zap.Error(err),
				zap.String("trip_id", tripID),
			)
		}
	}()

	return spec.PostTripsTripIDInvitesJSON201Response(nil)
}

// Get a trip links.
// (GET /trips/{tripId}/links)
func (api *API) GetTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDLinksJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	links, err := api.store.GetTripLinks(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDLinksJSON400Response(spec.Error{Message: "Links not found"})
		}

		api.logger.Error("Failed to get links from trip", zap.Error(err), zap.String("trip_id", id.String()))
		return spec.GetTripsTripIDLinksJSON400Response(spec.Error{Message: "Something went wrong finding links from trip, try again"})
	}

	var response spec.GetLinksResponse
	for _, l := range links {
		response.Links = append(response.Links, spec.GetLinksResponseArray{
			ID:    l.ID.String(),
			Title: l.Title,
			URL:   l.Url,
		},
		)
	}

	return spec.GetTripsTripIDLinksJSON200Response(response)
}

// Create a trip link.
// (POST /trips/{tripId}/links)
func (api *API) PostTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	var body spec.PostTripsTripIDLinksJSONBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Invalid JSON"})
	}

	err = api.validator.Struct(body)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}

	linkID, err := api.store.CreateTripLink(r.Context(),
		pgstore.CreateTripLinkParams{
			TripID: id,
			Title:  body.Title,
			Url:    body.URL,
		},
	)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Failed to add link to trip, try again"})
	}

	return spec.PostTripsTripIDLinksJSON201Response(spec.CreateLinkResponse{LinkID: linkID.String()})
}

// Get a trip participants.
// (GET /trips/{tripId}/participants)
func (api *API) GetTripsTripIDParticipants(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "invalid uuid"})
	}

	participants, err := api.store.GetParticipants(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Participants not found"})
		}

		api.logger.Error("Failed to get participants from trip", zap.Error(err), zap.String("trip_id", id.String()))
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Something went wrong finding participants from trip, try again"})
	}

	var response spec.GetTripParticipantsResponse
	for _, p := range participants {
		response.Participants = append(response.Participants, spec.GetTripParticipantsResponseArray{
			Email:       types.Email(p.Email),
			ID:          p.ID.String(),
			IsConfirmed: p.IsConfirmed,
		},
		)
	}

	return spec.GetTripsTripIDParticipantsJSON200Response(response)
}
