package token

import (
	"encoding/hex"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

const RandomSymmetricKey = "\x00"

type PasetoMaker struct {
	token        paseto.Token
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker(symmetricKey string) (*PasetoMaker, error) {
	maker := &PasetoMaker{
		token: paseto.NewToken(),
	}

	if symmetricKey == RandomSymmetricKey {
		maker.symmetricKey = paseto.NewV4SymmetricKey()
	} else {
		var err error
		maker.symmetricKey, err = paseto.V4SymmetricKeyFromHex(hex.EncodeToString([]byte(symmetricKey)))
		if err != nil {
			return nil, err
		}

	}

	return maker, nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	maker.token.SetIssuedAt(payload.IssuedAt)
	maker.token.SetExpiration(payload.ExpiredAt)
	maker.token.SetString("username", payload.Username)
	maker.token.SetString("uuid", payload.ID.String())

	return maker.token.V4Encrypt(maker.symmetricKey, nil), nil
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	parser := paseto.NewParser()

	parser.AddRule(paseto.NotExpired())

	t, err := parser.ParseV4Local(maker.symmetricKey, token, nil)
	if err != nil {
		return nil, err
	}

	payload.Username, err = t.GetString("username")
	if err != nil {
		return nil, err
	}

	var uuidString string
	uuidString, err = t.GetString("uuid")
	if err != nil {
		return nil, err
	}
	payload.ID, err = uuid.Parse(uuidString)
	if err != nil {
		return nil, err
	}

	payload.IssuedAt, err = t.GetIssuedAt()
	if err != nil {
		return nil, err
	}

	payload.ExpiredAt, err = t.GetExpiration()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
