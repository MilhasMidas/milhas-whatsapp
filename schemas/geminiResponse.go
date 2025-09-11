package schemas

// GeminiResponse represents the structured response from Gemini API
// when processing flight booking images
type GeminiResponse struct {
	Origin          GeminiOrigin           `json:"origin"`
	Destination     GeminiDestination      `json:"destination"`
	Airline         string                 `json:"airline"`
	ServiceClass    string                 `json:"serviceClass"`
	LoyaltyPrograms []GeminiLoyaltyProgram `json:"loyaltyPrograms"`
	Availability    GeminiAvailability     `json:"availability"`
	Connections     []GeminiConnection     `json:"connections"`
}

// GeminiOrigin represents the departure location
type GeminiOrigin struct {
	City        string `json:"city"`
	AirportCode string `json:"airportCode"`
}

// GeminiDestination represents the arrival location
type GeminiDestination struct {
	City        string `json:"city"`
	AirportCode string `json:"airportCode"`
}

// GeminiConnection represents a connection point along the route
type GeminiConnection struct {
	City        string `json:"city"`
	AirportCode string `json:"airportCode"`
}

// GeminiFees represents the fees associated with a loyalty program
type GeminiFees struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// GeminiLoyaltyProgram represents a loyalty program with miles and fees
type GeminiLoyaltyProgram struct {
	Name  string      `json:"name"`
	Miles float64     `json:"miles"`
	Fees  interface{} `json:"fees"` // Can be GeminiFees struct or string
}

// GeminiAvailability represents the search date and available dates
type GeminiAvailability struct {
	SearchDate     string   `json:"searchDate"`
	AvailableDates []string `json:"availableDates"`
}

// GeminiErrorResponse represents an error response from Gemini API
type GeminiErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// GeminiProcessingRequest represents a request to process an image with Gemini
type GeminiProcessingRequest struct {
	ImagePath string `json:"imagePath"`
	APIKey    string `json:"apiKey,omitempty"`
}

// GeminiProcessingResponse represents the response from processing an image
type GeminiProcessingResponse struct {
	Success   bool            `json:"success"`
	Data      *GeminiResponse `json:"data,omitempty"`
	Error     string          `json:"error,omitempty"`
	Processed bool            `json:"processed"`
}

// ImagePreAnalysisResponse represents the response from pre-analyzing an image
// to determine if it's a flight availability calendar
type ImagePreAnalysisResponse struct {
	IsCalendar bool `json:"isCalendar"`
}
