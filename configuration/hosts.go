package configuration

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/ElricleNecro/TOD/formatter"
	"github.com/ElricleNecro/TOD/host"
	"launchpad.net/goyaml"
)

// A type defining the user structure in the YAML file which we need
// in argument.
type MyHosts struct {
	Port     int    "port"
	Priority int    "priority"
	Threads  int    "threads"
	Protocol string "protocol"
	Timeout  int    "timeout"
}

// The map type passed to the read YAML procedure to get multiple users
// informations if we need.
type Hosts map[string]MyHosts

// A function to read hosts from the stdin.
func ReadHostsStdin(config *Config) *Hosts {

	// create a standard host configuration
	myhost := MyHosts{
		Port:     config.Port,
		Priority: 1,
		Threads:  1,
		Protocol: config.Protocol,
	}

	// create the result
	hosts := make(Hosts)

	// get data from stdin
	data, _ := ioutil.ReadAll(os.Stdin)

	// Create a list of hosts
	hostnames := []string(strings.Split(strings.TrimSpace(string(data)), "|"))

	// loop over hosts and create the structure
	for _, name := range hostnames {
		hosts[name] = myhost
	}

	// return the hosts
	return &hosts

}

// A function to read an hosts file in the YAML format and returns
// a dictionary in the same format as the structured file.
func ReadHostsYAML(
	filename string,
) *Hosts {

	// Start by reading the whole file in byte
	data, _ := ioutil.ReadFile(filename)

	// Create the variable handling the type of the user file
	t := &Hosts{}

	// Now read in the YAML file the structure of the file into
	// the structured dictionary
	err := goyaml.Unmarshal(
		data,
		t,
	)

	// Check error when reading the file
	if err != nil {
		formatter.ColoredPrintln(
			formatter.Red,
			false,
			"The file "+filename+" can't be read for accessing"+
				"the YAML structure!\n"+
				"Reason is: "+err.Error(),
		)
		return nil
	}

	// return the structured file and data
	return t

}

// This function converts the dictionary of hosts to the structure needed
// by the dispatcher.
func HostsToDispatcher(hosts Hosts) *host.Hosts {

	// init the hosts for output
	hst := new(host.Hosts)
	myhosts := make([]*host.Host, 0)

	// loop over elements in the map
	for hostname, fields := range hosts {

		// create a host object
		host := &host.Host{
			Hostname: hostname,
			Port:     fields.Port,
			Protocol: fields.Protocol,
			Timeout:  fields.Timeout,
		}

		// append to hosts
		myhosts = append(myhosts, host)
	}

	hst.Hosts = myhosts

	return hst
}
