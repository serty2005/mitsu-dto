// Package ofdclient provides a client for interacting with OFD (Operator of Fiscal Data) servers.
// It includes functionality for sending and receiving messages, handling errors, and managing
// transport connections. The package supports TCP transport with retry mechanisms and logging.
//
// Key Features:
//   - TCP transport with configurable timeouts and retry logic
//   - Message serialization and deserialization
//   - CRC calculation and validation
//   - Error handling and logging support
//   - Protocol-specific message handling
//
// Example Usage:
//
//	// Create a new TCP transport
//	transport := ofdclient.NewTCPTransport(
//	    10*time.Second,  // timeout
//	    3,                // retry count
//	    1*time.Second,    // retry delay
//	    func(msg string) { fmt.Println(msg) }, // logger
//	)
//
//	// Create a new OFD client
//	client := ofdclient.NewClient(transport)
//
//	// Send a message to the OFD server
//	response, err := client.SendMessage(context.Background(), "192.168.1.1:10000", message)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Process the response
//	fmt.Println("Received response:", response)
//
// The package is designed to be extensible, allowing for custom transport implementations
// and additional protocol support as needed.
package ofdclient
