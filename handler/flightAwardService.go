package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	schemas "github.com/Unicorn-s-Club/whats-unicorn/schemas"
)

// FlightAwardService handles the conversion and storage of Gemini responses to FlightAward data
type FlightAwardService struct {
	db *gorm.DB
}

// NewFlightAwardService creates a new FlightAwardService instance
func NewFlightAwardService(db *gorm.DB) *FlightAwardService {
	return &FlightAwardService{db: db}
}

// SaveGeminiResponse converts and saves a GeminiResponse to the database
func (s *FlightAwardService) SaveGeminiResponse(geminiResp *schemas.GeminiResponse) (*schemas.FlightAward, error) {
	// Start a transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create or get origin airport
	originAirport, err := s.createOrGetAirport(tx, geminiResp.Origin.City, geminiResp.Origin.AirportCode)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating/getting origin airport: %v", err)
	}

	// Create or get destination airport
	destAirport, err := s.createOrGetAirport(tx, geminiResp.Destination.City, geminiResp.Destination.AirportCode)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating/getting destination airport: %v", err)
	}

	// Create or get airline
	airline, err := s.createOrGetAirline(tx, geminiResp.Airline)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating/getting airline: %v", err)
	}

	// Parse search date
	searchDate, err := time.Parse("2006-01-02", geminiResp.Availability.SearchDate)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error parsing search date: %v", err)
	}

	// Convert service class
	serviceClassCode := s.convertServiceClass(geminiResp.ServiceClass)

	// Create FlightAward
	flightAward := &schemas.FlightAward{
		ID:                     uuid.New(),
		OriginAirportCode:      originAirport.AirportCode,
		DestinationAirportCode: destAirport.AirportCode,
		AirlineID:              airline.ID,
		ServiceClassCode:       serviceClassCode,
		ServiceClassRaw:        geminiResp.ServiceClass,
		SearchDate:             searchDate,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Save FlightAward
	if err := tx.Create(flightAward).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating flight award: %v", err)
	}

	// Save available dates
	if err := s.saveAvailableDates(tx, flightAward.ID, geminiResp.Availability.AvailableDates); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error saving available dates: %v", err)
	}

	// Save loyalty programs and costs
	if err := s.saveLoyaltyPrograms(tx, flightAward.ID, geminiResp.LoyaltyPrograms); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error saving loyalty programs: %v", err)
	}

	// Save connections
	if err := s.saveConnections(tx, flightAward.ID, geminiResp.Connections); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error saving connections: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return flightAward, nil
}

// createOrGetAirport creates a new airport or returns existing one
func (s *FlightAwardService) createOrGetAirport(tx *gorm.DB, city, airportCode string) (*schemas.Airport, error) {
	var airport schemas.Airport

	// Try to find existing airport
	if err := tx.Where("airport_code = ?", airportCode).First(&airport).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new airport
			airport = schemas.Airport{
				AirportCode: airportCode,
				City:        city,
			}
			if err := tx.Create(&airport).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &airport, nil
}

// createOrGetAirline creates a new airline or returns existing one
func (s *FlightAwardService) createOrGetAirline(tx *gorm.DB, airlineName string) (*schemas.Airline, error) {
	var airline schemas.Airline

	// Try to find existing airline
	if err := tx.Where("name = ?", airlineName).First(&airline).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new airline
			airline = schemas.Airline{
				ID:   uuid.New(),
				Name: airlineName,
			}
			if err := tx.Create(&airline).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &airline, nil
}

// createOrGetLoyaltyProgram creates a new loyalty program or returns existing one
func (s *FlightAwardService) createOrGetLoyaltyProgram(tx *gorm.DB, programName string) (*schemas.LoyaltyProgram, error) {
	var program schemas.LoyaltyProgram

	// Try to find existing program
	if err := tx.Where("name = ?", programName).First(&program).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new program
			program = schemas.LoyaltyProgram{
				ID:   uuid.New(),
				Name: programName,
			}
			if err := tx.Create(&program).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &program, nil
}

// convertServiceClass converts string service class to ServiceClassCode
func (s *FlightAwardService) convertServiceClass(serviceClass string) schemas.ServiceClassCode {
	serviceClass = strings.ToUpper(strings.TrimSpace(serviceClass))

	switch serviceClass {
	case "ECONOMY", "ECONÔMICA":
		return schemas.ServiceClassEconomy
	case "PREMIUM ECONOMY", "ECONOMIA PREMIUM", "PREMIUM_ECONOMY":
		return schemas.ServiceClassPremiumEconomy
	case "BUSINESS", "EXECUTIVA":
		return schemas.ServiceClassBusiness
	case "FIRST", "PRIMEIRA":
		return schemas.ServiceClassFirst
	default:
		return schemas.ServiceClassOther
	}
}

