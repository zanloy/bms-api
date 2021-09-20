#export VERSION := $(shell head -n 1 version.txt | tr -d '\n')
VERSION != head -n 1 version.txt | tr -d '\n'

NEXUS := container-registry.prod8.bip.va.gov
TAG := $(NEXUS)/bms-api:$(VERSION)

.PHONY: build push run

build:
	@echo "Building bms-api docker image..."
	docker build . -t bms-api:$(VERSION)
	@echo "Done."

push:
	@echo "Tagging our image for upload..."
	docker tag bms-api:$(VERSION) $(TAG)
	@echo "Pushing tagged image up to $(NEXUS)..."
	docker push $(TAG)
	@echo "Cleaning up tag..."
	docker rmi $(TAG)
	@echo "Done."

run:
	@docker run \
		--name bms-api \
		-p 8080:8080 \
		--rm \
		--read-only -v $(HOME)/.kube:/root/.kube \
		bms-api:$(VERSION)
