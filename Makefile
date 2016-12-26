BEATNAME=vspherebeat
BEAT_DIR=git.teamwork.net/BeatsTeamwork/vspherebeat
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.
VIRTUALENV_PARAMS="--python=/usr/bin/python3"
BUILD_DIR?=$(shell pwd)/build
COVERAGE_DIR=${BUILD_DIR}/coverage

.
# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile


.PHONY: vspherebeat

vspherebeat-secured:  $(GOFILES_ALL)
	make before-build
	go build -ldflags="-X git.teamwork.net/BeatsTeamwork/vspherebeat/beater.encryptionKey=${SALT_KEY}"
	go build encryptPassword/encryptpassword.go

.PHONY: full-clean
full-clean:
	make clean
	rm *.json
	rm *.yml
	rm -rf vendor
	rm encryptPassword

# Initial beat setup
.PHONY: setup
setup: copy-vendor
	make update

# Copy beats into vendor directory
.PHONY: copy-vendor
copy-vendor:
	mkdir -p vendor/github.com/elastic/
	cp -R ${GOPATH}/src/github.com/elastic/beats vendor/github.com/elastic/
	rm -rf vendor/github.com/elastic/beats/.git

.PHONY: git-init
git-init:
	git init
	git add README.md CONTRIBUTING.md
	git commit -m "Initial commit"
	git add LICENSE
	git commit -m "Add the LICENSE"
	git add .gitignore
	git commit -m "Add git settings"
	git add .
	git reset -- .travis.yml
	git commit -m "Add vspherebeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:
	python3 kibanaBuilder/builder.py --source _meta/kibana.raw --dest _meta/kibana

# Collects all dependencies and then calls update
.PHONY: collect
collect:

	PHONY: coverage-report
	coverage-report:
		python2 ${ES_BEATS}/dev-tools/aggregate_coverage.py -o ${COVERAGE_DIR}/full.cov ${COVERAGE_DIR}
		go tool cover -html=${COVERAGE_DIR}/full.cov -o ${COVERAGE_DIR}/full.html
		test ! -s ${COVERAGE_DIR}/integration.cov   || go tool cover -html=${COVERAGE_DIR}/integration.cov   -o ${COVERAGE_DIR}/integration.html
		test ! -s ${COVERAGE_DIR}/system.cov || go tool cover -html=${COVERAGE_DIR}/system.cov -o ${COVERAGE_DIR}/system.html
		test ! -s ${COVERAGE_DIR}/unit.cov   || go tool cover -html=${COVERAGE_DIR}/unit.cov   -o ${COVERAGE_DIR}/unit.html
