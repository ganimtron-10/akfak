# akfak: A Go-based Kafka Implementation

akfak is a Go-based implementation of the Kafka messaging system, built from scratch for understanding how things work under the hood.


## Features

* **Concurrent Clients:** Handles multiple client connections concurrently.
* **APIVersions API:** Implements Kafka's API versioning, allowing clients to specify the protocol version they're using.
* **Describe Topic Partitions API:** DescribeTopicPartitions allows clients to retrieve detailed information about topic and partitions.
* **Fetch API:** Fetch API enables clients to consume messages from Kafka topics.
* **Live Reloading:** Integrated `air` for rapid development with live code reloading.


## Project Structure

```
├── app
│   ├── api_version.go        # Handles APIVersion API.
│   ├── describe_partition.go # Handles DescribeTopicPartitions API.
│   ├── parser.go             # Parses incoming Kafka requests.
│   ├── server.go             # Manages the TCP server/client connections.
│   ├── type.go               # Defines custom data types for the project.
│   └── utils.go              # Utility functions.
├── go.mod                    # Go module definition.
├── go.sum                    # Go module checksums.
├── README.md                 # This file.
└── your_program.sh           # Script to build and start the application.
```


## Getting Started

1.  **Clone the Repository:**

    ```bash
    git clone https://github.com/ganimtron-10/akfak
    cd akfak
    ```

2.  **Build and Run:**

    Use the provided `your_program.sh` script to build and start the application:

    ```bash
    ./your_program.sh
    ```

    This script will compile the Go code and run the resulting binary.

3.  **Live Reloading:**

    If you have `air` installed, you can use it to automatically reload the application upon code changes.

    ```bash
    air
    ```

    If [air](https://github.com/air-verse/air) is not installed, install it by using the following command.

    ```bash
    go install github.com/air-verse/air@latest
    ```


## References

1.  Codecrafters Kafka: <https://app.codecrafters.io/courses/kafka/overview>
1.  ApiKeys - <https://kafka.apache.org/protocol.html#protocol_api_keys>
1.  Request and Response - <https://kafka.apache.org/protocol.html#protocol_messages>
1.  ApiVersions - <https://kafka.apache.org/protocol.html#The_Messages_ApiVersions>
1.  DescribePartitions - <https://kafka.apache.org/protocol.html#The_Messages_DescribeTopicPartitions>
1.  Fetch - <https://kafka.apache.org/protocol.html#The_Messages_Fetch>
1.  ErrorCode - <https://kafka.apache.org/protocol.html#protocol_error_codes>

