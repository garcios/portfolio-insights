# Test Client

In the context of service-to-service communication using gRPC, a test client typically refers to a client application or script used to interact with and test gRPC-based services. This client simulates the behavior of an actual service consumer, allowing developers to verify the functionality and performance of the service APIs.

## Purpose

The test client serves several key purposes:

### 1. **Functional Testing**
- Validates that gRPC service methods work as expected
- Tests request/response handling for various RPC calls
- Verifies correct implementation of service contracts defined in `.proto` files

### 2. **Integration Testing**
- Tests end-to-end communication between services
- Validates authentication and authorization mechanisms
- Ensures proper handling of metadata and headers

### 3. **Development and Debugging**
- Provides a quick way to manually test services during development
- Helps debug issues by sending controlled requests
- Allows inspection of responses and error messages

### 4. **Performance Testing**
- Can be used to load test gRPC services
- Measures response times and throughput
- Identifies bottlenecks in service communication

## Usage

Test clients in this directory can be used to:
- Send requests to individual microservices (user-service, transaction-service, portfolio-service, etc.)
- Test various RPC methods defined in the service proto files
- Validate service behavior under different scenarios
- Debug service-to-service communication issues

## Implementation

Test clients typically:
- Import the generated gRPC client stubs from proto definitions
- Establish connections to target services
- Send requests with various payloads
- Log and validate responses
- Handle errors and edge cases
