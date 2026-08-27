# Plata — Go Engineer — Test Assignment

## Overview

This is a test assignment submission for [Plata](https://bancoplata.mx/en).

This repo contains a service that provides an asynchronous interface for
fetching currency exchange rates. It uses
[exchangeratesapi.io](https://docs.apilayer.com/exchangeratesapi/docs/exchange-rates-api-v-1-0-0#/Endpoints/exchangeratesapiLatest)
to fetch the latest exchange rates. Only `USD`, `EUR`, and `MXN` currencies are
supported.

### Client flow

- Client requests an update via `POST /update?pair=XXX/YYY`. It must include the
  'Idempotency-Key' header with a valid UUIDv4 value. The JSON body returns
  'updateId'.
- Client can request their exchange rate via GET /quote/:updateId. The response
  can report multiple states, such as pending, completed, or failed.
- The latest known value of any supported pair can be fetched with
  `GET /quote/latest?pair=XXX/YYY`

## How to run

- Have Docker installed.
- Copy example environment variables with `cp .env.example .env`. Make sure to
  populate `EXCHANGERATESAPI_KEY` with a value from the
  [exchangeratesapi.io dashboard](https://manage.exchangeratesapi.io/dashboard).
- Run `docker compose up`. Check API health with
  `curl http://localhost:3000/health`
- (Optional) Run simple e2e tests with Node with `node minie2e.js`

## Internals

This project includes an HTTP server and two background workers. PostgreSQL
backs the underlying data storage. The workers poll the database at a
configurable rate. The term "update" refers to an internal data structure
representing an asynchronous exchange rate request.

The HTTP server creates new updates with a status `pending`, respecting
idempotency of the requests.

The update worker grabs `pending` updates and sets their status to processing.
Then the external api is queried for up-to-date exchange rates. Only one
external request is needed to perform conversion of all supported currencies.
Next, the worker performs the conversions and saves them to the database as
`completed`; any errors move the updates to a `failed` state.

The update could be stuck in the `processing` state, or become stale after a
configurable time. To address these issues, the cleanup worker moves old or
unfinished updates to a `failed` state.

# TODOs

Return fewer statuses to the client. Return pair too
