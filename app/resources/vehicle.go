package resources

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/bodenr/vehicle-app/db"
	"github.com/bodenr/vehicle-app/log"
	"github.com/bodenr/vehicle-app/svr"
	"gorm.io/gorm"
)

type Vehicle struct {
	gorm.Model
	VIN           string    `gorm:"primaryKey;unique;size:64" json:"vin,omitempty"`
	Make          string    `gorm:"size:64;not null" json:"make,omitempty"`
	Name          string    `gorm:"size:64;not null" json:"name,omitempty"`
	Year          uint16    `gorm:"not null" json:"year,omitempty"`
	ExteriorColor string    `gorm:"size:64;not null" json:"exterior_color,omitempty"`
	InteriorColor string    `gorm:"size:64;not null" json:"interior_color,omitempty"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
	DeletedAt     time.Time `json:"-"`
}

func BindVehicleRequestHandlers(router *mux.Router) {
	router.HandleFunc("/vehicles", ListHandler).Methods("GET")
	router.HandleFunc("/vehicles/{vin}", DeleteHandler).Methods("DELETE")
	router.HandleFunc("/vehicles/{vin}", GetHandler).Methods("GET")
	router.HandleFunc("/vehicles", CreateHandler).Methods("POST")
}

func ListHandler(writer http.ResponseWriter, request *http.Request) {
	vehicles, err := List()
	if err != nil {
		svr.HttpRespond(writer, request, http.StatusInternalServerError,
			svr.Error{Message: "Database error listing vehicles"})
		return
	}
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
	svr.HttpRespond(writer, request, http.StatusNoContent, vehicle)
}

func CreateHandler(writer http.ResponseWriter, request *http.Request) {
	var vehicle Vehicle
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
	// TODO: validate and sanitize body
	err = svr.Unmarshal(contentType, body, &vehicle)
	if err != nil {
		log.Log.Err(err).Msg("Error unmarshalling request body")
		svr.HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	if err = Create(&vehicle); err != nil {
		svr.HttpRespond(writer, request, http.StatusInternalServerError, nil)
		return
	}

	svr.HttpRespond(writer, request, http.StatusOK, vehicle)
}

func List() ([]Vehicle, error) {
	var vehicles []Vehicle
	store := db.GetDB()
	result := store.Find(&vehicles)
	if err := result.Error; err != nil {
		log.Log.Err(err).Msg("Database error listing vehicles")
		return vehicles, err
	}
	return vehicles, nil
}

func Get(vin string) (*Vehicle, error) {
	var vehicle Vehicle
	store := db.GetDB()
	result := store.First(&vehicle, vin)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error getting vehicle")
		return nil, err
	}
	return &vehicle, nil
}

func Delete(vin string) (bool, error) {
	store := db.GetDB()
	result := store.Delete(&Vehicle{VIN: vin})
	if err := result.Error; err != nil {
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error deleting vehicle")
		return false, err
	}
	return result.RowsAffected > 0, nil
}

func Create(vehicle *Vehicle) error {
	store := db.GetDB()
	result := store.Create(vehicle)
	if err := result.Error; err != nil {
		//  duplicate key value violates unique constraint \"vehicles_vin_key\" (SQLSTATE 23505)
		log.Log.Err(err).Msg("Database error creating vehicle")
		return err
	}
	return nil
}
