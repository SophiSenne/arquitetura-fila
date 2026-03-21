CREATE TABLE sensor_types (
    id       SERIAL PRIMARY KEY,
    name     VARCHAR(50)  NOT NULL UNIQUE,
    unit     VARCHAR(20)
);
 
CREATE TABLE devices (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    serial         VARCHAR(100) NOT NULL UNIQUE,
    description    TEXT,
    active         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
 
CREATE TABLE analog_readings (
    id             BIGSERIAL      PRIMARY KEY,
    device_id      UUID           NOT NULL REFERENCES devices(id),
    sensor_type_id INTEGER        NOT NULL REFERENCES sensor_types(id),
    value          NUMERIC(10, 4) NOT NULL,
    collected_at   TIMESTAMPTZ    NOT NULL,
    saved_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
 
CREATE TABLE discrete_readings (
    id             BIGSERIAL   PRIMARY KEY,
    device_id      UUID        NOT NULL REFERENCES devices(id),
    sensor_type_id INTEGER     NOT NULL REFERENCES sensor_types(id),
    value          INTEGER     NOT NULL,
    collected_at   TIMESTAMPTZ NOT NULL,
    saved_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
 
-- Indexes
CREATE INDEX idx_analog_device      ON analog_readings(device_id);
CREATE INDEX idx_analog_collected   ON analog_readings(collected_at DESC);
CREATE INDEX idx_analog_sensor      ON analog_readings(sensor_type_id);
CREATE INDEX idx_discrete_device    ON discrete_readings(device_id);
CREATE INDEX idx_discrete_collected ON discrete_readings(collected_at DESC);
CREATE INDEX idx_discrete_sensor    ON discrete_readings(sensor_type_id);
 
 
-- =============================================================
-- SEED — sensor_types - gerado por IA
-- =============================================================
 
INSERT INTO sensor_types (name, unit) VALUES
    ('temperatura',         '°C'),
    ('umidade',             '%'),
    ('presença',            NULL),      -- discreto: 0/1
    ('vibração',            'mm/s'),
    ('luminosidade',        'lux'),
    ('nível_reservatório',  '%');       -- 0-100 % de capacidade
 
 
INSERT INTO devices (id, serial, description, active) VALUES
    ('a1b2c3d4-0001-0001-0001-000000000001',
     'SN-TH-001', 'Sensor de temperatura e umidade – Sala de servidores A', TRUE),
 
    ('a1b2c3d4-0002-0002-0002-000000000002',
     'SN-TH-002', 'Sensor de temperatura e umidade – Sala de servidores B', TRUE),
 
    ('a1b2c3d4-0003-0003-0003-000000000003',
     'SN-PIR-001', 'Sensor de presença PIR – Corredor principal', TRUE),
 
    ('a1b2c3d4-0004-0004-0004-000000000004',
     'SN-VIB-001', 'Sensor de vibração – Compressor #1', TRUE),
 
    ('a1b2c3d4-0005-0005-0005-000000000005',
     'SN-LUX-001', 'Sensor de luminosidade – Área de produção', TRUE),
 
    ('a1b2c3d4-0006-0006-0006-000000000006',
     'SN-NIV-001', 'Sensor de nível – Reservatório de água tratada', TRUE),
 
    ('a1b2c3d4-0007-0007-0007-000000000007',
     'SN-TH-003', 'Sensor de temperatura – Câmara fria', FALSE);   -- dispositivo inativo