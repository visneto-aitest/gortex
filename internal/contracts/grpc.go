package contracts

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zzet/gortex/internal/graph"
)

// GRPCExtractor detects gRPC service definitions (providers) and client usage (consumers).
type GRPCExtractor struct{}

var (
	// Proto service definitions: service Foo { rpc Bar(...) returns (...) }
	protoServiceRe = regexp.MustCompile(`(?m)service\s+(\w+)\s*\{`)
	protoRPCRe     = regexp.MustCompile(`(?m)rpc\s+(\w+)\s*\(`)

	// Go consumers
	goGRPCNewClientRe = regexp.MustCompile(`pb\.New(\w+)Client\(`)

	// TypeScript consumers
	tsGRPCNewClientRe = regexp.MustCompile(`new\s+(\w+)Client\(`)

	// Python consumers
	pyGRPCStubRe = regexp.MustCompile(`(\w+)Stub\(channel`)
)

func (e *GRPCExtractor) SupportedLanguages() []string {
	return []string{"proto", "go", "typescript", "python"}
}

func (e *GRPCExtractor) Extract(filePath string, src []byte, nodes []*graph.Node, edges []*graph.Edge) []Contract {
	var contracts []Contract

	if strings.HasSuffix(filePath, ".proto") {
		contracts = append(contracts, e.extractProtoProviders(filePath, src)...)
	} else {
		contracts = append(contracts, e.extractConsumers(filePath, src)...)
	}

	return contracts
}

func (e *GRPCExtractor) extractProtoProviders(filePath string, src []byte) []Contract {
	var contracts []Contract
	text := string(src)
	lines := strings.Split(text, "\n")

	// Find service blocks and their RPC methods.
	serviceMatches := protoServiceRe.FindAllStringSubmatchIndex(text, -1)
	for _, sMatch := range serviceMatches {
		serviceName := text[sMatch[2]:sMatch[3]]
		// Find RPCs within the remainder of this service block.
		serviceStart := sMatch[0]
		rest := text[serviceStart:]
		rpcMatches := protoRPCRe.FindAllStringSubmatch(rest, -1)
		rpcLocs := protoRPCRe.FindAllStringIndex(rest, -1)
		for i, rpc := range rpcMatches {
			methodName := rpc[1]
			absOffset := serviceStart + rpcLocs[i][0]
			line := lineNumber(lines, absOffset)
			contracts = append(contracts, Contract{
				ID:         fmt.Sprintf("grpc::%s::%s", serviceName, methodName),
				Type:       ContractGRPC,
				Role:       RoleProvider,
				FilePath:   filePath,
				Line:       line,
				Meta:       map[string]any{"service": serviceName, "method": methodName},
				Confidence: 0.95,
			})
		}
	}

	return contracts
}

func (e *GRPCExtractor) extractConsumers(filePath string, src []byte) []Contract {
	var contracts []Contract
	text := string(src)
	lines := strings.Split(text, "\n")

	// Go: pb.NewServiceNameClient(conn)
	for _, m := range goGRPCNewClientRe.FindAllStringSubmatchIndex(text, -1) {
		svc := text[m[2]:m[3]]
		contracts = append(contracts, Contract{
			ID:         fmt.Sprintf("grpc::%s", svc),
			Type:       ContractGRPC,
			Role:       RoleConsumer,
			FilePath:   filePath,
			Line:       lineNumber(lines, m[0]),
			Meta:       map[string]any{"service": svc, "lang": "go"},
			Confidence: 0.9,
		})
	}

	// TS: new ServiceNameClient()
	for _, m := range tsGRPCNewClientRe.FindAllStringSubmatchIndex(text, -1) {
		svc := text[m[2]:m[3]]
		contracts = append(contracts, Contract{
			ID:         fmt.Sprintf("grpc::%s", svc),
			Type:       ContractGRPC,
			Role:       RoleConsumer,
			FilePath:   filePath,
			Line:       lineNumber(lines, m[0]),
			Meta:       map[string]any{"service": svc, "lang": "typescript"},
			Confidence: 0.85,
		})
	}

	// Python: ServiceNameStub(channel)
	for _, m := range pyGRPCStubRe.FindAllStringSubmatchIndex(text, -1) {
		svc := text[m[2]:m[3]]
		contracts = append(contracts, Contract{
			ID:         fmt.Sprintf("grpc::%s", svc),
			Type:       ContractGRPC,
			Role:       RoleConsumer,
			FilePath:   filePath,
			Line:       lineNumber(lines, m[0]),
			Meta:       map[string]any{"service": svc, "lang": "python"},
			Confidence: 0.85,
		})
	}

	return contracts
}

// lineNumber returns the 1-based line number for the given byte offset.
func lineNumber(lines []string, offset int) int {
	pos := 0
	for i, l := range lines {
		end := pos + len(l) + 1 // +1 for newline
		if offset < end {
			return i + 1
		}
		pos = end
	}
	return len(lines)
}
