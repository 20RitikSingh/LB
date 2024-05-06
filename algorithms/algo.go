package algorithms

import "net/url"

// RoundRobin returns the next server in a round-robin fashion
// func RoundRobin(servers []S, counter *int) Server {
// 	// Get the next server using the counter and update the counter
// 	server := servers[*counter]
// 	*counter = (*counter + 1) % len(servers)

// 	return server
// }

func leastConnections(targets []*url.URL) *url.URL {
	// Find the server with the least number of active connections
	// min := servers[0]
	// for _, server := range servers {
	// 	if server.ActiveConnections < min.ActiveConnections {
	// 		min = server
	// 	}
	// }

	// return min
}
