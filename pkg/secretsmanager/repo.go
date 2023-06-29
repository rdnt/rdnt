package secretsmanager

import (
	"context"
	"crypto/sha256"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/rdnt/rdnt/pkg/crypto"
)

type Secret struct {
	Name            string `bson:"name"`
	EncryptedSecret []byte `bson:"encrypted_secret"`
}

var ErrNotFound = errors.New("secret not found")

func (m *Manager) Set(name string, secret []byte) error {
	collection := m.mdb.Database(m.database).Collection("secrets")
	ctx := context.Background()

	b, err := crypto.Aes256CbcEncrypt(secret, m.encKey)
	if err != nil {
		return errors.WithMessage(err, "failed to encrypt secret")
	}

	mac := crypto.HmacSha256(m.signKey, b)

	// prepend mac to the ciphertext
	b = append(mac, b...)

	bs := Secret{Name: name, EncryptedSecret: b}

	_, err = collection.InsertOne(ctx, bs)
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

	mac := bs.EncryptedSecret[:sha256.Size]
	b := bs.EncryptedSecret[sha256.Size:]

	valid := crypto.VerifyHmacSha256(m.signKey, mac, b)
	if !valid {
		return nil, errors.WithMessage(err, "integrity check failed")
	}

	b, err = crypto.Aes256CbcDecrypt(b, m.encKey)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to encrypt secret")
	}

	return b, nil
}
