// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package commands

const buildInstructionsHelp = `
You can see the build instructions by running the 'config action' subcommand.
The instructions are set by the BUILD_INSTRUCTIONS environment variable, or a
simple default set of build instructions are used if that is not set.

This command fails if the build instructions do not write a file to BIN_PATH.

See the 'config env describe' subcommand for info on what environment variables are
available to your build instructions.

See the 'config env dump' subcommand to print out the values for all these variables.
`
