#!/bin/sh

mockgen \
    -source=endpoints/endpoint.go \
    -destination=endpoints/endpoint_mock.go \
    -package=endpoints \
    Endpoint
