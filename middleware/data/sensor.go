package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"middleware/model"

	_ "github.com/lib/pq"
)

const (
	ReadTypeAnalog   = "analog"
	ReadTypeDiscrete = "discrete"
)

type SensorRepository struct {
	db *sql.DB
}

func NewSensorRepository(db *sql.DB) *SensorRepository {
	return &SensorRepository{db: db}
}

func (r *SensorRepository) SaveReading(ctx context.Context, msg model.SensorMessage) error {
	deviceID, err := r.resolveDeviceID(ctx, msg.IDSensor)
	if err != nil {
		return fmt.Errorf("resolving device_id for sensor %d: %w", msg.IDSensor, err)
	}

	sensorTypeID, err := r.resolveSensorTypeID(ctx, msg.SensorType)
	if err != nil {
		return fmt.Errorf("resolving sensor_type_id for %q: %w", msg.SensorType, err)
	}

	collectedAt, err := time.Parse(time.RFC3339, msg.Timestamp)
	if err != nil {
		return fmt.Errorf("parsing timestamp %q: %w", msg.Timestamp, err)
	}

	switch msg.ReadType {
	case ReadTypeAnalog:
		return r.saveAnalogReading(ctx, deviceID, sensorTypeID, msg.Value, collectedAt)
	case ReadTypeDiscrete:
		return r.saveDiscreteReading(ctx, deviceID, sensorTypeID, int(msg.Value), collectedAt)
	default:
		return fmt.Errorf("unknown read_type %q: must be %q or %q", msg.ReadType, ReadTypeAnalog, ReadTypeDiscrete)
	}
}

func (r *SensorRepository) SaveReadingTx(ctx context.Context, tx *sql.Tx, msg model.SensorMessage) error {
	deviceID, err := r.resolveDeviceIDTx(ctx, tx, msg.IDSensor)
	if err != nil {
		return fmt.Errorf("resolving device_id for sensor %d: %w", msg.IDSensor, err)
	}

	sensorTypeID, err := r.resolveSensorTypeIDTx(ctx, tx, msg.SensorType)
	if err != nil {
		return fmt.Errorf("resolving sensor_type_id for %q: %w", msg.SensorType, err)
	}

	collectedAt, err := time.Parse(time.RFC3339, msg.Timestamp)
	if err != nil {
		return fmt.Errorf("parsing timestamp %q: %w", msg.Timestamp, err)
	}

	switch msg.ReadType {
	case ReadTypeAnalog:
		return r.saveAnalogReadingTx(ctx, tx, deviceID, sensorTypeID, msg.Value, collectedAt)
	case ReadTypeDiscrete:
		return r.saveDiscreteReadingTx(ctx, tx, deviceID, sensorTypeID, int(msg.Value), collectedAt)
	default:
		return fmt.Errorf("unknown read_type %q", msg.ReadType)
	}
}

func (r *SensorRepository) SaveReadingsBatch(ctx context.Context, msgs []model.SensorMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for i, msg := range msgs {
		if err = r.SaveReadingTx(ctx, tx, msg); err != nil {
			return fmt.Errorf("saving message at index %d: %w", i, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func (r *SensorRepository) saveAnalogReading(
	ctx context.Context,
	deviceID string,
	sensorTypeID int,
	value float64,
	collectedAt time.Time,
) error {
	const query = `
		INSERT INTO analog_readings (device_id, sensor_type_id, value, collected_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, deviceID, sensorTypeID, value, collectedAt)
	if err != nil {
		return fmt.Errorf("inserting analog reading: %w", err)
	}
	return nil
}

func (r *SensorRepository) saveAnalogReadingTx(
	ctx context.Context,
	tx *sql.Tx,
	deviceID string,
	sensorTypeID int,
	value float64,
	collectedAt time.Time,
) error {
	const query = `
		INSERT INTO analog_readings (device_id, sensor_type_id, value, collected_at)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.ExecContext(ctx, query, deviceID, sensorTypeID, value, collectedAt)
	if err != nil {
		return fmt.Errorf("inserting analog reading (tx): %w", err)
	}
	return nil
}

func (r *SensorRepository) saveDiscreteReading(
	ctx context.Context,
	deviceID string,
	sensorTypeID int,
	value int,
	collectedAt time.Time,
) error {
	const query = `
		INSERT INTO discrete_readings (device_id, sensor_type_id, value, collected_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, deviceID, sensorTypeID, value, collectedAt)
	if err != nil {
		return fmt.Errorf("inserting discrete reading: %w", err)
	}
	return nil
}

func (r *SensorRepository) saveDiscreteReadingTx(
	ctx context.Context,
	tx *sql.Tx,
	deviceID string,
	sensorTypeID int,
	value int,
	collectedAt time.Time,
) error {
	const query = `
		INSERT INTO discrete_readings (device_id, sensor_type_id, value, collected_at)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.ExecContext(ctx, query, deviceID, sensorTypeID, value, collectedAt)
	if err != nil {
		return fmt.Errorf("inserting discrete reading (tx): %w", err)
	}
	return nil
}

func (r *SensorRepository) resolveDeviceID(ctx context.Context, idSensor string) (string, error) {
	const query = `SELECT id FROM devices WHERE serial = $1 AND active = TRUE`
	serial := idSensor

	var id string
	err := r.db.QueryRowContext(ctx, query, serial).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("device with serial %q not found or inactive", serial)
	}
	if err != nil {
		return "", fmt.Errorf("querying device: %w", err)
	}
	return id, nil
}

func (r *SensorRepository) resolveDeviceIDTx(ctx context.Context, tx *sql.Tx, idSensor string) (string, error) {
	const query = `SELECT id FROM devices WHERE serial = $1 AND active = TRUE`
	serial := idSensor

	var id string
	err := tx.QueryRowContext(ctx, query, serial).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("device with serial %q not found or inactive", serial)
	}
	if err != nil {
		return "", fmt.Errorf("querying device (tx): %w", err)
	}
	return id, nil
}

func (r *SensorRepository) resolveSensorTypeID(ctx context.Context, sensorType string) (int, error) {
	const query = `SELECT id FROM sensor_types WHERE name = $1`

	var id int
	err := r.db.QueryRowContext(ctx, query, sensorType).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("sensor type %q not found", sensorType)
	}
	if err != nil {
		return 0, fmt.Errorf("querying sensor_type: %w", err)
	}
	return id, nil
}

func (r *SensorRepository) resolveSensorTypeIDTx(ctx context.Context, tx *sql.Tx, sensorType string) (int, error) {
	const query = `SELECT id FROM sensor_types WHERE name = $1`

	var id int
	err := tx.QueryRowContext(ctx, query, sensorType).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("sensor type %q not found", sensorType)
	}
	if err != nil {
		return 0, fmt.Errorf("querying sensor_type (tx): %w", err)
	}
	return id, nil
}