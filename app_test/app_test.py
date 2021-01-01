import os
import random
import requests
import time
import unittest
import uuid


ACCEPT_JSON = "application/json"


def get_env(key, default=None):
    return os.getenv(key) or default


def generate_vehicles(make, model, year, int_color, ext_color, count):
    vehicles = []
    for i in range(count):
        vehicle = {
            "vin": "%s" % (str(uuid.uuid4())),
            "make": make,
            "year": year,
            "model": model,
            "exterior_color": ext_color,
            "interior_color": int_color
        }
        vehicles.append(vehicle)

    return vehicles


class RequestContext(object):

    def __init__(self, base_url=None, headers=None, timeout=4):
        self.base_url = base_url or 'http://' + \
            get_env("API_HOSTNAME", "172.22.0.3") + \
            ":" + get_env("API_PORT", "8080") + '/api'
        if not self.base_url.endswith('/'):
            self.base_url += '/'

        self.headers = headers or {'Accept': ACCEPT_JSON}
        if 'Accept' not in self.headers:
            self.headers['Accept'] = ACCEPT_JSON

        self.timeout = timeout
        # TODO: retries

    def build_url(self, url):
        if url.startswith('/'):
            url = url[1:]
        return self.base_url + url


class BaseClient(object):

    def __init__(self, request_context=None):
        self._context = request_context or RequestContext()

    def _request(self, url, method, request_context=None, **kwargs):
        context = request_context or self._context
        headers = context.headers.copy()
        if 'headers' in kwargs:
            headers.update(kwargs['headers'])
            kwargs.pop('headers', None)
        request = getattr(requests, method)
        return request(context.build_url(url), timeout=context.timeout,
                       headers=headers,
                       **kwargs)

    def get(self, url, request_context=None, **kwargs):
        return self._request(url, 'get', request_context=request_context, **kwargs)

    def list(self, url, request_context=None, **kwargs):
        responses = []
        response = self.get(url, request_context=request_context, **kwargs)
        responses.append(response)

        while response.links.get('next'):
            response = self.get(response.links.get(
                'next')['url'], request_context=request_context, **kwargs)
            responses.append(response)

        return responses

    def post(self, url, body, request_context=None, **kwargs):
        return self._request(url, 'post', request_context=request_context, json=body, **kwargs)

    def put(self, url, body, request_context=None, **kwargs):
        return self._request(url, 'put', request_context=request_context, json=body, **kwargs)

    def delete(self, url, request_context=None, **kwargs):
        return self._request(url, 'delete', request_context=request_context, **kwargs)


class VehicleClient(BaseClient):

    def __init__(self, request_context=None):
        super().__init__(request_context=request_context)

    def get(self, vin, request_context=None, **kwargs):
        url = "vehicles/%s" % (vin)
        return super().get(url, request_context=request_context, **kwargs)

    def list(self, request_context=None, **kwargs):
        url = "vehicles"
        return super().get(url, request_context=request_context, **kwargs)

    def create(self, vehicle, request_context=None, **kwargs):
        url = "vehicles"
        return super().post(url, vehicle, request_context=request_context, **kwargs)

    def update(self, vin, vehicle, request_context=None, **kwargs):
        url = "vehicles/%s" % (vin)
        return super().put(url, vehicle, request_context=request_context, **kwargs)

    def delete(self, vin, request_context=None, **kwargs):
        url = "vehicles/%s" % (vin)
        return super().delete(url, request_context=request_context, **kwargs)


