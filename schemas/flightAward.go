package schemas

import (
	"time"

	"github.com/google/uuid"
)

type ServiceClassCode string

const (
	ServiceClassEconomy        ServiceClassCode = "ECONOMY"
	ServiceClassPremiumEconomy ServiceClassCode = "PREMIUM_ECONOMY"
	ServiceClassBusiness       ServiceClassCode = "BUSINESS"
	ServiceClassFirst          ServiceClassCode = "FIRST"
	ServiceClassOther          ServiceClassCode = "OTHER"
)

type Airport struct {
	AirportCode string `gorm:"column:airport_code;type:char(3);primaryKey"`
	City        string `gorm:"type:text;not null"`
}

type Airline struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `gorm:"type:text;not null;unique"`
}

type LoyaltyProgram struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `gorm:"type:text;not null;unique"`
}

type FlightAward struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	OriginAirportCode      string  `gorm:"column:origin_airport_code;type:char(3);not null;index:idx_fa_origin_dest,priority:1"`
	DestinationAirportCode string  `gorm:"column:destination_airport_code;type:char(3);not null;index:idx_fa_origin_dest,priority:2"`
	Origin                 Airport `gorm:"foreignKey:OriginAirportCode;references:AirportCode"`
	Destination            Airport `gorm:"foreignKey:DestinationAirportCode;references:AirportCode"`

	AirlineID uuid.UUID `gorm:"column:airline_id;type:uuid;not null;index"`
	Airline   Airline   `gorm:"foreignKey:AirlineID;references:ID"`

	ServiceClassCode ServiceClassCode `gorm:"column:service_class_code;type:service_class_code;not null"`
	ServiceClassRaw  string           `gorm:"column:service_class_raw;type:text;not null"`

	SearchDate time.Time `gorm:"column:search_date;type:date;not null;index:idx_fa_search_date"`

	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`

	AvailableDates []FlightAwardAvailableDate `gorm:"foreignKey:FlightAwardID;constraint:OnDelete:CASCADE"`
	ProgramCosts   []FlightAwardProgramCost   `gorm:"foreignKey:FlightAwardID;constraint:OnDelete:CASCADE"`
	Connections    []FlightConnection         `gorm:"foreignKey:FlightAwardID;constraint:OnDelete:CASCADE"`
}

type FlightAwardAvailableDate struct {
	ID uint64 `gorm:"primaryKey;autoIncrement"`

	FlightAwardID uuid.UUID   `gorm:"column:flight_award_id;type:uuid;not null;uniqueIndex:ux_faad_award_date"`
	FlightAward   FlightAward `gorm:"foreignKey:FlightAwardID;references:ID;constraint:OnDelete:CASCADE"`

	AvailableDate time.Time `gorm:"column:available_date;type:date;not null;index:idx_faad_date;uniqueIndex:ux_faad_award_date"`
}

type FlightAwardProgramCost struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	FlightAwardID uuid.UUID   `gorm:"column:flight_award_id;type:uuid;not null;index;uniqueIndex:ux_fapc_award_program"`
	FlightAward   FlightAward `gorm:"foreignKey:FlightAwardID;references:ID;constraint:OnDelete:CASCADE"`

	LoyaltyProgramID uuid.UUID      `gorm:"column:loyalty_program_id;type:uuid;not null;index;uniqueIndex:ux_fapc_award_program"`
	Program          LoyaltyProgram `gorm:"foreignKey:LoyaltyProgramID;references:ID"`

	Miles int `gorm:"type:int;not null"`

	FeeValue    *float64 `gorm:"column:fee_value;type:numeric(12,2)"`
	FeeCurrency *string  `gorm:"column:fee_currency;type:char(3)"`
	FeeText     *string  `gorm:"column:fee_text;type:text"`
}

type FlightConnection struct {
	ID uint64 `gorm:"primaryKey;autoIncrement"`

	FlightAwardID uuid.UUID   `gorm:"column:flight_award_id;type:uuid;not null;index:idx_conn_flight;uniqueIndex:ux_conn_flight_seq"`
	FlightAward   FlightAward `gorm:"foreignKey:FlightAwardID;references:ID;constraint:OnDelete:CASCADE"`

	AirportID string  `gorm:"column:airport_id;type:char(3);not null;index"`
	Airport   Airport `gorm:"foreignKey:AirportID;references:AirportCode"`

	SEQ uint16 `gorm:"column:seq;not null;uniqueIndex:ux_conn_flight_seq"`

	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
}
