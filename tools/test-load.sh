#!/usr/bin/env bash

# perform load testing with vegeta
pwd
vegeta attack -duration=20s -rate=5 -targets=tools/targets.list
