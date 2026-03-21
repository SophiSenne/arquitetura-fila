/**
 * k6-test.ts — Teste de carga para o serviço de ingestão de sensores
 *
 * Contrato:
 *   POST /sensor  (ou ajuste BASE_URL + path conforme sua rota)
 *   Body: SensorMessage { idSensor, timestamp, sensorType, readType, value }
 *
 * Execução local (sem Docker):
 *   k6 run k6-test.js
 *
 * Execução via Docker (ver Dockerfile):
 *   docker run --rm --network <rede> -e BASE_URL=http://go_backend:8080 k6-load-test run /tests/k6-test.js
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend, Counter } from "k6/metrics";
import { Options } from "k6/options";

// ─── Métricas customizadas ────────────────────────────────────────────────────

const errorRate = new Rate("sensor_error_rate");
const requestDuration = new Trend("sensor_request_duration", true);
const totalRequests = new Counter("sensor_total_requests");

// ─── Configuração do teste ────────────────────────────────────────────────────

export const options: Options = {
  /**
   * Cenário em rampa (ramp-up → sustentado → ramp-down)
   *
   *  0s  →  30s  : sobe de 0 → 10 VUs   (warm-up)
   * 30s  →  90s  : mantém 10 VUs        (carga base)
   * 90s  → 120s  : sobe de 10 → 50 VUs  (pico)
   * 120s → 180s  : mantém 50 VUs        (carga de pico)
   * 180s → 210s  : desce de 50 → 0      (cool-down)
   */
  stages: [
    { duration: "30s", target: 10 },
    { duration: "60s", target: 10 },
    { duration: "30s", target: 50 },
    { duration: "60s", target: 50 },
    { duration: "30s", target: 0 },
  ],

  thresholds: {
    // 95% das requisições devem responder em menos de 500 ms
    http_req_duration: ["p(95)<500"],
    // Taxa de erros deve ficar abaixo de 1%
    sensor_error_rate: ["rate<0.01"],
    // Duração customizada: mediana abaixo de 200 ms
    sensor_request_duration: ["med<200"],
  },
};

// ─── Dados de domínio ─────────────────────────────────────────────────────────

type SensorType =
  | "temperatura"
  | "umidade"
  | "presença"
  | "vibração"
  | "luminosidade"
  | "nível_reservatório";

type ReadType = "analog" | "discrete";

interface SensorMessage {
  idSensor: string;
  timestamp: string;
  sensorType: SensorType;
  readType: ReadType;
  value: number;
}

/** Mapeamento sensor → tipo de leitura e faixa de valores plausível */
const SENSOR_PROFILES: Array<{
  idSensor: string;
  sensorType: SensorType;
  readType: ReadType;
  valueMin: number;
  valueMax: number;
}> = [
  { idSensor: "SN-TH-001", sensorType: "temperatura",        readType: "analog",   valueMin: 15,  valueMax: 45  },
  { idSensor: "SN-TH-002", sensorType: "temperatura",        readType: "analog",   valueMin: 15,  valueMax: 45  },
  { idSensor: "SN-TH-003", sensorType: "umidade",            readType: "analog",   valueMin: 20,  valueMax: 95  },
  { idSensor: "SN-PIR-001", sensorType: "presença",          readType: "discrete", valueMin: 0,   valueMax: 1   },
  { idSensor: "SN-VIB-001", sensorType: "vibração",          readType: "analog",   valueMin: 0,   valueMax: 10  },
  { idSensor: "SN-LUX-001", sensorType: "luminosidade",      readType: "analog",   valueMin: 0,   valueMax: 1000},
  { idSensor: "SN-NIV-001", sensorType: "nível_reservatório", readType: "analog",  valueMin: 0,   valueMax: 100 },
];

// ─── Utilitários ──────────────────────────────────────────────────────────────

/** Número aleatório entre min e max (inclusive) */
function randFloat(min: number, max: number, decimals = 2): number {
  const raw = Math.random() * (max - min) + min;
  return parseFloat(raw.toFixed(decimals));
}

/** Escolhe um elemento aleatório de um array */
function pick<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

/** Timestamp ISO-8601 atual */
function nowISO(): string {
  return new Date().toISOString();
}

/** Gera um payload realista sorteando um perfil de sensor */
function generatePayload(): SensorMessage {
  const profile = pick(SENSOR_PROFILES);
  return {
    idSensor: profile.idSensor,
    timestamp: nowISO(),
    sensorType: profile.sensorType,
    readType: profile.readType,
    // Sensores discretos (presença) retornam 0 ou 1 sem casas decimais
    value:
      profile.readType === "discrete"
        ? randFloat(profile.valueMin, profile.valueMax, 0)
        : randFloat(profile.valueMin, profile.valueMax),
  };
}

// ─── Configuração de ambiente ─────────────────────────────────────────────────

const BASE_URL: string = __ENV.BASE_URL ?? "http://go_backend:8080";
const SENSOR_ENDPOINT = `${BASE_URL}/sensorData`;

const HEADERS = {
  "Content-Type": "application/json",
  Accept: "application/json",
};

// ─── Função principal (executada por cada VU em cada iteração) ────────────────

export default function (): void {
  const payload = generatePayload();
  const body = JSON.stringify(payload);

  const res = http.post(SENSOR_ENDPOINT, body, { headers: HEADERS });

  // Registra métricas customizadas
  requestDuration.add(res.timings.duration);
  totalRequests.add(1);

  const success = check(res, {
    "status 2xx": (r) => r.status >= 200 && r.status < 300,
    "response time < 500ms": (r) => r.timings.duration < 500,
    "body não vazio": (r) => (r.body as string)?.length > 0,
  });

  errorRate.add(!success);

  // Log resumido em caso de falha (visível no output do k6)
  if (!success) {
    console.warn(
      `[FALHA] ${payload.idSensor} | status=${res.status} | dur=${res.timings.duration.toFixed(0)}ms | body=${res.body}`
    );
  }

  // Pausa entre 100 ms e 500 ms para simular inter-arrival time real
  sleep(randFloat(0.1, 0.5, 3));
}

// ─── Lifecycle hooks ──────────────────────────────────────────────────────────

/** Executado uma vez antes do teste iniciar */
export function setup(): void {
  console.log(`🚀 Iniciando teste de carga`);
  console.log(`   Endpoint : ${SENSOR_ENDPOINT}`);
  console.log(`   Sensores : ${SENSOR_PROFILES.length} perfis cadastrados`);
}

/** Executado uma vez após o término do teste */
export function teardown(): void {
  console.log(`✅ Teste finalizado. Verifique os thresholds acima.`);
}