VERSION ?= $(shell poetry version | cut -d' ' -f2)

.PHONY: protobuf
protobuf: src/sophrosyne/grpc/checks/checks.proto src/sophrosyne/grpc/checks/checks_pb2.py src/sophrosyne/grpc/checks/checks_pb2_grpc.py src/sophrosyne/grpc/checks/checks_pb2.pyi

src/sophrosyne/grpc/checks/checks_pb2.py: src/sophrosyne/grpc/checks/checks.proto
	poetry run python -m grpc_tools.protoc -I./src --python_out=./src ./src/sophrosyne/grpc/checks/checks.proto

src/sophrosyne/grpc/checks/checks_pb2_grpc.py: src/sophrosyne/grpc/checks/checks.proto
	poetry run python -m grpc_tools.protoc -I./src --grpc_python_out=./src ./src/sophrosyne/grpc/checks/checks.proto

src/sophrosyne/grpc/checks/checks_pb2.pyi: src/sophrosyne/grpc/checks/checks.proto
	poetry run python -m grpc_tools.protoc -I./src --pyi_out=./src ./src/sophrosyne/grpc/checks/checks.proto

build/requirements.txt:
	@mkdir -p $(@D)
	poetry export --without-hashes --format=requirements.txt --with-credentials --output $@

dist/sophrosyne-$(VERSION)-py3-none-any.whl: src/sophrosyne/* src/sophrosyne/grpc/checks/checks.proto src/sophrosyne/grpc/checks/checks_pb2.py src/sophrosyne/grpc/checks/checks_pb2_grpc.py src/sophrosyne/grpc/checks/checks_pb2.pyi
	@mkdir -p $(@D)
	poetry build --format=wheel

dist/sophrosyne.tar: dist/sophrosyne-$(VERSION)-py3-none-any.whl build/requirements.txt
	mkdir -p $(@D)
	docker build --build-arg="dist_file=sophrosyne-$(VERSION)-py3-none-any.whl" --secret id=requirements,src=build/requirements.txt --no-cache --tag sophrosyne:$(VERSION) --attest=type=provenance,mode=max --attest=type=sbom --platform=linux/arm64 --output type=oci,dest=- . > $@

.PHONY: alembic/stamp
alembic/stamp:
	poetry run alembic stamp "head"

.PHONY: alembic/upgrade
alembic/upgrade:
	poetry run alembic upgrade "head"

.PHONY: alembic/auto
alembic/auto:
	poetry run alembic revision --autogenerate

.PHONY: alembic/revision
alembic/revision:
	poetry run alembic revision

.PHONY: dev/run
dev/run: build/.certificate_sentinel
	docker compose -f docker-compose.development.yml up -d
	SOPH__CONFIG_YAML_FILE=configurations/dev.yaml poetry run python src/sophrosyne/main.py run

.PHONY: dev/db/up
dev/db/up:
	docker compose -f docker-compose.development.yml up -d

.PHONY: dev/db/down
dev/db/down:
	docker compose -f docker-compose.development.yml down

build/.image_loaded_sentinel: dist/sophrosyne.tar
	mkdir -p $(@D)
	docker load --input dist/sophrosyne.tar
	@# For some reason the previous command doesn't include a newline in its output
	@printf "\n"
	touch $@

build/integration/root_token:
	mkdir -p $(@D)
	openssl rand -hex 128 > $@


.PHONY: test/integration
test/integration: test/integration/healthy_instance test/integration/auth01 test/integration/auth_required

.PHONY: test/integration/%
test/integration/%: build/.certificate_sentinel build/.image_loaded_sentinel build/integration/root_token
	$(MAKE) destroy/test/integration/$*
	VERSION=$(VERSION) ROOT_TOKEN="$$(cat build/integration/root_token)" docker compose -f tests/integration/$*/docker-compose.yml up --exit-code-from tester
	$(MAKE) destroy/test/integration/$*

.PHONY: destroy/test/integration/%
destroy/test/integration/%:
	VERSION="" ROOT_TOKEN="" docker compose -f tests/integration/$*/docker-compose.yml down

.PHONY: clean
clean:
	rm -rf src/sophrosyne/grpc/checks/checks_pb2.py src/sophrosyne/grpc/checks/checks_pb2_grpc.py src/sophrosyne/grpc/checks/checks_pb2.pyi
	rm -rf dist
	rm -rf build
	find . -name __pycache__ -exec rm -rf {} +
	rm -rf .pytest_cache
	rm -rf .mypy_cache
	rm -rf .ruff_cache
	-$(MAKE) dev/db/down

build/server.key: build/.certificate_sentinel

build/server.crt: build/.certificate_sentinel

build/.certificate_sentinel:
	@mkdir -p $(@D)
	openssl req -x509 -nodes -days 3650 -newkey ec -pkeyopt ec_paramgen_curve:secp384r1 -keyout build/server.key -out build/server.crt -subj '/CN=localhost' -addext 'subjectAltName = DNS:localhost,IP:127.0.0.1,IP:0.0.0.0,DNS:api'
	chmod 0777 build/server.key
	chmod 0777 build/server.crt
	touch $@

.PHONY:
dev/install:
	poetry install --with dev,test
