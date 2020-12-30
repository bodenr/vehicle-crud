## Forward

This repo is a WIP; not intended for consumption!

## Routes

- GET         /api/vehicles  <-- get all vehicles
- GET         /api/vehicles?exterior_color=red&make=dodge <-- - search vehicles
- POST        /api/vehicles <-- create a new vehicle
- GET         /api/vehicles/{vin} <-- get specific vehicle by VIN
- DELETE      /api/vehicles/{vin} <-- delete vehicle by VIN
- PUT         /api/vehicles/{vin} <-- update a vehicle

## Vehicle format

A sample vehicle is shown below in `JSON` format; `vin` is the primary key and must be unique.

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