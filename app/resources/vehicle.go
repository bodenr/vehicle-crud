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

func CreateSchema() {
	db.GetDB().MustExec(schema)
}

var allowedQueryParams = map[string]bool{
	"make":           true,
	"model":          true,
	"year":           true,
	"exterior_color": true,
	"interior_color": true,
}

type Vehicle struct {
	VIN           string    `json:"vin,omitempty" xml:"vin"`
	Make          string    `json:"make,omitempty" xml:"make"`
	Model         string    `json:"model,omitempty" xml:"model"`
	Year          uint16    `json:"year,omitempty" xml:"year"`
	ExteriorColor string    `db:"exterior_color" json:"exterior_color,omitempty" xml:"exterior_color"`
	InteriorColor string    `db:"interior_color" json:"interior_color,omitempty" xml:"interior_color"`
	UpdatedAt     time.Time `db:"updated_at" json:"-" xml:"-"`
}

func (encoded Vehicle) Validate(method string) error {
	if encoded.VIN == "" && method != http.MethodPut {
		return fmt.Errorf("A vin is required")
	}
	if encoded.Make == "" {
		return fmt.Errorf("A make is required")
	}
	if encoded.Model == "" {
		return fmt.Errorf("A model is required")
	}
	if encoded.Year == 0 {
		return fmt.Errorf("A year is required")
	}
	if encoded.ExteriorColor == "" {
		return fmt.Errorf("An exterior_color is required")
	}
	if encoded.InteriorColor == "" {
		return fmt.Errorf("An interior_color is required")
	}
	return nil
}

func vehiclesToStoredResource(vehicles []Vehicle) []svr.StoredResource {
	resources := make([]svr.StoredResource, len(vehicles))
	for i, v := range vehicles {
		resources[i] = svr.StoredResource(v)
	}
	return resources
}

func (vehicle Vehicle) BindRoutes(router *mux.Router, handler svr.RestfulHandler) {
	router.HandleFunc("/vehicles", handler.List).Methods("GET")
	router.HandleFunc("/vehicles/{vin}", handler.Delete).Methods("DELETE")
	router.HandleFunc("/vehicles/{vin}", handler.Get).Methods("GET")
	router.HandleFunc("/vehicles/{vin}", handler.Update).Methods("PUT")
	router.HandleFunc("/vehicles", handler.Create).Methods("POST")
}

func (vehicle Vehicle) Unmarshal(contentType string, resource []byte) (svr.StoredResource, error) {
	v := Vehicle{}
	err := svr.Unmarshal(contentType, resource, &v)
	return v, err
}

func (vehicle Vehicle) Search(queryParams *url.Values) ([]svr.StoredResource, *svr.StoreError) {
	vehicles := make([]Vehicle, 0)
	statement := "SELECT * FROM vehicles WHERE"
	var inStatements []string

	for col, vals := range *queryParams {
		_, exists := allowedQueryParams[col]
		if !exists {
			return vehiclesToStoredResource(vehicles), &svr.StoreError{
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
		return vehiclesToStoredResource(vehicles), &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return vehiclesToStoredResource(vehicles), nil
}

func (vehicle Vehicle) List() ([]svr.StoredResource, *svr.StoreError) {
	vehicles := make([]Vehicle, 0)
	store := db.GetDB()
	err := store.Select(&vehicles, "SELECT * FROM vehicles")
	if err != nil {
		log.Log.Err(err).Msg("Database error listing vehicles")
		return vehiclesToStoredResource(vehicles), &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return vehiclesToStoredResource(vehicles), nil
}

func (vehicle Vehicle) Get(requestVars svr.RequestVars) (svr.StoredResource, *svr.StoreError) {
	v := Vehicle{}
	vin := requestVars["vin"]
	store := db.GetDB()
	err := store.Get(&v, "SELECT * FROM vehicles WHERE vin=$1", vin)
	if err != nil {
		if err == sql.ErrNoRows {
			return v, &svr.StoreError{
				Error:      err,
				StatusCode: http.StatusNotFound,
			}
		}
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error getting vehicle")
		return v, &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return v, nil
}

func (vehicle Vehicle) Delete(requestVars svr.RequestVars) *svr.StoreError {
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

func (vehicle Vehicle) Create() *svr.StoreError {
	store := db.GetDB()
	_, err := store.Exec(`INSERT INTO vehicles (vin, make, model, year, exterior_color, 
		interior_color, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7)`,
		vehicle.VIN, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.ExteriorColor,
		vehicle.InteriorColor, time.Now())
	if err != nil {
		log.Log.Err(err).Msg("Database error creating vehicle")

		// TODO: don't check error string
		if strings.Contains(err.Error(), "duplicate key value") {
			return &svr.StoreError{
				Error:      err,
				StatusCode: http.StatusBadRequest,
			}
		}

		return &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return nil
}

func (vehicle Vehicle) Update(requestVars svr.RequestVars) *svr.StoreError {
	vin := requestVars["vin"]
	store := db.GetDB()

	_, err := store.Exec("UPDATE vehicles SET make=$1, model=$2, year=$3, exterior_color=$4, interior_color=$5, updated_at=$6 WHERE vin=$7",
		vehicle.Make, vehicle.Model, vehicle.Year, vehicle.ExteriorColor, vehicle.InteriorColor, time.Now(), vin)
	if err != nil {
		log.Log.Err(err).Msg("Database error updating vehicle")
		return &svr.StoreError{
			Error:      err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	vehicle.VIN = vin
	return nil
}
