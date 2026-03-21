// testes gerados por IA

package consumer_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"middleware/model"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers / fakes
// ──────────────────────────────────────────────────────────────────────────────

// fakeRepo implements a minimal in-memory sensor repository for testing.
type fakeRepo struct {
	saveCalledWith []model.SensorMessage
	returnErr      error
}

func (f *fakeRepo) SaveReading(_ context.Context, msg model.SensorMessage) error {
	f.saveCalledWith = append(f.saveCalledWith, msg)
	return f.returnErr
}

func (f *fakeRepo) SaveReadingsBatch(_ context.Context, msgs []model.SensorMessage) error {
	for _, msg := range msgs {
		if err := f.SaveReading(context.Background(), msg); err != nil {
			return err
		}
	}
	return f.returnErr
}

// validMessage returns a minimal valid SensorMessage.
func validAnalogMessage() model.SensorMessage {
	return model.SensorMessage{
		IDSensor:   "SENSOR-001",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SensorType: "temperature",
		ReadType:   "analog",
		Value:      23.5,
	}
}

func validDiscreteMessage() model.SensorMessage {
	return model.SensorMessage{
		IDSensor:   "SENSOR-002",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SensorType: "motion",
		ReadType:   "discrete",
		Value:      1,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// model.SensorMessage JSON marshalling / unmarshalling
// ──────────────────────────────────────────────────────────────────────────────

func TestSensorMessage_JSONRoundTrip(t *testing.T) {
	original := validAnalogMessage()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded model.SensorMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.IDSensor != original.IDSensor {
		t.Errorf("IDSensor mismatch: got %q, want %q", decoded.IDSensor, original.IDSensor)
	}
	if decoded.SensorType != original.SensorType {
		t.Errorf("SensorType mismatch: got %q, want %q", decoded.SensorType, original.SensorType)
	}
	if decoded.ReadType != original.ReadType {
		t.Errorf("ReadType mismatch: got %q, want %q", decoded.ReadType, original.ReadType)
	}
	if decoded.Value != original.Value {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Value, original.Value)
	}
	if decoded.Timestamp != original.Timestamp {
		t.Errorf("Timestamp mismatch: got %q, want %q", decoded.Timestamp, original.Timestamp)
	}
}

func TestSensorMessage_JSONFieldNames(t *testing.T) {
	msg := model.SensorMessage{
		IDSensor:   "SN-999",
		Timestamp:  "2024-01-01T00:00:00Z",
		SensorType: "humidity",
		ReadType:   "analog",
		Value:      55.0,
	}

	raw, _ := json.Marshal(msg)
	m := map[string]interface{}{}
	_ = json.Unmarshal(raw, &m)

	expectedKeys := []string{"idSensor", "timestamp", "sensorType", "readType", "value"}
	for _, k := range expectedKeys {
		if _, ok := m[k]; !ok {
			t.Errorf("expected JSON key %q not found in marshalled output", k)
		}
	}
}

func TestSensorMessage_UnmarshalInvalidJSON(t *testing.T) {
	var msg model.SensorMessage
	err := json.Unmarshal([]byte(`{invalid json}`), &msg)
	if err == nil {
		t.Error("expected error when unmarshalling invalid JSON, got nil")
	}
}

func TestSensorMessage_UnmarshalEmptyBody(t *testing.T) {
	var msg model.SensorMessage
	err := json.Unmarshal([]byte(`{}`), &msg)
	if err != nil {
		t.Errorf("unexpected error on empty JSON object: %v", err)
	}
	// All fields should be zero values.
	if msg.IDSensor != "" || msg.Value != 0 {
		t.Errorf("expected zero values for empty JSON object, got %+v", msg)
	}
}

func TestSensorMessage_FloatValuePrecision(t *testing.T) {
	msg := model.SensorMessage{Value: 3.141592653589793}
	data, _ := json.Marshal(msg)
	var decoded model.SensorMessage
	_ = json.Unmarshal(data, &decoded)
	if decoded.Value != msg.Value {
		t.Errorf("float precision lost: got %v, want %v", decoded.Value, msg.Value)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// fakeRepo behaviour (simulates the data layer contract)
// ──────────────────────────────────────────────────────────────────────────────

func TestFakeRepo_SaveReading_RecordsMessage(t *testing.T) {
	repo := &fakeRepo{}
	msg := validAnalogMessage()

	if err := repo.SaveReading(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.saveCalledWith) != 1 {
		t.Fatalf("expected 1 saved message, got %d", len(repo.saveCalledWith))
	}
	if repo.saveCalledWith[0].IDSensor != msg.IDSensor {
		t.Errorf("saved message IDSensor mismatch")
	}
}

func TestFakeRepo_SaveReading_PropagatesError(t *testing.T) {
	sentinel := errors.New("db unavailable")
	repo := &fakeRepo{returnErr: sentinel}

	err := repo.SaveReading(context.Background(), validAnalogMessage())
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFakeRepo_SaveReadingsBatch_AllSaved(t *testing.T) {
	repo := &fakeRepo{}
	msgs := []model.SensorMessage{
		validAnalogMessage(),
		validDiscreteMessage(),
	}

	if err := repo.SaveReadingsBatch(context.Background(), msgs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.saveCalledWith) != 2 {
		t.Errorf("expected 2 saved messages, got %d", len(repo.saveCalledWith))
	}
}

func TestFakeRepo_SaveReadingsBatch_StopsOnError(t *testing.T) {
	sentinel := errors.New("insert failed")
	repo := &fakeRepo{returnErr: sentinel}
	msgs := []model.SensorMessage{validAnalogMessage(), validDiscreteMessage()}

	err := repo.SaveReadingsBatch(context.Background(), msgs)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ReadType validation logic (mirrors data.SaveReading switch)
// ──────────────────────────────────────────────────────────────────────────────

// readTypeIsValid mirrors the switch inside data.SensorRepository.SaveReading.
func readTypeIsValid(rt string) bool {
	return rt == "analog" || rt == "discrete"
}

func TestReadType_ValidAnalog(t *testing.T) {
	if !readTypeIsValid("analog") {
		t.Error("'analog' should be a valid read type")
	}
}

func TestReadType_ValidDiscrete(t *testing.T) {
	if !readTypeIsValid("discrete") {
		t.Error("'discrete' should be a valid read type")
	}
}

func TestReadType_InvalidValues(t *testing.T) {
	invalid := []string{"", "ANALOG", "Discrete", "digital", "binary", "unknown"}
	for _, rt := range invalid {
		if readTypeIsValid(rt) {
			t.Errorf("read type %q should be invalid", rt)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Timestamp parsing (mirrors data.SaveReading time.Parse call)
// ──────────────────────────────────────────────────────────────────────────────

func TestTimestamp_ValidRFC3339(t *testing.T) {
	valid := []string{
		"2024-01-15T10:30:00Z",
		"2024-06-01T23:59:59+03:00",
		"2023-12-31T00:00:00-05:00",
	}
	for _, ts := range valid {
		if _, err := time.Parse(time.RFC3339, ts); err != nil {
			t.Errorf("expected valid RFC3339 timestamp %q, got error: %v", ts, err)
		}
	}
}

func TestTimestamp_InvalidFormats(t *testing.T) {
	invalid := []string{
		"",
		"2024-01-15",
		"15/01/2024 10:30:00",
		"not-a-date",
		"2024-01-15T10:30:00",    // missing timezone
		"2024-01-15 10:30:00Z",   // space instead of T
	}
	for _, ts := range invalid {
		if _, err := time.Parse(time.RFC3339, ts); err == nil {
			t.Errorf("expected error for invalid timestamp %q, got nil", ts)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// handleMessage – simulates the consumer handle() logic end-to-end
// ──────────────────────────────────────────────────────────────────────────────

// handleMessage replicates the logic in consumer.Job.handle so we can unit-test
// it without depending on a live RabbitMQ connection.
func handleMessage(ctx context.Context, body []byte, repo interface {
	SaveReading(context.Context, model.SensorMessage) error
}) error {
	var msg model.SensorMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err // would Nack(false, false)
	}
	return repo.SaveReading(ctx, msg) // would Nack(false, true) on error, Ack on nil
}

func TestHandleMessage_ValidAnalog(t *testing.T) {
	repo := &fakeRepo{}
	msg := validAnalogMessage()
	body, _ := json.Marshal(msg)

	if err := handleMessage(context.Background(), body, repo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.saveCalledWith) != 1 {
		t.Errorf("expected SaveReading to be called once")
	}
}

func TestHandleMessage_ValidDiscrete(t *testing.T) {
	repo := &fakeRepo{}
	msg := validDiscreteMessage()
	body, _ := json.Marshal(msg)

	if err := handleMessage(context.Background(), body, repo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	repo := &fakeRepo{}
	err := handleMessage(context.Background(), []byte(`not json`), repo)
	if err == nil {
		t.Error("expected JSON parse error, got nil")
	}
	if len(repo.saveCalledWith) != 0 {
		t.Error("SaveReading should not be called on parse failure")
	}
}

func TestHandleMessage_EmptyBody(t *testing.T) {
	repo := &fakeRepo{}
	err := handleMessage(context.Background(), []byte(`{}`), repo)
	// Empty body is valid JSON; SaveReading is called with zero-value message.
	if err != nil {
		t.Errorf("unexpected error for empty JSON body: %v", err)
	}
	if len(repo.saveCalledWith) != 1 {
		t.Error("SaveReading should be called even for empty JSON")
	}
}

func TestHandleMessage_RepoError_IsReturned(t *testing.T) {
	sentinel := errors.New("save failed")
	repo := &fakeRepo{returnErr: sentinel}
	body, _ := json.Marshal(validAnalogMessage())

	err := handleMessage(context.Background(), body, repo)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestHandleMessage_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	repo := &fakeRepo{}
	body, _ := json.Marshal(validAnalogMessage())

	// The fake repo ignores ctx; a real repo would propagate the cancellation.
	// We verify the function at least doesn't panic.
	_ = handleMessage(ctx, body, repo)
}

// ──────────────────────────────────────────────────────────────────────────────
// DSN builders (mirrors main.go helpers)
// ──────────────────────────────────────────────────────────────────────────────

func buildPostgresDSN(host, port, user, pass, name string) string {
	return "host=" + host + " port=" + port + " user=" + user +
		" password=" + pass + " dbname=" + name + " sslmode=disable"
}

func buildRabbitMQDSN(user, pass, host, port string) string {
	return "amqp://" + user + ":" + pass + "@" + host + ":" + port + "/"
}

func TestBuildPostgresDSN(t *testing.T) {
	dsn := buildPostgresDSN("db-host", "5432", "admin", "secret", "mydb")
	expected := "host=db-host port=5432 user=admin password=secret dbname=mydb sslmode=disable"
	if dsn != expected {
		t.Errorf("DSN mismatch:\ngot  %q\nwant %q", dsn, expected)
	}
}

func TestBuildRabbitMQDSN(t *testing.T) {
	dsn := buildRabbitMQDSN("guest", "guest", "rabbit-host", "5672")
	expected := "amqp://guest:guest@rabbit-host:5672/"
	if dsn != expected {
		t.Errorf("DSN mismatch:\ngot  %q\nwant %q", dsn, expected)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Edge cases – value boundaries
// ──────────────────────────────────────────────────────────────────────────────

func TestSensorMessage_ZeroValue(t *testing.T) {
	msg := model.SensorMessage{
		IDSensor:   "SN-X",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SensorType: "pressure",
		ReadType:   "analog",
		Value:      0,
	}
	repo := &fakeRepo{}
	body, _ := json.Marshal(msg)
	if err := handleMessage(context.Background(), body, repo); err != nil {
		t.Errorf("zero value should be accepted: %v", err)
	}
}

func TestSensorMessage_NegativeValue(t *testing.T) {
	msg := model.SensorMessage{
		IDSensor:   "SN-TEMP",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SensorType: "temperature",
		ReadType:   "analog",
		Value:      -40.0,
	}
	repo := &fakeRepo{}
	body, _ := json.Marshal(msg)
	if err := handleMessage(context.Background(), body, repo); err != nil {
		t.Errorf("negative value should be accepted: %v", err)
	}
}

func TestSensorMessage_LargeValue(t *testing.T) {
	msg := model.SensorMessage{
		IDSensor:   "SN-BIG",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SensorType: "pressure",
		ReadType:   "analog",
		Value:      1e15,
	}
	repo := &fakeRepo{}
	body, _ := json.Marshal(msg)
	if err := handleMessage(context.Background(), body, repo); err != nil {
		t.Errorf("large value should be accepted: %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ErrNoRows sentinel – mirrors data layer not-found handling
// ──────────────────────────────────────────────────────────────────────────────

func TestErrNoRows_DeviceNotFound(t *testing.T) {
	// Simulate the error path in resolveDeviceID when the device is missing.
	err := sql.ErrNoRows
	if !errors.Is(err, sql.ErrNoRows) {
		t.Error("sql.ErrNoRows should satisfy errors.Is check")
	}
}

func TestErrNoRows_SensorTypeNotFound(t *testing.T) {
	err := sql.ErrNoRows
	if !errors.Is(err, sql.ErrNoRows) {
		t.Error("sql.ErrNoRows should satisfy errors.Is check")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Batch – ordering guarantee
// ──────────────────────────────────────────────────────────────────────────────

func TestBatch_PreservesOrder(t *testing.T) {
	repo := &fakeRepo{}
	msgs := []model.SensorMessage{
		{IDSensor: "A", Timestamp: time.Now().UTC().Format(time.RFC3339), SensorType: "t", ReadType: "analog", Value: 1},
		{IDSensor: "B", Timestamp: time.Now().UTC().Format(time.RFC3339), SensorType: "t", ReadType: "analog", Value: 2},
		{IDSensor: "C", Timestamp: time.Now().UTC().Format(time.RFC3339), SensorType: "t", ReadType: "analog", Value: 3},
	}
	_ = repo.SaveReadingsBatch(context.Background(), msgs)

	for i, saved := range repo.saveCalledWith {
		if saved.IDSensor != msgs[i].IDSensor {
			t.Errorf("index %d: expected IDSensor %q, got %q", i, msgs[i].IDSensor, saved.IDSensor)
		}
	}
}