package exec

import (
	"github.com/ElricleNecro/TOD/checker"
	"github.com/ElricleNecro/TOD/formatter"
	"github.com/ElricleNecro/TOD/ssh"
)

// This function loop over hosts to launch commands in concurrent mode.
func RunCommands(hosts []*formatter.Host, ncommands int) {

	var RunOnHost func(
		hosts []*formatter.Host,
		host *formatter.Host,
	)

	// number of hosts
	nhosts := len(hosts)

	// check that there is some hosts
	if nhosts == 0 {
		formatter.ColoredPrintln(
			formatter.Red,
			false,
			"There is no hosts given to run commands !",
		)
	}

	// A channel to wait for dispatching
	disconnected := make(chan *formatter.Host)

	// A channel to wait for end of program
	ender := make(chan bool, ncommands)

	// Function to dispatch an host on other. Set variables to allow a good synchronisation
	// between go routines.
	go func() {

		for {

			// display
			formatter.ColoredPrintln(
				formatter.Green,
				false,
				"Waiting for a disconnected host !",
			)

			// wait for a signal from a disconnected host
			host := <-disconnected

			// display
			formatter.ColoredPrintln(
				formatter.Green,
				false,
				"Dispatch the jobs of ", host.Hostname,
				" to other connected hosts !",
			)

			// mark the host as not connected
			host.IsConnected = false

			// dispatch remaining work to other hosts
			formatter.Dispatcher(
				host.Commands[host.CommandNumber:],
				hosts,
				false,
			)

			// display
			formatter.ColoredPrintln(
				formatter.Green,
				false,
				"Dispatching done for ", host.Hostname, " !",
			)
		}

	}()

	// This function is used to run a command on a host
	// with supplied informations.
	RunOnHost = func(
		hosts []*formatter.Host,
		host *formatter.Host,
	) {

	loop:
		// Do an infinite loop for waiting when ended
		for {

			// check the size of commands to execute before
			if len(host.Commands) != 0 {

				// display
				formatter.ColoredPrintln(
					formatter.Blue,
					true,
					"Executing ",
					len(host.Commands),
					" commands for ", host.Hostname,
				)

				// loop over commands on this hosts
				for i := host.CommandNumber; i < len(host.Commands); i++ {

					// number of the command
					host.CommandNumber = i

					// create a session object
					session := ssh.New(
						host.Commands[i].User,
						host,
					)

					// check the host can be called
					if is, err := checker.IsConnected(host); !is || (err != nil) {

						// display
						formatter.ColoredPrintln(
							formatter.Red,
							false,
							"Can't connect to host ", host.Hostname,
						)

						// dispatch remaining work to other hosts
						select {
						case disconnected <- host:
						default:
						}

						// exit the loop
						break loop
					}

					// Attempt a connection to the host
					err := session.Connect()

					// check the host can be called
					if err != nil {

						// display
						formatter.ColoredPrintln(
							formatter.Red,
							false,
							"Can't connect to host ", host.Hostname,
						)

						// dispatch remaining work to other hosts
						select {
						case disconnected <- host:
						default:
						}

						// exit the loop
						break loop
					}

					// add a session to connect to host
					_, err = session.AddSession()
					if err != nil {

						// display
						formatter.ColoredPrintln(
							formatter.Red,
							false,
							"Problem when adding a session to the host !",
						)

						// dispatch remaining work to other hosts
						select {
						case disconnected <- host:
						default:
						}

						// exit the loop
						break loop
					}

					// execute the command on the host
					formatter.ColoredPrintln(
						formatter.Green,
						false,
						"Execute command on ", host.Hostname,
					)
					output, err2 := session.Run(host.Commands[i].Command)
					if err2 != nil {

						// display
						formatter.ColoredPrintln(
							formatter.Red,
							false,
							"An error occurred during the execution ",
							"of the command !",
						)
						formatter.ColoredPrintln(
							formatter.Red,
							false,
							"The command was: ", host.Commands[i].Command,
						)
						formatter.ColoredPrintln(
							formatter.Red,
							false,
							"Error information: ", err.Error(),
						)

						// exit the loop
						break loop
					}

					// The command has been executed correctly, say it to other
					ender <- true

					// Close the session
					session.Close()

					// for now print the result of the command
					formatter.ColoredPrintln(
						formatter.Magenta,
						false,
						output,
					)

					// wait here for new jobs
					if i == len(host.Commands)-1 {

						// display
						formatter.ColoredPrintln(
							formatter.Magenta,
							true,
							"Waiting more jobs for ", host.Hostname,
						)

						// Now wait for new job
						<-*(host.Waiter)

						// display
						formatter.ColoredPrintln(
							formatter.Magenta,
							true,
							host.Hostname+" has more jobs !",
						)
						formatter.ColoredPrintln(
							formatter.Green,
							true,
							"Number of commands for ", host.Hostname, " :",
							len(host.Commands),
						)

					}

				}

			}

		}

	}

	// loop over hosts and run the command
	for _, host := range hosts {

		// in several goroutine
		go RunOnHost(hosts, host)

	}

	// Wait for the end of goroutines
	for i := 0; i < ncommands; i++ {
		<-ender
	}

}

//vim: spelllang=en
