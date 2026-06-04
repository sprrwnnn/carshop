package pgdb

import (
	dberrors "carshop/internal/database/errors"
	dbtypes "carshop/internal/database/types"
	"carshop/internal/events"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
)

const (
	carCreatedEventType = "car.created"
	carUpdatedEventType = "car.updated"
	carDeletedEventType = "car.deleted"
)

type storedCarEvent struct {
	EventType string
	Payload   []byte
}

type carProjection struct {
	ID        dbtypes.CarID
	Name      string
	Colour    string
	Price     float64
	BuildDate time.Time
	DeletedAt *time.Time
	Created   bool
}

func (c *client) ExecuteCreateCarCommand(
	ctx context.Context,
	req *dbtypes.CarCreateCommand,
) (*dbtypes.CarCreateCommandResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", dberrors.ErrBadRequest)
	}

	conn, err := c.pool.PeekWriteConn()
	if err != nil {
		return nil, fmt.Errorf("%w: error getting write connection: %v", dberrors.ErrInternal, err)
	}

	defer conn.Release()

	var id dbtypes.CarID

	if err := conn.QueryRow(ctx, "select nextval('core.car_id_seq')").Scan(&id); err != nil {
		return nil, fmt.Errorf("%w: error generating car id: %v", dberrors.ErrInternal, err)
	}

	event := events.CarCreatedEvent{
		ID:        id,
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: req.BuildDate.Format(time.RFC3339),
	}

	if err := c.appendCarEvent(ctx, conn, id, carCreatedEventType, event); err != nil {
		return nil, err
	}

	return &dbtypes.CarCreateCommandResponse{ID: id}, nil
}

func (c *client) ExecuteUpdateCarCommand(
	ctx context.Context,
	req *dbtypes.CarUpdateCommand,
) (*dbtypes.CarUpdateCommandResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", dberrors.ErrBadRequest)
	}

	conn, err := c.pool.PeekWriteConn()
	if err != nil {
		return nil, fmt.Errorf("%w: error getting write connection: %v", dberrors.ErrInternal, err)
	}

	defer conn.Release()

	exists, err := c.carAggregateExists(ctx, conn, req.ID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, dberrors.ErrNotFound
	}

	var buildDate *string
	if req.BuildDate != nil {
		formatted := req.BuildDate.Format(time.RFC3339)
		buildDate = &formatted
	}

	event := events.CarUpdatedEvent{
		ID:        req.ID,
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: buildDate,
	}

	if err := c.appendCarEvent(ctx, conn, req.ID, carUpdatedEventType, event); err != nil {
		return nil, err
	}

	return &dbtypes.CarUpdateCommandResponse{}, nil
}

func (c *client) ExecuteDeleteCarCommand(
	ctx context.Context,
	req *dbtypes.CarDeleteCommand,
) (*dbtypes.CarDeleteCommandResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", dberrors.ErrBadRequest)
	}

	conn, err := c.pool.PeekWriteConn()
	if err != nil {
		return nil, fmt.Errorf("%w: error getting write connection: %v", dberrors.ErrInternal, err)
	}

	defer conn.Release()

	exists, err := c.carAggregateExists(ctx, conn, req.ID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, dberrors.ErrNotFound
	}

	event := events.CarDeletedEvent{
		ID:        req.ID,
		DeletedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := c.appendCarEvent(ctx, conn, req.ID, carDeletedEventType, event); err != nil {
		return nil, err
	}

	return &dbtypes.CarDeleteCommandResponse{}, nil
}

func (c *client) RebuildCarReadModels(ctx context.Context) error {
	conn, err := c.pool.PeekWriteConn()
	if err != nil {
		return fmt.Errorf("%w: error getting write connection: %v", dberrors.ErrInternal, err)
	}

	defer conn.Release()

	events, err := c.getCarEvents(ctx, conn)
	if err != nil {
		return err
	}

	projections := make(map[dbtypes.CarID]carProjection)

	for _, event := range events {
		if err := applyStoredCarEvent(projections, event); err != nil {
			return err
		}
	}

	if _, err := conn.Exec(ctx, "truncate table core.car_read_models"); err != nil {
		return fmt.Errorf("%w: error truncating car read models: %v", dberrors.ErrInternal, err)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	for _, projection := range projections {
		if !projection.Created {
			continue
		}

		insert := psql.Insert(carReadModelsTableName).
			Columns("id", "name", "colour", "price", "build_date", "deleted_at").
			Values(
				projection.ID,
				projection.Name,
				projection.Colour,
				projection.Price,
				projection.BuildDate,
				projection.DeletedAt,
			)

		query, args, err := insert.ToSql()
		if err != nil {
			return fmt.Errorf("%w: error building read model insert: %v", dberrors.ErrInternal, err)
		}

		if _, err := conn.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("%w: error inserting car read model: %v", dberrors.ErrInternal, err)
		}
	}

	return nil
}

