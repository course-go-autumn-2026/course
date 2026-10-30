# Lab 2 Architecture and Observability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Preserve the Lab 1 HTTP contract while restructuring the service into layers and adding required tests, linting, logs, metrics, and traces.

**Architecture:** cmd/trip-service owns process lifecycle only. internal/container creates concrete dependencies; HTTP handlers call usecases through consumer-side interfaces; usecases apply domain rules and use interfaces; PostgreSQL and telemetry live in infra or repository.

**Tech Stack:** Go 1.27, chi, pgxpool, squirrel, OpenTelemetry OTLP/HTTP, otelhttp, otelpgx, slog, gomock, httptest, PostgreSQL, goose, golangci-lint.

**Spec:** docs/superpowers/specs/2026-10-27-lab2-architecture-observability-design.md

## Global Constraints

- Do not add HTTP operations, edit the OpenAPI contract, or create migrations.
- Generated OpenAPI code belongs in internal/generated and is replaced only by make generate.
- internal/domain imports no HTTP, chi, pgx, squirrel, OpenTelemetry, or generated API packages.
- internal/usecase imports no HTTP, chi, pgx, squirrel, or generated API packages.
- internal/http imports no concrete package under internal/repository or internal/infra.
- Dependency interfaces are non-exported and stay in the consumer's contract.go.
- Mocks are generated with mockgen into adjacent mocks_test.go files and committed.
- make test runs with -race and reports at least 50% coverage in domain and usecase packages.
- Integration tests use PostgreSQL from tripgoctl environment start and have the integration build tag.
- Logs never include bodies, credentials, or database URLs.

---

## File Structure

| Path | Responsibility |
| --- | --- |
| internal/domain/trip | Entity, status transition, domain errors, domain tests |
| internal/model | Coordinates, create input, and history values |
| internal/usecase/trip/{create,get,finish} | One operation, contracts, unit tests, mocks |
| internal/repository/trip | SQL, conversion, and integration tests |
| internal/infra/postgres | pgx pool and transaction manager |
| internal/infra/telemetry | OpenTelemetry setup, metrics, trace-aware logger |
| internal/http/trip/{create,get,finish} | Generated-type conversion and handlers |
| internal/http/domainerror | One domain-error-to-Problem mapper |
| internal/http/middleware | Recover, tracing, logging, metrics |
| internal/container | Concrete dependency construction |
| internal/generated/api.gen.go | Generated OpenAPI code |

## Task 1: Configure Lab 2 tooling and environment

**Files:**
- Create: .golangci.yml, internal/config/config_test.go, internal/generated/api.gen.go
- Modify: api/oapi-codegen.yaml, .env.example, environment.toml, internal/config/config.go, Makefile, go.mod, go.sum
- Delete: api/trip_service.gen.go

**Produces:** config.TelemetryConfig with ServiceName, Endpoint, Protocol, ResourceAttributes, and TracesSampler; targets make generate, make lint, make test, make test-integration.

- [ ] **Step 1: Write the failing telemetry configuration test**

~~~
func TestLoadTelemetry(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "trip-service")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=tripgo,deployment.environment=local")
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_always_on")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Telemetry.ServiceName != "trip-service" {
		t.Fatalf("service name = %q", cfg.Telemetry.ServiceName)
	}
}
~~~

Set the current HTTP and database environment variables in the same test helper.

- [ ] **Step 2: Run the test before implementation**

Run: go test ./internal/config -run TestLoadTelemetry -count=1

Expected: failure because Config has no Telemetry field.

- [ ] **Step 3: Add configuration, environment, and generator output**

Add:

~~~
type TelemetryConfig struct {
	ServiceName        string
	Endpoint           string
	Protocol           string
	ResourceAttributes string
	TracesSampler      string
}
~~~

Load all five OTEL values through the current required-value helper. Copy homework/tasks/02-architecture-and-observability/environment.toml to the root and copy its Lab 2 OTel values from homework/config/trip-service.env.example to .env.example. Change generator configuration to package api and output internal/generated/api.gen.go. Keep the existing allowed operation IDs.

- [ ] **Step 4: Add tools and Make targets**

