package contenthub

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/golang/mock/gomock"
)

func TestContentHubService_CreateOrUpdate(t *testing.T) {
	control := gomock.NewController(t)
	repository := NewMockContentHubRepository(control)

	tests := []struct {
		name       string
		behavior   func(contenthub Entity)
		contenthub Entity
		wantErr    bool
	}{
		{
			name: "should create new content hub",
			behavior: func(contenthub Entity) {
				repository.EXPECT().GetContentHub(contenthub.Variable).Return(Entity{}, errorutils.NewNotFoundError("contenthub"))
				repository.EXPECT().SaveContentHub(contenthub).Return(nil)
			},
			contenthub: Entity{
				Variable: "test1",
				Active:   true,
			},
			wantErr: false,
		},
		{
			name: "should update existing content hub",
			behavior: func(contenthub Entity) {
				existing := Entity{
					Variable: "test1",
					Active:   false,
				}
				repository.EXPECT().GetContentHub(contenthub.Variable).Return(existing, nil)
				existing.Active = contenthub.Active
				repository.EXPECT().SaveContentHub(existing).Return(nil)
			},
			contenthub: Entity{
				Variable: "test1",
				Active:   true,
			},
			wantErr: false,
		},
		{
			name: "should return error on repository failure",
			behavior: func(contenthub Entity) {
				repository.EXPECT().GetContentHub(contenthub.Variable).Return(Entity{}, errors.New("repository error"))
			},
			contenthub: Entity{
				Variable: "test1",
				Active:   true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := Service{
				repository: repository,
			}
			tt.behavior(tt.contenthub)
			if err := ch.CreateOrUpdate(context.Background(), tt.contenthub); (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContentHubService_RemoveContentHub(t *testing.T) {
	control := gomock.NewController(t)
	repository := NewMockContentHubRepository(control)

	tests := []struct {
		name     string
		key      string
		behavior func(key string)
		wantErr  bool
	}{
		{
			name: "should remove content hub",
			key:  "test1",
			behavior: func(key string) {
				repository.EXPECT().DeleteContentHub(key).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "should return error on repository failure",
			key:  "test1",
			behavior: func(key string) {
				repository.EXPECT().DeleteContentHub(key).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := Service{
				repository: repository,
			}
			tt.behavior(tt.key)
			if err := ch.RemoveContentHub(tt.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveContentHub() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContentHubService_GetAllContentHub(t *testing.T) {
	control := gomock.NewController(t)
	repository := NewMockContentHubRepository(control)

	tests := []struct {
		name     string
		behavior func()
		want     map[string]Entity
		wantErr  bool
	}{
		{
			name: "should return all content hubs",
			behavior: func() {
				contenthubs := map[string]Entity{
					"test1": {Variable: "test1", Active: true},
					"test2": {Variable: "test2", Active: false},
				}
				repository.EXPECT().GetAllContentHub().Return(contenthubs, nil)
			},
			want: map[string]Entity{
				"test1": {Variable: "test1", Active: true},
				"test2": {Variable: "test2", Active: false},
			},
			wantErr: false,
		},
		{
			name: "should return error on repository failure",
			behavior: func() {
				repository.EXPECT().GetAllContentHub().Return(nil, errors.New("repository error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := Service{
				repository: repository,
			}
			tt.behavior()
			got, err := ch.GetAllContentHub()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllContentHub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllContentHub() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentHubService_GetContentHub(t *testing.T) {
	control := gomock.NewController(t)
	repository := NewMockContentHubRepository(control)

	tests := []struct {
		name     string
		key      string
		behavior func(key string)
		want     Entity
		wantErr  bool
	}{
		{
			name: "should return content hub",
			key:  "test1",
			behavior: func(key string) {
				contenthub := Entity{Variable: "test1", Active: true}
				repository.EXPECT().GetContentHub(key).Return(contenthub, nil)
			},
			want:    Entity{Variable: "test1", Active: true},
			wantErr: false,
		},
		{
			name: "should return error on repository failure",
			key:  "test1",
			behavior: func(key string) {
				repository.EXPECT().GetContentHub(key).Return(Entity{}, errors.New("repository error"))
			},
			want:    Entity{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := Service{
				repository: repository,
			}
			tt.behavior(tt.key)
			got, err := ch.GetContentHub(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContentHub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContentHub() got = %v, want %v", got, tt.want)
			}
		})
	}
}