func (c *client) appendCarEvent(
	ctx context.Context,
	driver Driver,
	aggregateID dbtypes.CarID,
	eventType string,
	payload any,
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: error marshalling car event: %v", dberrors.ErrInternal, err)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	insert := psql.Insert(carEventsTableName).
		Columns("aggregate_id", "event_type", "payload").
		Values(aggregateID, eventType, squirrel.Expr("?::jsonb", string(body)))

	query, args, err := insert.ToSql()
	if err != nil {
		return fmt.Errorf("%w: error building query: %v", dberrors.ErrInternal, err)
	}

	if _, err := driver.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("%w: error executing query: %v", dberrors.ErrInternal, err)
	}

	return nil
}

func (c *client) carAggregateExists(ctx context.Context, driver Driver, aggregateID dbtypes.CarID) (bool, error) {
	events, err := c.getCarEvents(ctx, driver, aggregateID)
	if err != nil {
		return false, err
	}

	projections := make(map[dbtypes.CarID]carProjection)

	for _, event := range events {
		if err := applyStoredCarEvent(projections, event); err != nil {
			return false, err
		}
	}

	projection, ok := projections[aggregateID]

	return ok && projection.Created && projection.DeletedAt == nil, nil
}

func (c *client) getCarEvents(
	ctx context.Context,
	driver Driver,
	aggregateID ...dbtypes.CarID,
) ([]storedCarEvent, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	getEvents := psql.Select(
		"event_type", "payload",
	).
		From(carEventsTableName).
		OrderBy("sequence_id asc")

	if len(aggregateID) > 0 {
		getEvents = getEvents.Where(squirrel.Eq{"aggregate_id": aggregateID[0]})
	}

	query, args, err := getEvents.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: error building query: %v", dberrors.ErrInternal, err)
	}

	rows, err := driver.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: error executing query: %v", dberrors.ErrInternal, err)
	}

	defer rows.Close()

	events := make([]storedCarEvent, 0)

	for rows.Next() {
		var event storedCarEvent

		if err := rows.Scan(&event.EventType, &event.Payload); err != nil {
			return nil, fmt.Errorf("%w: error scanning row: %v", dberrors.ErrInternal, err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: error iterating rows: %v", dberrors.ErrInternal, err)
	}

	return events, nil
}

func applyStoredCarEvent(projections map[dbtypes.CarID]carProjection, event storedCarEvent) error {
	switch event.EventType {
	case carCreatedEventType:
		var payload events.CarCreatedEvent
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("%w: error unmarshalling car created event: %v", dberrors.ErrInternal, err)
		}

		buildDate, err := parseEventDate(payload.BuildDate)
		if err != nil {
			return err
		}

		projections[payload.ID] = carProjection{
			ID:        payload.ID,
			Name:      payload.Name,
			Colour:    payload.Colour,
			Price:     payload.Price,
			BuildDate: buildDate,
			Created:   true,
		}

	case carUpdatedEventType:
		var payload events.CarUpdatedEvent
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("%w: error unmarshalling car updated event: %v", dberrors.ErrInternal, err)
		}

		projection, ok := projections[payload.ID]
		if !ok || !projection.Created || projection.DeletedAt != nil {
			return nil
		}

		if payload.Name != nil {
			projection.Name = *payload.Name
		}

		if payload.Colour != nil {
			projection.Colour = *payload.Colour
		}

		if payload.Price != nil {
			projection.Price = *payload.Price
		}

		if payload.BuildDate != nil {
			buildDate, err := parseEventDate(*payload.BuildDate)
			if err != nil {
				return err
			}

			projection.BuildDate = buildDate
		}

		projections[payload.ID] = projection

	case carDeletedEventType:
		var payload events.CarDeletedEvent
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("%w: error unmarshalling car deleted event: %v", dberrors.ErrInternal, err)
		}

		projection, ok := projections[payload.ID]
		if !ok || !projection.Created || projection.DeletedAt != nil {
			return nil
		}

		deletedAt, err := time.Parse(time.RFC3339, payload.DeletedAt)
		if err != nil {
			return fmt.Errorf("%w: invalid deleted_at %q", dberrors.ErrInternal, payload.DeletedAt)
		}

		projection.DeletedAt = &deletedAt
		projections[payload.ID] = projection

	default:
		return fmt.Errorf("%w: unknown car event type %q", dberrors.ErrInternal, event.EventType)
	}

	return nil
}

func parseEventDate(value string) (time.Time, error) {
	res, err := time.Parse(time.DateOnly, value)
	if err == nil {
		return res, nil
	}

	res, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return res, nil
	}

	return time.Time{}, fmt.Errorf("%w: invalid event date %q", dberrors.ErrInternal, value)
}