Install the module tool:

~~~
go get -tool go.uber.org/mock/mockgen
~~~

Copy homework/config/golangci.yml to .golangci.yml. Install golangci-lint using its official binary installer. Add Make targets equivalent to:

~~~
generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml
	go generate ./internal/...

lint:
	golangci-lint run ./...

test:
	go test -race -cover ./...

test-integration:
	set -a; . ./.env; set +a; go test -tags=integration ./internal/repository/...
~~~

- [ ] **Step 5: Generate and commit tooling**

Run:

~~~
gofmt -w internal/config/config.go internal/config/config_test.go
make generate
go test ./internal/config
git add .golangci.yml .env.example environment.toml Makefile api/oapi-codegen.yaml go.mod go.sum internal/config internal/generated
git rm api/trip_service.gen.go
git commit -m "chore: configure lab 2 tooling"
~~~

## Task 2: Create domain model and validation

**Files:**
- Create: internal/domain/trip/trip.go, internal/domain/trip/errors.go, internal/domain/trip/trip_test.go
- Create: internal/model/coordinates.go, internal/model/create_trip.go, internal/model/status_change.go
- Delete: internal/trip/trip.go, internal/trip/errors.go

**Produces:** trip.New(id uuid.UUID, input model.CreateTripInput, startedAt time.Time) (Trip, error), Trip.Finish(time.Time) error, ValidationError, ErrTripNotFound, ErrDriverBusy, and ErrTripCompleted.

- [ ] **Step 1: Write failing table-driven domain tests**

Test Finish for active and completed trips:

~~~
cases := []struct {
	name    string
	status  trip.Status
	wantErr error
}{
	{name: "active becomes completed", status: trip.StatusActive},
	{name: "completed stays completed", status: trip.StatusCompleted, wantErr: trip.ErrTripCompleted},
}
~~~

On success assert FinishedAt equals the supplied time. On failure assert both status and FinishedAt remain unchanged. Add New cases for nil user ID, nil driver ID, invalid latitude, invalid longitude, and negative price.

- [ ] **Step 2: Confirm domain tests fail**

Run: go test ./internal/domain/trip -count=1

Expected: failure because the package does not exist.

- [ ] **Step 3: Implement pure types and rules**

model.Coordinates owns latitude and longitude; model.CreateTripInput owns user, driver, points, and price. domain/trip.Trip owns lifecycle fields. New returns ValidationError with Field and Message. Finish mutates only an active trip and returns ErrTripCompleted otherwise.

- [ ] **Step 4: Verify and commit the inner layer**

Run:

~~~
go test -race -cover ./internal/domain/trip
go list -deps ./internal/domain/... | grep -E 'jackc/pgx|go-chi|net/http' && echo "VIOLATION" || true
git add internal/domain/trip internal/model
git rm internal/trip/trip.go internal/trip/errors.go
git commit -m "refactor: introduce trip domain model"
~~~

Expected: no violation is printed.

## Task 3: Implement create-trip usecase with gomock

**Files:**
- Create: internal/usecase/trip/create/contract.go, usecase.go, usecase_test.go, mocks_test.go

**Consumes:** trip.New, model.CreateTripInput, model.StatusChange.

**Produces:** Usecase.Create(ctx context.Context, input model.CreateTripInput) (trip.Trip, error).

- [ ] **Step 1: Define consumer-side contracts and generate mocks**

Put this in contract.go:

~~~
//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package create_test
type tripRepository interface {
	Save(context.Context, trip.Trip) error
}
type statusHistoryRepository interface {
	Add(context.Context, model.StatusChange) error
}
type txManager interface {
	Do(context.Context, func(context.Context) error) error
}
type metrics interface {
	TripCreated(context.Context)
}
~~~

Add an injected slog.Logger argument to the constructor. Run go generate ./internal/usecase/trip/create.

- [ ] **Step 2: Write failing behavior tests**

Create TestUsecaseCreateSuccess, TestUsecaseCreateDriverBusy, and TestUsecaseCreateHistoryFailure. In success, Do invokes its closure; expect Save, then Add with reason trip created, then TripCreated. If Save returns ErrDriverBusy, history and metric calls must not occur. If Add fails, metric must not occur.

