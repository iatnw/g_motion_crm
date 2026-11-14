package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var ErrVINFormat = errors.New("enter a complete 17-character VIN using letters other than I, O or Q")
var ErrVINNotDecoded = errors.New("vehicle details could not be decoded reliably; enter them manually")
var vinPattern = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)

type DecodedVehicle struct {
	VIN       string `json:"vin"`
	Make      string `json:"make"`
	Model     string `json:"model"`
	TrimLevel string `json:"trim_level"`
	ModelYear string `json:"model_year"`
}

type VINDecoder struct{ client *http.Client }

func NewVINDecoder(client *http.Client) *VINDecoder {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &VINDecoder{client: client}
}

func (d *VINDecoder) Decode(ctx context.Context, vin string) (DecodedVehicle, error) {
	vin = strings.ToUpper(strings.Join(strings.Fields(vin), ""))
	if !vinPattern.MatchString(vin) {
		return DecodedVehicle{}, ErrVINFormat
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://vpic.nhtsa.dot.gov/api/vehicles/DecodeVinValues/"+vin+"?format=json", nil)
	if err != nil {
		return DecodedVehicle{}, err
	}
	res, err := d.client.Do(req)
	if err != nil {
		return DecodedVehicle{}, fmt.Errorf("VIN lookup: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return DecodedVehicle{}, fmt.Errorf("VIN lookup returned status %d", res.StatusCode)
	}
	var payload struct {
		Results []struct{ VIN, Make, Model, Trim, ModelYear, ErrorCode string }
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&payload); err != nil {
		return DecodedVehicle{}, fmt.Errorf("decode VIN response: %w", err)
	}
	if len(payload.Results) != 1 {
		return DecodedVehicle{}, ErrVINNotDecoded
	}
	vehicle := payload.Results[0]
	if strings.TrimSpace(vehicle.ErrorCode) != "0" || !strings.EqualFold(vehicle.VIN, vin) || strings.TrimSpace(vehicle.Make) == "" || strings.TrimSpace(vehicle.Model) == "" {
		return DecodedVehicle{}, ErrVINNotDecoded
	}
	return DecodedVehicle{VIN: vin, Make: strings.TrimSpace(vehicle.Make), Model: strings.TrimSpace(vehicle.Model), TrimLevel: strings.TrimSpace(vehicle.Trim), ModelYear: strings.TrimSpace(vehicle.ModelYear)}, nil
}
