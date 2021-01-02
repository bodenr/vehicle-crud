## Forward

Implements a basic CRUD containerized service that exposes (crude) vehicle resources over a REST and gRPC API.

The following services/containers are created as part of the docker-compose file:

- `db` A postgres database to store the vehicle data.
- `app` The API service written in GO providing a REST and gRPC API.
- `test` A basic set of python tests that verify the app REST API (with JSON). 


## REST API

- `GET         /api/vehicles`                                <-- get all vehicles
- `GET         /api/vehicles?exterior_color=red&make=dodge`  <-- - search vehicles
- `POST        /api/vehicles`                                <-- create a new vehicle
- `GET         /api/vehicles/{vin}`                          <-- get specific vehicle by VIN; ETag supported
- `DELETE      /api/vehicles/{vin}`                          <-- delete vehicle by VIN; ETag supported
- `PUT         /api/vehicles/{vin}`                          <-- update a vehicle; ETag supported

ETag support is provided for getting, updating, and deleting a specific vehicle.

The following content types are supported:

- `application/json`
- `application/xml`
- `application/x-protobuf`

## gRPC Resources

See `svr/proto/vehicle.proto`

## Vehicle format

A sample vehicle is shown below in `JSON` format; `vin` is the primary key and must be unique and all properties are required.

```json
{
    "vin": "abc124",
    "make": "Honda",
    "year": 2019,
    "model": "Accord",
    "exterior_color": "Red",
    "interior_color": "Tan"
}
```

API responses also include an `updated_at` property reflecting the Unix milliseconds (UTC) timestamp the resource was last updated.

## Running the App

Simply clone the repo, optionally updating any `environment` settings in the `docker-compose.yaml`, and run it with `docker-compose`.

1. Clone the repo: `git clone git@github.com:bodenr/vehicle-crud.git && cd ./vehicle-crud`
2. (Optional) Update the `docker-compose.yaml` environment settings to your liking.
3. (Optional) Edit the `.env` file and change the `BUILD_TYPE` to `test` for running the app with the go race detector enabled.
4. Run the app: `docker-compose up`

Note that when starting a basic set of integration tests are run via the `test` container to ensure the REST API is kosher.
