package model
 
type SensorMessage struct {
	IDSensor   string  `json:"idSensor"`
	Timestamp  string  `json:"timestamp"`
	SensorType string  `json:"sensorType"`
	ReadType   string  `json:"readType"`
	Value      float64 `json:"value"`
}