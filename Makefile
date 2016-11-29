.PHONY: all iperf \
	help \
	deps update-deps fakes \
	test test-iperf lint \
	release \
	clean clean-iperf

all:
	CGO_ENABLED=0 go build -ldflags "-s -d -w" -o clique-agent ./cmd/clique-agent

iperf: iperf/vendor/src/.lib
	go build -o clique-agent -tags "withIperf" ./cmd/clique-agent

###### Help ###################################################################

help:
	@echo '    all ................................. builds clique-agent'
	@echo '    iperf ............................... builds clique-agent with Iperf'
	@echo '    deps ................................ installs dependencies'
	@echo '    update-deps ......................... updates dependencies'
	@echo '    fakes ............................... run go generate'
	@echo '    test ................................ runs tests'
	@echo '    test-iperf .......................... runs iperf transfer tests'
	@echo '    lint ................................ lint the Go code'
	@echo '    release ............................. make Github release'
	@echo '    clean ............................... clean the built artifact'
	@echo '    clean-iperf ......................... clean the built artifacts for Iperf'

###### Dependencies ###########################################################

iperf/vendor/Makefile:
	cd iperf/vendor; ./configure

iperf/vendor/src/.lib: iperf/vendor/Makefile
	cd iperf/vendor; make

deps:
	glide install
	git submodule update --init --recursive

update-deps:
	glide update
	cd iperf/vendor; git checkout master; git pull

fakes:
	go generate `go list ./... | grep -v vendor`

###### Testing ################################################################

test: all
	CLIQUE_AGENT_PATH=${PWD}/clique-agent ginkgo -randomizeAllSpecs -p acceptance
	ginkgo -randomizeAllSpecs -r -p -race -skipPackage acceptance,ctl,vendor,iperf
	ginkgo -randomizeAllSpecs ctl

test-iperf: iperf
	LD_LIBRARY_PATH=${PWD}/iperf/vendor/src/.libs \
	DYLD_LIBRARY_PATH=${PWD}/iperf/vendor/src/.libs \
	TEST_WITH_IPERF=1 \
	CLIQUE_AGENT_PATH=${PWD}/clique-agent \
		ginkgo -randomizeAllSpecs -p acceptance
	LD_LIBRARY_PATH=${PWD}/iperf/vendor/src/.libs \
	DYLD_LIBRARY_PATH=${PWD}/iperf/vendor/src/.libs \
		ginkgo -randomizeSuites -p -race iperf

###### Code quality ###########################################################

lint:
	go vet `go list ./... | grep -v vendor`

###### Github release #########################################################

release: iperf/vendor/src/.lib
	mkdir release
	make
	mv ./clique-agent ./release/clique-agent-simple
	make iperf
	mv ./clique-agent ./release/clique-agent-iperf
	cp ./iperf/vendor/src/.libs/libiperf.so.0 ./release

###### Cleanup ################################################################

clean:
	rm -Rf ./clique-agent

clean-iperf:
	cd vendor/iperf; make clean
