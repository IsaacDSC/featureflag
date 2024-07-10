package ctxutils

import (
	"context"
	"reflect"
	"testing"
)

func TestCtxUtils(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value string
	}

	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "should be return",
			args: args{
				ctx:   context.Background(),
				key:   "key",
				value: "value",
			},
			want: "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := SetContext(tt.args.ctx, tt.args.key, tt.args.value)
			if got := GetValueCtx(ctx, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValueCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}
