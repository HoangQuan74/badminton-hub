# Backend test entrypoint

Run each Go module from its own directory:

```bash
(cd ../platform && go test ./...)
(cd ../services/identity && go test ./...)
```

The remaining service modules currently contain scaffolding only and report
`no packages to test` until their implementation is added:

```text
services/booking
services/catalog
services/consumer
services/cron
services/notification
services/payment
services/socket
```

Owner tests live in `platform/test` or `services/<service>/test`. Shared
integration, contract, and end-to-end tests that span owners belong in this
directory.
