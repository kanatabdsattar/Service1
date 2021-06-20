package commands

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"os"
	"time"
)

package commands

import (
"context"
"crypto/rsa"
"fmt"
"io"
"os"
"time"


)


func GenToken(traceID string, log *zap.SugaredLogger, cfg database.Config, id string, privateKeyFile string, algorithm string) error {
	if id == "" || privateKeyFile == "" || algorithm == "" {
		fmt.Println("help: gentoken <id> <private_key_file> <algorithm>")
		fmt.Println("algorithm: RS256, HS256")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := user.NewStore(log, db)


	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: id,
		},
		Roles: []string{auth.RoleAdmin},
	}

	usr, err := store.QueryByID(ctx, traceID, claims, id)
	if err != nil {
		return errors.Wrap(err, "retrieve user")
	}


	pkf, err := os.Open(privateKeyFile)
	if err != nil {
		return errors.Wrap(err, "opening PEM private key file")
	}
	defer pkf.Close()
	privatePEM, err := io.ReadAll(io.LimitReader(pkf, 1024*1024))
	if err != nil {
		return errors.Wrap(err, "reading PEM private key file")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing PEM into private key")
	}


	a, err := auth.New(algorithm, keystore.NewMap(map[string]*rsa.PrivateKey{id: privateKey}))
	if err != nil {
		return errors.Wrap(err, "constructing auth")
	}


	claims = auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   usr.ID,
			ExpiresAt: jwt.At(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.Now(),
		},
		Roles: usr.Roles,
	}

	token, err := a.GenerateToken(id, claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", token)
	return nil
}
