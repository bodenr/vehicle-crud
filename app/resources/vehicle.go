package resources

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/bodenr/vehicle-app/db"
	"github.com/bodenr/vehicle-app/log"
	"github.com/bodenr/vehicle-app/svr"
	"gorm.io/gorm"
)

type VehicleModel struct {
	gorm.Model
	VIN           string `gorm:"primaryKey;unique;size:64"`
	Make          string `gorm:"size:64;not null"`
	Name          string `gorm:"size:64;not null"`
	Year          uint16 `gorm:"not null"`
	ExteriorColor string `gorm:"size:64;not null"`
	InteriorColor string `gorm:"size:64;not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     time.Time
}

func (model *VehicleModel) ToEncoding() VehicleEncoding {
	return VehicleEncoding{
		VIN:           model.VIN,
		Make:          model.Make,
		Name:          model.Name,
		Year:          model.Year,
		ExteriorColor: model.ExteriorColor,
		InteriorColor: model.InteriorColor,
	}
}

func EncodeVehicles(vehicles []VehicleModel) []VehicleEncoding {
	encoded := make([]VehicleEncoding, 0)
	for _, v := range vehicles {
		encoded = append(encoded, v.ToEncoding())
	}
	return encoded
}

type VehicleEncoding struct {
	VIN           string `json:"vin,omitempty"`
	Make          string `json:"make,omitempty"`
	Name          string `json:"name,omitempty"`
	Year          uint16 `json:"year,omitempty"`
	ExteriorColor string `json:"exterior_color,omitempty"`
	InteriorColor string `json:"interior_color,omitempty"`
}

func (encoded *VehicleEncoding) ToModel() *VehicleModel {
	return &VehicleModel{
		VIN:           encoded.VIN,
		Make:          encoded.Make,
		Name:          encoded.Name,
		Year:          encoded.Year,
		ExteriorColor: encoded.ExteriorColor,
		InteriorColor: encoded.InteriorColor,
	}
}

func (encoded *VehicleEncoding) Validate() error {
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
	vehicles, err := List()
	if err != nil {
		svr.HttpRespond(writer, request, http.StatusInternalServerError,
			svr.Error{Message: "Error listing vehicles"})
		return
	}
	encoded := EncodeVehicles(vehicles)
	log.Log.Info().Interface("vehicles", encoded).Msg("encoded vehicles")
	svr.HttpRespond(writer, request, http.StatusOK, encoded)
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
	svr.HttpRespond(writer, request, http.StatusOK, vehicle.ToEncoding())
}

func CreateHandler(writer http.ResponseWriter, request *http.Request) {
	var vehicle VehicleEncoding
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
	// TODO: validate and sanitize body
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

	if err = Create(vehicle.ToModel()); err != nil {
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

func List() ([]VehicleModel, error) {
	var vehicles []VehicleModel
	store := db.GetDB()
	result := store.Find(&vehicles)
	if err := result.Error; err != nil {
		log.Log.Err(err).Msg("Database error listing vehicles")
		return vehicles, err
	}
	log.Log.Info().Int64("rows", result.RowsAffected).Interface("models", vehicles).Msg("Listing vehicles")
	return vehicles, nil
}

func Get(vin string) (*VehicleModel, error) {
	var vehicle VehicleModel
	store := db.GetDB()
	result := store.Where("vin = ?", vin).First(&vehicle)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error getting vehicle")
		return nil, err
	}
	log.Log.Info().Str("vin", vin).Interface("model", vehicle).Msg("Get vehicle")
	return &vehicle, nil
}

func Delete(vin string) (bool, error) {
	store := db.GetDB()
	result := store.Delete(&VehicleModel{VIN: vin})
	if err := result.Error; err != nil {
		log.Log.Err(err).Str(log.VIN, vin).Msg("Database error deleting vehicle")
		return false, err
	}
	return result.RowsAffected > 0, nil
}

func Create(vehicle *VehicleModel) error {
	store := db.GetDB()
	result := store.Create(vehicle)
	if err := result.Error; err != nil {
		//  duplicate key value violates unique constraint \"vehicles_vin_key\" (SQLSTATE 23505)
		log.Log.Err(err).Msg("Database error creating vehicle")
		return err
	}
	log.Log.Debug().Int64("rows", result.RowsAffected).Uint("row_id", vehicle.ID).Interface("vehicle", vehicle).Msg("Created vehicle")
	return nil
}
