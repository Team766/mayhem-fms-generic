Mechanical M-Ayhem FMS (Generic Game) - based on [Team 254's Cheesy Arena](https://github.com/Team254/cheesy-arena).
============
This repo contains a fork of [Cheesy Arena](https://github.com/Team254/cheesy-arena), to be used as the 
Field Management System for [Team 766's Mechanical M-Ayhem](https://www.team766.com/mechanical-m-ayhem-1) rookie competition.
To make year-specific game customization easier, this codebase implements a "generic" game and is meant to be forked and
customized.  The bulk of the changes implement a generic, two gamepiece game, with each gamepiece meant to be scored in a
specific structure.  This generic version also adds optional support for 2v2 gameplay, as Mechanical M-Ayhem is currently a
2v2 game (but may grow to 3v3 in the future).

## License

Teams may use M-Ayhem FMS freely for practice, scrimmages, and off-season events. See [LICENSE](LICENSE) for more
details.

## Installing

**From source**

1. Download [Go](https://golang.org/dl/) (version 1.22 or later required)
1. Clone this GitHub repository to a location of your choice
1. Navigate to the repository's directory in the terminal
1. Compile the code with `go build`
1. Run the `cheesy-arena` or `cheesy-arena.exe` binary
1. Navigate to http://localhost:8080 in your browser (Google Chrome recommended)

**IP address configuration**

When running Cheesy Arena on a playing field with robots, set the IP address of the computer running Cheesy Arena to
10.0.100.5. By a convention baked into the FRC Driver Station software, driver stations will broadcast their presence on
the network to this hardcoded address so that the FMS does not need to discover them by some other method.

When running Cheesy Arena without robots for testing or development, any IP address can be used.

## Under the hood

Cheesy Arena is written using [Go](https://golang.org), a language developed by Google and first released in 2009. Go
excels in the areas of concurrency, networking, performance, and portability, which makes it ideal for a field
management system.

Cheesy Arena is implemented as a web server, with all human interaction done via browser. The graphical interfaces are
implemented in HTML, JavaScript, and CSS. There are many advantages to this approach &ndash; development of new
graphical elements is rapid, and no software needs to be installed other than on the server. Client web pages send
commands and receive updates using WebSockets.

[Bolt](https://github.com/etcd-io/bbolt) is used as the datastore, and making backups or transferring data from one
installation to another is as simple as copying the database file.

Schedule generation is fast because pre-generated schedules are included with the code. Each schedule contains a certain
number of matches per team for placeholder teams 1 through N, so generating the actual match schedule becomes a simple
exercise in permuting the mapping of real teams to placeholder teams. The pre-generated schedules are checked into this
repository and can be vetted in advance of any events for deviations from the randomness (and other) requirements.

Cheesy Arena includes support for, but doesn't require, networking hardware similar to that used in official FRC events.
Teams are issued their own SSIDs and WPA keys, and when connected to Cheesy Arena are isolated to a VLAN which prevents
any communication other than between the driver station, robot, and event server. The network hardware is reconfigured
via SSH and Telnet commands for the new set of teams when each mach is loaded.

## PLC integration

Mechanical M-Ayhem uses a custom designed Arduino-based PLC, protocol compatible with a subset of the PLC Cheesy Arena uses.
PLC functionality used by M-Ayhem consists of team e-stops and a-stops and the stack lights showing alliance readiness.

The PLC code can be found [here](https://github.com/Team766/fakeplc-arduino).

## Advanced networking

See the [Advanced Networking wiki page](https://github.com/Team254/cheesy-arena/wiki/Advanced-Networking-Concepts) for
instructions on what equipment to obtain and how to configure it in order to support advanced network security.
