package resources

import (
	"database/sql"
	"fmt"
	"io/ioutil"
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
	name VARCHAR(64) NOT NULL,
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
	"name":           true,
	"year":           true,
	"exterior_color": true,
	"interior_color": true,
}

type Vehicle struct {
	VIN           string    `json:"vin,omitempty"`
	Make          string    `json:"make,omitempty"`
	Name          string    `json:"name,omitempty"`
	Year          uint16    `json:"year,omitempty"`
	ExteriorColor string    `db:"exterior_color" json:"exterior_color,omitempty"`
	InteriorColor string    `db:"interior_color" json:"interior_color,omitempty"`
	UpdatedAt     time.Time `db:"updated_at" json:"-"`
}

func (encoded *Vehicle) Validate() error {
	if encoded.VIN == "" {
		return fmt.Errorf("A vin is required")
	}
	if encoded.Make == "" {
		return fmt.Errorf("A make is required")
	}
	if encoded.Name == "" {
		return fmt.Errorf("A name is required")
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

func BindVehicleRequestHandlers(router *mux.Router) {
	router.HandleFunc("/vehicles", ListHandler).Methods("GET")
	router.HandleFunc("/vehicles/{vin}", DeleteHandler).Methods("DELETE")
	router.HandleFunc("/vehicles/{vin}", GetHandler).Methods("GET")
	router.HandleFunc("/vehicles", CreateHandler).Methods("POST")
}

func ListHandler(writer http.ResponseWriter, request *http.Request) {
	vehicles := []Vehicle{}
	var err error

	queryParams := request.URL.Query()
	if len(queryParams) == 0 {
		vehicles, err = List()
	} else {
		vehicles, err = Search(&queryParams)
	}

	if err != nil {
		svr.HttpRespond(writer, request, http.StatusInternalServerError,
			svr.Error{Message: "Error listing vehicles"})
		return
	}
	log.Log.Info().Interface("vehicles", vehicles).Msg("encoded vehicles")
	svr.HttpRespond(writer, request, http.StatusOK, vehicles)
}

func DeleteHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	vin := vars["vin"]
	deleted, err := Delete(vin)
	if err != nil {
		svr.HttpRespond(writer, request, http.StatusInternalServerError, nil)
		return
	}
	if deleted != true {
		svr.HttpRespond(writer, request, http.StatusNotFound,
			svr.Error{Message: fmt.Sprintf("Vehicle with VIN %s doesn't exist", vin)})
		return
	}
	svr.HttpRespond(writer, request, http.StatusNoContent, nil)
}

func GetHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	vin := vars["vin"]
	vehicle, err := Get(vin)
	if err != nil {
		svr.HttpRespond(writer, request, http.StatusInternalServerError, nil)
		return
	}
	if vehicle == nil {
		svr.HttpRespond(writer, request, http.StatusNotFound, nil)
		return
	}
	log.Log.Info().Interface("vehicle", vehicle).Msg("Got vehicle")
	svr.HttpRespond(writer, request, http.StatusOK, vehicle)
}

func CreateHandler(writer http.ResponseWriter, request *http.Request) {
	var vehicle Vehicle
	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error ready request body")
		svr.HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	contentType := svr.GetRequestContentType(request)
	if contentType == "" {
		svr.HttpRespond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation
	err = svr.Unmarshal(contentType, body, &vehicle)
	if err != nil {
		log.Log.Err(err).Msg("Error unmarshalling request body")
		svr.HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	if err = vehicle.Validate(); err != nil {
		log.Log.Err(err).Msg("Invalid vehicle format")
		svr.HttpRespond(writer, request, http.StatusBadRequest, svr.Error{Message: err.Error()})
		return
	}

	if err = Create(&vehicle); err != nil {
		// TODO: better handling of existing vin
		if strings.Contains(err.Error(), "duplicate key") {
			svr.HttpRespond(writer, request, http.StatusBadRequest,
				svr.Error{Message: "Vehicle with that vin already exists"})
			return
		}
		svr.HttpRespond(writer, request, http.StatusInternalServerError, nil)
		return
	}

	svr.HttpRespond(writer, request, http.StatusOK, vehicle)
}

func Search(queryParams *url.Values) ([]Vehicle, error) {
	vehicles := []Vehicle{}
	statement := "SELECT * FROM vehicles WHERE"
	var inStatements []string

	for col, vals := range *queryParams {
		_, exists := allowedQueryParams[col]
		if !exists {
			// TODO: this is a bad request not internal error
			return nil, fmt.Errorf("Invalid query param: %s", col)
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
		return vehicles, err
	}
	return vehicles, nil
}

func List() ([]Vehicle, error) {
	vehicles := []Vehicle{}
	store := db.GetDB()
	err := store.Select(&vehicles, "SELECT * FROM vehicles")
	if err != nil {
		log.Log.Err(err).Msg("Database error listing vehicles")
		return vehicles, err
	}
	return vehicles, nil
}

func Get(vin string) (*Vehicle, error) {
	vehicle := Vehicle{}
	store := db.GetDB()
	err := store.Get(&vehicle, "SELECT * FROM vehicles WHERE vin=$1", vin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error getting vehicle")
		return nil, err
	}
	return &vehicle, nil
}

func Delete(vin string) (bool, error) {
	store := db.GetDB()
	result, err := store.Exec("DELETE FROM vehicles WHERE vin=$1", vin)
	if err != nil {
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error deleting vehicle")
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func Create(vehicle *Vehicle) error {
	store := db.GetDB()
	_, err := store.Exec(`INSERT INTO vehicles (vin, make, name, year, exterior_color, 
		interior_color, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7)`,
		vehicle.VIN, vehicle.Make, vehicle.Name, vehicle.Year, vehicle.ExteriorColor,
		vehicle.InteriorColor, time.Now())
	if err != nil {
		//  duplicate key value violates unique constraint \"vehicles_vin_key\" (SQLSTATE 23505)
		log.Log.Err(err).Msg("Database error creating vehicle")
		return err
	}
	return nil
}
