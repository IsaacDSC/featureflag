package domain

import (
	"errors"
	"ff/internal/domain/entity"
	"ff/internal/domain/interfaces"
	"ff/internal/errorutils"
	mock "ff/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
)

func TestFeatureflagService_CreateOrUpdate(t *testing.T) {
	type fields struct {
		repository interfaces.FeatureFlagRepository
	}

	type args struct {
		featureflag entity.Featureflag
		behavior    func(featureflag entity.Featureflag)
	}

	control := gomock.NewController(t)
	repository := mock.NewMockFeatureFlagRepository(control)

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
				behavior: func(featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(gomock.Any()).Return(featureflag, nil)
					repository.EXPECT().SaveFF(featureflag).Return(nil)
				},
				featureflag: entity.Featureflag{
					ID:         uuid.New(),
					FlagName:   "teste1",
					Strategies: entity.Strategy{},
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
				behavior: func(featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(gomock.Any()).Return(entity.Featureflag{}, errorutils.NewNotFoundError("ff"))
					repository.EXPECT().SaveFF(featureflag).Return(nil)
				},
				featureflag: entity.Featureflag{
					ID:       uuid.New(),
					FlagName: "teste1",
					Strategies: entity.Strategy{
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
				behavior: func(featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(gomock.Any()).Return(entity.Featureflag{}, errors.New("error read file"))
				},
				featureflag: entity.Featureflag{
					ID:       uuid.New(),
					FlagName: "teste1",
					Strategies: entity.Strategy{
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
			repo := FeatureflagService{
				repository: tt.fields.repository,
			}
			tt.args.behavior(tt.args.featureflag)
			if err := repo.CreateOrUpdate(tt.args.featureflag); (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatureflagService_GetFeatureFlag(t *testing.T) {
	control := gomock.NewController(t)
	repository := mock.NewMockFeatureFlagRepository(control)

	type fields struct {
		repository interfaces.FeatureFlagRepository
	}
	type args struct {
		key       string
		sessionID string
		behavior  func(key string, sessionID string, featureflag entity.Featureflag)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entity.Featureflag
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
				behavior: func(key string, sessionID string, featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(key).Return(featureflag, nil)
				},
			},
			want: entity.Featureflag{
				ID:         uuid.New(),
				FlagName:   "teste1",
				Strategies: entity.Strategy{},
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
				behavior: func(key string, sessionID string, featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(key).Return(featureflag, nil)
					featureflag.Strategies.QtdCall += 1
					featureflag.Active = false
					repository.EXPECT().SaveFF(featureflag).Return(nil)
				},
			},
			want: entity.Featureflag{
				ID:       uuid.New(),
				FlagName: "teste2",
				Strategies: entity.Strategy{
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
				behavior: func(key string, sessionID string, featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(key).Return(featureflag, nil)
					featureflag.Strategies.QtdCall += 1
					repository.EXPECT().SaveFF(featureflag).Return(nil)
				},
			},
			want: entity.Featureflag{
				ID:       uuid.New(),
				FlagName: "teste2",
				Strategies: entity.Strategy{
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
				behavior: func(key string, sessionID string, featureflag entity.Featureflag) {
					repository.EXPECT().GetFF(key).Return(featureflag, errors.New("os read file error"))
				},
			},
			want:    entity.Featureflag{},
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
				behavior: func(key string, sessionID string, featureflag entity.Featureflag) {
					featureflag = entity.Featureflag{
						ID:       uuid.New(),
						FlagName: "teste2",
						Strategies: entity.Strategy{
							WithStrategy: true,
							SessionsID:   map[string]bool{"01J2BQ9Y19SHS6F6PMZQCH9Z70": true},
						},
						Active:    true,
						CreatedAt: time.Now(),
					}
					repository.EXPECT().GetFF(key).Return(featureflag, nil)
					featureflag.Strategies.QtdCall += 1
					repository.EXPECT().SaveFF(featureflag).Return(errors.New("os write error file"))
				},
			},
			want:    entity.Featureflag{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := FeatureflagService{
				repository: tt.fields.repository,
			}

			tt.args.behavior(tt.args.key, tt.args.sessionID, tt.want)

			got, err := ff.GetFeatureFlag(tt.args.key, tt.args.sessionID)
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
