// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package integration contains integration tests for Sophrosyne. The integration tests is done in the style of
// black-box testing and thus doesn't try to test private/internal functionality that doesn't affect the public
// footprint of Sophrosyne.
//
// Our integration tests relies heavily on [github.com/testcontainers/testcontainers-go] in order to spin up a
// realistic, yet ephemeral and reproducible test environment. There have previously been issues with running
// testcontainers-go with Podman and Colima, so for the time being, using Docker is preferred.
//
// The tests take the form of an outside observer / client, and as such the actual Sophrosyne application will be
// started as a container and the tests will interact with this container, usually via HTTP. In every test, the first
// order of business is to run the [setupEnv] function, as this will bootstrap everything. It is not recommended to
// spin up a new environment for every test, setting up the environment and taking it down again takes 5-10 seconds.
// Only create a brand-new environment if absolutely necessary.
//
// [setupEnv] returns a [testEnv] struct that contains everything necessary to talk to the integration test environment.
//
// The test environment will attempt to unmarshall every log from Sophrosyne as JSON, and if this isn't possible, the
// running test will fail.
//
// Before running the integration tests, the code has to be build first. This does not happen as part of running the
// test. As it is running in Docker, even if running on MacOS, you will need to ensure the binary is build for Linux.
// On a MacOS M-series machine, the software can be build and a docker image created and loaded by running:
//
//	mise run build:dist --goos=linux --goarch=arm64
//	mise run build:docker
//	cat build/sophrosyne.tar | docker load
//
// This also applies if making changes to the code, and you want to test these changes.
package integration
