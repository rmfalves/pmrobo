# Overview
PMRobo is a cloud-based, multi-threaded project scheduling engine written in Go. Its purpose is to integrate with project management tools in order to provide them with auto-scheduling capabilities as simply as possible. Given project constraints such as task dependencies, resource capacities, and task demands, it establishes task start dates so that all constraints are met and the project is completed in the shortest possible timeframe.

REST web services and XML data exchange are used for the integration. Input must include task durations, task dependencies, resource capacities, and task resource consumption. The output includes the minimum feasible timeframe attained and the corresponding task dates, so that task dependencies are met and the total resource consumption by tasks does not exceed any resource capacity at any time.

# Installation
The *binaries* folder contains binary files for Linux and Windows.  If you want to download and compile the source code, do the following:
 1. Install the [Golang framework](https://go.dev/doc/install) if it is not already installed.
 2. Download the folders `goproj` and `pmrobo` to your $GO/src directory.
 3. `cd $GO/src/pmrobo`
 4. `go build `
 5. Execute the `pmrobo` executable file.
 6. If everything is in order, you should now have a service operating and listening for XML requests.
 # Configuration
PMRobo is configured as follows via the *pmrobo.xml* file:

**threads:** number of threads used by the solving procedure. It affects the performance directly.  
**port:** the TCP/IP port on which to listen for requests,  the default value is 9100.

There are other parameters reserved for developers who know the details of the solving process. They impact directly the performance and the solver behaviour, so you must be certain that you understand what you are doing before modifying them.
 # Usage
 ## Request structure
|REST Parameter|Value|
|--|--|
|URL|`server`/schedule:`port` where `server` is the IP/domain of the host on which the *pmrobo* service is running, and `port` is the assigned TCP/IP port, which by default is 9100|
|Action|POST|
|Content|XML string containing project specifications. For more information regarding the input XML data, please refer to [this tutorial](https://github.com/rmfalves/pmrobo/blob/main/TUTORIAL.md)|
## Result

XML string containing the project schedule. For more information regarding the returned XML data, please refer to [this tutorial](https://github.com/rmfalves/pmrobo/blob/main/TUTORIAL.md)
 
# Acknowledgements and License

The [Gin-Gonic library](https://github.com/gin-gonic/gin) on which this project depends to implement REST web services is [MIT licensed](https://opensource.org/license/mit/).

The PMRobo code is AGPL licensed as follows:

GNU Affero General Public License version 3

Copyright (C) 2023  Rui Alves

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see https://www.gnu.org/licenses.

THERE IS NO WARRANTY FOR THE PROGRAM, TO THE EXTENT PERMITTED BY APPLICABLE LAW. EXCEPT WHEN OTHERWISE STATED IN WRITING THE COPYRIGHT HOLDERS AND/OR OTHER PARTIES PROVIDE THE PROGRAM “AS IS” WITHOUT WARRANTY OF ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE. THE ENTIRE RISK AS TO THE QUALITY AND PERFORMANCE OF THE PROGRAM IS WITH YOU. SHOULD THE PROGRAM PROVE DEFECTIVE, YOU ASSUME THE COST OF ALL NECESSARY SERVICING, REPAIR OR CORRECTION.

IN NO EVENT UNLESS REQUIRED BY APPLICABLE LAW OR AGREED TO IN WRITING WILL ANY COPYRIGHT HOLDER, OR ANY OTHER PARTY WHO MODIFIES AND/OR CONVEYS THE PROGRAM AS PERMITTED ABOVE, BE LIABLE TO YOU FOR DAMAGES, INCLUDING ANY GENERAL, SPECIAL, INCIDENTAL OR CONSEQUENTIAL DAMAGES ARISING OUT OF THE USE OR INABILITY TO USE THE PROGRAM (INCLUDING BUT NOT LIMITED TO LOSS OF DATA OR DATA BEING RENDERED INACCURATE OR LOSSES SUSTAINED BY YOU OR THIRD PARTIES OR A FAILURE OF THE PROGRAM TO OPERATE WITH ANY OTHER PROGRAMS), EVEN IF SUCH HOLDER OR OTHER PARTY HAS BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.
