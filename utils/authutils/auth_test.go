package authutils

import (
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/google/uuid"
	"testing"
)

func TestCreateToken(t *testing.T) {
	env.Override(env.Environment{SecretKey: "secret"})

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
				data:          "isaacdsc",
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