- [ ] **Step 3: Run failing test and implement the transaction**

Run: go test ./internal/usecase/trip/create -count=1

Expected: failure because Usecase is absent.

Implement a constructor with the contracts, logger, and func() time.Time. Create a domain trip, save it and active history inside txManager.Do, then increment the metric and log trip created with trip_id only after the transaction succeeds.

- [ ] **Step 4: Verify and commit create**

~~~
go test -race -cover ./internal/usecase/trip/create
git add internal/usecase/trip/create
git commit -m "refactor: add create trip usecase"
~~~

## Task 4: Implement get and finish usecases

**Files:**
- Create: internal/usecase/trip/get/{contract.go,usecase.go,usecase_test.go,mocks_test.go}
- Create: internal/usecase/trip/finish/{contract.go,usecase.go,usecase_test.go,mocks_test.go}
- Delete: internal/trip/service.go, internal/trip/repository.go, internal/trip/transaction.go

**Produces:** get.Usecase.Get and finish.Usecase.Finish.

- [ ] **Step 1: Define, test, and implement get**

Use:

~~~
//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package get_test
type tripRepository interface {
	GetByID(context.Context, uuid.UUID) (trip.Trip, error)
}
~~~

Test successful passthrough and ErrTripNotFound passthrough. Implement a constructor accepting only this interface and delegate Get. Run go generate ./internal/usecase/trip/get and go test -race -cover ./internal/usecase/trip/get.

- [ ] **Step 2: Define finish contracts and write failing tests**

Use:

~~~
//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package finish_test
type tripRepository interface {
	GetForUpdate(context.Context, uuid.UUID) (trip.Trip, error)
	Update(context.Context, trip.Trip) error
}
~~~

Add history, transaction, and TripCompleted metrics interfaces. Write TestUsecaseFinishActiveTrip, TestUsecaseFinishCompletedTrip, TestUsecaseFinishMissingTrip, TestUsecaseFinishUpdateFailure, and TestUsecaseFinishHistoryFailure. The success case expects locked read, domain Finish, Update, status history with trip completed, and metric after a successful transaction.

- [ ] **Step 3: Implement finish with a locked read**

Inside one Do callback, call GetForUpdate, then value.Finish(clock()), Update, and Add history. The lock makes a concurrent request wait; it then sees the completed entity and returns ErrTripCompleted.

- [ ] **Step 4: Verify and commit both usecases**

~~~
go generate ./internal/usecase/trip/get ./internal/usecase/trip/finish
go test -race -cover ./internal/usecase/trip/get ./internal/usecase/trip/finish
git add internal/usecase/trip/get internal/usecase/trip/finish
git rm internal/trip/service.go internal/trip/repository.go internal/trip/transaction.go
git commit -m "refactor: split trip get and finish usecases"
~~~

## Task 5: Move PostgreSQL details to infra and repository

