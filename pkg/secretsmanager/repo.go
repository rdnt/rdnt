package secretsmanager

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Secret struct {
	Name            string `bson:"name"`
	EncryptedSecret []byte `bson:"encrypted_secret"`
}

var ErrNotFound = errors.New("secret not found")

func (m *Manager) Set(name string, secret []byte) error {
	collection := m.mdb.Database(m.database).Collection("secrets")
	ctx := context.Background()

	bs := Secret{Name: name, EncryptedSecret: secret}

	_, err := collection.InsertOne(ctx, bs)
	if err != nil {
		return errors.Wrap(err, "failed to set secret")
	}

	return nil
}

func (m *Manager) Get(name string) ([]byte, error) {
	collection := m.mdb.Database(m.database).Collection("secrets")
	ctx := context.Background()

	res := collection.FindOne(ctx, bson.D{primitive.E{Key: "name", Value: name}})
	err := res.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to find invitation")
	}

	var bs Secret
	err = res.Decode(&bs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode invitation")
	}

	return bs.EncryptedSecret, nil
}