# TODO: test xml/protobuf payloads + negative tests
class TestVehicleCrud(unittest.TestCase):

    def setUp(self):
        super().setUp()
        self.client = VehicleClient()

    def _retry_connect(self):
        connected = False
        for i in range(5):
            print("Unable to connect to API... Sleeping 1 second")
            time.sleep(1)
            resp = self.client.list()
            if resp.status_code == 200:
                connected = True
                break

        if not connected:
            raise Exception("Unabled to connect to client after 5 retries")

    def tearDown(self):
        super().tearDown()
        resp = self.client.list()
        self.assertEqual(resp.status_code, 200)
        for v in resp.json():
            resp = self.client.delete(v["vin"])
            self.assertEqual(resp.status_code, 204)
        resp = self.client.list()
        self.assertEqual(resp.status_code, 200)
        self.assertEqual([], resp.json())

    def assert_vehicle_equal(self, expected, actual):
        for k, v in expected.items():
            self.assertEqual(v, actual.get(k))
        self.assertIsNotNone(actual.get('updated_at'))

    def test_basic_create(self):
        vehicles = generate_vehicles(
            "Honda", "Accord", 2018, "Black", "Red", 10)
        for v in vehicles:
            resp = self.client.create(v)
            self.assertEqual(resp.status_code, 200)
            created_vehicle = resp.json()
            self.assert_vehicle_equal(v, created_vehicle)

    def test_basic_update(self):
        vehicle = generate_vehicles("Ford", "F150", 2020, "White", "Tan", 1)[0]
        resp = self.client.create(vehicle)
        self.assertEqual(resp.status_code, 200)
        created = resp.json()
        self.assert_vehicle_equal(vehicle, created)

        vehicle["year"] = 2021
        vehicle["interior_color"] = "Black"
        vehicle["exterior_color"] = "Green"
        resp = self.client.update(vehicle["vin"], vehicle)
        self.assertEqual(resp.status_code, 200)
        created = resp.json()
        self.assert_vehicle_equal(vehicle, created)

        resp = self.client.get(vehicle["vin"])
        self.assertEqual(resp.status_code, 200)
        created = resp.json()
        self.assert_vehicle_equal(vehicle, created)

    def test_search(self):
        vehicles = generate_vehicles(
            "VW", "Jetta", 2018, "Black", "Red", 5)
        vehicles.extend(generate_vehicles(
            "VW", "Jetta", 2019, "Black", "Red", 5))
        vehicles.extend(generate_vehicles(
            "VW", "Jetta", 2019, "Tan", "Black", 5))
        vehicles.extend(generate_vehicles(
            "VW", "Jetta", 2020, "Tan", "Black", 5))
        for v in vehicles:
            resp = self.client.create(v)
            self.assertEqual(resp.status_code, 200)

        resp = self.client.list(request_context=None, params={"year": 2020})
        self.assertEqual(resp.status_code, 200)
        self.assertEqual(5, len(resp.json()))

        resp = self.client.list(request_context=None, params={
                                "year": [2020, 2019]})
        self.assertEqual(resp.status_code, 200)
        self.assertEqual(15, len(resp.json()))

        resp = self.client.list(request_context=None, params={
                                "interior_color": "Black"})
        self.assertEqual(resp.status_code, 200)
        self.assertEqual(10, len(resp.json()))

        resp = self.client.list(request_context=None, params={
                                "year": 2019, "interior_color": "Black"})
        self.assertEqual(resp.status_code, 200)
        self.assertEqual(5, len(resp.json()))

    def test_etags(self):
        vehicle = generate_vehicles("Ford", "F150", 2020, "White", "Tan", 1)[0]
        resp = self.client.create(vehicle)
        self.assertEqual(resp.status_code, 200)
        created = resp.json()
        self.assert_vehicle_equal(vehicle, created)

        self.assertEqual(resp.headers.get("Pragma"), "no-cache")
        etag = resp.headers.get("ETag")
        self.assertIsNotNone(etag)

        resp = self.client.get(vehicle["vin"], request_context=None, headers={
            "If-None-Match": etag})
        self.assertEqual(resp.status_code, 304)

        vehicle["year"] = 2019
        resp = self.client.update(vehicle["vin"], vehicle, request_context=None, headers={
            "If-None-Match": "nope"})
        self.assertEqual(resp.status_code, 412)
        resp = self.client.update(vehicle["vin"], vehicle, request_context=None, headers={
            "If-None-Match": etag})
        self.assertEqual(resp.status_code, 200)
        updated = resp.json()
        self.assert_vehicle_equal(vehicle, updated)
        etag = resp.headers.get("ETag")

        resp = self.client.delete(vehicle["vin"], request_context=None, headers={
            "If-None-Match": "nope"})
        self.assertEqual(resp.status_code, 412)
        resp = self.client.delete(vehicle["vin"], request_context=None, headers={
            "If-None-Match": etag})
        self.assertEqual(resp.status_code, 204)


if __name__ == '__main__':
    unittest.main()
