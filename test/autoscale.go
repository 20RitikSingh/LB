package test

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/20ritiksingh/LoadBalancer/servers"
)

func findAvailablePort() (string, error) {
	// Listen on port ":0" to let the OS choose an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer listener.Close()

	// Get the address of the listener
	addr := listener.Addr().(*net.TCPAddr)

	// Return the port as a string
	return fmt.Sprintf(":%d", addr.Port), nil
}
func StartAutoScaling(instances int) {

	for i := 0; i < instances; i++ {
		port, err := findAvailablePort()
		if err != nil {
			i--
			continue
		}
		exec.Command("docker", "build", "-t", "scaleout", ".").Run()
		exec.Command("docker", "run", "-d", "--name", "scaleout", "-p", port, ":8081", "scaleout").Run()
		servers.AddtoYAML("servers.yaml", "http://loacalhost:"+port)
	}
}
