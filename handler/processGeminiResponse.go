package handler

import (
	"fmt"
	"net/http"

	schemas "github.com/Unicorn-s-Club/whats-unicorn/schemas"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProcessGeminiResponseRequest represents the request to process a Gemini response
type ProcessGeminiResponseRequest struct {
	GeminiResponse schemas.GeminiResponse `json:"geminiResponse" binding:"required"`
}

// ProcessGeminiResponseResponse represents the response after processing
type ProcessGeminiResponseResponse struct {
	Success     bool                 `json:"success"`
	FlightAward *schemas.FlightAward `json:"flightAward,omitempty"`
	Error       string               `json:"error,omitempty"`
	Message     string               `json:"message,omitempty"`
}

// ProcessGeminiResponse handles the processing and saving of Gemini responses
func ProcessGeminiResponse(c *gin.Context) {
	var req ProcessGeminiResponseRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		Logger.Errorf("Error binding request: %v", err)
		c.JSON(http.StatusBadRequest, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Validate the Gemini response
	if err := validateGeminiResponse(&req.GeminiResponse); err != nil {
		Logger.Errorf("Error validating Gemini response: %v", err)
		c.JSON(http.StatusBadRequest, ProcessGeminiResponseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Create service instance
	service := NewFlightAwardService(Db)

	// Save the Gemini response to database
	flightAward, err := service.SaveGeminiResponse(&req.GeminiResponse)
	if err != nil {
		Logger.Errorf("Error saving Gemini response: %v", err)
		c.JSON(http.StatusInternalServerError, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Failed to save flight award data",
		})
		return
	}

	Logger.Infof("Successfully saved flight award with ID: %s", flightAward.ID)

	// Return success response
	c.JSON(http.StatusOK, ProcessGeminiResponseResponse{
		Success:     true,
		FlightAward: flightAward,
		Message:     "Flight award data saved successfully",
	})
}

// GetFlightAwardByID retrieves a flight award by ID
func GetFlightAwardByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Flight award ID is required",
		})
		return
	}

	// Parse UUID
	flightAwardID, err := parseUUID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Invalid flight award ID format",
		})
		return
	}

	// Create service instance
	service := NewFlightAwardService(Db)

	// Retrieve flight award
	flightAward, err := service.GetFlightAwardByID(flightAwardID)
	if err != nil {
		Logger.Errorf("Error retrieving flight award: %v", err)
		c.JSON(http.StatusNotFound, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Flight award not found",
		})
		return
	}

	c.JSON(http.StatusOK, ProcessGeminiResponseResponse{
		Success:     true,
		FlightAward: flightAward,
	})
}

// GetFlightAwardsByRoute retrieves flight awards by origin and destination
func GetFlightAwardsByRoute(c *gin.Context) {
	originCode := c.Query("origin")
	destCode := c.Query("destination")

	if originCode == "" || destCode == "" {
		c.JSON(http.StatusBadRequest, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Origin and destination airport codes are required",
		})
		return
	}

	// Create service instance
	service := NewFlightAwardService(Db)

	// Retrieve flight awards by route
	flightAwards, err := service.GetFlightAwardsByRoute(originCode, destCode)
	if err != nil {
		Logger.Errorf("Error retrieving flight awards by route: %v", err)
		c.JSON(http.StatusInternalServerError, ProcessGeminiResponseResponse{
			Success: false,
			Error:   "Failed to retrieve flight awards",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"flightAwards": flightAwards,
		"count":        len(flightAwards),
		"origin":       originCode,
		"destination":  destCode,
	})
}

// validateGeminiResponse validates the Gemini response data
func validateGeminiResponse(resp *schemas.GeminiResponse) error {
	if resp.Origin.AirportCode == "" {
		return fmt.Errorf("origin airport code is required")
	}

	if resp.Destination.AirportCode == "" {
		return fmt.Errorf("destination airport code is required")
	}

	if resp.Airline == "" {
		return fmt.Errorf("airline is required")
	}

	if resp.ServiceClass == "" {
		return fmt.Errorf("service class is required")
	}

	if resp.Availability.SearchDate == "" {
		return fmt.Errorf("search date is required")
	}

	return nil
}

// parseUUID parses a string UUID
func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
