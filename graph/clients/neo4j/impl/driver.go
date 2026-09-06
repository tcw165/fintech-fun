package impl

import (
	"context"
	"fmt"

	official "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
)

// Driver is the official Bolt client. Composition root constructs this; tests keep Recording.
type Driver struct {
	inner official.DriverWithContext
}

func Open(uri, user, password string) (*Driver, error) {
	inner, err := official.NewDriverWithContext(uri, official.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}
	return &Driver{inner: inner}, nil
}

func OpenFromEnv() (*Driver, error) {
	return Open(URIFromEnv(), UserFromEnv(), PasswordFromEnv())
}

func (d *Driver) Close(ctx context.Context) error {
	if d == nil || d.inner == nil {
		return nil
	}
	return d.inner.Close(ctx)
}

func (d *Driver) Verify(ctx context.Context) error {
	if d == nil || d.inner == nil {
		return fmt.Errorf("neo4j driver is nil")
	}
	return d.inner.VerifyConnectivity(ctx)
}

func (d *Driver) Run(cypher string, params map[string]any) ([]map[string]any, error) {
	ctx := context.Background()
	result, err := official.ExecuteQuery(ctx, d.inner, cypher, params, official.EagerResultTransformer)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, len(result.Records))
	for _, record := range result.Records {
		rows = append(rows, record.AsMap())
	}
	return rows, nil
}

var _ neo4j.Client = (*Driver)(nil)
