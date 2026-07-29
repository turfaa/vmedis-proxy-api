<!--
Title should follow Conventional Commits, e.g. feat(procurement): add supplier recap v2 endpoint
Merging to main triggers semver tagging, changelog generation, an image publish, and deployment.
-->

## Summary

<!-- What this changes and why, in a few sentences. -->

## AI attribution

<!--
Required. If any part of this change was written by an AI model, name the model here —
the same name used in the commit's Co-Authored-By trailer, e.g. "Claude Opus 5".
Name every model if more than one was used. Write "None" if written entirely by hand.
-->

## Changes

<!-- One bullet per package or area touched, with a short note on what changed there. -->

-

## Verification

<!-- What you actually ran, and what it said. Note anything that could not be run here. -->

- [ ] `go build ./...`
- [ ] `go test ./...`
- [ ] `gofmt -l .` and `go vet ./...` clean
- [ ] Exercised the affected endpoint or command against a real Vmedis / DB

## Notes

<!-- Delete the lines that do not apply. -->

- **API**: endpoints added or changed — `docs/openapi.yaml` updated to match. New `/api/v2` responses use `cui` types.
- **Database**: new models registered in `database/db.go` for `AutoMigrate`; note any column or constraint change and how existing rows behave.
- **Dependencies**: new services or handlers wired in `cmd/dependencies.go`.
- **Scraping**: parser changes tied to a Vmedis HTML/API layout change — say which page or gateway call.
- **Consumers**: Kafka topic or message shape changes, and whether producer and consumer can be deployed independently.
- **Frontend**: paired `apotek-dashboard` change — link the PR.
