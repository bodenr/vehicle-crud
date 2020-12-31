import os
import random
import requests


ACCEPT_JSON = "application/json"


def get_env(key, default=None):
    return os.getenv(key) or default


class RequestContext(object):

    def __init__(self, base_url=None, headers=None, timeout=5):
        self.base_url = base_url or 'http://' + \
            get_env("API_HOSTNAME") + ":" + get_env("API_PORT") + '/api'
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

    def _get(self, url, request_context, params=None):
        response = requests.get(url, params=params,
                                timeout=request_context.timeout,
                                headers=request_context.headers)

        if response.status_code >= 300:
            response.raise_for_status()

        return response

    def get(self, url, request_context=None, params=None):
        context = request_context or self._context
        response = self._get(context.build_url(url), context,
                             params=params)
        return response.json()

    def list(self, url, request_context=None, params=None):
        context = request_context or self._context
        response = self._get(context.build_url(url), context, params=params)
        resources = response.json()

        while response.links.get('next'):
            response = self._get(response.links.get('next')['url'], context)
            resources.extend(response.json())

        return resources

    def post(self, url, body, request_context=None, params=None):
        context = request_context or self._context
        return requests.post(context.build_url(url), json=body,
                             timeout=context.timeout,
                             headers=context.headers)

    def put(self, url, body, request_context=None, params=None):
        context = request_context or self._context
        return requests.put(context.build_url(url), json=body,
                            timeout=context.timeout,
                            headers=context.headers)

    def delete(self, url, request_context=None, params=None):
        context = request_context or self._context
        return requests.delete(context.build_url(url),
                               timeout=context.timeout,
                               headers=context.headers)


class VehicleClient(BaseClient):

    def __init__(self, request_context=None):
        super().__init__(request_context=request_context)

    def get(self, vin, request_context=None, params=None):
        url = "vehicles/%s" % (vin)
        return super().get(url, request_context=request_context, params=params)

    def list(self, request_context=None, params=None):
        url = "vehicles"
        return super().list(url, request_context=request_context, params=params)

    def create(self, vehicle, request_context=None, params=None):
        url = "vehicles"
        return super().post(url, vehicle, request_context=request_context, params=params)

    def update(self, vin, vehicle, request_context=None, params=None):
        url = "vehicles/%s" % (vin)
        return super().put(url, vehicle, request_context=request_context, params=params)

    def delete(self, vin, request_context=None, params=None):
        url = "vehicles/%s" % (vin)
        return super().delete(url, request_context=request_context, params=params)


class VehicleData(object):

    def __init__(self, client):
        self.client = client

    def generate_vehicles(self, make, model, int_color, ext_color, count):
        vehicles = []
        for i in range(count):
            vehicle = {
                "vin": "%s.%s.%s" % (make, model, str(i)),
                "make": make,
                "year": 2019,
                "model": model,
                "exterior_color": ext_color,
                "interior_color": int_color
            }
            vehicles.append(vehicle)

        return vehicles

    def create_all(self, vehicles):
        for v in vehicles:
            resp = self.client.create(v)
            print("Created %s ... %s" % (v["vin"], resp))
            if resp.status_code != 200:
                return resp

    def delete_all(self):
        for v in self.client.list():
            print("Deleting %s ... %s" %
                  (v["vin"], self.client.delete(v["vin"])))

    def get_all(self, vehicles):
        for v in vehicles:
            resp = self.client.get(v["vin"])
            print("Getting %s ... %s" % (v["vin"], resp))
            if not resp:
                return v["vin"]


client = VehicleClient(request_context=RequestContext(
    base_url="http://172.22.0.3:8080/api"))

data = VehicleData(client)
hondas = data.generate_vehicles("Honda", "Accord", "Black", "Red", 10)
data.create_all(hondas)
data.get_all(hondas)
data.delete_all()
