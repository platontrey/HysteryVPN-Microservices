To enable quick testing of the new MasterService endpoints, you can run the orchestrator service locally:

```bash
make orchestrator-run
```

The service will start on port 8081 with gRPC services enabled. You can test the MasterService endpoints using a gRPC client like grpcurl or by connecting agent services.