# Changelog

## v0.3.1 (12 February 2017)

* Fixes nasty bug with handling of outgoing transfers while serving incoming.

## v0.3.0 (29 November 2016)

* Use Iperf to do the transfers.
* Add RTT field in transfer results.

## v0.2.0 (06 August 2016)

* Rename the `GET /transfers` and `GET /transfers/:IP` to
  `GET /transfer_results` and `GET /transfer_results/:IP` respectively.
* Add `GET /transfers/:state` endpoint to get list of transfers filtered by
  state.

## v0.1.0 (29 December 2015)

The first release of clique includes:

* Working agent
* Transfer server and client
* Scheduler to schedule transfers
* REST API to access results and initiate new transfers + API client (Go)
* Control script to start/stop and check if the agent is running
