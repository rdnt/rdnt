package secretsmanager

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Manager struct {
	mdb      *mongo.Client
	database string
	encKey   []byte
}

func New(addr, database string, encKey []byte) (*Manager, error) {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(addr))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mongodb")
	}

	return &Manager{
		mdb:      client,
		database: database,
		encKey:   encKey,
	}, nil
}
