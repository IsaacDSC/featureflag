package featureflag

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/IsaacDSC/featureflag/internal/strategy"
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestFeatureflagService_CreateOrUpdate(t *testing.T) {
	type fields struct {
		repository Adapter
	}

	type args struct {
		featureflag Entity
		behavior    func(featureflag Entity)
	}

	control := gomock.NewController(t)
	repository := NewMockFeatureFlagRepository(control)
	publisher := NewMockPublisher(control)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should be create new feature flag",
			fields: fields{
				repository: repository,
			},
			args: args{
				behavior: func(ff Entity) {
					repository.EXPECT().GetFF(gomock.Any(), gomock.Any()).Return(ff, nil)
					repository.EXPECT().SaveFF(gomock.Any(), ff).Return(nil)
					publisher.EXPECT().Publish(gomock.Any(), "featureflag", gomock.Any()).Return(nil)
				},
				featureflag: Entity{
					ID:         uuid.New(),
					FlagName:   "teste1",
					Strategies: strategy.Strategy{},
					Active:     true,
					CreatedAt:  time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "should be update feature flag",
			fields: fields{
				repository: repository,
			},
			args: args{
				behavior: func(ff Entity) {
					repository.EXPECT().GetFF(gomock.Any(), gomock.Any()).Return(Entity{}, errorutils.NewNotFoundError("featureflag"))
					repository.EXPECT().SaveFF(gomock.Any(), ff).Return(nil)
				},
				featureflag: Entity{
					ID:       uuid.New(),
					FlagName: "teste1",
					Strategies: strategy.Strategy{
						WithStrategy: true,
						SessionsID:   map[string]bool{"01J2BNZV8E3866GHMFFHDBZ3CD": true},
					},
					Active:    true,
					CreatedAt: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "should be not create ff, return error get ff",
			fields: fields{
				repository: repository,
			},
			args: args{
				behavior: func(ff Entity) {
					repository.EXPECT().GetFF(gomock.Any(), gomock.Any()).Return(Entity{}, errors.New("error read file"))
				},
				featureflag: Entity{
					ID:       uuid.New(),
					FlagName: "teste1",
					Strategies: strategy.Strategy{
						WithStrategy: true,
						SessionsID:   map[string]bool{"01J2BNZV8E3866GHMFFHDBZ3CD": true},
					},
					Active:    true,
					CreatedAt: time.Now(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := Service{
				repository: tt.fields.repository,
				pub:        publisher,
			}
			tt.args.behavior(tt.args.featureflag)
			if err := repo.CreateOrUpdate(context.Background(), tt.args.featureflag); (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatureflagService_GetFeatureFlag(t *testing.T) {
	control := gomock.NewController(t)
	repository := NewMockFeatureFlagRepository(control)
	publisher := NewMockPublisher(control)

	type fields struct {
		repository Adapter
	}
	type args struct {
		key       string
		sessionID string
		behavior  func(key string, sessionID string, featureflag Entity)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Entity
		wantErr bool
	}{
		{
			name: "should be return ff without strategy active",
			fields: fields{
				repository: repository,
			},
			args: args{
				key:       "teste1",
				sessionID: "",
				behavior: func(key string, sessionID string, featureflag Entity) {
					repository.EXPECT().GetFF(gomock.Any(), key).Return(featureflag, nil)
				},
			},
			want: Entity{
				ID:         uuid.New(),
				FlagName:   "teste1",
				Strategies: strategy.Strategy{},
				Active:     true,
				CreatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "should be return ff with strategy inactive",
			fields: fields{
				repository: repository,
			},
			args: args{
				key:       "teste2",
				sessionID: "",
				behavior: func(key string, sessionID string, featureflag Entity) {
					repository.EXPECT().GetFF(gomock.Any(), key).Return(featureflag, nil)
					featureflag.Strategies.QtdCall += 1
					featureflag.Active = false
					repository.EXPECT().SaveFF(gomock.Any(), featureflag).Return(nil)
				},
			},
			want: Entity{
				ID:       uuid.New(),
				FlagName: "teste2",
				Strategies: strategy.Strategy{
					WithStrategy: true,
					Percent:      50,
				},
				Active:    true,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "should be return ff with strategy active",
			fields: fields{
				repository: repository,
			},
			args: args{
				key:       "teste2",
				sessionID: "01J2BQ9Y19SHS6F6PMZQCH9Z70",
				behavior: func(key string, sessionID string, featureflag Entity) {
					repository.EXPECT().GetFF(gomock.Any(), key).Return(featureflag, nil)
					featureflag.Strategies.QtdCall += 1
					repository.EXPECT().SaveFF(gomock.Any(), featureflag).Return(nil)
				},
			},
			want: Entity{
				ID:       uuid.New(),
				FlagName: "teste2",
				Strategies: strategy.Strategy{
					WithStrategy: true,
					SessionsID:   map[string]bool{"01J2BQ9Y19SHS6F6PMZQCH9Z70": true},
				},
				Active:    true,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "should be return error retrieve ff",
			fields: fields{
				repository: repository,
			},
			args: args{
				key:       "teste2",
				sessionID: "01J2BQ9Y19SHS6F6PMZQCH9Z70",
				behavior: func(key string, sessionID string, featureflag Entity) {
					repository.EXPECT().GetFF(gomock.Any(), key).Return(featureflag, errors.New("os read file error"))
				},
			},
			want:    Entity{},
			wantErr: true,
		},
		{
			name: "should be return error to save ff with strategy",
			fields: fields{
				repository: repository,
			},
			args: args{
				key:       "teste2",
				sessionID: "01J2BQ9Y19SHS6F6PMZQCH9Z70",
				behavior: func(key string, sessionID string, ff Entity) {
					ff = Entity{
						ID:       uuid.New(),
						FlagName: "teste2",
						Strategies: strategy.Strategy{
							WithStrategy: true,
							SessionsID:   map[string]bool{"01J2BQ9Y19SHS6F6PMZQCH9Z70": true},
						},
						Active:    true,
						CreatedAt: time.Now(),
					}
					repository.EXPECT().GetFF(gomock.Any(), key).Return(ff, nil)
					ff.Strategies.QtdCall += 1
					repository.EXPECT().SaveFF(gomock.Any(), ff).Return(errors.New("os write error file"))
				},
			},
			want:    Entity{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := Service{
				repository: tt.fields.repository,
				pub:        publisher,
			}

			tt.args.behavior(tt.args.key, tt.args.sessionID, tt.want)

			got, err := ff.GetFeatureFlag(context.Background(), tt.args.key, tt.args.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeatureFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeatureFlag() got = %v, want %v", got, tt.want)
			}
		})
	}
}
