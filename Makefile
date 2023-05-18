# FROM env variable
REGION := $(TF_VAR_region)
BASE_URL := $(BASE_URL)

BASE_REPO := $(REGION)-docker.pkg.dev/$(GOOGLE_CLOUD_PROJECT)/my-app
TOPIC := test

.PHONY: all
all: delivering requesting

.PHONY: repo
repo:
	gcloud artifacts repositories create --repository-format=docker --location=$(REGION) my-app
	gcloud auth configure-docker $(REGION)-docker.pkg.dev

.PHONY: requesting
requesting:
	docker build -t $(BASE_REPO)/requesting -f ./requesting/Dockerfile ./requesting
	docker push $(BASE_REPO)/requesting
	gcloud run deploy requesting --region=$(REGION) --set-env-vars=TOPIC=$(TOPIC) --image=$(BASE_REPO)/requesting --allow-unauthenticated

.PHONY: delivering
delivering:
	docker build -t $(BASE_REPO)/delivering -f ./delivering/Dockerfile ./delivering
	docker push $(BASE_REPO)/delivering
	gcloud run deploy delivering --region=$(REGION) --set-env-vars=BASE_URL=$(BASE_URL),PROJECT_ID=$(GOOGLE_CLOUD_PROJECT) --image=$(BASE_REPO)/delivering --allow-unauthenticated

.PHONY: test-client
test-client:
	( cd requesting-clients && \
		go build -o test-client . )