// saveAvailableDates saves available dates for a flight award
func (s *FlightAwardService) saveAvailableDates(tx *gorm.DB, flightAwardID uuid.UUID, availableDates []string) error {
	for _, dateStr := range availableDates {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Skip invalid dates
			continue
		}

		availableDate := schemas.FlightAwardAvailableDate{
			FlightAwardID: flightAwardID,
			AvailableDate: date,
		}

		// Use FirstOrCreate to avoid duplicates
		if err := tx.Where("flight_award_id = ? AND available_date = ?", flightAwardID, date).
			FirstOrCreate(&availableDate).Error; err != nil {
			return err
		}
	}

	return nil
}

// saveLoyaltyPrograms saves loyalty programs and their costs
func (s *FlightAwardService) saveLoyaltyPrograms(tx *gorm.DB, flightAwardID uuid.UUID, programs []schemas.GeminiLoyaltyProgram) error {
	for _, program := range programs {
		// Create or get loyalty program
		loyaltyProgram, err := s.createOrGetLoyaltyProgram(tx, program.Name)
		if err != nil {
			return err
		}

		// Parse fees
		var feeValue *float64
		var feeCurrency *string
		var feeText *string

		if program.Fees != nil {
			switch fees := program.Fees.(type) {
			case schemas.GeminiFees:
				feeValue = &fees.Value
				feeCurrency = &fees.Currency
			case string:
				feeText = &fees
			case map[string]interface{}:
				if value, ok := fees["value"].(float64); ok {
					feeValue = &value
				}
				if currency, ok := fees["currency"].(string); ok {
					feeCurrency = &currency
				}
			}
		}

		// Create program cost
		programCost := schemas.FlightAwardProgramCost{
			ID:               uuid.New(),
			FlightAwardID:    flightAwardID,
			LoyaltyProgramID: loyaltyProgram.ID,
			Miles:            int(program.Miles),
			FeeValue:         feeValue,
			FeeCurrency:      feeCurrency,
			FeeText:          feeText,
		}

		// Use FirstOrCreate to avoid duplicates
		if err := tx.Where("flight_award_id = ? AND loyalty_program_id = ?", flightAwardID, loyaltyProgram.ID).
			FirstOrCreate(&programCost).Error; err != nil {
			return err
		}
	}

	return nil
}

// saveConnections saves flight connections
func (s *FlightAwardService) saveConnections(tx *gorm.DB, flightAwardID uuid.UUID, connections []schemas.GeminiConnection) error {
	for seq, connection := range connections {
		// Create or get connection airport
		airport, err := s.createOrGetAirport(tx, connection.City, connection.AirportCode)
		if err != nil {
			return err
		}

		flightConnection := schemas.FlightConnection{
			FlightAwardID: flightAwardID,
			AirportID:     airport.AirportCode,
			SEQ:           uint16(seq + 1), // SEQ starts from 1
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Use FirstOrCreate to avoid duplicates
		if err := tx.Where("flight_award_id = ? AND seq = ?", flightAwardID, flightConnection.SEQ).
			FirstOrCreate(&flightConnection).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetFlightAwardByID retrieves a flight award by ID with all related data
func (s *FlightAwardService) GetFlightAwardByID(id uuid.UUID) (*schemas.FlightAward, error) {
	var flightAward schemas.FlightAward

	if err := s.db.Preload("Origin").
		Preload("Destination").
		Preload("Airline").
		Preload("AvailableDates").
		Preload("ProgramCosts.Program").
		Preload("Connections.Airport").
		First(&flightAward, id).Error; err != nil {
		return nil, err
	}

	return &flightAward, nil
}

// GetFlightAwardsByRoute retrieves flight awards by origin and destination
func (s *FlightAwardService) GetFlightAwardsByRoute(originCode, destCode string) ([]schemas.FlightAward, error) {
	var flightAwards []schemas.FlightAward

	if err := s.db.Where("origin_airport_code = ? AND destination_airport_code = ?", originCode, destCode).
		Preload("Origin").
		Preload("Destination").
		Preload("Airline").
		Preload("AvailableDates").
		Preload("ProgramCosts.Program").
		Preload("Connections.Airport").
		Find(&flightAwards).Error; err != nil {
		return nil, err
	}

	return flightAwards, nil
}
