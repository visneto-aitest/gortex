package contracts

import (
	"testing"
)

func TestGRPCExtractor_ProtoProvider(t *testing.T) {
	ext := &GRPCExtractor{}
	src := []byte(`
syntax = "proto3";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {}
}
`)
	contracts := ext.Extract("user.proto", src, nil, nil)
	if len(contracts) != 2 {
		t.Fatalf("expected 2 contracts, got %d", len(contracts))
	}
	assertContract(t, contracts[0], "grpc::UserService::GetUser", ContractGRPC, RoleProvider)
	assertContract(t, contracts[1], "grpc::UserService::ListUsers", ContractGRPC, RoleProvider)
}

func TestGRPCExtractor_GoConsumer(t *testing.T) {
	ext := &GRPCExtractor{}
	src := []byte(`
package main

func main() {
	client := pb.NewUserServiceClient(conn)
}
`)
	contracts := ext.Extract("main.go", src, nil, nil)
	if len(contracts) < 1 {
		t.Fatalf("expected at least 1 contract, got %d", len(contracts))
	}
	assertContract(t, contracts[0], "grpc::UserService", ContractGRPC, RoleConsumer)
}

func TestGRPCExtractor_PythonConsumer(t *testing.T) {
	ext := &GRPCExtractor{}
	src := []byte(`
stub = UserServiceStub(channel)
response = stub.GetUser(request)
`)
	contracts := ext.Extract("client.py", src, nil, nil)
	if len(contracts) < 1 {
		t.Fatalf("expected at least 1 contract, got %d", len(contracts))
	}
	assertContract(t, contracts[0], "grpc::UserService", ContractGRPC, RoleConsumer)
}

func assertContract(t *testing.T, c Contract, id string, ctype ContractType, role Role) {
	t.Helper()
	if c.ID != id {
		t.Errorf("expected ID %q, got %q", id, c.ID)
	}
	if c.Type != ctype {
		t.Errorf("expected Type %q, got %q", ctype, c.Type)
	}
	if c.Role != role {
		t.Errorf("expected Role %q, got %q", role, c.Role)
	}
}
