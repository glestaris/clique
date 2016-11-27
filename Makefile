.PHONY: all \
	help \
	deps update-deps fakes \
	test lint \
	clean

all:
	CGO_ENABLED=0 go build -ldflags "-s -d -w" -o clique-agent ./cmd/clique-agent

###### Help ###################################################################

help:
	@echo '    all ................................. builds clique-agent'
	@echo '    deps ................................ installs dependencies'
	@echo '    update-deps ......................... updates dependencies'
	@echo '    fakes ............................... run go generate'
	@echo '    test ................................ runs tests'
	@echo '    lint ................................ lint the Go code'
	@echo '    clean ............................... clean the built artifact'

###### Dependencies ###########################################################

deps:
	glide install

update-deps:
	glide update

fakes:
	go generate `go list ./... | grep -v vendor`

###### Testing ################################################################

test: all
	CLIQUE_AGENT_PATH=${PWD}/clique-agent ginkgo -randomizeAllSpecs -p acceptance
	ginkgo -randomizeAllSpecs -randomizeSuites -r -p -race -skipPackage acceptance,ctl,vendor
	ginkgo -randomizeAllSpecs ctl

###### Code quality ###########################################################

lint:
	go vet `go list ./... | grep -v vendor`

###### Cleanup ################################################################

clean:
	rm -Rf ./clique-agent
