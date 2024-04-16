# Sophrosyne

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=MadsRC_sophrosyne&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=MadsRC_sophrosyne)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=MadsRC_sophrosyne&metric=bugs)](https://sonarcloud.io/summary/new_code?id=MadsRC_sophrosyne)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=MadsRC_sophrosyne&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=MadsRC_sophrosyne)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=MadsRC_sophrosyne&metric=coverage)](https://sonarcloud.io/summary/new_code?id=MadsRC_sophrosyne)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=MadsRC_sophrosyne&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=MadsRC_sophrosyne)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/MadsRC/sophrosyne/badge)](https://securityscorecards.dev/viewer/?uri=github.com/MadsRC/sophrosyne)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8804/badge)](https://www.bestpractices.dev/projects/8804)
[![CodeQL](https://github.com/MadsRC/sophrosyne/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/MadsRC/sophrosyne/actions/workflows/github-code-scanning/codeql)
[![Semgrep](https://github.com/MadsRC/sophrosyne/actions/workflows/semgrep.yml/badge.svg)](https://github.com/MadsRC/sophrosyne/actions/workflows/semgrep.yml)

Sophrosyne is a horizontally scaleable content moderation API built for the
age of Generative AI.

The API allows you to register upstream modules to perform artibrary checks
on input data and return a `go`/`no-go` verdict. Checks are associated with
profiles, allowing several checks to be run on a piece of input data.

The application provides the API, but does not include any checks. Checks
are expected to be implemented as self-contained services communicating with
sophrosyne via gRPC. Reference implementations and documentation for how these
services should function will be provided.

## Stability

This project follows semantic versioning, and will introduce breaking changes
several times before reaching version 1.0.0.