**Files:**
- Create: internal/infra/postgres/pool.go, internal/infra/postgres/transaction/manager.go
- Create: internal/repository/trip/{repository.go,save.go,fetch.go,update.go,history.go,convert.go,repository_integration_test.go}
- Delete: internal/postgres/*.go

**Produces:** transaction.Manager.Do with nested reuse and rollback; concrete repository implementing all usecase contracts.

- [ ] **Step 1: Write integration scaffold and failing scenarios**

Start repository_integration_test.go with:

~~~
//go:build integration

package trip_test
~~~

Load DATABASE_URL and fail clearly if missing. Each test uses a generated UUID and cleans its history then trip rows. Implement failing TestRepositorySaveAndGet, TestRepositoryGetForUpdateAndUpdate, and TestTransactionRollbackRemovesTripAndHistory.

The rollback test saves both records inside Manager.Do, returns a sentinel error, then verifies both tables have zero rows for the ID.

- [ ] **Step 2: Move pool and transaction implementation**

Move pool creation to infra/postgres and context transaction behavior to infra/postgres/transaction. Preserve ReadCommitted, nested reuse, rollback after error or panic, and query timeouts.

- [ ] **Step 3: Implement focused repository files**

save.go inserts a trip and maps constraint trips_driver_active_unique_idx to ErrDriverBusy. fetch.go implements GetByID and GetForUpdate with FOR UPDATE. update.go persists an entity already transitioned by the usecase. history.go inserts status history. convert.go owns Scan and PostgreSQL conversions.

- [ ] **Step 4: Verify real database behavior and commit**

~~~
tripgoctl environment start
make migrate
make test-integration
make test-integration
git add internal/infra/postgres internal/repository/trip
git rm -r internal/postgres
git commit -m "refactor: isolate postgres repository and transactions"
~~~

Expected: both integration runs pass, proving cleanup and repeatability.

## Task 6: Split HTTP transport and centralize errors

**Files:**
- Create: internal/http/domainerror/{domainerror.go,domainerror_test.go}
- Create: internal/http/trip/create/{contract.go,convert.go,handler.go,handler_test.go,mocks_test.go}
- Create: internal/http/trip/get/{contract.go,handler.go,handler_test.go,mocks_test.go}
- Create: internal/http/trip/finish/{contract.go,handler.go,handler_test.go,mocks_test.go}
- Create: internal/http/health.go
- Delete: internal/httpapi/*.go

**Produces:** domainerror.Problem(err, instance) and generated-interface-compatible endpoint handlers.

- [ ] **Step 1: Write and implement mapper tests**

Table-test ValidationError, ErrTripNotFound, ErrDriverBusy, ErrTripCompleted, and errors.New("database unavailable"). Assert 400, 404, 409, 409, and 500. Assert that the 500 problem does not contain database unavailable. Implement all errors.Is and errors.As logic in domainerror only.

- [ ] **Step 2: Define create contract and test strict input conversion**

Use:

~~~
//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package create_test
type usecase interface {
	Create(context.Context, model.CreateTripInput) (trip.Trip, error)
}
~~~

Test valid JSON, unknown JSON field, missing field, empty UUID, invalid coordinate, negative price, and a second JSON document. Every invalid case must return 400 with code invalid_request.

- [ ] **Step 3: Implement handlers**

Use json.Decoder.DisallowUnknownFields and ensure only one JSON object. Convert to model.CreateTripInput. Create writes 201, Location, and generated api.Trip. Get tests success plus 404. Finish tests success plus 409. Every arbitrary usecase error must return application/problem+json with 500 and no internal error text.

- [ ] **Step 4: Move health and readiness behind a narrow interface**

Keep health and readiness in internal/http/health.go. Readiness receives:

~~~
type pinger interface {
	Ping(context.Context) error
}
~~~

It never receives a concrete pgxpool type. Preserve current health and readiness response contracts.

- [ ] **Step 5: Generate, verify, and commit transport**

~~~
go generate ./internal/http/...
go test -race ./internal/http/...
git add internal/http
git rm -r internal/httpapi
git commit -m "refactor: split HTTP transport by operation"
~~~

## Task 7: Add OpenTelemetry, metrics, logs, and middleware

**Files:**
- Create: internal/infra/telemetry/{telemetry.go,logger.go,metrics.go}
- Create: internal/http/middleware/{recover.go,tracing.go,logging.go,metrics.go,response_writer.go,middleware_test.go}
- Modify: internal/infra/postgres/pool.go

**Produces:** telemetry.Setup, business metric recorder, and the required middleware chain.

- [ ] **Step 1: Write middleware tests first**

Test a handler returning 201 and a handler that panics. Assert the response recorder holds 201; recovery returns 500 internal_error and does not panic. Test that a request with a UUID route records /api/v1/trips/{tripId}, never the concrete ID.

- [ ] **Step 2: Implement telemetry initialization**

Create a resource from OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES. Use OTLP/HTTP trace and metric exporters because the Lab 2 environment sets http/protobuf. Honor endpoint and sampler values. Return Shutdown(context.Context) error that flushes both providers. Exporter failures are logged; an unavailable collector does not stop request handling.

- [ ] **Step 3: Implement trace-aware slog and custom metrics**

Wrap slog JSON handler so LogAttrs(ctx, ...) adds service, trace_id, and span_id when the span is valid. Create metrics TripCreated and TripCompleted. Record http.server.request.duration in seconds and http.server.active_requests with attributes http.request.method, http.route, and http.response.status_code.

- [ ] **Step 4: Instrument database and compose middleware**

Attach otelpgx to pgxpool.Config.ConnConfig.Tracer so child spans have db.operation.name and db.collection.name. Compose exactly:

~~~
Recover(Tracing(Logging(Metrics(router))))
~~~

Logging emits one line after every response: method, template route, status, duration, trace fields; use Info for 2xx/4xx and Error for 5xx. Usecases separately log trip created and trip completed without body data.

- [ ] **Step 5: Verify and commit observability**

~~~
go test -race ./internal/http/middleware ./internal/infra/telemetry
git add internal/infra/telemetry internal/infra/postgres/pool.go internal/http/middleware
git commit -m "feat: add service observability"
~~~

## Task 8: Compose dependencies and simplify main

**Files:**
- Create: internal/container/container.go, internal/container/container_test.go
- Modify: cmd/trip-service/main.go and imports affected by generated API move

**Produces:** container.New(ctx, cfg) (*Application, error), where Application has Server, Logger, and Close.

- [ ] **Step 1: Write a constructor boundary test**

Test malformed telemetry configuration that fails before opening external connections. Keep all tests requiring PostgreSQL in integration-tag files.

- [ ] **Step 2: Build concrete dependencies only in container**

Construct telemetry runtime and logger; pool; transaction manager; repository; create/get/finish usecases; health and trip handlers; generated chi server; middleware; http.Server. If a later constructor fails, close resources already created.

- [ ] **Step 3: Reduce main to lifecycle code**

main loads configuration, creates a signal context, calls container.New, starts Application.Server, calls Server.Shutdown with SHUTDOWN_TIMEOUT, and closes Application. It must not construct pgx, repositories, or usecases.

- [ ] **Step 4: Compile and commit**

~~~
gofmt -w cmd internal
go test -race ./...
go build ./...
git add cmd/trip-service/main.go internal/container
git commit -m "refactor: compose trip service in container"
~~~

## Task 9: Verify, document, and prepare review evidence

**Files:**
- Modify: README.md
- Create: docs/lab2/jaeger-trace.png, docs/lab2/grafana-metrics.png

**Produces:** README architecture section and safe Jaeger/Grafana evidence.

- [ ] **Step 1: Update README**

Document layer direction, consumer-side interfaces, container as composition root, make generate/lint/test/test-integration/migrate/run, and the requirement to run tripgoctl environment start before integration tests.

- [ ] **Step 2: Run automated checks**

~~~
make generate
make lint
make test
make test-integration
git diff --check
~~~

Expected: all commands exit zero and output shows domain/usecase coverage no lower than 50%.

- [ ] **Step 3: Verify service and telemetry**

Run tripgoctl environment start, make migrate, and make run. Create, get, and finish a trip using the Lab 1 requests. In Jaeger find the HTTP parent span and PostgreSQL child span. In Grafana find http_server_request_duration and tripgo_trips_created. Find the trace in Jaeger using trace_id from the JSON request log.

- [ ] **Step 4: Capture safe evidence, commit, and push**

Save one Jaeger trace tree and one Grafana metric screenshot, without credentials, database URL, body, or unrelated tabs.

~~~
git add README.md docs/lab2 Makefile
git commit -m "docs: describe lab 2 architecture and verification"
git push
~~~

Create a PR from Smoz1e:homework/02 to course-go-autumn-2026:main. The description includes layer scheme, consumer-side interface choice, FOR UPDATE decision, command results, and both screenshots.

## Plan Self-Review

- Architecture and dependency direction: Tasks 2–8.
- Domain transition and centralized error mapping: Tasks 2, 4, and 6.
- Required domain, usecase, mapper, handler, PostgreSQL, and rollback tests: Tasks 2–6.
- Linter and depguard: Task 1; final verification: Task 9.
- HTTP and DB traces, metrics, logs, recovery, and middleware order: Task 7.
- README, screenshots, PR material, and unchanged API: Task 9.
