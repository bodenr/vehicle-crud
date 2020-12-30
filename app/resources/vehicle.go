package resources

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/bodenr/vehicle-api/db"
	"github.com/bodenr/vehicle-api/log"
	"github.com/bodenr/vehicle-api/svr"
)

type StoredVehicle struct{}

type Vehicle struct {
	VIN           string    `json:"vin,omitempty" xml:"vin"`
	Make          string    `json:"make,omitempty" xml:"make"`
	Model         string    `json:"model,omitempty" xml:"model"`
	Year          uint16    `json:"year,omitempty" xml:"year"`
	ExteriorColor string    `db:"exterior_color" json:"exterior_color,omitempty" xml:"exterior_color"`
	InteriorColor string    `db:"interior_color" json:"interior_color,omitempty" xml:"interior_color"`
	UpdatedAt     time.Time `db:"updated_at" json:"-" xml:"-"`
}

// TODO: make query checking more generic
var allowedQueryParams = map[string]bool{
	"make":           true,
	"model":          true,
	"year":           true,
	"exterior_color": true,
	"interior_color": true,
}

// TODO: move to sql file
var schema = `
CREATE TABLE vehicles (
	vin VARCHAR(64) UNIQUE NOT NULL PRIMARY KEY,
	make VARCHAR(64) NOT NULL,
	model VARCHAR(64) NOT NULL,
	year integer NOT NULL,
	exterior_color VARCHAR(64) NOT NULL,
	interior_color VARCHAR(64) NOT NULL,
	updated_at TIMESTAMP NOT NULL
);
`

func vehiclesToInterfaces(vehicles []Vehicle) []interface{} {
	// https://golang.org/doc/faq#convert_slice_of_interface
	interfaces := make([]interface{}, len(vehicles))
	for i, v := range vehicles {
		interfaces[i] = v
	}
	return interfaces
}

func (v StoredVehicle) CreateSchema() {
	db.GetDB().MustExec(schema)
}

func (v StoredVehicle) BindRoutes(router *mux.Router) {
	handler := svr.NewRestfulResource(StoredVehicle{})
	router.HandleFunc("/vehicles", handler.List).Methods("GET")
	router.HandleFunc("/vehicles/{vin}", handler.Delete).Methods("DELETE")
	router.HandleFunc("/vehicles/{vin}", handler.Get).Methods("GET")
	router.HandleFunc("/vehicles/{vin}", handler.Update).Methods("PUT")
	router.HandleFunc("/vehicles", handler.Create).Methods("POST")
}

func (v StoredVehicle) Validate(resource interface{}, httpMethod string) error {
	vehicle := resource.(Vehicle)
	if vehicle.VIN == "" && httpMethod != http.MethodPut {
		return fmt.Errorf("A vin is required")
	}
	if vehicle.Make == "" {
		return fmt.Errorf("A make is required")
	}
	if vehicle.Model == "" {
		return fmt.Errorf("A model is required")
	}
	if vehicle.Year == 0 {
		return fmt.Errorf("A year is required")
	}
	if vehicle.ExteriorColor == "" {
		return fmt.Errorf("An exterior_color is required")
	}
	if vehicle.InteriorColor == "" {
		return fmt.Errorf("An interior_color is required")
	}
	return nil
}

func (v StoredVehicle) Unmarshal(contentType string, resource []byte) (interface{}, error) {
	vehicle := Vehicle{}
	err := svr.Unmarshal(contentType, resource, &vehicle)
	return vehicle, err
}

func (v StoredVehicle) Marshal(contentType string, resource interface{}) ([]byte, error) {
	return svr.Marshal(contentType, resource)
}

