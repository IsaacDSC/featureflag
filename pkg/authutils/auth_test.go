package authutils

import (
	"testing"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestCreateToken(t *testing.T) {
	env.Override(env.Environment{SecretKey: "82244db2-4346-4560-a768-14bf12aa5b81"})

	type args struct {
		data          any
		validateToken func(input string) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should be return",
			args: args{
				data:          "16db7723-bdd2-44b8-8a0d-9598ea45fb96",
				validateToken: VerifyToken,
			},
			wantErr: false,
		},
		{
			name: "should be return",
			args: args{
				data: struct {
					ID       uuid.UUID
					Username string
				}{
					ID:       uuid.New(),
					Username: "isaacdsc",
				},
				validateToken: VerifyToken,
			},
			wantErr: false,
		},
		{
			name: "should be return",
			args: args{
				data: "isaacdsc",
				validateToken: func(input string) error {
					return VerifyToken(input + "invalid")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateToken(tt.args.data)
			if err != nil {
				t.Errorf("CreateToken() error = %v", err)
				return
			}

			if (tt.args.validateToken(got) != nil) != tt.wantErr {
				t.Errorf("CreateToken() invalid token")
			}
		})
	}
}

// TestVerifyToken_RejectsAlgNone é um teste de regressão para a correção do
// "algorithm confusion": VerifyToken deve rejeitar um token assinado com
// alg=none, mesmo que o payload/claims sejam idênticos a um token válido.
func TestVerifyToken_RejectsAlgNone(t *testing.T) {
	env.Override(env.Environment{SecretKey: "82244db2-4346-4560-a768-14bf12aa5b81"})

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"data": "isaacdsc",
	})

	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build alg=none token: %v", err)
	}

	if err := VerifyToken(tokenString); err == nil {
		t.Error("VerifyToken() accepted a token signed with alg=none, want an error")
	}
}
