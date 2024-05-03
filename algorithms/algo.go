package algorithms

// RoundRobin returns the next server in a round-robin fashion
func RoundRobin(servers []Server, counter *int) Server {
	// Get the next server using the counter and update the counter
	server := servers[*counter]
	*counter = (*counter + 1) % len(servers)

	return server
}