func (v StoredVehicle) Search(queryParams url.Values) ([]interface{}, *svr.StoreError) {
	vehicles := make([]Vehicle, 0)
	statement := "SELECT * FROM vehicles WHERE"
	var inStatements []string

	for col, vals := range queryParams {
		_, exists := allowedQueryParams[col]
		if !exists {
			return vehiclesToInterfaces(vehicles), &svr.StoreError{
				Error:      fmt.Errorf("Invalid query param: %s", col),
				StatusCode: http.StatusBadRequest,
			}
		}
		s := fmt.Sprintf("%s IN ('%s')", col, strings.Join(vals[:], ","))
		inStatements = append(inStatements, s)
	}
	statement = fmt.Sprintf("%s %s", statement, strings.Join(inStatements[:], " AND "))
	log.Log.Debug().Str(log.Query, statement).Msg("Search query")

	store := db.GetDB()
	err := store.Select(&vehicles, statement)
	if err != nil {
		log.Log.Err(err).Msg("Database error listing vehicles")
		return vehiclesToInterfaces(vehicles), &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return vehiclesToInterfaces(vehicles), nil
}

func (v StoredVehicle) List() ([]interface{}, *svr.StoreError) {
	vehicles := make([]Vehicle, 0)
	store := db.GetDB()
	err := store.Select(&vehicles, "SELECT * FROM vehicles")
	if err != nil {
		log.Log.Err(err).Msg("Database error listing vehicles")
		return vehiclesToInterfaces(vehicles), &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return vehiclesToInterfaces(vehicles), nil
}

func (v StoredVehicle) Get(requestVars svr.RequestVars) (interface{}, *svr.StoreError) {
	vehicle := Vehicle{}
	vin := requestVars["vin"]
	store := db.GetDB()
	err := store.Get(&vehicle, "SELECT * FROM vehicles WHERE vin=$1", vin)
	if err != nil {
		// TODO: refactor DB common logic
		if err == sql.ErrNoRows {
			return vehicle, &svr.StoreError{
				Error:      err,
				StatusCode: http.StatusNotFound,
			}
		}
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error getting vehicle")
		return vehicle, &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return vehicle, nil
}

func (v StoredVehicle) Delete(requestVars svr.RequestVars) *svr.StoreError {
	vin := requestVars["vin"]
	store := db.GetDB()
	result, err := store.Exec("DELETE FROM vehicles WHERE vin=$1", vin)
	if err != nil {
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error deleting vehicle")
		return &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	affected, err := result.RowsAffected()
	if err != nil {
		log.Log.Err(err).Msg("Error deleting vehicle")
		return &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	if affected == 0 {
		return &svr.StoreError{
			Error:      fmt.Errorf("Vehicle with VIN %s doesn't exist", vin),
			StatusCode: http.StatusNotFound,
		}
	}
	return nil
}

func (v StoredVehicle) Create(resource interface{}) (interface{}, *svr.StoreError) {
	vehicle := resource.(Vehicle)
	store := db.GetDB()
	_, err := store.Exec(`INSERT INTO vehicles (vin, make, model, year, exterior_color, 
		interior_color, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7)`,
		vehicle.VIN, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.ExteriorColor,
		vehicle.InteriorColor, time.Now())
	if err != nil {
		log.Log.Err(err).Msg("Database error creating vehicle")

		// TODO: find a better way to detect db specific errors
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, &svr.StoreError{
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}
		}

		return nil, &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return vehicle, nil
}

func (v StoredVehicle) Update(resource interface{}, requestVars svr.RequestVars) (interface{}, *svr.StoreError) {
	vehicle := resource.(Vehicle)
	vin := requestVars["vin"]
	store := db.GetDB()

	_, err := store.Exec("UPDATE vehicles SET make=$1, model=$2, year=$3, exterior_color=$4, interior_color=$5, updated_at=$6 WHERE vin=$7",
		vehicle.Make, vehicle.Model, vehicle.Year, vehicle.ExteriorColor, vehicle.InteriorColor, time.Now(), vin)
	if err != nil {
		log.Log.Err(err).Msg("Database error updating vehicle")
		return nil, &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	vehicle.VIN = vin
	return vehicle, nil
}
