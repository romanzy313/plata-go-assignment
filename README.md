# Plata — Go Engineer — Test Assignment

## Architecture

- Client requests an update via `POST /update`. Update is added to the database
  in a `pending` state and `updateId` is returned.
- Background worker fetches all supported rate parts every X minutes. It fills
  all pending updates with the exchange rate.
- Client can request their exchange rate via `GET /quote/:updateId`, or latest
  value can be fetched using `GET /quote/latest`

## Libraries

- Http server: Echo
- Http requests [resty](https://github.com/go-resty/resty/blob/v2/README.md)

## Resources:

- [exchangeratesapi.io](https://docs.apilayer.com/exchangeratesapi/docs/exchange-rates-api-v-1-0-0#/Endpoints/exchangeratesapiLatest)

## TODO

- Application config
- Database implementation
- Background worker (pseudo cron)
- Failure state on the update
- Idepmotency of an update endpoint
- Graceful shutdown
