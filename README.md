# AI Librarian

Library QA: unidentified files, wrong metadata matches, missing episodes, and duplicate paths. Extra AI feature alongside the requested subtitle / recommend / ticket / filter modules.

## Ports

| Service | Default |
|---------|---------|
| gRPC | `127.0.0.1:9770` |
| HTTP | `127.0.0.1:9771` |

`POST /v1/scan` with `{ "library": [ ... ] }`.
