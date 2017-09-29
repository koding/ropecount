package pkg

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// JWTData holds the data for JWT signing
type JWTData struct {
	Source string
	Target string
}

// Claim holds authentication claims
type Claim struct {
	// Source is the request sender
	Src string `json:"src"`

	// Target is the request executor
	Tgt string `json:"tgt"`

	jwt.StandardClaims
}

// Valid checks if the Claim is valid or not.
func (c *Claim) Valid() error {
	return c.StandardClaims.Valid()
}

var mySigningKey = []byte("AllYourBaseAreBelongToMe")

// SignJWT signs the given JWTData with the private key.
func SignJWT(d *JWTData) (string, error) {

	// Create the Claims
	claims := &Claim{
		Src: d.Source,
		Tgt: d.Target,
		StandardClaims: jwt.StandardClaims{
			Issuer: "auther", // string iss
			// Audience  string aud
			// ExpiresAt int64  exp
			// Id        string jti
			IssuedAt: time.Now().Unix(), //  int64 iat
			// NotBefore int64  nbf
			// Subject   string sub
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

// ParseJWT parses the given JWT and outputs debug log messages based on the
// validation errors.
func ParseJWT(logger log.Logger, s string) (*JWTData, error) {
	token, err := jwt.ParseWithClaims(s, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				level.Debug(logger).Log("operation", "ParseJWT", "token", s, "msg", "that's not even a token", "err", err)
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				level.Debug(logger).Log("operation", "ParseJWT", "token", s, "msg", "time is not valid.", "err", err)
			} else {
				level.Debug(logger).Log("operation", "ParseJWT", "token", s, "msg", "token error.", "err", err)
			}
		} else {
			level.Debug(logger).Log("operation", "ParseJWT", "token", s, "msg", "other error", "err", err)
		}

		return nil, err
	}

	if token.Valid {

		claims, ok := token.Claims.(*Claim)
		if !ok {
			return nil, fmt.Errorf("invalid data type in Claims %T", token.Claims)
		}

		return &JWTData{
			Source: claims.Src,
			Target: claims.Tgt,
		}, nil
	}

	return nil, fmt.Errorf("ParseJWT: invalid token, no error for %q", s)
}
