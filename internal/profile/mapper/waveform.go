package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
)

func MarshalWaveformData(waveformData []float32) (null.JSON, error) {
	if waveformData == nil {
		return null.JSON{}, nil
	}

	raw, err := json.Marshal(waveformData)
	if err != nil {
		return null.JSON{}, fmt.Errorf("marshal waveform data: %w", err)
	}

	return null.JSONFrom(raw), nil
}

func UnmarshalWaveformData(raw null.JSON) ([]float32, error) {
	if !raw.Valid {
		return nil, nil
	}

	var waveformData []float32
	if err := raw.Unmarshal(&waveformData); err != nil {
		return nil, fmt.Errorf("unmarshal waveform data: %w", err)
	}

	return waveformData, nil
}
