package cars

import (
	"strings"
	"testing"
	"time"
)

func TestCarCreateRequestValidate(t *testing.T) {
	validDate := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		req     *CarCreateRequest
		wantErr string
	}{
		{
			name: "valid request",
			req: &CarCreateRequest{
				Name:      "BMW M3",
				Colour:    "#1122AA",
				Price:     75000,
				BuildDate: validDate,
			},
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: "request is nil",
		},
		{
			name: "blank name",
			req: &CarCreateRequest{
				Name:      " ",
				Colour:    "#1122AA",
				Price:     75000,
				BuildDate: validDate,
			},
			wantErr: "name is required",
		},
		{
			name: "invalid colour",
			req: &CarCreateRequest{
				Name:      "BMW M3",
				Colour:    "blue",
				Price:     75000,
				BuildDate: validDate,
			},
			wantErr: "colour must be in hex format",
		},
		{
			name: "non-positive price",
			req: &CarCreateRequest{
				Name:      "BMW M3",
				Colour:    "#1122AA",
				Price:     0,
				BuildDate: validDate,
			},
			wantErr: "price must be greater than 0",
		},
		{
			name: "missing build date",
			req: &CarCreateRequest{
				Name:   "BMW M3",
				Colour: "#1122AA",
				Price:  75000,
			},
			wantErr: "build_date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}

				return
			}

			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestCarsGetRequestValidateRejectsInvertedPriceRange(t *testing.T) {
	priceFrom := 100.0
	priceTo := 50.0

	err := (&CarsGetRequest{
		PriceFrom: &priceFrom,
		PriceTo:   &priceTo,
	}).Validate()

	if err == nil {
		t.Fatal("Validate() error = nil, want price range error")
	}

	if !strings.Contains(err.Error(), "price_from must be less than price_to") {
		t.Fatalf("Validate() error = %q", err.Error())
	}
}

func TestCarUpdateRequestValidate(t *testing.T) {
	blankName := " "
	invalidColour := "black"
	negativePrice := -1.0
	zeroDate := time.Time{}

	tests := []struct {
		name    string
		req     *CarUpdateRequest
		wantErr string
	}{
		{
			name: "valid partial update",
			req: &CarUpdateRequest{
				ID:   1,
				Name: ptr("Audi RS6"),
			},
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: "request is nil",
		},
		{
			name:    "missing id",
			req:     &CarUpdateRequest{},
			wantErr: "id is required",
		},
		{
			name: "blank name",
			req: &CarUpdateRequest{
				ID:   1,
				Name: &blankName,
			},
			wantErr: "name is required",
		},
		{
			name: "invalid colour",
			req: &CarUpdateRequest{
				ID:     1,
				Colour: &invalidColour,
			},
			wantErr: "colour must be in hex format",
		},
		{
			name: "non-positive price",
			req: &CarUpdateRequest{
				ID:    1,
				Price: &negativePrice,
			},
			wantErr: "price must be greater than 0",
		},
		{
			name: "zero build date",
			req: &CarUpdateRequest{
				ID:        1,
				BuildDate: &zeroDate,
			},
			wantErr: "build_date is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}

				return
			}

			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}
