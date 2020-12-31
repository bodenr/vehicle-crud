// Package resources provides stored resource implementations.
package resources

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/bodenr/vehicle-api/db"
	"github.com/bodenr/vehicle-api/log"
	"github.com/bodenr/vehicle-api/svr"
	"github.com/bodenr/vehicle-api/svr/proto"
	"github.com/bodenr/vehicle-api/util"
	protobuf "github.com/golang/protobuf/proto"
)

// StoredVehicle implements the StoredResource interface for vehicle resources.
type StoredVehicle struct{}

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
	updated_at bigint NOT NULL
);
`

func vehiclesToInterfaces(vehicles []proto.Vehicle) []interface{} {
	// https://golang.org/doc/faq#convert_slice_of_interface
	interfaces := make([]interface{}, len(vehicles))
	for i, v := range vehicles {
		interfaces[i] = v
	}
	return interfaces
}

// CreateSchema creates the database table schema for vehicles.
func (v StoredVehicle) CreateSchema() {
	db.GetDB().MustExec(schema)
}

// BindRoutes bind the vehicle routes to a router.
func (v StoredVehicle) BindRoutes(router *mux.Router) {
	handler := svr.NewRestfulResource(StoredVehicle{})
	router.HandleFunc("/vehicles", handler.List).Methods(http.MethodGet)
	router.HandleFunc("/vehicles/{vin}", handler.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/vehicles/{vin}", handler.Get).Methods(http.MethodGet)
	router.HandleFunc("/vehicles/{vin}", handler.Update).Methods(http.MethodPut)
	router.HandleFunc("/vehicles", handler.Create).Methods(http.MethodPost)
}

// Validate validates the said vehicle struct.
func (v StoredVehicle) Validate(resource interface{}, httpMethod string) error {
	vehicle := resource.(proto.Vehicle)
	if vehicle.Vin == "" && httpMethod != http.MethodPut {
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

// Unmarshal converts bytes into a vehicle using the said content type.
func (v StoredVehicle) Unmarshal(contentType string, resource []byte) (interface{}, error) {
	vehicle := proto.Vehicle{}
	if contentType == svr.ContentAppProtobuf {
		err := protobuf.Unmarshal(resource, &vehicle)
		return vehicle, err
	}
	err := svr.Unmarshal(contentType, resource, &vehicle)
	return vehicle, err
}

// Marshal converts a vehicle into bytes for the said content type.
func (v StoredVehicle) Marshal(contentType string, resource interface{}) ([]byte, error) {
	if contentType == svr.ContentAppProtobuf {
		v := resource.(proto.Vehicle)
		return protobuf.Marshal(&v)
	}
	return svr.Marshal(contentType, resource)
}

// Search searches the database for vehicles using the said query params.
func (v StoredVehicle) Search(queryParams url.Values) ([]interface{}, *svr.StoreError) {
	vehicles := make([]proto.Vehicle, 0)
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

// List returns all vehicles in the database.
func (v StoredVehicle) List() ([]interface{}, *svr.StoreError) {
	vehicles := make([]proto.Vehicle, 0)
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

// Get returns a specific vehicles as per the request vars if it exists.
func (v StoredVehicle) Get(requestVars svr.RequestVars) (interface{}, *svr.StoreError) {
	vehicle := proto.Vehicle{}
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

// Delete deletes a vehicle as specified by the request vars.
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

// Create creates a vehicle.
func (v StoredVehicle) Create(resource interface{}) (interface{}, *svr.StoreError) {
	vehicle := resource.(proto.Vehicle)
	store := db.GetDB()
	ts := util.TimeMillis()
	_, err := store.Exec(`INSERT INTO vehicles (vin, make, model, year, exterior_color, 
		interior_color, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7)`,
		vehicle.Vin, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.ExteriorColor,
		vehicle.InteriorColor, ts)
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
	vehicle.UpdatedAt = ts
	return vehicle, nil
}

// Update updates an existing vehicle.
func (v StoredVehicle) Update(resource interface{}, requestVars svr.RequestVars) (interface{}, *svr.StoreError) {
	vehicle := resource.(proto.Vehicle)
	vin := requestVars["vin"]
	store := db.GetDB()

	_, err := store.Exec("UPDATE vehicles SET make=$1, model=$2, year=$3, exterior_color=$4, interior_color=$5, updated_at=$6 WHERE vin=$7",
		vehicle.Make, vehicle.Model, vehicle.Year, vehicle.ExteriorColor, vehicle.InteriorColor, util.TimeMillis(), vin)
	if err != nil {
		log.Log.Err(err).Msg("Database error updating vehicle")
		return nil, &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	vehicle.Vin = vin
	return vehicle, nil
}

// GetETag builds an eTag by finding the vehicle in the request vars.
func (v StoredVehicle) GetETag(requestVars svr.RequestVars) (string, *svr.StoreError) {
	// TODO: refactor interface to return etag on Get/Update/Create
	vehicle := proto.Vehicle{}
	vin := requestVars["vin"]
	store := db.GetDB()
	err := store.Get(&vehicle, "SELECT updated_at FROM vehicles WHERE vin=$1", vin)
	if err != nil {
		// TODO: refactor DB common logic
		if err == sql.ErrNoRows {
			return "", &svr.StoreError{
				Error:      err,
				StatusCode: http.StatusNotFound,
			}
		}
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error getting vehicle")
		return "", &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	return buildETag(vin, vehicle.UpdatedAt), nil
}

// BuildETag builds and eTag from the given vehicle resource.
func (v StoredVehicle) BuildETag(resource interface{}) (string, error) {
	vehicle := resource.(proto.Vehicle)
	if vehicle.Vin == "" {
		return "", fmt.Errorf("Vehicle does not contain a VIN")
	}
	if vehicle.UpdatedAt == 0 {
		return "", fmt.Errorf("Vehicle does not contain an Updated at timestamp")
	}
	return buildETag(vehicle.Vin, vehicle.UpdatedAt), nil
}

func buildETag(vin string, updatedAt int64) string {
	data := []byte(fmt.Sprintf("%s.%d", vin, updatedAt))
	return fmt.Sprintf("%x", md5.Sum(data))
}
